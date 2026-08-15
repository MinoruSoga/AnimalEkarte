package auth

import (
	"context"
	"time"
)

const (
	loginFailureResponseFloor       = 750 * time.Millisecond
	loginFailureResponseJitterLimit = 250 * time.Millisecond
)

type loginFailureResponseTiming struct {
	now    func() time.Time
	sleep  func(context.Context, time.Duration) error
	jitter func() time.Duration
}

func (t loginFailureResponseTiming) withDefaults() loginFailureResponseTiming {
	if t.now == nil {
		t.now = time.Now
	}
	if t.sleep == nil {
		t.sleep = sleepWithContext
	}
	if t.jitter == nil {
		t.jitter = randomForgotPasswordJitter
	}
	return t
}

func (t loginFailureResponseTiming) wait(
	ctx context.Context,
	startedAt time.Time,
) {
	t = t.withDefaults()
	elapsed := t.now().Sub(startedAt)
	jitter := t.jitter()
	if jitter < 0 {
		jitter = 0
	}
	if jitter > loginFailureResponseJitterLimit {
		jitter = loginFailureResponseJitterLimit
	}
	remaining := loginFailureResponseFloor + jitter - elapsed
	if remaining < 0 {
		remaining = 0
	}
	_ = t.sleep(ctx, remaining)
}
