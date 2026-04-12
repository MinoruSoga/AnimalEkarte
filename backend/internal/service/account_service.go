package service

import (
	"context"
	"fmt"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type AccountService interface {
	FindByEmail(ctx context.Context, email string) (*model.Account, error)
	GetByID(ctx context.Context, id uint64) (*model.Account, error)
	// UpdatePasswordHash はアカウントのパスワードハッシュを更新する。
	UpdatePasswordHash(ctx context.Context, accountID uint64, newHash string) error
}

type accountService struct {
	repo repository.AccountRepository
}

func NewAccountService(repo repository.AccountRepository) AccountService {
	return &accountService{repo: repo}
}

func (s *accountService) FindByEmail(ctx context.Context, email string) (*model.Account, error) {
	account, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find account by email")
	}
	return account, nil
}

func (s *accountService) GetByID(ctx context.Context, id uint64) (*model.Account, error) {
	account, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.Wrap(err, fmt.Sprintf("failed to get account: %d", id))
	}
	return account, nil
}

func (s *accountService) UpdatePasswordHash(ctx context.Context, accountID uint64, newHash string) error {
	if err := s.repo.Update(ctx, accountID, map[string]any{"password_hash": newHash}); err != nil {
		return apperrors.Wrap(err, "failed to update password hash")
	}
	slog.InfoContext(ctx, "password hash updated",
		slog.Uint64("account_id", accountID))
	return nil
}
