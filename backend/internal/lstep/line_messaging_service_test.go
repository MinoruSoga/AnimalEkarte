package lstep

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/infra/line"
)

// lineMessagingRoundTripFunc lets a test control the HTTP response for
// LineMessagingService.PushText without touching production code
// (LineMessagingService.httpClient is a plain *http.Client, so we can swap its Transport from
// within the same package).
type lineMessagingRoundTripFunc func(req *http.Request) (*http.Response, error)

func (f lineMessagingRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newTestLineMessagingService(channelToken string, rt lineMessagingRoundTripFunc) *LineMessagingService {
	return &LineMessagingService{
		channelToken: channelToken,
		httpClient:   &http.Client{Transport: rt, Timeout: 10 * time.Second},
	}
}

func TestLineMessagingService_PushText(t *testing.T) {
	t.Run("no-op when channel token is empty", func(t *testing.T) {
		called := false
		svc := newTestLineMessagingService("", func(_ *http.Request) (*http.Response, error) {
			called = true
			return nil, errors.New("should not be called")
		})
		err := svc.PushText(context.Background(), "U123", "hello")
		assert.NoError(t, err)
		assert.False(t, called, "HTTP request must not be sent when channel token is empty")
	})

	t.Run("sends the request with bearer auth and returns nil on 200 OK", func(t *testing.T) {
		var gotAuth, gotContentType, gotMethod, gotURL string
		svc := newTestLineMessagingService("secret-token", func(req *http.Request) (*http.Response, error) {
			gotAuth = req.Header.Get("Authorization")
			gotContentType = req.Header.Get("Content-Type")
			gotMethod = req.Method
			gotURL = req.URL.String()
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(nil),
			}, nil
		})
		err := svc.PushText(context.Background(), "U123", "hello world")
		assert.NoError(t, err)
		assert.Equal(t, "Bearer secret-token", gotAuth)
		assert.Equal(t, "application/json", gotContentType)
		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, line.PushEndpoint, gotURL)
	})

	t.Run("returns wrapped error when the HTTP client fails", func(t *testing.T) {
		svc := newTestLineMessagingService("secret-token", func(_ *http.Request) (*http.Response, error) {
			return nil, errors.New("connection refused")
		})
		err := svc.PushText(context.Background(), "U123", "hello")
		assert.Error(t, err)
	})

	t.Run("returns internal server error when LINE responds with non-200 status", func(t *testing.T) {
		svc := newTestLineMessagingService("secret-token", func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(nil),
			}, nil
		})
		err := svc.PushText(context.Background(), "U123", "hello")
		assert.Error(t, err)
	})

	t.Run("returns error when context is already cancelled", func(t *testing.T) {
		svc := newTestLineMessagingService("secret-token", func(_ *http.Request) (*http.Response, error) {
			return nil, errors.New("should not be reached")
		})
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := svc.PushText(ctx, "U123", "hello")
		assert.Error(t, err)
	})
}

func TestNewLineMessagingService(t *testing.T) {
	svc := NewLineMessagingService("token-abc")
	assert.NotNil(t, svc)
	assert.Equal(t, "token-abc", svc.channelToken)
	assert.NotNil(t, svc.httpClient)
}
