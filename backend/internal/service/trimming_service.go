package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type TrimmingService interface {
	List(ctx context.Context, clinicID uuid.UUID, petID *uuid.UUID, ownerID *uuid.UUID, page, limit int) ([]model.TrimmingRecord, int64, error)
	GetByID(ctx context.Context, clinicID, id uuid.UUID) (*model.TrimmingRecord, error)
	Create(ctx context.Context, clinicID uuid.UUID, trimming *model.TrimmingRecord) error
	Update(ctx context.Context, clinicID uuid.UUID, trimming *model.TrimmingRecord) error
	Delete(ctx context.Context, clinicID, id uuid.UUID) error
}

type trimmingService struct {
	repo repository.TrimmingRepository
}

func NewTrimmingService(repo repository.TrimmingRepository) TrimmingService {
	return &trimmingService{repo: repo}
}

func (s *trimmingService) List(ctx context.Context, clinicID uuid.UUID, petID *uuid.UUID, ownerID *uuid.UUID, page, limit int) ([]model.TrimmingRecord, int64, error) {
	return s.repo.FindAll(ctx, clinicID, petID, ownerID, page, limit)
}

func (s *trimmingService) GetByID(ctx context.Context, clinicID, id uuid.UUID) (*model.TrimmingRecord, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *trimmingService) Create(ctx context.Context, clinicID uuid.UUID, trimming *model.TrimmingRecord) error {
	return s.repo.Create(ctx, clinicID, trimming)
}

func (s *trimmingService) Update(ctx context.Context, clinicID uuid.UUID, trimming *model.TrimmingRecord) error {
	return s.repo.Update(ctx, clinicID, trimming)
}

func (s *trimmingService) Delete(ctx context.Context, clinicID, id uuid.UUID) error {
	return s.repo.Delete(ctx, clinicID, id)
}
