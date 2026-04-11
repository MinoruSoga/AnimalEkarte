// Package service provides business logic implementations for Trimming entity.
package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// CreateTrimmingInput はトリミング記録作成の入力DTO。
type CreateTrimmingInput struct {
	Date            time.Time
	PetID           *uint64
	StaffID         *uint64
	CourseID        *uint64
	Status          model.TrimmingStatus
	StyleRequest    string
	BodyWeight      *float64
	BWUnit          model.BodyWeightUnit
	BodyTemperature *float64
	UsedShampoo     string
	UsedRibbon      string
	Remarks         string
	StyleImage      string
	CompletedImage  string
	OptionIDs       []uint64
}

// UpdateTrimmingInput はトリミング記録部分更新の入力DTO。nil = 未送信フィールド。
// OptionIDs: nil = 変更なし、non-nil（空スライス含む）= 全置換
type UpdateTrimmingInput struct {
	Date            *time.Time
	PetID           *uint64
	StaffID         *uint64
	CourseID        *uint64
	Status          *model.TrimmingStatus
	StyleRequest    *string
	BodyWeight      **float64
	BWUnit          *model.BodyWeightUnit
	BodyTemperature **float64
	UsedShampoo     *string
	UsedRibbon      *string
	Remarks         *string
	StyleImage      *string
	CompletedImage  *string
	OptionIDs       *[]uint64
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
	items, total, err := s.repo.FindAll(ctx, clinicID, petID, ownerID, startDate, endDate, page, limit)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list trimming records")
	}
	return items, total, nil
}

func (s *trimmingService) GetByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingRecord, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get trimming record")
	}
	return result, nil
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
		ClinicID:        clinicID,
		Date:            input.Date,
		PetID:           input.PetID,
		StaffID:         input.StaffID,
		CourseID:        input.CourseID,
		Status:          status,
		StyleRequest:    input.StyleRequest,
		BodyWeight:      input.BodyWeight,
		BWUnit:          bwUnit,
		BodyTemperature: input.BodyTemperature,
		UsedShampoo:     input.UsedShampoo,
		UsedRibbon:      input.UsedRibbon,
		Remarks:         input.Remarks,
		StyleImage:      input.StyleImage,
		CompletedImage:  input.CompletedImage,
	}
	if err := s.repo.Create(ctx, clinicID, trimming); err != nil {
		return nil, apperrors.Wrap(err, "failed to create trimming record")
	}
	if len(input.OptionIDs) > 0 {
		if err := s.repo.SetOptions(ctx, trimming.ID, input.OptionIDs); err != nil {
			return nil, apperrors.Wrap(err, "failed to set trimming options")
		}
	}
	slog.InfoContext(ctx, "trimming record created",
		slog.Uint64("trimming_id", trimming.ID),
		slog.Uint64("clinic_id", clinicID))
	result, err := s.repo.FindByID(ctx, clinicID, trimming.ID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get trimming record after create")
	}
	return result, nil
}

func (s *trimmingService) Update(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingInput) (*model.TrimmingRecord, error) {
	existing, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get trimming record")
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
	if input.BodyWeight != nil {
		existing.BodyWeight = *input.BodyWeight
	}
	if input.BWUnit != nil {
		existing.BWUnit = *input.BWUnit
	}
	if input.BodyTemperature != nil {
		existing.BodyTemperature = *input.BodyTemperature
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
		return nil, apperrors.Wrap(err, "failed to update trimming record")
	}
	if input.OptionIDs != nil {
		if err := s.repo.SetOptions(ctx, id, *input.OptionIDs); err != nil {
			return nil, apperrors.Wrap(err, "failed to set trimming options")
		}
	}
	slog.InfoContext(ctx, "trimming record updated",
		slog.Uint64("trimming_id", id),
		slog.Uint64("clinic_id", clinicID))
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get trimming record after update")
	}
	return result, nil
}

func (s *trimmingService) Delete(ctx context.Context, clinicID, id uint64) error {
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete trimming record")
	}
	slog.InfoContext(ctx, "trimming record deleted",
		slog.Uint64("trimming_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}
