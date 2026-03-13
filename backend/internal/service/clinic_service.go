package service

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type ClinicService interface {
	ListClinics(ctx context.Context) ([]model.Clinic, error)
	GetClinicByID(ctx context.Context, id uint64) (*model.Clinic, error)
	CreateClinic(ctx context.Context, clinic *model.Clinic) (*model.Clinic, error)
	UpdateClinic(ctx context.Context, id uint64, clinic *model.Clinic) (*model.Clinic, error)
	DeleteClinic(ctx context.Context, id uint64) error
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

func (s *clinicService) GetClinicByID(ctx context.Context, id uint64) (*model.Clinic, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *clinicService) CreateClinic(ctx context.Context, clinic *model.Clinic) (*model.Clinic, error) {
	// company はシングルトンなので自動設定する
	company, err := s.repo.GetCompany(ctx)
	if err != nil {
		return nil, err
	}
	clinic.CompanyID = company.ID
	if err := s.repo.Create(ctx, clinic); err != nil {
		return nil, err
	}
	return clinic, nil
}

func (s *clinicService) UpdateClinic(ctx context.Context, id uint64, input *model.Clinic) (*model.Clinic, error) {
	clinic, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// immutable フィールドを既存レコードから引き継ぐ
	input.ID = clinic.ID
	input.CompanyID = clinic.CompanyID
	if err := s.repo.Update(ctx, input); err != nil {
		return nil, err
	}
	return input, nil
}

func (s *clinicService) DeleteClinic(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *clinicService) GetCompany(ctx context.Context) (*model.Company, error) {
	return s.repo.GetCompany(ctx)
}

func (s *clinicService) UpdateCompany(ctx context.Context, company *model.Company) error {
	return s.repo.UpdateCompany(ctx, company)
}
