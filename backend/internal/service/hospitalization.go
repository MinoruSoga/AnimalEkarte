package service

import (
	"context"

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
		return nil, apperrors.WrapNotFound("hospitalization with id %s not found", id)
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
	hosp := &model.Hospitalization{
		ID:          uuid.New(),
		PetID:       req.PetID,
		OwnerID:     req.OwnerID,
		CageID:      req.CageID,
		Type:        req.Type,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Status:      "予約",
		OwnerRequest: req.OwnerRequest,
		StaffNotes:  req.StaffNotes,
		Memo:        req.Memo,
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
		return nil, apperrors.WrapNotFound("hospitalization with id %s not found", id)
	}

	// Update fields
	if req.Type != "" {
		hosp.Type = req.Type
	}
	if req.CageID != nil {
		hosp.CageID = req.CageID
	}
	if req.EndDate != nil {
		hosp.EndDate = *req.EndDate
	}
	if req.Status != "" {
		hosp.Status = req.Status
	}
	if req.OwnerRequest != "" {
		hosp.OwnerRequest = req.OwnerRequest
	}
	if req.StaffNotes != "" {
		hosp.StaffNotes = req.StaffNotes
	}
	if req.Memo != "" {
		hosp.Memo = req.Memo
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
