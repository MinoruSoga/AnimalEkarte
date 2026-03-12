package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type AccountingService interface {
	List(ctx context.Context, clinicID uuid.UUID, petID *uuid.UUID, ownerID *uuid.UUID, status *string, page, limit int) ([]model.Billing, int64, error)
	GetByID(ctx context.Context, clinicID, id uuid.UUID) (*model.Billing, error)
	Create(ctx context.Context, clinicID uuid.UUID, accounting *model.Billing) error
	Update(ctx context.Context, clinicID uuid.UUID, accounting *model.Billing) error
	Delete(ctx context.Context, clinicID, id uuid.UUID) error
}

type accountingService struct {
	repo repository.AccountingRepository
}

func NewAccountingService(repo repository.AccountingRepository) AccountingService {
	return &accountingService{repo: repo}
}

func (s *accountingService) List(ctx context.Context, clinicID uuid.UUID, petID *uuid.UUID, ownerID *uuid.UUID, status *string, page, limit int) ([]model.Billing, int64, error) {
	return s.repo.FindAll(ctx, clinicID, petID, ownerID, status, page, limit)
}

func (s *accountingService) GetByID(ctx context.Context, clinicID, id uuid.UUID) (*model.Billing, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *accountingService) Create(ctx context.Context, clinicID uuid.UUID, accounting *model.Billing) error {
	return s.repo.Create(ctx, clinicID, accounting)
}

func (s *accountingService) Update(ctx context.Context, clinicID uuid.UUID, accounting *model.Billing) error {
	return s.repo.Update(ctx, clinicID, accounting)
}

func (s *accountingService) Delete(ctx context.Context, clinicID, id uuid.UUID) error {
	return s.repo.Delete(ctx, clinicID, id)
}
