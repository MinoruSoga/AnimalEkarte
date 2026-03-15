// Package service provides business logic implementations for Insurance entity.
package service

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- InsuranceService ----

type InsuranceService interface {
	List(ctx context.Context) ([]model.Insurance, error)
	GetByID(ctx context.Context, id uint64) (*model.Insurance, error)
	Create(ctx context.Context, insurance *model.Insurance) error
	Update(ctx context.Context, insurance *model.Insurance) error
	Delete(ctx context.Context, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type insuranceService struct {
	repo repository.InsuranceRepository
}

func NewInsuranceService(repo repository.InsuranceRepository) InsuranceService {
	return &insuranceService{repo: repo}
}

func (s *insuranceService) List(ctx context.Context) ([]model.Insurance, error) {
	return s.repo.FindAll(ctx)
}
func (s *insuranceService) GetByID(ctx context.Context, id uint64) (*model.Insurance, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *insuranceService) Create(ctx context.Context, insurance *model.Insurance) error {
	return s.repo.Create(ctx, insurance)
}
func (s *insuranceService) Update(ctx context.Context, insurance *model.Insurance) error {
	return s.repo.Update(ctx, insurance)
}
func (s *insuranceService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *insuranceService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return s.repo.Reorder(ctx, clinicID, ids)
}
