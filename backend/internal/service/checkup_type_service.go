// Package service provides business logic implementations for CheckupType entity.
package service

import (
	"context"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- CheckupTypeService ----

type CheckupTypeService interface {
	List(ctx context.Context) ([]model.CheckupType, error)
	GetByID(ctx context.Context, id uint64) (*model.CheckupType, error)
	Create(ctx context.Context, checkupType *model.CheckupType) error
	Update(ctx context.Context, checkupType *model.CheckupType) error
	Delete(ctx context.Context, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type checkupTypeService struct {
	repo repository.CheckupTypeRepository
}

func NewCheckupTypeService(repo repository.CheckupTypeRepository) CheckupTypeService {
	return &checkupTypeService{repo: repo}
}

func (s *checkupTypeService) List(ctx context.Context) ([]model.CheckupType, error) {
	return s.repo.FindAll(ctx)
}
func (s *checkupTypeService) GetByID(ctx context.Context, id uint64) (*model.CheckupType, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *checkupTypeService) Create(ctx context.Context, checkupType *model.CheckupType) error {
	return s.repo.Create(ctx, checkupType)
}
func (s *checkupTypeService) Update(ctx context.Context, checkupType *model.CheckupType) error {
	return s.repo.Update(ctx, checkupType)
}
func (s *checkupTypeService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *checkupTypeService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	return s.repo.Reorder(ctx, clinicID, ids)
}
