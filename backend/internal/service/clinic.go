package service

import (
	"context"

	"github.com/google/uuid"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ClinicService クリニックサービスインターフェース
type ClinicService interface {
	GetAllClinics(ctx context.Context) ([]model.Clinic, error)
	GetClinicByID(ctx context.Context, id string) (*model.Clinic, error)
	CreateClinic(ctx context.Context, req *model.CreateClinicInfoRequest) (*model.Clinic, error)
	UpdateClinic(ctx context.Context, id string, req *model.UpdateClinicInfoRequest) (*model.Clinic, error)
	DeleteClinic(ctx context.Context, id string) error
}

// GetAllClinics 全てのクリニックを取得
func (s *Service) GetAllClinics(ctx context.Context) ([]model.Clinic, error) {
	return s.clinicRepo.GetAllClinics(ctx)
}

// GetClinicByID IDでクリニックを取得
func (s *Service) GetClinicByID(ctx context.Context, id string) (*model.Clinic, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid clinic ID format")
	}

	clinic, err := s.clinicRepo.GetClinicByID(ctx, uid.String())
	if err != nil {
		return nil, err
	}

	if clinic == nil {
		return nil, apperrors.WrapNotFound("clinic", id)
	}

	return clinic, nil
}

// CreateClinic クリニックを作成
func (s *Service) CreateClinic(ctx context.Context, req *model.CreateClinicInfoRequest) (*model.Clinic, error) {
	clinic := &model.Clinic{
		ID:                 uuid.New(),
		Name:               req.Name,
		BranchName:         req.BranchName,
		PostalCode:         req.PostalCode,
		Address:            req.Address,
		PhoneNumber:        req.PhoneNumber,
		FaxNumber:          req.FaxNumber,
		RegistrationNumber: req.RegistrationNumber,
		DirectorName:       req.DirectorName,
		Email:              req.Email,
		Website:            req.Website,
		LogoURL:            req.LogoURL,
	}

	if err := s.clinicRepo.CreateClinic(ctx, clinic); err != nil {
		return nil, err
	}

	return clinic, nil
}

// UpdateClinic クリニックを更新
func (s *Service) UpdateClinic(ctx context.Context, id string, req *model.UpdateClinicInfoRequest) (*model.Clinic, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid clinic ID format")
	}

	clinic, err := s.clinicRepo.GetClinicByID(ctx, uid.String())
	if err != nil {
		return nil, err
	}

	if clinic == nil {
		return nil, apperrors.WrapNotFound("clinic", id)
	}

	// Update fields
	if req.Name != nil {
		clinic.Name = *req.Name
	}
	if req.BranchName != nil {
		clinic.BranchName = *req.BranchName
	}
	if req.PostalCode != nil {
		clinic.PostalCode = *req.PostalCode
	}
	if req.Address != nil {
		clinic.Address = *req.Address
	}
	if req.PhoneNumber != nil {
		clinic.PhoneNumber = *req.PhoneNumber
	}
	if req.FaxNumber != nil {
		clinic.FaxNumber = *req.FaxNumber
	}
	if req.RegistrationNumber != nil {
		clinic.RegistrationNumber = *req.RegistrationNumber
	}
	if req.DirectorName != nil {
		clinic.DirectorName = *req.DirectorName
	}
	if req.Email != nil {
		clinic.Email = *req.Email
	}
	if req.Website != nil {
		clinic.Website = *req.Website
	}
	if req.LogoURL != nil {
		clinic.LogoURL = req.LogoURL
	}

	if err := s.clinicRepo.UpdateClinic(ctx, clinic); err != nil {
		return nil, err
	}

	return clinic, nil
}

// DeleteClinic クリニックを削除
func (s *Service) DeleteClinic(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return apperrors.WrapInvalidInput("invalid clinic ID format")
	}

	return s.clinicRepo.DeleteClinic(ctx, uid.String())
}
