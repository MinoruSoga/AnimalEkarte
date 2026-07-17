package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type mockAccountRepository struct {
	repository.AccountRepository
	findByIDFn    func(ctx context.Context, id uint64) (*model.Account, error)
	findByEmailFn func(ctx context.Context, email string) (*model.Account, error)
	updateFn      func(ctx context.Context, id uint64, fields map[string]any) error
}

func (m *mockAccountRepository) FindByID(ctx context.Context, id uint64) (*model.Account, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return &model.Account{ID: id}, nil
}

func (m *mockAccountRepository) FindByEmail(ctx context.Context, email string) (*model.Account, error) {
	if m.findByEmailFn != nil {
		return m.findByEmailFn(ctx, email)
	}
	return &model.Account{Email: email}, nil
}

func (m *mockAccountRepository) Update(ctx context.Context, id uint64, fields map[string]any) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, fields)
	}
	return nil
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

func TestAccountService_UpdatePasswordHash(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := &mockAccountRepository{
			updateFn: func(_ context.Context, id uint64, fields map[string]any) error {
				assert.Equal(t, uint64(123), id)
				assert.Equal(t, "new_hash", fields["password_hash"])
				return nil
			},
		}
		svc := NewAccountService(repo)
		err := svc.UpdatePasswordHash(ctx, 123, "new_hash")
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		repo := &mockAccountRepository{
			updateFn: func(_ context.Context, _ uint64, _ map[string]any) error {
				return errors.New("db error")
			},
		}
		svc := NewAccountService(repo)
		err := svc.UpdatePasswordHash(ctx, 123, "new_hash")
		assert.Error(t, err)
	})
}
