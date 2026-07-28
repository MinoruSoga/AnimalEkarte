package auth

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type loginFailureAuthServiceStub struct {
	err error
}

type blockingLoginFailureAudit struct {
	started chan struct{}
	release chan struct{}
}

func (a *blockingLoginFailureAudit) LogAuthLogin(
	context.Context,
	*uint64,
	*uint64,
	string,
	string,
	string,
) error {
	a.started <- struct{}{}
	<-a.release
	return nil
}

func (*blockingLoginFailureAudit) LogEntry(
	context.Context,
	AuthAuditEntry,
) error {
	return nil
}

func (s loginFailureAuthServiceStub) AuthenticateUser(
	context.Context,
	string,
	string,
) (*model.Account, *model.Staff, error) {
	return nil, nil, s.err
}

func (loginFailureAuthServiceStub) ResolveClinicInfo(
	[]model.StaffClinicAssignment,
) (string, []uint64) {
	return "", nil
}

func (loginFailureAuthServiceStub) ResolveSystemAdminMainClinicID(
	string,
	bool,
	[]model.Clinic,
) string {
	return ""
}

func (loginFailureAuthServiceStub) CalculateEffectivePermissions(
	context.Context,
	bool,
	uint64,
	uint64,
) AuthEffectivePermissions {
	return nil
}

func TestHTTPHandler_AuthenticateUser_AppliesUniformCredentialFailureFloor(
	t *testing.T,
) {
	startedAt := time.Unix(1_700_000_000, 0)
	for _, test := range []struct {
		name      string
		err       error
		wantSleep bool
	}{
		{
			name:      "unknown account",
			err:       invalidCredentialsError(),
			wantSleep: true,
		},
		{
			name: "known account wrong password",
			err: &wrongPasswordError{
				accountID: 7,
				err:       invalidCredentialsError(),
			},
			wantSleep: true,
		},
		{
			name:      "internal failure is not disguised",
			err:       errors.New("database unavailable"),
			wantSleep: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			nowCalls := 0
			sleepCalls := 0
			var slept time.Duration
			handler := &HTTPHandler{
				deps: HTTPDependencies{
					Auth: loginFailureAuthServiceStub{err: test.err},
				},
				loginFailureTiming: loginFailureResponseTiming{
					now: func() time.Time {
						nowCalls++
						return startedAt
					},
					sleep: func(_ context.Context, delay time.Duration) error {
						sleepCalls++
						slept = delay
						return nil
					},
					jitter: func() time.Duration { return 0 },
				},
			}

			account, staff, err := handler.AuthenticateUser(
				context.Background(),
				"private@example.test",
				"wrong-password",
				"127.0.0.1",
				"test",
			)
			handler.Wait()

			assert.Nil(t, account)
			assert.Nil(t, staff)
			require.Error(t, err)
			if test.wantSleep {
				assert.ErrorIs(t, err, apperrors.ErrUnauthorized)
				assert.Equal(t, 2, nowCalls)
				assert.Equal(t, 1, sleepCalls)
				assert.Equal(t, loginFailureResponseFloor, slept)
			} else {
				assert.Equal(t, 1, nowCalls)
				assert.Zero(t, sleepCalls)
			}
		})
	}
}

func TestHTTPHandler_AuthenticateUser_FloorHonorsContextCancellation(
	t *testing.T,
) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	sleepObservedCancellation := false
	handler := &HTTPHandler{
		deps: HTTPDependencies{
			Auth: loginFailureAuthServiceStub{err: invalidCredentialsError()},
		},
		loginFailureTiming: loginFailureResponseTiming{
			now: func() time.Time { return time.Unix(1_700_000_000, 0) },
			sleep: func(ctx context.Context, _ time.Duration) error {
				sleepObservedCancellation = errors.Is(ctx.Err(), context.Canceled)
				return ctx.Err()
			},
			jitter: func() time.Duration { return 0 },
		},
	}

	_, _, err := handler.AuthenticateUser(
		ctx,
		"private@example.test",
		"wrong-password",
		"127.0.0.1",
		"test",
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrUnauthorized)
	assert.True(t, sleepObservedCancellation)
	handler.Wait()
}

func TestHTTPHandler_AuthenticateUser_BoundsAndDetachesFailureAuditWorkers(
	t *testing.T,
) {
	var logs bytes.Buffer
	originalLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
	})

	audit := &blockingLoginFailureAudit{
		started: make(chan struct{}, loginFailureAuditWorkers),
		release: make(chan struct{}),
	}
	handler := NewHTTPHandler(HTTPDependencies{
		Auth: loginFailureAuthServiceStub{err: &wrongPasswordError{
			accountID: 41,
			err:       invalidCredentialsError(),
		}},
		Staff: sessionStaffReader{
			findByAccountIDFn: func(
				context.Context,
				uint64,
			) (*model.Staff, error) {
				return &model.Staff{ID: 17}, nil
			},
		},
		StaffAssignments: sessionAssignmentReader{
			assignments: []model.StaffClinicAssignment{
				{StaffID: 17, ClinicID: 23, IsMain: true},
			},
		},
		Audit: audit,
	}, CookieConfigForProduction(false))
	handler.loginFailureTiming = loginFailureResponseTiming{
		now: func() time.Time { return time.Unix(1_700_000_000, 0) },
		sleep: func(context.Context, time.Duration) error {
			return nil
		},
		jitter: func() time.Duration { return 0 },
	}

	for range loginFailureAuditWorkers {
		result := make(chan error, 1)
		go func() {
			_, _, err := handler.AuthenticateUser(
				context.Background(),
				"private@example.test",
				"wrong-password",
				"127.0.0.1",
				"test",
			)
			result <- err
		}()

		select {
		case <-audit.started:
		case <-time.After(time.Second):
			t.Fatal("login failure audit worker did not start")
		}
		select {
		case err := <-result:
			require.ErrorIs(t, err, apperrors.ErrUnauthorized)
		case <-time.After(time.Second):
			t.Fatal("credential response waited for the audit writer")
		}
	}

	_, _, overflowErr := handler.AuthenticateUser(
		context.Background(),
		"private@example.test",
		"wrong-password",
		"198.51.100.42",
		"test-user-agent",
	)
	require.ErrorIs(t, overflowErr, apperrors.ErrUnauthorized)
	assert.Len(t, handler.loginAuditSlots, loginFailureAuditWorkers)
	assert.Contains(t, logs.String(), model.AuditActionAuthLoginFailure)
	assert.Contains(t, logs.String(), "account_id=41")
	assert.Contains(t, logs.String(), "reason=worker_saturated")
	assert.NotContains(t, logs.String(), "private@example.test")
	assert.NotContains(t, logs.String(), "wrong-password")
	assert.NotContains(t, logs.String(), "198.51.100.42")
	assert.NotContains(t, logs.String(), "test-user-agent")

	close(audit.release)
	handler.Wait()
	assert.Empty(t, handler.loginAuditSlots)
}

func TestHTTPHandler_WaitClosesLoginAuditRegistration(t *testing.T) {
	var logs bytes.Buffer
	originalLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
	})

	audit := &blockingLoginFailureAudit{
		started: make(chan struct{}, 1),
		release: make(chan struct{}),
	}
	handler := NewHTTPHandler(HTTPDependencies{
		Auth: loginFailureAuthServiceStub{err: &wrongPasswordError{
			accountID: 41,
			err:       invalidCredentialsError(),
		}},
		Staff: sessionStaffReader{
			findByAccountIDFn: func(
				context.Context,
				uint64,
			) (*model.Staff, error) {
				return &model.Staff{ID: 17}, nil
			},
		},
		StaffAssignments: sessionAssignmentReader{
			assignments: []model.StaffClinicAssignment{
				{ClinicID: 23, IsMain: true},
			},
		},
		Audit: audit,
	}, CookieConfigForProduction(false))
	handler.loginFailureTiming = loginFailureResponseTiming{
		now:    func() time.Time { return time.Unix(1_700_000_000, 0) },
		sleep:  func(context.Context, time.Duration) error { return nil },
		jitter: func() time.Duration { return 0 },
	}

	handler.Wait()

	_, _, err := handler.AuthenticateUser(
		context.Background(),
		"private@example.test",
		"wrong-password",
		"127.0.0.1",
		"test-user-agent",
	)
	require.ErrorIs(t, err, apperrors.ErrUnauthorized)

	select {
	case <-audit.started:
		t.Fatal("login audit worker started after shutdown drain began")
	case <-time.After(25 * time.Millisecond):
	}
	assert.Contains(t, logs.String(), model.AuditActionAuthLoginFailure)
	assert.Contains(t, logs.String(), "account_id=41")
	assert.Contains(t, logs.String(), "reason=registration_closed")
	assert.NotContains(t, logs.String(), "private@example.test")
	assert.NotContains(t, logs.String(), "wrong-password")
	assert.NotContains(t, logs.String(), "127.0.0.1")
	assert.NotContains(t, logs.String(), "test-user-agent")

	close(audit.release)
}
