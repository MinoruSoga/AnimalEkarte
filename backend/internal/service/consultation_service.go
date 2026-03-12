// Package service provides business logic implementations for Consultation entity.
package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- ConsultationService ----

type ConsultationService interface {
	List(ctx context.Context) ([]model.Consultation, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Consultation, error)
	Create(ctx context.Context, consultation *model.Consultation) error
	Update(ctx context.Context, consultation *model.Consultation) error
	Delete(ctx context.Context, id uuid.UUID) error
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
func (s *consultationService) GetByID(ctx context.Context, id uuid.UUID) (*model.Consultation, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *consultationService) Create(ctx context.Context, consultation *model.Consultation) error {
	return s.repo.Create(ctx, consultation)
}
func (s *consultationService) Update(ctx context.Context, consultation *model.Consultation) error {
	return s.repo.Update(ctx, consultation)
}
func (s *consultationService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
