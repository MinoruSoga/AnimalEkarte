package service

import (
	"context"
	"fmt"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type AccountService interface {
	FindByEmail(ctx context.Context, email string) (*model.Account, error)
	GetByID(ctx context.Context, id uint64) (*model.Account, error)
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
	account, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperrors.Wrap(err, fmt.Sprintf("failed to get account: %d", id))
	}
	return account, nil
}
