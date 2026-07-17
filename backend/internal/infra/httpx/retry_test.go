package httpx

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type trackingBody struct {
	io.Reader
	closedPtr *bool
}

func (b *trackingBody) Close() error {
	*b.closedPtr = true
	return nil
}

func TestDoWithRetry_SucceedsWithoutRetryOnNon429(t *testing.T) {
	calls := 0
	do := func() (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("ok"))}, nil
	}

	resp, err := DoWithRetry(context.Background(), nil, 3, time.Millisecond, do)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}

func TestDoWithRetry_RetriesOn429ThenSucceedsAndDrainsBody(t *testing.T) {
	calls := 0
	closedBodies := make([]bool, 2) // 固定長で確保（append によるスライス再確保でポインタが無効化されるのを防ぐ）
	do := func() (*http.Response, error) {
		calls++
		if calls <= 2 {
			idx := calls - 1
			return &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Body:       &trackingBody{Reader: strings.NewReader("rate limited"), closedPtr: &closedBodies[idx]},
			}, nil
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("ok"))}, nil
	}

	resp, err := DoWithRetry(context.Background(), nil, 3, time.Millisecond, do)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if calls != 3 {
		t.Fatalf("calls = %d, want 3", calls)
	}
	for i, closed := range closedBodies {
		if !closed {
			t.Errorf("429 response body[%d] was not closed (drain+close before retry)", i)
		}
	}
}

func TestDoWithRetry_ExhaustsRetriesAndReturnsErrRateLimit(t *testing.T) {
	calls := 0
	do := func() (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: http.StatusTooManyRequests, Body: io.NopCloser(strings.NewReader("rate limited"))}, nil
	}

	_, err := DoWithRetry(context.Background(), nil, 2, time.Millisecond, do)

	if !errors.Is(err, ErrRateLimit) {
		t.Fatalf("err = %v, want ErrRateLimit", err)
	}
	if calls != 3 {
		t.Fatalf("calls = %d, want maxRetries+1=3", calls)
	}
}

func TestDoWithRetry_ReturnsCtxErrorWhenCancelledDuringBackoff(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	do := func() (*http.Response, error) {
		calls++
		if calls == 1 {
			cancel()
		}
		return &http.Response{StatusCode: http.StatusTooManyRequests, Body: io.NopCloser(strings.NewReader("rate limited"))}, nil
	}

	_, err := DoWithRetry(ctx, nil, 3, 50*time.Millisecond, do)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1 (cancelled before 2nd attempt)", calls)
	}
}

func TestDoWithRetry_AppliesExponentialBackoff(t *testing.T) {
	var timestamps []time.Time
	calls := 0
	do := func() (*http.Response, error) {
		timestamps = append(timestamps, time.Now())
		calls++
		if calls <= 2 {
			return &http.Response{StatusCode: http.StatusTooManyRequests, Body: io.NopCloser(strings.NewReader("x"))}, nil
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("ok"))}, nil
	}
	initialWait := 20 * time.Millisecond

	_, err := DoWithRetry(context.Background(), nil, 3, initialWait, do)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(timestamps) != 3 {
		t.Fatalf("calls = %d, want 3", len(timestamps))
	}
	firstGap := timestamps[1].Sub(timestamps[0])
	secondGap := timestamps[2].Sub(timestamps[1])
	if firstGap < initialWait {
		t.Errorf("first backoff gap = %v, want >= %v", firstGap, initialWait)
	}
	if secondGap < 2*initialWait {
		t.Errorf("second backoff gap = %v, want >= %v (exponential doubling)", secondGap, 2*initialWait)
	}
}
