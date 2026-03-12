package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type OwnerService interface {
	List(ctx context.Context, clinicID uuid.UUID, page, limit int, search string) ([]model.Owner, int64, error)
	GetByID(ctx context.Context, clinicID, id uuid.UUID) (*model.Owner, error)
	Create(ctx context.Context, owner *model.Owner) error
	Update(ctx context.Context, owner *model.Owner) error
	Delete(ctx context.Context, clinicID, id uuid.UUID) error
}

type ownerService struct {
	repo repository.OwnerRepository
}

func NewOwnerService(repo repository.OwnerRepository) OwnerService {
	return &ownerService{repo: repo}
}

func (s *ownerService) List(ctx context.Context, clinicID uuid.UUID, page, limit int, search string) ([]model.Owner, int64, error) {
	return s.repo.FindAll(ctx, clinicID, page, limit, search)
}

func (s *ownerService) GetByID(ctx context.Context, clinicID, id uuid.UUID) (*model.Owner, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *ownerService) Create(ctx context.Context, owner *model.Owner) error {
	return s.repo.Create(ctx, owner)
}

func (s *ownerService) Update(ctx context.Context, owner *model.Owner) error {
	return s.repo.Update(ctx, owner)
}

func (s *ownerService) Delete(ctx context.Context, clinicID, id uuid.UUID) error {
	return s.repo.Delete(ctx, clinicID, id)
}
