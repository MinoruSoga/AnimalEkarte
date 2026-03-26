// Package service provides business logic implementations for Trimming entity.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// CreateTrimmingInput はトリミング記録作成の入力DTO。
type CreateTrimmingInput struct {
	Date           time.Time
	PetID          *uint64
	StaffID        *uint64
	CourseID       *uint64
	Status         model.TrimmingStatus
	StyleRequest   string
	BW             *float64
	BWUnit         model.BodyWeightUnit
	BT             *float64
	UsedShampoo    string
	UsedRibbon     string
	Remarks        string
	StyleImage     string
	CompletedImage string
	OptionIDs      []uint64
}

// UpdateTrimmingInput はトリミング記録部分更新の入力DTO。nil = 未送信フィールド。
// OptionIDs: nil = 変更なし、non-nil（空スライス含む）= 全置換
type UpdateTrimmingInput struct {
	Date           *time.Time
	PetID          *uint64
	StaffID        *uint64
	CourseID       *uint64
	Status         *model.TrimmingStatus
	StyleRequest   *string
	BW             **float64
	BWUnit         *model.BodyWeightUnit
	BT             **float64
	UsedShampoo    *string
	UsedRibbon     *string
	Remarks        *string
	StyleImage     *string
	CompletedImage *string
	OptionIDs      *[]uint64
}

type TrimmingService interface {
	List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.TrimmingRecord, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingRecord, error)
	Create(ctx context.Context, clinicID uint64, input *CreateTrimmingInput) (*model.TrimmingRecord, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingInput) (*model.TrimmingRecord, error)
	Delete(ctx context.Context, clinicID, id uint64) error
}

type trimmingService struct {
	repo repository.TrimmingRepository
}

func NewTrimmingService(repo repository.TrimmingRepository) TrimmingService {
	return &trimmingService{repo: repo}
}

func (s *trimmingService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.TrimmingRecord, int64, error) {
	return s.repo.FindAll(ctx, clinicID, petID, ownerID, startDate, endDate, page, limit)
}

func (s *trimmingService) GetByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingRecord, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *trimmingService) Create(ctx context.Context, clinicID uint64, input *CreateTrimmingInput) (*model.TrimmingRecord, error) {
	status := model.TrimmingStatusReserved
	if input.Status != "" {
		status = input.Status
	}
	bwUnit := model.BodyWeightUnitKg
	if input.BWUnit != "" {
		bwUnit = input.BWUnit
	}
	trimming := &model.TrimmingRecord{
		ClinicID:       clinicID,
		Date:           input.Date,
		PetID:          input.PetID,
		StaffID:        input.StaffID,
		CourseID:       input.CourseID,
		Status:         status,
		StyleRequest:   input.StyleRequest,
		BW:             input.BW,
		BWUnit:         bwUnit,
		BT:             input.BT,
		UsedShampoo:    input.UsedShampoo,
		UsedRibbon:     input.UsedRibbon,
		Remarks:        input.Remarks,
		StyleImage:     input.StyleImage,
		CompletedImage: input.CompletedImage,
	}
	if err := s.repo.Create(ctx, clinicID, trimming); err != nil {
		return nil, fmt.Errorf("failed to create trimming record: %w", err)
	}
	if len(input.OptionIDs) > 0 {
		if err := s.repo.SetOptions(ctx, trimming.ID, input.OptionIDs); err != nil {
			return nil, fmt.Errorf("failed to set trimming options: %w", err)
		}
	}
	return s.repo.FindByID(ctx, clinicID, trimming.ID)
}

func (s *trimmingService) Update(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingInput) (*model.TrimmingRecord, error) {
	existing, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get trimming record: %w", err)
	}
	if input.Date != nil {
		existing.Date = *input.Date
	}
	if input.PetID != nil {
		existing.PetID = input.PetID
	}
	if input.StaffID != nil {
		existing.StaffID = input.StaffID
	}
	if input.CourseID != nil {
		existing.CourseID = input.CourseID
	}
	if input.Status != nil {
		existing.Status = *input.Status
	}
	if input.StyleRequest != nil {
		existing.StyleRequest = *input.StyleRequest
	}
	if input.BW != nil {
		existing.BW = *input.BW
	}
	if input.BWUnit != nil {
		existing.BWUnit = *input.BWUnit
	}
	if input.BT != nil {
		existing.BT = *input.BT
	}
	if input.UsedShampoo != nil {
		existing.UsedShampoo = *input.UsedShampoo
	}
	if input.UsedRibbon != nil {
		existing.UsedRibbon = *input.UsedRibbon
	}
	if input.Remarks != nil {
		existing.Remarks = *input.Remarks
	}
	if input.StyleImage != nil {
		existing.StyleImage = *input.StyleImage
	}
	if input.CompletedImage != nil {
		existing.CompletedImage = *input.CompletedImage
	}
	if err := s.repo.Update(ctx, clinicID, existing); err != nil {
		return nil, fmt.Errorf("failed to update trimming record: %w", err)
	}
	if input.OptionIDs != nil {
		if err := s.repo.SetOptions(ctx, id, *input.OptionIDs); err != nil {
			return nil, fmt.Errorf("failed to set trimming options: %w", err)
		}
	}
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *trimmingService) Delete(ctx context.Context, clinicID, id uint64) error {
	return s.repo.Delete(ctx, clinicID, id)
}
