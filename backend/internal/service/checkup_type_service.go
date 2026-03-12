// Package service provides business logic implementations for CheckupType entity.
package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- CheckupTypeService ----

type CheckupTypeService interface {
	List(ctx context.Context) ([]model.CheckupType, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.CheckupType, error)
	Create(ctx context.Context, checkupType *model.CheckupType) error
	Update(ctx context.Context, checkupType *model.CheckupType) error
	Delete(ctx context.Context, id uuid.UUID) error
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
func (s *checkupTypeService) GetByID(ctx context.Context, id uuid.UUID) (*model.CheckupType, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *checkupTypeService) Create(ctx context.Context, checkupType *model.CheckupType) error {
	return s.repo.Create(ctx, checkupType)
}
func (s *checkupTypeService) Update(ctx context.Context, checkupType *model.CheckupType) error {
	return s.repo.Update(ctx, checkupType)
}
func (s *checkupTypeService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
