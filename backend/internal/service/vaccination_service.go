package service

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type VaccinationService interface {
	List(ctx context.Context, petID *uint64, ownerID *uint64, page, limit int) ([]model.Vaccination, int64, error)
	GetByID(ctx context.Context, id uint64) (*model.Vaccination, error)
	Create(ctx context.Context, vaccination *model.Vaccination) error
	Update(ctx context.Context, vaccination *model.Vaccination) error
	Delete(ctx context.Context, id uint64) error
}

type vaccinationService struct {
	repo repository.VaccinationRepository
}

func NewVaccinationService(repo repository.VaccinationRepository) VaccinationService {
	return &vaccinationService{repo: repo}
}

func (s *vaccinationService) List(ctx context.Context, petID *uint64, ownerID *uint64, page, limit int) ([]model.Vaccination, int64, error) {
	return s.repo.FindAll(ctx, petID, ownerID, page, limit)
}

func (s *vaccinationService) GetByID(ctx context.Context, id uint64) (*model.Vaccination, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *vaccinationService) Create(ctx context.Context, vaccination *model.Vaccination) error {
	return s.repo.Create(ctx, vaccination)
}

func (s *vaccinationService) Update(ctx context.Context, vaccination *model.Vaccination) error {
	return s.repo.Update(ctx, vaccination)
}

func (s *vaccinationService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}
