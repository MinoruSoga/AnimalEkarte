package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// VaccinationService ワクチン接種サービスインターフェース
type VaccinationService interface {
	GetAllVaccinations(ctx context.Context) ([]model.VaccinationRecord, error)
	GetVaccinationByID(ctx context.Context, id string) (*model.VaccinationRecord, error)
	GetVaccinationsByPetID(ctx context.Context, petID string) ([]model.VaccinationRecord, error)
	GetVaccinationsByOwnerID(ctx context.Context, ownerID string) ([]model.VaccinationRecord, error)
	CreateVaccination(ctx context.Context, req *model.CreateVaccinationRecordRequest) (*model.VaccinationRecord, error)
	UpdateVaccination(ctx context.Context, id string, req *model.UpdateVaccinationRecordRequest) (*model.VaccinationRecord, error)
	DeleteVaccination(ctx context.Context, id string) error
}

// GetAllVaccinations 全てのワクチン接種記録を取得
func (s *Service) GetAllVaccinations(ctx context.Context) ([]model.VaccinationRecord, error) {
	return s.vaccinationRepo.GetAllVaccinations(ctx)
}

// GetVaccinationByID IDでワクチン接種記録を取得
func (s *Service) GetVaccinationByID(ctx context.Context, id string) (*model.VaccinationRecord, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid vaccination ID format")
	}

	vac, err := s.vaccinationRepo.GetVaccinationByID(ctx, uid.String())
	if err != nil {
		return nil, err
	}

	if vac == nil {
		return nil, apperrors.WrapNotFound("vaccination", id)
	}

	return vac, nil
}

// GetVaccinationsByPetID ペットIDでワクチン接種記録を取得
func (s *Service) GetVaccinationsByPetID(ctx context.Context, petID string) ([]model.VaccinationRecord, error) {
	uid, err := uuid.Parse(petID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid pet ID format")
	}

	return s.vaccinationRepo.GetVaccinationsByPetID(ctx, uid.String())
}

// GetVaccinationsByOwnerID 飼い主IDでワクチン接種記録を取得
func (s *Service) GetVaccinationsByOwnerID(ctx context.Context, ownerID string) ([]model.VaccinationRecord, error) {
	uid, err := uuid.Parse(ownerID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid owner ID format")
	}

	return s.vaccinationRepo.GetVaccinationsByOwnerID(ctx, uid.String())
}

// CreateVaccination ワクチン接種記録を作成
func (s *Service) CreateVaccination(ctx context.Context, req *model.CreateVaccinationRecordRequest) (*model.VaccinationRecord, error) {
	medicalRecordID, err := uuid.Parse(req.MedicalRecordID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid medical record ID format")
	}

	petID, err := uuid.Parse(req.PetID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid pet ID format")
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid date format, expected YYYY-MM-DD")
	}

	vac := &model.VaccinationRecord{
		ID:              uuid.New(),
		MedicalRecordID: medicalRecordID,
		PetID:           petID,
		OwnerName:       req.OwnerName,
		PetName:         req.PetName,
		VaccineName:     req.VaccineName,
		Date:            date,
		Doctor:          req.Doctor,
		Supplemental:    req.Supplemental,
		Lot1:            req.Lot1,
		Lot2:            req.Lot2,
		Lot3:            req.Lot3,
		Lot4:            req.Lot4,
		Remarks:         req.Remarks,
	}

	if req.NextDate != nil && *req.NextDate != "" {
		t, err := time.Parse("2006-01-02", *req.NextDate)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid next_date format, expected YYYY-MM-DD")
		}
		vac.NextDate = &t
	}

	if req.NextScheduleType != nil {
		vac.NextScheduleType = req.NextScheduleType
	}

	if err := s.vaccinationRepo.CreateVaccination(ctx, vac); err != nil {
		return nil, err
	}

	return vac, nil
}

// UpdateVaccination ワクチン接種記録を更新
func (s *Service) UpdateVaccination(ctx context.Context, id string, req *model.UpdateVaccinationRecordRequest) (*model.VaccinationRecord, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid vaccination ID format")
	}

	vac, err := s.vaccinationRepo.GetVaccinationByID(ctx, uid.String())
	if err != nil {
		return nil, err
	}

	if vac == nil {
		return nil, apperrors.WrapNotFound("vaccination", id)
	}

	// Update fields
	if req.VaccineName != nil {
		vac.VaccineName = *req.VaccineName
	}
	if req.Date != nil && *req.Date != "" {
		t, err := time.Parse("2006-01-02", *req.Date)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid date format, expected YYYY-MM-DD")
		}
		vac.Date = t
	}
	if req.NextDate != nil && *req.NextDate != "" {
		t, err := time.Parse("2006-01-02", *req.NextDate)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid next_date format, expected YYYY-MM-DD")
		}
		vac.NextDate = &t
	}
	if req.NextScheduleType != nil {
		vac.NextScheduleType = req.NextScheduleType
	}
	if req.Doctor != nil {
		vac.Doctor = *req.Doctor
	}
	if req.Supplemental != nil {
		vac.Supplemental = *req.Supplemental
	}
	if req.Lot1 != nil {
		vac.Lot1 = *req.Lot1
	}
	if req.Lot2 != nil {
		vac.Lot2 = *req.Lot2
	}
	if req.Lot3 != nil {
		vac.Lot3 = *req.Lot3
	}
	if req.Lot4 != nil {
		vac.Lot4 = *req.Lot4
	}
	if req.Remarks != nil {
		vac.Remarks = *req.Remarks
	}

	if err := s.vaccinationRepo.UpdateVaccination(ctx, vac); err != nil {
		return nil, err
	}

	return vac, nil
}

// DeleteVaccination ワクチン接種記録を削除
func (s *Service) DeleteVaccination(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return apperrors.WrapInvalidInput("invalid vaccination ID format")
	}

	return s.vaccinationRepo.DeleteVaccination(ctx, uid.String())
}
