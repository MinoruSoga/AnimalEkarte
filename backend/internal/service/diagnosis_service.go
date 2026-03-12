// Package service provides business logic implementations for DiagnosisCategory and DiagnosisName entities.
package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- DiagnosisCategoryService ----

type DiagnosisCategoryService interface {
	List(ctx context.Context) ([]model.DiagnosisCategory, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.DiagnosisCategory, error)
	Create(ctx context.Context, category *model.DiagnosisCategory) error
	Update(ctx context.Context, category *model.DiagnosisCategory) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type diagnosisCategoryService struct {
	repo repository.DiagnosisCategoryRepository
}

func NewDiagnosisCategoryService(repo repository.DiagnosisCategoryRepository) DiagnosisCategoryService {
	return &diagnosisCategoryService{repo: repo}
}

func (s *diagnosisCategoryService) List(ctx context.Context) ([]model.DiagnosisCategory, error) {
	return s.repo.FindAll(ctx)
}
func (s *diagnosisCategoryService) GetByID(ctx context.Context, id uuid.UUID) (*model.DiagnosisCategory, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *diagnosisCategoryService) Create(ctx context.Context, category *model.DiagnosisCategory) error {
	return s.repo.Create(ctx, category)
}
func (s *diagnosisCategoryService) Update(ctx context.Context, category *model.DiagnosisCategory) error {
	return s.repo.Update(ctx, category)
}
func (s *diagnosisCategoryService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// ---- DiagnosisNameService ----

type DiagnosisNameService interface {
	List(ctx context.Context) ([]model.DiagnosisName, error)
	ListByCategoryID(ctx context.Context, categoryID uuid.UUID) ([]model.DiagnosisName, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.DiagnosisName, error)
	Create(ctx context.Context, name *model.DiagnosisName) error
	Update(ctx context.Context, name *model.DiagnosisName) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type diagnosisNameService struct {
	repo repository.DiagnosisNameRepository
}

func NewDiagnosisNameService(repo repository.DiagnosisNameRepository) DiagnosisNameService {
	return &diagnosisNameService{repo: repo}
}

func (s *diagnosisNameService) List(ctx context.Context) ([]model.DiagnosisName, error) {
	return s.repo.FindAll(ctx)
}
func (s *diagnosisNameService) ListByCategoryID(ctx context.Context, categoryID uuid.UUID) ([]model.DiagnosisName, error) {
	return s.repo.FindByCategoryID(ctx, categoryID)
}
func (s *diagnosisNameService) GetByID(ctx context.Context, id uuid.UUID) (*model.DiagnosisName, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *diagnosisNameService) Create(ctx context.Context, name *model.DiagnosisName) error {
	return s.repo.Create(ctx, name)
}
func (s *diagnosisNameService) Update(ctx context.Context, name *model.DiagnosisName) error {
	return s.repo.Update(ctx, name)
}
func (s *diagnosisNameService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
