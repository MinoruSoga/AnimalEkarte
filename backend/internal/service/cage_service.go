// Package service provides business logic implementations for Cage entity.
package service

import (
	"context"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- CageService ----

type CageService interface {
	List(ctx context.Context, cageType *string) ([]model.Cage, error)
	GetByID(ctx context.Context, id uint64) (*model.Cage, error)
	Create(ctx context.Context, cage *model.Cage) error
	Update(ctx context.Context, cage *model.Cage) error
	Delete(ctx context.Context, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type cageService struct{ repo repository.CageRepository }

func NewCageService(repo repository.CageRepository) CageService {
	return &cageService{repo: repo}
}

func (s *cageService) List(ctx context.Context, cageType *string) ([]model.Cage, error) {
	return s.repo.FindAll(ctx, cageType)
}
func (s *cageService) GetByID(ctx context.Context, id uint64) (*model.Cage, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *cageService) Create(ctx context.Context, cage *model.Cage) error {
	return s.repo.Create(ctx, cage)
}
func (s *cageService) Update(ctx context.Context, cage *model.Cage) error {
	return s.repo.Update(ctx, cage)
}
func (s *cageService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *cageService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	return s.repo.Reorder(ctx, clinicID, ids)
}
