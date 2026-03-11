package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// TrimmingService トリミングサービスインターフェース
type TrimmingService interface {
	GetAllTrimmings(ctx context.Context) ([]model.TrimmingRecord, error)
	GetTrimmingByID(ctx context.Context, id string) (*model.TrimmingRecord, error)
	GetTrimmingsByPetID(ctx context.Context, petID string) ([]model.TrimmingRecord, error)
	GetTrimmingsByOwnerID(ctx context.Context, ownerID string) ([]model.TrimmingRecord, error)
	GetTrimmingsByStatus(ctx context.Context, status string) ([]model.TrimmingRecord, error)
	CreateTrimming(ctx context.Context, req *model.CreateTrimmingRequest) (*model.TrimmingRecord, error)
	UpdateTrimming(ctx context.Context, id string, req *model.UpdateTrimmingRequest) (*model.TrimmingRecord, error)
	DeleteTrimming(ctx context.Context, id string) error
}

// GetAllTrimmings 全てのトリミングを取得
func (s *Service) GetAllTrimmings(ctx context.Context) ([]model.TrimmingRecord, error) {
	return s.trimmingRepo.GetAllTrimmings(ctx)
}

// GetTrimmingByID IDでトリミングを取得
func (s *Service) GetTrimmingByID(ctx context.Context, id string) (*model.TrimmingRecord, error) {
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
func (s *Service) GetTrimmingsByPetID(ctx context.Context, petID string) ([]model.TrimmingRecord, error) {
	uid, err := uuid.Parse(petID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid pet ID format")
	}

	return s.trimmingRepo.GetTrimmingsByPetID(ctx, uid.String())
}

// GetTrimmingsByOwnerID 飼い主IDでトリミングを取得
func (s *Service) GetTrimmingsByOwnerID(ctx context.Context, ownerID string) ([]model.TrimmingRecord, error) {
	uid, err := uuid.Parse(ownerID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid owner ID format")
	}

	return s.trimmingRepo.GetTrimmingsByOwnerID(ctx, uid.String())
}

// GetTrimmingsByStatus ステータスでトリミングを取得
func (s *Service) GetTrimmingsByStatus(ctx context.Context, status string) ([]model.TrimmingRecord, error) {
	return s.trimmingRepo.GetTrimmingsByStatus(ctx, status)
}

// CreateTrimming トリミングを作成
func (s *Service) CreateTrimming(ctx context.Context, req *model.CreateTrimmingRequest) (*model.TrimmingRecord, error) {
	petID, err := uuid.Parse(req.PetID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid pet ID format")
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid date format, expected YYYY-MM-DD")
	}

	trim := &model.TrimmingRecord{
		ID:           uuid.New(),
		PetID:        petID,
		Date:         date,
		PetNumber:    req.PetNumber,
		PetName:      req.PetName,
		OwnerName:    req.OwnerName,
		Species:      req.Species,
		Weight:       req.Weight,
		StyleRequest: req.StyleRequest,
		Staff:        req.Staff,
		Status:       req.Status,
		Bw:           req.Bw,
		BwUnit:       req.BwUnit,
		Bt:           req.Bt,
		UsedShampoo:  req.UsedShampoo,
		UsedRibbon:   req.UsedRibbon,
		Treatment:    req.Treatment,
		Medicine:     req.Medicine,
		Charge:       req.Charge,
		Remarks:      req.Remarks,
		FinalCheck:   req.FinalCheck,
	}

	if trim.Status == "" {
		trim.Status = "予約"
	}

	if req.CourseID != nil {
		courseUID, err := uuid.Parse(*req.CourseID)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid course ID format")
		}
		trim.CourseID = &courseUID
	}

	if err := s.trimmingRepo.CreateTrimming(ctx, trim); err != nil {
		return nil, err
	}

	return trim, nil
}

// UpdateTrimming トリミングを更新
func (s *Service) UpdateTrimming(ctx context.Context, id string, req *model.UpdateTrimmingRequest) (*model.TrimmingRecord, error) {
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
	if req.Date != nil {
		t, err := time.Parse("2006-01-02", *req.Date)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid date format, expected YYYY-MM-DD")
		}
		trim.Date = t
	}
	if req.PetName != nil {
		trim.PetName = *req.PetName
	}
	if req.OwnerName != nil {
		trim.OwnerName = *req.OwnerName
	}
	if req.Species != nil {
		trim.Species = *req.Species
	}
	if req.Weight != nil {
		trim.Weight = *req.Weight
	}
	if req.StyleRequest != nil {
		trim.StyleRequest = *req.StyleRequest
	}
	if req.Staff != nil {
		trim.Staff = *req.Staff
	}
	if req.Status != nil {
		trim.Status = *req.Status
	}
	if req.CourseID != nil {
		courseUID, err := uuid.Parse(*req.CourseID)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid course ID format")
		}
		trim.CourseID = &courseUID
	}
	if req.Bw != nil {
		trim.Bw = *req.Bw
	}
	if req.BwUnit != nil {
		trim.BwUnit = *req.BwUnit
	}
	if req.Bt != nil {
		trim.Bt = *req.Bt
	}
	if req.UsedShampoo != nil {
		trim.UsedShampoo = *req.UsedShampoo
	}
	if req.UsedRibbon != nil {
		trim.UsedRibbon = *req.UsedRibbon
	}
	if req.Treatment != nil {
		trim.Treatment = *req.Treatment
	}
	if req.Medicine != nil {
		trim.Medicine = *req.Medicine
	}
	if req.Charge != nil {
		trim.Charge = *req.Charge
	}
	if req.Remarks != nil {
		trim.Remarks = *req.Remarks
	}
	if req.FinalCheck != nil {
		trim.FinalCheck = *req.FinalCheck
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
