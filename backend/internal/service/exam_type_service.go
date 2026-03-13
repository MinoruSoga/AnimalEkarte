// Package service provides business logic implementations for ExamType entity.
package service

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- ExamTypeService ----

type ExamTypeService interface {
	List(ctx context.Context) ([]model.ExamType, error)
	GetByID(ctx context.Context, id uint64) (*model.ExamType, error)
	Create(ctx context.Context, exType *model.ExamType) error
	Update(ctx context.Context, exType *model.ExamType) error
	Delete(ctx context.Context, id uint64) error
}

type examTypeService struct{ repo repository.ExamTypeRepository }

func NewExamTypeService(repo repository.ExamTypeRepository) ExamTypeService {
	return &examTypeService{repo: repo}
}

func (s *examTypeService) List(ctx context.Context) ([]model.ExamType, error) {
	return s.repo.FindAll(ctx)
}
func (s *examTypeService) GetByID(ctx context.Context, id uint64) (*model.ExamType, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *examTypeService) Create(ctx context.Context, exType *model.ExamType) error {
	return s.repo.Create(ctx, exType)
}
func (s *examTypeService) Update(ctx context.Context, exType *model.ExamType) error {
	return s.repo.Update(ctx, exType)
}
func (s *examTypeService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}
