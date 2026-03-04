package service

import (
	"context"

	"github.com/google/uuid"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// VaccinationService ワクチン接種サービスインターフェース
type VaccinationService interface {
	GetAllVaccinations(ctx context.Context) ([]model.Vaccination, error)
	GetVaccinationByID(ctx context.Context, id string) (*model.Vaccination, error)
	GetVaccinationsByPetID(ctx context.Context, petID string) ([]model.Vaccination, error)
	GetVaccinationsByOwnerID(ctx context.Context, ownerID string) ([]model.Vaccination, error)
	CreateVaccination(ctx context.Context, req *model.CreateVaccinationRequest) (*model.Vaccination, error)
	UpdateVaccination(ctx context.Context, id string, req *model.UpdateVaccinationRequest) (*model.Vaccination, error)
	DeleteVaccination(ctx context.Context, id string) error
}

// GetAllVaccinations 全てのワクチン接種を取得
func (s *Service) GetAllVaccinations(ctx context.Context) ([]model.Vaccination, error) {
	return s.vaccinationRepo.GetAllVaccinations(ctx)
}

// GetVaccinationByID IDでワクチン接種を取得
func (s *Service) GetVaccinationByID(ctx context.Context, id string) (*model.Vaccination, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid vaccination ID format")
	}

	vac, err := s.vaccinationRepo.GetVaccinationByID(ctx, uid.String())
	if err != nil {
		return nil, err
	}

	if vac == nil {
		return nil, apperrors.WrapNotFound("vaccination with id %s not found", id)
	}

	return vac, nil
}

// GetVaccinationsByPetID ペットIDでワクチン接種を取得
func (s *Service) GetVaccinationsByPetID(ctx context.Context, petID string) ([]model.Vaccination, error) {
	uid, err := uuid.Parse(petID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid pet ID format")
	}

	return s.vaccinationRepo.GetVaccinationsByPetID(ctx, uid.String())
}

// GetVaccinationsByOwnerID 飼い主IDでワクチン接種を取得
func (s *Service) GetVaccinationsByOwnerID(ctx context.Context, ownerID string) ([]model.Vaccination, error) {
	uid, err := uuid.Parse(ownerID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid owner ID format")
	}

	return s.vaccinationRepo.GetVaccinationsByOwnerID(ctx, uid.String())
}

// CreateVaccination ワクチン接種を作成
func (s *Service) CreateVaccination(ctx context.Context, req *model.CreateVaccinationRequest) (*model.Vaccination, error) {
	vac := &model.Vaccination{
		ID:              uuid.New(),
		PetID:           req.PetID,
		OwnerID:         req.OwnerID,
		DoctorID:        req.DoctorID,
		VaccineMasterID: req.VaccineMasterID,
		VaccineName:     req.VaccineName,
		VaccinationDate: req.VaccinationDate,
		NextDate:        req.NextDate,
		LotNumber:       req.LotNumber,
		Notes:           req.Notes,
	}

	if err := s.vaccinationRepo.CreateVaccination(ctx, vac); err != nil {
		return nil, err
	}

	return vac, nil
}

// UpdateVaccination ワクチン接種を更新
func (s *Service) UpdateVaccination(ctx context.Context, id string, req *model.UpdateVaccinationRequest) (*model.Vaccination, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid vaccination ID format")
	}

	vac, err := s.vaccinationRepo.GetVaccinationByID(ctx, uid.String())
	if err != nil {
		return nil, err
	}

	if vac == nil {
		return nil, apperrors.WrapNotFound("vaccination with id %s not found", id)
	}

	// Update fields
	if req.VaccineName != "" {
		vac.VaccineName = req.VaccineName
	}
	if req.VaccinationDate != nil {
		vac.VaccinationDate = *req.VaccinationDate
	}
	if req.NextDate != nil {
		vac.NextDate = req.NextDate
	}
	if req.LotNumber != "" {
		vac.LotNumber = req.LotNumber
	}
	if req.Notes != "" {
		vac.Notes = req.Notes
	}

	if err := s.vaccinationRepo.UpdateVaccination(ctx, vac); err != nil {
		return nil, err
	}

	return vac, nil
}

// DeleteVaccination ワクチン接種を削除
func (s *Service) DeleteVaccination(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return apperrors.WrapInvalidInput("invalid vaccination ID format")
	}

	return s.vaccinationRepo.DeleteVaccination(ctx, uid.String())
}
