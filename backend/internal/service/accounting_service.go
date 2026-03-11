package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type AccountingService interface {
	List(ctx context.Context, status *string, page, limit int) ([]model.Accounting, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Accounting, error)
	Create(ctx context.Context, accounting *model.Accounting) error
	Update(ctx context.Context, accounting *model.Accounting) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type accountingService struct {
	repo repository.AccountingRepository
}

func NewAccountingService(repo repository.AccountingRepository) AccountingService {
	return &accountingService{repo: repo}
}

func (s *accountingService) List(ctx context.Context, status *string, page, limit int) ([]model.Accounting, int64, error) {
	return s.repo.FindAll(ctx, status, page, limit)
}

func (s *accountingService) GetByID(ctx context.Context, id uuid.UUID) (*model.Accounting, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *accountingService) Create(ctx context.Context, accounting *model.Accounting) error {
	return s.repo.Create(ctx, accounting)
}

func (s *accountingService) Update(ctx context.Context, accounting *model.Accounting) error {
	return s.repo.Update(ctx, accounting)
}

func (s *accountingService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
