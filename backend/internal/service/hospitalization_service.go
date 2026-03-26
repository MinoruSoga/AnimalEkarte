package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type HospitalizationService interface {
	List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Hospitalization, int64, error)
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

func (s *hospitalizationService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Hospitalization, int64, error) {
	return s.repo.FindAll(ctx, clinicID, petID, ownerID, status, startDate, endDate, page, limit)
}

func (s *hospitalizationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *hospitalizationService) Create(ctx context.Context, hospitalization *model.Hospitalization) error {
	if err := s.repo.Create(ctx, hospitalization); err != nil {
		return fmt.Errorf("failed to create hospitalization: %w", err)
	}
	slog.InfoContext(ctx, "hospitalization created",
		slog.Uint64("hospitalization_id", hospitalization.ID),
		slog.Uint64("clinic_id", hospitalization.ClinicID))
	return nil
}

func (s *hospitalizationService) Update(ctx context.Context, hospitalization *model.Hospitalization) error {
	if err := s.repo.Update(ctx, hospitalization); err != nil {
		return fmt.Errorf("failed to update hospitalization: %w", err)
	}
	slog.InfoContext(ctx, "hospitalization updated",
		slog.Uint64("hospitalization_id", hospitalization.ID),
		slog.Uint64("clinic_id", hospitalization.ClinicID))
	return nil
}

func (s *hospitalizationService) Delete(ctx context.Context, clinicID, id uint64)
error {
	return s.repo.Delete(ctx, clinicID, id)
}
