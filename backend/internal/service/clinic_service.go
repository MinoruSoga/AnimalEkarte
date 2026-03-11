package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type ClinicService interface {
	ListClinics(ctx context.Context) ([]model.Clinic, error)
	GetClinicByID(ctx context.Context, id uuid.UUID) (*model.Clinic, error)
	GetClinicInfo(ctx context.Context) (*model.ClinicInfo, error)
	UpdateClinicInfo(ctx context.Context, info *model.ClinicInfo) error
}

type clinicService struct {
	repo repository.ClinicRepository
}

func NewClinicService(repo repository.ClinicRepository) ClinicService {
	return &clinicService{repo: repo}
}

func (s *clinicService) ListClinics(ctx context.Context) ([]model.Clinic, error) {
	return s.repo.FindAll(ctx)
}

func (s *clinicService) GetClinicByID(ctx context.Context, id uuid.UUID) (*model.Clinic, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *clinicService) GetClinicInfo(ctx context.Context) (*model.ClinicInfo, error) {
	return s.repo.GetClinicInfo(ctx)
}

func (s *clinicService) UpdateClinicInfo(ctx context.Context, info *model.ClinicInfo) error {
	return s.repo.UpdateClinicInfo(ctx, info)
}
