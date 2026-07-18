package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func newAuthServiceForAuthenticateTest(accountRepo *mockAccountRepository, staffRepo *mockStaffRepository) AuthService {
	return NewAuthService(
		NewAccountService(accountRepo),
		NewStaffService(staffRepo, nil, nil, nil, nil, nil, nil, nil, nil),
		nil,
	)
}

func TestAuthService_Authenticate(t *testing.T) {
	ctx := context.Background()

	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	require.NoError(t, err)
	passwordHash := string(hash)

	t.Run("correct password returns account and staff", func(t *testing.T) {
		svc := newAuthServiceForAuthenticateTest(
			&mockAccountRepository{
				findByEmailFn: func(_ context.Context, email string) (*model.Account, error) {
					assert.Equal(t, "user@test.com", email)
					return &model.Account{ID: 1, Email: "user@test.com", IsActive: true, PasswordHash: passwordHash}, nil
				},
			},
			&mockStaffRepository{
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

	t.Run("wrong password returns unauthorized with wrong-password sentinel", func(t *testing.T) {
		svc := newAuthServiceForAuthenticateTest(
			&mockAccountRepository{
				findByEmailFn: func(_ context.Context, _ string) (*model.Account, error) {
					return &model.Account{ID: 1, IsActive: true, PasswordHash: passwordHash}, nil
				},
			},
			&mockStaffRepository{},
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
		svc := newAuthServiceForAuthenticateTest(
			&mockAccountRepository{
				findByEmailFn: func(_ context.Context, _ string) (*model.Account, error) {
					return &model.Account{ID: 1, IsActive: true, PasswordHash: passwordHash}, nil
				},
			},
			&mockStaffRepository{
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
