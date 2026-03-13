// Package service provides business logic implementations for Vaccine entity.
package service

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- VaccineService ----

type VaccineService interface {
	List(ctx context.Context, species *string) ([]model.Vaccine, error)
	GetByID(ctx context.Context, id uint64) (*model.Vaccine, error)
	Create(ctx context.Context, vaccine *model.Vaccine) error
	Update(ctx context.Context, vaccine *model.Vaccine) error
	Delete(ctx context.Context, id uint64) error
}

type vaccineService struct{ repo repository.VaccineRepository }

func NewVaccineService(repo repository.VaccineRepository) VaccineService {
	return &vaccineService{repo: repo}
}

func (s *vaccineService) List(ctx context.Context, species *string) ([]model.Vaccine, error) {
	return s.repo.FindAll(ctx, species)
}
func (s *vaccineService) GetByID(ctx context.Context, id uint64) (*model.Vaccine, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *vaccineService) Create(ctx context.Context, vaccine *model.Vaccine) error {
	return s.repo.Create(ctx, vaccine)
}
func (s *vaccineService) Update(ctx context.Context, vaccine *model.Vaccine) error {
	return s.repo.Update(ctx, vaccine)
}
func (s *vaccineService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}
