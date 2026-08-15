package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestPasswordResetService_ForgotPassword_AppliesSameResponseFloorToEveryOutcome(
	t *testing.T,
) {
	lookupFailure := errors.New("lookup failed")
	transactionFailure := errors.New("transaction failed")
	tests := []struct {
		name        string
		accountRepo AccountRepository
		transactor  Transactor
		wantErr     bool
	}{
		{
			name: "existing account",
			accountRepo: &mockAccountRepository{
				findByEmailFn: func(context.Context, string) (*model.Account, error) {
					return &model.Account{ID: 7}, nil
				},
			},
			transactor: immediatePasswordResetTransactor(),
		},
		{
			name: "unknown account",
			accountRepo: &mockAccountRepository{
				findByEmailFn: func(context.Context, string) (*model.Account, error) {
					return nil, apperrors.WrapNotFound("account", "lookup")
				},
			},
			transactor: immediatePasswordResetTransactor(),
		},
		{
			name: "nil account result",
			accountRepo: &mockAccountRepository{
				findByEmailFn: func(context.Context, string) (*model.Account, error) {
					return nil, nil
				},
			},
			transactor: immediatePasswordResetTransactor(),
			wantErr:    true,
		},
		{
			name: "lookup failure",
			accountRepo: &mockAccountRepository{
				findByEmailFn: func(context.Context, string) (*model.Account, error) {
					return nil, lookupFailure
				},
			},
			transactor: immediatePasswordResetTransactor(),
			wantErr:    true,
		},
		{
			name: "transaction failure",
			accountRepo: &mockAccountRepository{
				findByEmailFn: func(context.Context, string) (*model.Account, error) {
					return &model.Account{ID: 7}, nil
				},
			},
			transactor: passwordResetTransactorFunc(
				func(context.Context, func(context.Context) error) error {
					return transactionFailure
				},
			),
			wantErr: true,
		},
		{
			name:    "missing transaction boundary",
			wantErr: true,
		},
	}

	startedAt := time.Unix(1_727_123_456, 0)
	elapsed := 125 * time.Millisecond
	jitter := 75 * time.Millisecond
	expectedDelay := forgotPasswordResponseFloor + jitter - elapsed

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			type responseFloorContextKey struct{}
			ctx := context.WithValue(
				context.Background(),
				responseFloorContextKey{},
				test.name,
			)
			nowCalls := 0
			sleepCalls := 0
			var slept time.Duration
			service := &passwordResetService{
				cfg:         &PasswordResetConfig{FrontendURL: "https://example.com"},
				accountRepo: test.accountRepo,
				tokenRepo:   &mockPasswordResetTokenRepository{},
				transactor:  test.transactor,
				comparePassword: func([]byte, []byte) error {
					return nil
				},
				responseTiming: forgotPasswordResponseTiming{
					now: func() time.Time {
						nowCalls++
						if nowCalls == 1 {
							return startedAt
						}
						return startedAt.Add(elapsed)
					},
					sleep: func(ctx context.Context, delay time.Duration) error {
						assert.Equal(t, test.name, ctx.Value(responseFloorContextKey{}))
						sleepCalls++
						slept = delay
						return nil
					},
					jitter: func() time.Duration {
						return jitter
					},
				},
			}

			err := service.ForgotPassword(ctx, "private-owner@example.test")
			service.Wait()

			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, 2, nowCalls)
			assert.Equal(t, 1, sleepCalls)
			assert.Equal(t, expectedDelay, slept)
		})
	}
}

func TestPasswordResetService_ForgotPassword_ResponseFloorIsContextAware(
	t *testing.T,
) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	sleepCalls := 0
	startedAt := time.Unix(1_727_123_456, 0)
	service := &passwordResetService{
		cfg: &PasswordResetConfig{FrontendURL: "https://example.com"},
		accountRepo: &mockAccountRepository{
			findByEmailFn: func(context.Context, string) (*model.Account, error) {
				return nil, apperrors.WrapNotFound("account", "lookup")
			},
		},
		tokenRepo:       &mockPasswordResetTokenRepository{},
		transactor:      immediatePasswordResetTransactor(),
		comparePassword: func([]byte, []byte) error { return nil },
		responseTiming: forgotPasswordResponseTiming{
			now: func() time.Time {
				return startedAt
			},
			sleep: func(received context.Context, _ time.Duration) error {
				sleepCalls++
				return received.Err()
			},
			jitter: func() time.Duration { return 0 },
		},
	}

	err := service.ForgotPassword(ctx, "private-owner@example.test")

	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 1, sleepCalls)
}

func TestBoundedForgotPasswordJitter(t *testing.T) {
	assert.Zero(t, boundedForgotPasswordJitter(-time.Second))
	assert.Equal(
		t,
		forgotPasswordResponseJitterLimit,
		boundedForgotPasswordJitter(forgotPasswordResponseJitterLimit+time.Second),
	)
	assert.Equal(t, 50*time.Millisecond, boundedForgotPasswordJitter(50*time.Millisecond))
}

func TestSleepWithContext_ReturnsCancellationWithoutWaitingForFloor(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := sleepWithContext(ctx, forgotPasswordResponseFloor)

	require.ErrorIs(t, err, context.Canceled)
}

func TestSleepWithContext_ZeroDelayReturnsImmediately(t *testing.T) {
	require.NoError(t, sleepWithContext(context.Background(), 0))
}

func TestRandomForgotPasswordJitter_IsBounded(t *testing.T) {
	for range 32 {
		jitter := randomForgotPasswordJitter()
		assert.GreaterOrEqual(t, jitter, time.Duration(0))
		assert.LessOrEqual(t, jitter, forgotPasswordResponseJitterLimit)
	}
}
