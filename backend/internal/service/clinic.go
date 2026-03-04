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
	CreateClinic(ctx context.Context, req *model.CreateClinicRequest) (*model.Clinic, error)
	UpdateClinic(ctx context.Context, id string, req *model.UpdateClinicRequest) (*model.Clinic, error)
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
		return nil, apperrors.WrapNotFound("clinic with id %s not found", id)
	}

	return clinic, nil
}

// CreateClinic クリニックを作成
func (s *Service) CreateClinic(ctx context.Context, req *model.CreateClinicRequest) (*model.Clinic, error) {
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
func (s *Service) UpdateClinic(ctx context.Context, id string, req *model.UpdateClinicRequest) (*model.Clinic, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid clinic ID format")
	}

	clinic, err := s.clinicRepo.GetClinicByID(ctx, uid.String())
	if err != nil {
		return nil, err
	}

	if clinic == nil {
		return nil, apperrors.WrapNotFound("clinic with id %s not found", id)
	}

	// Update fields
	if req.Name != "" {
		clinic.Name = req.Name
	}
	if req.BranchName != "" {
		clinic.BranchName = req.BranchName
	}
	if req.PostalCode != "" {
		clinic.PostalCode = req.PostalCode
	}
	if req.Address != "" {
		clinic.Address = req.Address
	}
	if req.PhoneNumber != "" {
		clinic.PhoneNumber = req.PhoneNumber
	}
	if req.FaxNumber != "" {
		clinic.FaxNumber = req.FaxNumber
	}
	if req.RegistrationNumber != "" {
		clinic.RegistrationNumber = req.RegistrationNumber
	}
	if req.DirectorName != "" {
		clinic.DirectorName = req.DirectorName
	}
	if req.Email != "" {
		clinic.Email = req.Email
	}
	if req.Website != "" {
		clinic.Website = req.Website
	}
	if req.LogoURL != "" {
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

// StaffService スタッフサービスインターフェース
type StaffService interface {
	GetAllStaff(ctx context.Context) ([]model.Staff, error)
	GetStaffByID(ctx context.Context, id string) (*model.Staff, error)
	GetStaffByClinicID(ctx context.Context, clinicID string) ([]model.Staff, error)
	CreateStaff(ctx context.Context, req *model.CreateStaffRequest) (*model.Staff, error)
	UpdateStaff(ctx context.Context, id string, req *model.UpdateStaffRequest) (*model.Staff, error)
	DeleteStaff(ctx context.Context, id string) error
}

// GetAllStaff 全てのスタッフを取得
func (s *Service) GetAllStaff(ctx context.Context) ([]model.Staff, error) {
	return s.staffRepo.GetAllStaff(ctx)
}

// GetStaffByID IDでスタッフを取得
func (s *Service) GetStaffByID(ctx context.Context, id string) (*model.Staff, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid staff ID format")
	}

	staff, err := s.staffRepo.GetStaffByID(ctx, uid.String())
	if err != nil {
		return nil, err
	}

	if staff == nil {
		return nil, apperrors.WrapNotFound("staff with id %s not found", id)
	}

	return staff, nil
}

// GetStaffByClinicID クリニックIDでスタッフを取得
func (s *Service) GetStaffByClinicID(ctx context.Context, clinicID string) ([]model.Staff, error) {
	uid, err := uuid.Parse(clinicID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid clinic ID format")
	}

	return s.staffRepo.GetStaffByClinicID(ctx, uid.String())
}

// CreateStaff スタッフを作成
func (s *Service) CreateStaff(ctx context.Context, req *model.CreateStaffRequest) (*model.Staff, error) {
	staff := &model.Staff{
		ID:       uuid.New(),
		ClinicID: req.ClinicID,
		Name:     req.Name,
		Role:     req.Role,
		Email:    req.Email,
		Phone:    req.Phone,
		IsActive: req.IsActive,
	}

	if err := s.staffRepo.CreateStaff(ctx, staff); err != nil {
		return nil, err
	}

	return staff, nil
}

// UpdateStaff スタッフを更新
func (s *Service) UpdateStaff(ctx context.Context, id string, req *model.UpdateStaffRequest) (*model.Staff, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid staff ID format")
	}

	staff, err := s.staffRepo.GetStaffByID(ctx, uid.String())
	if err != nil {
		return nil, err
	}

	if staff == nil {
		return nil, apperrors.WrapNotFound("staff with id %s not found", id)
	}

	// Update fields
	if req.Name != "" {
		staff.Name = req.Name
	}
	if req.Role != "" {
		staff.Role = req.Role
	}
	if req.Email != "" {
		staff.Email = req.Email
	}
	if req.Phone != "" {
		staff.Phone = req.Phone
	}
	if req.IsActive != nil {
		staff.IsActive = *req.IsActive
	}

	if err := s.staffRepo.UpdateStaff(ctx, staff); err != nil {
		return nil, err
	}

	return staff, nil
}

// DeleteStaff スタッフを削除
func (s *Service) DeleteStaff(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return apperrors.WrapInvalidInput("invalid staff ID format")
	}

	return s.staffRepo.DeleteStaff(ctx, uid.String())
}
