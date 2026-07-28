package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

type mockAccountRepository struct {
	AccountRepository
	findByIDFn          func(ctx context.Context, id uint64) (*model.Account, error)
	findByIDForUpdateFn func(ctx context.Context, id uint64) (*model.Account, error)
	findByEmailFn       func(ctx context.Context, email string) (*model.Account, error)
	compareAndSwapFn    func(
		ctx context.Context,
		id uint64,
		expectedHash, newHash string,
		updatedAt time.Time,
	) (bool, error)
	updatePasswordHashFn func(
		ctx context.Context,
		id uint64,
		newHash string,
		updatedAt time.Time,
	) error
}

func (m *mockAccountRepository) FindByIDForUpdate(
	ctx context.Context,
	id uint64,
) (*model.Account, error) {
	if m.findByIDForUpdateFn != nil {
		return m.findByIDForUpdateFn(ctx, id)
	}
	return m.FindByID(ctx, id)
}

func (m *mockAccountRepository) FindByID(ctx context.Context, id uint64) (*model.Account, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return &model.Account{ID: id, UpdatedAt: time.Unix(1, 0)}, nil
}

func (m *mockAccountRepository) FindByEmail(ctx context.Context, email string) (*model.Account, error) {
	if m.findByEmailFn != nil {
		return m.findByEmailFn(ctx, email)
	}
	return &model.Account{Email: email}, nil
}

func (m *mockAccountRepository) CompareAndSwapPasswordHash(
	ctx context.Context,
	id uint64,
	expectedHash, newHash string,
	updatedAt time.Time,
) (bool, error) {
	if m.compareAndSwapFn != nil {
		return m.compareAndSwapFn(ctx, id, expectedHash, newHash, updatedAt)
	}
	return true, nil
}

func (m *mockAccountRepository) UpdatePasswordHash(
	ctx context.Context,
	id uint64,
	newHash string,
	updatedAt time.Time,
) error {
	if m.updatePasswordHashFn != nil {
		return m.updatePasswordHashFn(ctx, id, newHash, updatedAt)
	}
	return nil
}

type accountPasswordChangerTestPort interface {
	ChangePassword(
		ctx context.Context,
		accountID uint64,
		currentPassword, newPassword string,
		audit CredentialMutationAudit,
	) error
}

type accountResetTokenInvalidatorFunc func(context.Context, uint64) error

func (f accountResetTokenInvalidatorFunc) DeleteByAccountID(
	ctx context.Context,
	accountID uint64,
) error {
	return f(ctx, accountID)
}

type accountChangeTransactorFunc func(
	context.Context,
	func(context.Context) error,
) error

func (f accountChangeTransactorFunc) WithTx(
	ctx context.Context,
	fn func(context.Context) error,
) error {
	return f(ctx, fn)
}

func newTestAccountPasswordService(
	repo AccountRepository,
	invalidate accountResetTokenInvalidatorFunc,
) AccountService {
	if invalidate == nil {
		invalidate = func(context.Context, uint64) error { return nil }
	}
	return NewAccountServiceWithCredentialAudit(
		repo,
		invalidate,
		accountChangeTransactorFunc(func(
			ctx context.Context,
			fn func(context.Context) error,
		) error {
			return fn(ctx)
		}),
		noopCredentialAuditTxLogger{},
	)
}

func TestAccountService_FindByEmail(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := &mockAccountRepository{
			findByEmailFn: func(_ context.Context, email string) (*model.Account, error) {
				return &model.Account{Email: email}, nil
			},
		}
		svc := NewAccountService(repo)
		acc, err := svc.FindByEmail(ctx, "test@example.com")
		assert.NoError(t, err)
		assert.Equal(t, "test@example.com", acc.Email)
	})

	t.Run("error", func(t *testing.T) {
		repo := &mockAccountRepository{
			findByEmailFn: func(_ context.Context, _ string) (*model.Account, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewAccountService(repo)
		acc, err := svc.FindByEmail(ctx, "test@example.com")
		assert.Error(t, err)
		assert.Nil(t, acc)
	})
}

func TestAccountService_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := &mockAccountRepository{
			findByIDFn: func(_ context.Context, id uint64) (*model.Account, error) {
				return &model.Account{ID: id}, nil
			},
		}
		svc := NewAccountService(repo)
		acc, err := svc.GetByID(ctx, 123)
		assert.NoError(t, err)
		assert.Equal(t, uint64(123), acc.ID)
	})

	t.Run("error", func(t *testing.T) {
		repo := &mockAccountRepository{
			findByIDFn: func(_ context.Context, _ uint64) (*model.Account, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewAccountService(repo)
		acc, err := svc.GetByID(ctx, 123)
		assert.Error(t, err)
		assert.Nil(t, acc)
	})
}

func TestAccountService_ChangePassword_UsesPasswordHashCompareAndSwap(t *testing.T) {
	ctx := context.Background()
	const (
		accountID       = uint64(123)
		currentPassword = "oldPassw0rd"
		newPassword     = "newPassw0rd"
	)
	currentHash, err := bcrypt.GenerateFromPassword(
		[]byte(currentPassword),
		bcrypt.MinCost,
	)
	require.NoError(t, err)
	previousUpdatedAt := time.Now().Add(-time.Minute).Truncate(time.Microsecond)

	events := make([]string, 0, 3)
	repo := &mockAccountRepository{
		findByIDFn: func(_ context.Context, id uint64) (*model.Account, error) {
			events = append(events, "find")
			assert.Equal(t, accountID, id)
			return &model.Account{
				ID:           id,
				PasswordHash: string(currentHash),
				UpdatedAt:    previousUpdatedAt,
			}, nil
		},
		compareAndSwapFn: func(
			_ context.Context,
			id uint64,
			expectedHash, newHash string,
			updatedAt time.Time,
		) (bool, error) {
			events = append(events, "compare-and-swap")
			assert.Equal(t, accountID, id)
			assert.Equal(t, string(currentHash), expectedHash)
			assert.True(t, updatedAt.After(previousUpdatedAt))
			assert.NoError(
				t,
				bcrypt.CompareHashAndPassword([]byte(newHash), []byte(newPassword)),
			)
			cost, costErr := bcrypt.Cost([]byte(newHash))
			require.NoError(t, costErr)
			assert.Equal(t, config.BcryptCost, cost)
			return true, nil
		},
	}

	service := newTestAccountPasswordService(
		repo,
		func(_ context.Context, id uint64) error {
			assert.Equal(t, accountID, id)
			events = append(events, "invalidate-reset-tokens")
			return nil
		},
	)
	changer, ok := service.(accountPasswordChangerTestPort)
	require.True(t, ok, "account service must expose the atomic password-change use case")

	require.NoError(t, changer.ChangePassword(
		ctx,
		accountID,
		currentPassword,
		newPassword,
		testPasswordChangeAudit(17),
	))
	assert.Equal(
		t,
		[]string{"find", "compare-and-swap", "invalidate-reset-tokens"},
		events,
	)
}

func TestAccountService_ChangePassword_FailsClosed(t *testing.T) {
	currentHash, err := bcrypt.GenerateFromPassword(
		[]byte("oldPassw0rd"),
		bcrypt.MinCost,
	)
	require.NoError(t, err)
	databaseUnavailable := errors.New("database unavailable")

	tests := []struct {
		name              string
		findResult        *model.Account
		findError         error
		currentPassword   string
		compareAndSwapOK  bool
		compareAndSwapErr error
		wantTarget        error
		wantCode          string
		wantCASCalls      int
	}{
		{
			name:            "account lookup failure",
			findError:       apperrors.WrapNotFound("account", "123"),
			currentPassword: "oldPassw0rd",
			wantTarget:      apperrors.ErrNotFound,
		},
		{
			name:            "nil account result",
			currentPassword: "oldPassw0rd",
			wantCode:        "INTERNAL",
		},
		{
			name: "wrong current password",
			findResult: &model.Account{
				ID:           123,
				PasswordHash: string(currentHash),
			},
			currentPassword: "wrongPassw0rd",
			wantTarget:      apperrors.ErrUnauthorized,
		},
		{
			name: "concurrent password change wins compare and swap",
			findResult: &model.Account{
				ID:           123,
				PasswordHash: string(currentHash),
			},
			currentPassword:  "oldPassw0rd",
			compareAndSwapOK: false,
			wantTarget:       apperrors.ErrUnauthorized,
			wantCASCalls:     1,
		},
		{
			name: "compare and swap persistence failure",
			findResult: &model.Account{
				ID:           123,
				PasswordHash: string(currentHash),
			},
			currentPassword:   "oldPassw0rd",
			compareAndSwapErr: databaseUnavailable,
			wantTarget:        databaseUnavailable,
			wantCASCalls:      1,
		},
		{
			name: "compare and swap cancellation preserves context error",
			findResult: &model.Account{
				ID:           123,
				PasswordHash: string(currentHash),
			},
			currentPassword:   "oldPassw0rd",
			compareAndSwapErr: context.Canceled,
			wantTarget:        context.Canceled,
			wantCASCalls:      1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			casCalls := 0
			repo := &mockAccountRepository{
				findByIDFn: func(context.Context, uint64) (*model.Account, error) {
					return test.findResult, test.findError
				},
				compareAndSwapFn: func(
					context.Context,
					uint64,
					string,
					string,
					time.Time,
				) (bool, error) {
					casCalls++
					return test.compareAndSwapOK, test.compareAndSwapErr
				},
			}
			service := newTestAccountPasswordService(repo, nil)
			changer, ok := service.(accountPasswordChangerTestPort)
			require.True(t, ok, "account service must expose the atomic password-change use case")

			changeErr := changer.ChangePassword(
				context.Background(),
				123,
				test.currentPassword,
				"newPassw0rd",
				testPasswordChangeAudit(17),
			)

			require.Error(t, changeErr)
			if test.wantTarget != nil {
				assert.ErrorIs(t, changeErr, test.wantTarget)
			}
			if test.wantCode != "" {
				var appErr *apperrors.AppError
				require.ErrorAs(t, changeErr, &appErr)
				assert.Equal(t, test.wantCode, appErr.Code)
			}
			assert.Equal(t, test.wantCASCalls, casCalls)
		})
	}
}

func TestAccountService_ChangePassword_TokenRevocationFailureRollsBackCredential(
	t *testing.T,
) {
	const (
		accountID       = uint64(123)
		currentPassword = "oldPassw0rd"
		newPassword     = "newPassw0rd"
	)
	currentHash, err := bcrypt.GenerateFromPassword(
		[]byte(currentPassword),
		bcrypt.MinCost,
	)
	require.NoError(t, err)
	committedHash := string(currentHash)
	stagedHash := ""
	rolledBack := false
	revokeErr := errors.New("reset token deletion failed")

	repo := &mockAccountRepository{
		findByIDFn: func(context.Context, uint64) (*model.Account, error) {
			return &model.Account{
				ID:           accountID,
				PasswordHash: committedHash,
				UpdatedAt:    time.Now().Add(-time.Minute),
			}, nil
		},
		compareAndSwapFn: func(
			context.Context,
			uint64,
			string,
			string,
			time.Time,
		) (bool, error) {
			stagedHash = newPassword
			return true, nil
		},
	}
	service := NewAccountServiceWithCredentialAudit(
		repo,
		accountResetTokenInvalidatorFunc(func(context.Context, uint64) error {
			return revokeErr
		}),
		accountChangeTransactorFunc(func(
			ctx context.Context,
			fn func(context.Context) error,
		) error {
			if txErr := fn(ctx); txErr != nil {
				rolledBack = true
				stagedHash = ""
				return txErr
			}
			committedHash = stagedHash
			return nil
		}),
		noopCredentialAuditTxLogger{},
	)
	changer := service.(AccountPasswordChanger)

	changeErr := changer.ChangePassword(
		context.Background(),
		accountID,
		currentPassword,
		newPassword,
		testPasswordChangeAudit(17),
	)

	require.Error(t, changeErr)
	assert.ErrorIs(t, changeErr, revokeErr)
	assert.True(t, rolledBack)
	assert.Empty(t, stagedHash)
	assert.Equal(t, string(currentHash), committedHash)
}

func TestNextAccountSessionEpoch_AlwaysAdvancesPersistedEpoch(t *testing.T) {
	previous := time.Unix(1_727_123_456, 789_012_000)

	assert.Equal(
		t,
		previous.Add(time.Microsecond),
		nextAccountSessionEpoch(previous, previous),
	)
	assert.Equal(
		t,
		previous.Add(time.Microsecond),
		nextAccountSessionEpoch(previous, previous.Add(-time.Second)),
	)
	assert.Equal(
		t,
		previous.Add(time.Second),
		nextAccountSessionEpoch(previous, previous.Add(time.Second)),
	)
}
