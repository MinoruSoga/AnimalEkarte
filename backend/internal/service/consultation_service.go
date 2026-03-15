// Package service provides business logic implementations for Consultation entity.
package service

import (
	"context"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- ConsultationService ----

type ConsultationService interface {
	List(ctx context.Context) ([]model.Consultation, error)
	GetByID(ctx context.Context, id uint64) (*model.Consultation, error)
	Create(ctx context.Context, consultation *model.Consultation) error
	Update(ctx context.Context, consultation *model.Consultation) error
	Delete(ctx context.Context, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type consultationService struct {
	repo repository.ConsultationRepository
}

func NewConsultationService(repo repository.ConsultationRepository) ConsultationService {
	return &consultationService{repo: repo}
}

func (s *consultationService) List(ctx context.Context) ([]model.Consultation, error) {
	return s.repo.FindAll(ctx)
}
func (s *consultationService) GetByID(ctx context.Context, id uint64) (*model.Consultation, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *consultationService) Create(ctx context.Context, consultation *model.Consultation) error {
	return s.repo.Create(ctx, consultation)
}
func (s *consultationService) Update(ctx context.Context, consultation *model.Consultation) error {
	return s.repo.Update(ctx, consultation)
}
func (s *consultationService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *consultationService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	return s.repo.Reorder(ctx, clinicID, ids)
}
