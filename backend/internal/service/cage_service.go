// Package service provides business logic implementations for Cage entity.
package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- CageService ----

type CageService interface {
	List(ctx context.Context, cageType *string) ([]model.Cage, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Cage, error)
	Create(ctx context.Context, cage *model.Cage) error
	Update(ctx context.Context, cage *model.Cage) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type cageService struct{ repo repository.CageRepository }

func NewCageService(repo repository.CageRepository) CageService {
	return &cageService{repo: repo}
}

func (s *cageService) List(ctx context.Context, cageType *string) ([]model.Cage, error) {
	return s.repo.FindAll(ctx, cageType)
}
func (s *cageService) GetByID(ctx context.Context, id uuid.UUID) (*model.Cage, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *cageService) Create(ctx context.Context, cage *model.Cage) error {
	return s.repo.Create(ctx, cage)
}
func (s *cageService) Update(ctx context.Context, cage *model.Cage) error {
	return s.repo.Update(ctx, cage)
}
func (s *cageService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
