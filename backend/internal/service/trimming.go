package service

import (
	"context"

	"github.com/google/uuid"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// TrimmingService トリミングサービスインターフェース
type TrimmingService interface {
	GetAllTrimmings(ctx context.Context) ([]model.Trimming, error)
	GetTrimmingByID(ctx context.Context, id string) (*model.Trimming, error)
	GetTrimmingsByPetID(ctx context.Context, petID string) ([]model.Trimming, error)
	GetTrimmingsByOwnerID(ctx context.Context, ownerID string) ([]model.Trimming, error)
	GetTrimmingsByStatus(ctx context.Context, status string) ([]model.Trimming, error)
	CreateTrimming(ctx context.Context, req *model.CreateTrimmingRequest) (*model.Trimming, error)
	UpdateTrimming(ctx context.Context, id string, req *model.UpdateTrimmingRequest) (*model.Trimming, error)
	DeleteTrimming(ctx context.Context, id string) error
}

// GetAllTrimmings 全てのトリミングを取得
func (s *Service) GetAllTrimmings(ctx context.Context) ([]model.Trimming, error) {
	return s.trimmingRepo.GetAllTrimmings(ctx)
}

// GetTrimmingByID IDでトリミングを取得
func (s *Service) GetTrimmingByID(ctx context.Context, id string) (*model.Trimming, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid trimming ID format")
	}

	trim, err := s.trimmingRepo.GetTrimmingByID(ctx, uid.String())
	if err != nil {
		return nil, err
	}

	if trim == nil {
		return nil, apperrors.WrapNotFound("trimming with id %s not found", id)
	}

	return trim, nil
}

// GetTrimmingsByPetID ペットIDでトリミングを取得
func (s *Service) GetTrimmingsByPetID(ctx context.Context, petID string) ([]model.Trimming, error) {
	uid, err := uuid.Parse(petID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid pet ID format")
	}

	return s.trimmingRepo.GetTrimmingsByPetID(ctx, uid.String())
}

// GetTrimmingsByOwnerID 飼い主IDでトリミングを取得
func (s *Service) GetTrimmingsByOwnerID(ctx context.Context, ownerID string) ([]model.Trimming, error) {
	uid, err := uuid.Parse(ownerID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid owner ID format")
	}

	return s.trimmingRepo.GetTrimmingsByOwnerID(ctx, uid.String())
}

// GetTrimmingsByStatus ステータスでトリミングを取得
func (s *Service) GetTrimmingsByStatus(ctx context.Context, status string) ([]model.Trimming, error) {
	return s.trimmingRepo.GetTrimmingsByStatus(ctx, status)
}

// CreateTrimming トリミングを作成
func (s *Service) CreateTrimming(ctx context.Context, req *model.CreateTrimmingRequest) (*model.Trimming, error) {
	trim := &model.Trimming{
		ID:              uuid.New(),
		PetID:           req.PetID,
		OwnerID:         req.OwnerID,
		StaffID:         req.StaffID,
		AppointmentDate: req.AppointmentDate,
		Course:          req.Course,
		Options:         req.Options,
		StyleRequest:    req.StyleRequest,
		Status:          "予約",
		Notes:           req.Notes,
	}

	if err := s.trimmingRepo.CreateTrimming(ctx, trim); err != nil {
		return nil, err
	}

	return trim, nil
}

// UpdateTrimming トリミングを更新
func (s *Service) UpdateTrimming(ctx context.Context, id string, req *model.UpdateTrimmingRequest) (*model.Trimming, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid trimming ID format")
	}

	trim, err := s.trimmingRepo.GetTrimmingByID(ctx, uid.String())
	if err != nil {
		return nil, err
	}

	if trim == nil {
		return nil, apperrors.WrapNotFound("trimming with id %s not found", id)
	}

	// Update fields
	if req.Status != "" {
		trim.Status = req.Status
	}
	if req.AppointmentDate != nil {
		trim.AppointmentDate = *req.AppointmentDate
	}
	if req.Course != "" {
		trim.Course = req.Course
	}
	if req.Options != "" {
		trim.Options = req.Options
	}
	if req.StyleRequest != "" {
		trim.StyleRequest = req.StyleRequest
	}
	if req.TotalPrice != nil {
		trim.TotalPrice = req.TotalPrice
	}
	if req.Notes != "" {
		trim.Notes = req.Notes
	}

	if err := s.trimmingRepo.UpdateTrimming(ctx, trim); err != nil {
		return nil, err
	}

	return trim, nil
}

// DeleteTrimming トリミングを削除
func (s *Service) DeleteTrimming(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return apperrors.WrapInvalidInput("invalid trimming ID format")
	}

	return s.trimmingRepo.DeleteTrimming(ctx, uid.String())
}
