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
	GetCompany(ctx context.Context) (*model.Company, error)
	UpdateCompany(ctx context.Context, company *model.Company) error
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

func (s *clinicService) GetCompany(ctx context.Context) (*model.Company, error) {
	return s.repo.GetCompany(ctx)
}

func (s *clinicService) UpdateCompany(ctx context.Context, company *model.Company) error {
	return s.repo.UpdateCompany(ctx, company)
}
