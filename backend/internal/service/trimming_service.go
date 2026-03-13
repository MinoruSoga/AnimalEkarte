package service

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type TrimmingService interface {
	List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, page, limit int) ([]model.TrimmingRecord, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingRecord, error)
	Create(ctx context.Context, clinicID uint64, trimming *model.TrimmingRecord) error
	Update(ctx context.Context, clinicID uint64, trimming *model.TrimmingRecord) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type trimmingService struct {
	repo repository.TrimmingRepository
}

func NewTrimmingService(repo repository.TrimmingRepository) TrimmingService {
	return &trimmingService{repo: repo}
}

func (s *trimmingService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, page, limit int) ([]model.TrimmingRecord, int64, error) {
	return s.repo.FindAll(ctx, clinicID, petID, ownerID, page, limit)
}

func (s *trimmingService) GetByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingRecord, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *trimmingService) Create(ctx context.Context, clinicID uint64, trimming *model.TrimmingRecord) error {
	return s.repo.Create(ctx, clinicID, trimming)
}

func (s *trimmingService) Update(ctx context.Context, clinicID uint64, trimming *model.TrimmingRecord) error {
	return s.repo.Update(ctx, clinicID, trimming)
}

func (s *trimmingService) Delete(ctx context.Context, clinicID, id uint64) error {
	return s.repo.Delete(ctx, clinicID, id)
}
