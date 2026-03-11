package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// HospitalizationService 入院/ホテルサービスインターフェース
type HospitalizationService interface {
	GetAllHospitalizations(ctx context.Context) ([]model.Hospitalization, error)
	GetHospitalizationByID(ctx context.Context, id string) (*model.Hospitalization, error)
	GetHospitalizationsByPetID(ctx context.Context, petID string) ([]model.Hospitalization, error)
	GetHospitalizationsByOwnerID(ctx context.Context, ownerID string) ([]model.Hospitalization, error)
	GetHospitalizationsByStatus(ctx context.Context, status string) ([]model.Hospitalization, error)
	CreateHospitalization(ctx context.Context, req *model.CreateHospitalizationRequest) (*model.Hospitalization, error)
	UpdateHospitalization(ctx context.Context, id string, req *model.UpdateHospitalizationRequest) (*model.Hospitalization, error)
	DeleteHospitalization(ctx context.Context, id string) error
}

// GetAllHospitalizations 全ての入院/ホテルを取得
func (s *Service) GetAllHospitalizations(ctx context.Context) ([]model.Hospitalization, error) {
	return s.hospitalizationRepo.GetAllHospitalizations(ctx)
}

// GetHospitalizationByID IDで入院/ホテルを取得
func (s *Service) GetHospitalizationByID(ctx context.Context, id string) (*model.Hospitalization, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid hospitalization ID format")
	}

	hosp, err := s.hospitalizationRepo.GetHospitalizationByID(ctx, uid.String())
	if err != nil {
		return nil, err
	}

	if hosp == nil {
		return nil, apperrors.WrapNotFound("hospitalization", id)
	}

	return hosp, nil
}

// GetHospitalizationsByPetID ペットIDで入院/ホテルを取得
func (s *Service) GetHospitalizationsByPetID(ctx context.Context, petID string) ([]model.Hospitalization, error) {
	uid, err := uuid.Parse(petID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid pet ID format")
	}

	return s.hospitalizationRepo.GetHospitalizationsByPetID(ctx, uid.String())
}

// GetHospitalizationsByOwnerID 飼い主IDで入院/ホテルを取得
func (s *Service) GetHospitalizationsByOwnerID(ctx context.Context, ownerID string) ([]model.Hospitalization, error) {
	uid, err := uuid.Parse(ownerID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid owner ID format")
	}

	return s.hospitalizationRepo.GetHospitalizationsByOwnerID(ctx, uid.String())
}

// GetHospitalizationsByStatus ステータスで入院/ホテルを取得
func (s *Service) GetHospitalizationsByStatus(ctx context.Context, status string) ([]model.Hospitalization, error) {
	return s.hospitalizationRepo.GetHospitalizationsByStatus(ctx, status)
}

// CreateHospitalization 入院/ホテルを作成
func (s *Service) CreateHospitalization(ctx context.Context, req *model.CreateHospitalizationRequest) (*model.Hospitalization, error) {
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid start_date format, expected YYYY-MM-DD")
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid end_date format, expected YYYY-MM-DD")
	}

	hosp := &model.Hospitalization{
		ID:                  uuid.New(),
		OwnerName:           req.OwnerName,
		PetName:             req.PetName,
		Species:             req.Species,
		HospitalizationType: req.HospitalizationType,
		StartDate:           startDate,
		EndDate:             endDate,
		Status:              req.Status,
		DoctorName:          req.DoctorName,
		OwnerRequest:        req.OwnerRequest,
		StaffNotes:          req.StaffNotes,
		Memo:                req.Memo,
	}

	if hosp.Status == "" {
		hosp.Status = "予約"
	}

	if req.OwnerID != nil {
		ownerUID, err := uuid.Parse(*req.OwnerID)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid owner ID format")
		}
		hosp.OwnerID = &ownerUID
	}

	if req.PetID != nil {
		petUID, err := uuid.Parse(*req.PetID)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid pet ID format")
		}
		hosp.PetID = &petUID
	}

	if req.CageID != nil {
		cageUID, err := uuid.Parse(*req.CageID)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid cage ID format")
		}
		hosp.CageID = &cageUID
	}

	if err := s.hospitalizationRepo.CreateHospitalization(ctx, hosp); err != nil {
		return nil, err
	}

	return hosp, nil
}

// UpdateHospitalization 入院/ホテルを更新
func (s *Service) UpdateHospitalization(ctx context.Context, id string, req *model.UpdateHospitalizationRequest) (*model.Hospitalization, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid hospitalization ID format")
	}

	hosp, err := s.hospitalizationRepo.GetHospitalizationByID(ctx, uid.String())
	if err != nil {
		return nil, err
	}

	if hosp == nil {
		return nil, apperrors.WrapNotFound("hospitalization", id)
	}

	// Update fields
	if req.HospitalizationType != nil {
		hosp.HospitalizationType = *req.HospitalizationType
	}
	if req.OwnerName != nil {
		hosp.OwnerName = *req.OwnerName
	}
	if req.PetName != nil {
		hosp.PetName = *req.PetName
	}
	if req.Species != nil {
		hosp.Species = *req.Species
	}
	if req.StartDate != nil {
		t, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid start_date format, expected YYYY-MM-DD")
		}
		hosp.StartDate = t
	}
	if req.EndDate != nil {
		t, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid end_date format, expected YYYY-MM-DD")
		}
		hosp.EndDate = t
	}
	if req.Status != nil {
		hosp.Status = *req.Status
	}
	if req.DoctorName != nil {
		hosp.DoctorName = req.DoctorName
	}
	if req.CageID != nil {
		cageUID, err := uuid.Parse(*req.CageID)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid cage ID format")
		}
		hosp.CageID = &cageUID
	}
	if req.OwnerRequest != nil {
		hosp.OwnerRequest = *req.OwnerRequest
	}
	if req.StaffNotes != nil {
		hosp.StaffNotes = *req.StaffNotes
	}
	if req.Memo != nil {
		hosp.Memo = *req.Memo
	}

	if err := s.hospitalizationRepo.UpdateHospitalization(ctx, hosp); err != nil {
		return nil, err
	}

	return hosp, nil
}

// DeleteHospitalization 入院/ホテルを削除
func (s *Service) DeleteHospitalization(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return apperrors.WrapInvalidInput("invalid hospitalization ID format")
	}

	return s.hospitalizationRepo.DeleteHospitalization(ctx, uid.String())
}
