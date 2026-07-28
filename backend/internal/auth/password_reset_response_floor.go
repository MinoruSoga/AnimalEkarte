package auth

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"time"
)

const (
	forgotPasswordResponseFloor       = 750 * time.Millisecond
	forgotPasswordResponseJitterLimit = 250 * time.Millisecond
)

type forgotPasswordResponseTiming struct {
	now    func() time.Time
	sleep  func(context.Context, time.Duration) error
	jitter func() time.Duration
}

func (t forgotPasswordResponseTiming) withDefaults() forgotPasswordResponseTiming {
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

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func randomForgotPasswordJitter() time.Duration {
	var value [8]byte
	if _, err := rand.Read(value[:]); err != nil {
		return 0
	}
	limit := uint64(forgotPasswordResponseJitterLimit) + 1
	return time.Duration(binary.LittleEndian.Uint64(value[:]) % limit)
}

func boundedForgotPasswordJitter(jitter time.Duration) time.Duration {
	if jitter < 0 {
		return 0
	}
	if jitter > forgotPasswordResponseJitterLimit {
		return forgotPasswordResponseJitterLimit
	}
	return jitter
}
