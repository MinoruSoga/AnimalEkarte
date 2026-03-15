// Package service provides business logic implementations for HospitalizationPlan entity.
package service

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- HospitalizationPlanService ----

type HospitalizationPlanService interface {
	List(ctx context.Context) ([]model.HospitalizationPlan, error)
	GetByID(ctx context.Context, id uint64) (*model.HospitalizationPlan, error)
	Create(ctx context.Context, plan *model.HospitalizationPlan) error
	Update(ctx context.Context, plan *model.HospitalizationPlan) error
	Delete(ctx context.Context, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type hospitalizationPlanService struct {
	repo repository.HospitalizationPlanRepository
}

func NewHospitalizationPlanService(repo repository.HospitalizationPlanRepository) HospitalizationPlanService {
	return &hospitalizationPlanService{repo: repo}
}

func (s *hospitalizationPlanService) List(ctx context.Context) ([]model.HospitalizationPlan, error) {
	return s.repo.FindAll(ctx)
}
func (s *hospitalizationPlanService) GetByID(ctx context.Context, id uint64) (*model.HospitalizationPlan, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *hospitalizationPlanService) Create(ctx context.Context, plan *model.HospitalizationPlan) error {
	return s.repo.Create(ctx, plan)
}
func (s *hospitalizationPlanService) Update(ctx context.Context, plan *model.HospitalizationPlan) error {
	return s.repo.Update(ctx, plan)
}
func (s *hospitalizationPlanService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *hospitalizationPlanService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return s.repo.Reorder(ctx, clinicID, ids)
}
