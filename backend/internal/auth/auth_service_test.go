package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/seedlogin"
)

type mockStaffAccountFinder struct {
	findByAccountIDFn func(ctx context.Context, accountID uint64) (*model.Staff, error)
}

type authServiceEffectivePermissionStub struct {
	getFn func(ctx context.Context, staffID, clinicID uint64) ([]model.PermissionGroupRule, error)
}

func (s authServiceEffectivePermissionStub) GetEffectivePermissions(
	ctx context.Context,
	staffID, clinicID uint64,
) ([]model.PermissionGroupRule, error) {
	return s.getFn(ctx, staffID, clinicID)
}

func (m *mockStaffAccountFinder) FindByAccountID(
	ctx context.Context,
	accountID uint64,
) (*model.Staff, error) {
	if m.findByAccountIDFn != nil {
		return m.findByAccountIDFn(ctx, accountID)
	}
	return &model.Staff{ID: accountID, IsActive: true}, nil
}

func newServiceForAuthenticateTest(
	accountRepo *mockAccountRepository,
	staff *mockStaffAccountFinder,
) Service {
	return NewService(
		NewAccountService(accountRepo),
		staff,
		nil,
	)
}

func TestService_Authenticate(t *testing.T) {
	ctx := context.Background()

	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	require.NoError(t, err)
	passwordHash := string(hash)

	t.Run("correct password returns account and staff", func(t *testing.T) {
		svc := newServiceForAuthenticateTest(
			&mockAccountRepository{
				findByEmailFn: func(_ context.Context, email string) (*model.Account, error) {
					assert.Equal(t, "user@test.com", email)
					return &model.Account{ID: 1, Email: "user@test.com", IsActive: true, PasswordHash: passwordHash}, nil
				},
			},
			&mockStaffAccountFinder{
				findByAccountIDFn: func(_ context.Context, accountID uint64) (*model.Staff, error) {
					assert.Equal(t, uint64(1), accountID)
					return &model.Staff{ID: 10, IsActive: true}, nil
				},
			},
		)

		account, staff, err := svc.AuthenticateUser(ctx, "user@test.com", "correct-password")

		require.NoError(t, err)
		require.NotNil(t, account)
		require.NotNil(t, staff)
		assert.Equal(t, uint64(1), account.ID)
		assert.Equal(t, uint64(10), staff.ID)
	})

	t.Run("staging catalog email accepts shared demo password when hash differs", func(t *testing.T) {
		t.Setenv("APP_ENV", "staging")
		svc := newServiceForAuthenticateTest(
			&mockAccountRepository{
				findByEmailFn: func(_ context.Context, email string) (*model.Account, error) {
					assert.Equal(t, seedlogin.Catalog()[0].Email, email)
					return &model.Account{
						ID:           1,
						Email:        email,
						IsActive:     true,
						PasswordHash: passwordHash,
					}, nil
				},
			},
			&mockStaffAccountFinder{
				findByAccountIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
					return &model.Staff{ID: 10, IsActive: true}, nil
				},
			},
		)

		account, staff, err := svc.AuthenticateUser(
			ctx,
			seedlogin.Catalog()[0].Email,
			seedlogin.SharedPassword,
		)
		require.NoError(t, err)
		require.NotNil(t, account)
		require.NotNil(t, staff)
	})

	t.Run("production rejects shared demo password when hash differs", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		svc := newServiceForAuthenticateTest(
			&mockAccountRepository{
				findByEmailFn: func(_ context.Context, email string) (*model.Account, error) {
					return &model.Account{
						ID:           1,
						Email:        email,
						IsActive:     true,
						PasswordHash: passwordHash,
					}, nil
				},
			},
			&mockStaffAccountFinder{},
		)

		account, staff, err := svc.AuthenticateUser(
			ctx,
			seedlogin.Catalog()[0].Email,
			seedlogin.SharedPassword,
		)
		assert.Nil(t, account)
		assert.Nil(t, staff)
		require.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrUnauthorized))
	})

	t.Run("staging does not accept shared password for non-catalog email", func(t *testing.T) {
		t.Setenv("APP_ENV", "staging")
		svc := newServiceForAuthenticateTest(
			&mockAccountRepository{
				findByEmailFn: func(_ context.Context, _ string) (*model.Account, error) {
					return &model.Account{
						ID:           1,
						Email:        "stg-operator@example.test",
						IsActive:     true,
						PasswordHash: passwordHash,
					}, nil
				},
			},
			&mockStaffAccountFinder{},
		)

		account, staff, err := svc.AuthenticateUser(
			ctx,
			"stg-operator@example.test",
			seedlogin.SharedPassword,
		)
		assert.Nil(t, account)
		assert.Nil(t, staff)
		require.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrUnauthorized))
	})

	t.Run("wrong password returns unauthorized with wrong-password sentinel", func(t *testing.T) {
		svc := newServiceForAuthenticateTest(
			&mockAccountRepository{
				findByEmailFn: func(_ context.Context, _ string) (*model.Account, error) {
					return &model.Account{ID: 1, IsActive: true, PasswordHash: passwordHash}, nil
				},
			},
			&mockStaffAccountFinder{},
		)

		account, staff, err := svc.AuthenticateUser(ctx, "user@test.com", "wrong-password")

		assert.Nil(t, account)
		assert.Nil(t, staff)
		require.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrUnauthorized))
		accountID, ok := IsAuthenticateWrongPassword(err)
		assert.True(t, ok)
		assert.Equal(t, uint64(1), accountID)
	})

	t.Run("inactive staff returns unauthorized", func(t *testing.T) {
		svc := newServiceForAuthenticateTest(
			&mockAccountRepository{
				findByEmailFn: func(_ context.Context, _ string) (*model.Account, error) {
					return &model.Account{ID: 1, IsActive: true, PasswordHash: passwordHash}, nil
				},
			},
			&mockStaffAccountFinder{
				findByAccountIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
					return &model.Staff{ID: 10, IsActive: false}, nil
				},
			},
		)

		account, staff, err := svc.AuthenticateUser(ctx, "user@test.com", "correct-password")

		assert.Nil(t, account)
		assert.Nil(t, staff)
		require.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrUnauthorized))
		_, ok := IsAuthenticateWrongPassword(err)
		assert.False(t, ok)
	})
}

func TestService_Authenticate_InvalidCredentialResponsesAreUniform(t *testing.T) {
	ctx := context.Background()
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	require.NoError(t, err)
	const publicMessage = "メールアドレスまたはパスワードが正しくありません"

	tests := []struct {
		name        string
		account     *model.Account
		accountErr  error
		staff       *model.Staff
		password    string
		expectAudit bool
	}{
		{
			name:       "nonexistent account",
			accountErr: apperrors.WrapNotFound("account", "lookup"),
			password:   "correct-password",
		},
		{
			name: "bad password",
			account: &model.Account{
				ID:           1,
				IsActive:     true,
				PasswordHash: string(hash),
			},
			staff:       &model.Staff{ID: 10, IsActive: true},
			password:    "wrong-password",
			expectAudit: true,
		},
		{
			name: "inactive account",
			account: &model.Account{
				ID:           1,
				IsActive:     false,
				PasswordHash: string(hash),
			},
			password: "correct-password",
		},
		{
			name: "inactive staff",
			account: &model.Account{
				ID:           1,
				IsActive:     true,
				PasswordHash: string(hash),
			},
			staff:    &model.Staff{ID: 10, IsActive: false},
			password: "correct-password",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc := newServiceForAuthenticateTest(
				&mockAccountRepository{
					findByEmailFn: func(context.Context, string) (*model.Account, error) {
						return test.account, test.accountErr
					},
				},
				&mockStaffAccountFinder{
					findByAccountIDFn: func(context.Context, uint64) (*model.Staff, error) {
						return test.staff, nil
					},
				},
			)

			account, staff, authenticateErr := svc.AuthenticateUser(
				ctx,
				"private@example.test",
				test.password,
			)

			assert.Nil(t, account)
			assert.Nil(t, staff)
			require.Error(t, authenticateErr)
			assert.True(t, errors.Is(authenticateErr, apperrors.ErrUnauthorized))
			assert.Contains(t, authenticateErr.Error(), publicMessage)
			assert.Equal(t, test.expectAudit, func() bool {
				_, ok := IsAuthenticateWrongPassword(authenticateErr)
				return ok
			}())
		})
	}
}

func TestService_Authenticate_UnknownAndInactiveAccountsPerformOneCost12Comparison(
	t *testing.T,
) {
	const submittedPassword = "probe-password-123"
	const realPasswordHash = "real-account-password-hash"

	cases := []struct {
		name         string
		account      *model.Account
		accountError error
		expectedHash string
	}{
		{
			name:         "nonexistent account",
			accountError: apperrors.WrapNotFound("account", "lookup"),
			expectedHash: dummyPasswordHash,
		},
		{
			name: "inactive account",
			account: &model.Account{
				ID:           1,
				IsActive:     false,
				PasswordHash: realPasswordHash,
			},
			expectedHash: dummyPasswordHash,
		},
		{
			name: "deleted account",
			account: &model.Account{
				ID:           1,
				IsActive:     true,
				PasswordHash: realPasswordHash,
				DeletedAt:    gorm.DeletedAt{Valid: true},
			},
			expectedHash: dummyPasswordHash,
		},
		{
			name: "real wrong password",
			account: &model.Account{
				ID:           1,
				IsActive:     true,
				PasswordHash: realPasswordHash,
			},
			expectedHash: realPasswordHash,
		},
	}

	cost, err := bcrypt.Cost([]byte(dummyPasswordHash))
	require.NoError(t, err)
	require.Equal(t, config.BcryptCost, cost)
	require.NoError(t, defaultPasswordCompare(
		[]byte(dummyPasswordHash),
		[]byte(dummyPasswordCandidate),
	))

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			service := newServiceForAuthenticateTest(
				&mockAccountRepository{
					findByEmailFn: func(context.Context, string) (*model.Account, error) {
						return test.account, test.accountError
					},
				},
				&mockStaffAccountFinder{},
			).(*authService)
			comparisonCount := 0
			service.comparePassword = func(hashedPassword, password []byte) error {
				comparisonCount++
				assert.Equal(t, test.expectedHash, string(hashedPassword))
				assert.Equal(t, submittedPassword, string(password))
				return bcrypt.ErrMismatchedHashAndPassword
			}

			account, staff, authenticateErr := service.AuthenticateUser(
				context.Background(),
				"private@example.test",
				submittedPassword,
			)

			assert.Nil(t, account)
			assert.Nil(t, staff)
			require.Error(t, authenticateErr)
			assert.ErrorIs(t, authenticateErr, apperrors.ErrUnauthorized)
			assert.Equal(t, 1, comparisonCount)
		})
	}
}

func TestService_CalculateEffectivePermissions_ErrorReturnsEmptyMap(t *testing.T) {
	svc := NewService(nil, nil, authServiceEffectivePermissionStub{
		getFn: func(_ context.Context, staffID, clinicID uint64) ([]model.PermissionGroupRule, error) {
			assert.Equal(t, uint64(11), staffID)
			assert.Equal(t, uint64(22), clinicID)
			return nil, errors.New("permission repository unavailable")
		},
	})

	permissions := svc.CalculateEffectivePermissions(
		context.Background(),
		false,
		11,
		22,
	)

	require.NotNil(t, permissions)
	assert.Empty(t, permissions)
}

func TestService_ClinicResolutionHelpers(t *testing.T) {
	svc := NewService(nil, nil, nil)

	t.Run("main assignment wins while every clinic remains in token scope", func(t *testing.T) {
		mainClinicID, clinicIDs := svc.ResolveClinicInfo([]model.StaffClinicAssignment{
			{ClinicID: 8},
			{ClinicID: 4, IsMain: true},
			{ClinicID: 9},
		})

		assert.Equal(t, "4", mainClinicID)
		assert.Equal(t, []uint64{8, 4, 9}, clinicIDs)
	})

	t.Run("first assignment is deterministic fallback", func(t *testing.T) {
		mainClinicID, clinicIDs := svc.ResolveClinicInfo([]model.StaffClinicAssignment{
			{ClinicID: 8},
			{ClinicID: 4},
		})

		assert.Equal(t, "8", mainClinicID)
		assert.Equal(t, []uint64{8, 4}, clinicIDs)
	})

	t.Run("unassigned system admin uses first clinic", func(t *testing.T) {
		mainClinicID := svc.ResolveSystemAdminMainClinicID(
			"",
			true,
			[]model.Clinic{{ID: 31, IsActive: true}, {ID: 12, IsActive: true}},
		)

		assert.Equal(t, "31", mainClinicID)
	})

	t.Run("system admin fallback skips zero and inactive clinics", func(t *testing.T) {
		mainClinicID := svc.ResolveSystemAdminMainClinicID(
			"",
			true,
			[]model.Clinic{
				{ID: 0, IsActive: true},
				{ID: 31, IsActive: false},
				{ID: 12, IsActive: true},
			},
		)

		assert.Equal(t, "12", mainClinicID)
	})

	t.Run("system admin fallback fails closed without an active clinic", func(t *testing.T) {
		mainClinicID := svc.ResolveSystemAdminMainClinicID(
			"",
			true,
			[]model.Clinic{
				{ID: 0, IsActive: true},
				{ID: 31, IsActive: false},
			},
		)

		assert.Empty(t, mainClinicID)
	})

	t.Run("non-admin never broadens scope and admin main must remain active", func(t *testing.T) {
		assert.Empty(t, svc.ResolveSystemAdminMainClinicID(
			"",
			false,
			[]model.Clinic{{ID: 31}},
		))
		assert.Equal(t, "31", svc.ResolveSystemAdminMainClinicID(
			"12",
			true,
			[]model.Clinic{
				{ID: 12, IsActive: false},
				{ID: 31, IsActive: true},
			},
		))
		assert.Equal(t, "12", svc.ResolveSystemAdminMainClinicID(
			"12",
			true,
			[]model.Clinic{
				{ID: 12, IsActive: true},
				{ID: 31, IsActive: true},
			},
		))
	})
}
