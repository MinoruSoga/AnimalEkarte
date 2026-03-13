package service

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type HospitalizationService interface {
	List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status *string, page, limit int) ([]model.Hospitalization, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error)
	Create(ctx context.Context, hospitalization *model.Hospitalization) error
	Update(ctx context.Context, hospitalization *model.Hospitalization) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type hospitalizationService struct {
	repo repository.HospitalizationRepository
}

func NewHospitalizationService(repo repository.HospitalizationRepository) HospitalizationService {
	return &hospitalizationService{repo: repo}
}

func (s *hospitalizationService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status *string, page, limit int) ([]model.Hospitalization, int64, error) {
	return s.repo.FindAll(ctx, clinicID, petID, ownerID, status, page, limit)
}

func (s *hospitalizationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *hospitalizationService) Create(ctx context.Context, hospitalization *model.Hospitalization) error {
	return s.repo.Create(ctx, hospitalization)
}

func (s *hospitalizationService) Update(ctx context.Context, hospitalization *model.Hospitalization) error {
	return s.repo.Update(ctx, hospitalization)
}

func (s *hospitalizationService) Delete(ctx context.Context, clinicID, id uint64) error {
	return s.repo.Delete(ctx, clinicID, id)
}
