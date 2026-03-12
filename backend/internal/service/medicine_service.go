// Package service provides business logic implementations for Medicine entity.
package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- MedicineService ----

type MedicineService interface {
	List(ctx context.Context) ([]model.Medicine, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Medicine, error)
	Create(ctx context.Context, medicine *model.Medicine) error
	Update(ctx context.Context, medicine *model.Medicine) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type medicineService struct{ repo repository.MedicineRepository }

func NewMedicineService(repo repository.MedicineRepository) MedicineService {
	return &medicineService{repo: repo}
}

func (s *medicineService) List(ctx context.Context) ([]model.Medicine, error) {
	return s.repo.FindAll(ctx)
}
func (s *medicineService) GetByID(ctx context.Context, id uuid.UUID) (*model.Medicine, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *medicineService) Create(ctx context.Context, medicine *model.Medicine) error {
	return s.repo.Create(ctx, medicine)
}
func (s *medicineService) Update(ctx context.Context, medicine *model.Medicine) error {
	return s.repo.Update(ctx, medicine)
}
func (s *medicineService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
