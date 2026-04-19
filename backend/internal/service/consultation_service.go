// Package service provides business logic implementations for Consultation entity.
package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

const (
	colConsultationName          = "name"
	colConsultationPrice         = "price"
	colConsultationIsActive      = "is_active"
	colConsultationDescription   = "description"
	colConsultationTimeCondition = "time_condition"
	colConsultationDuration      = "duration"
	colConsultationParentID      = "parent_id"
	colConsultationSortOrder     = "sort_order"
	colConsultationTaxType       = "tax_type"
	colConsultationTaxRate       = "tax_rate"
)

// ---- ConsultationService ----

// CreateConsultationInput は診察種別作成のサービス入力 DTO
type CreateConsultationInput struct {
	Name          string
	Price         *int64
	IsActive      bool
	Description   string
	TimeCondition string
	Duration      *int
	ParentID      *uint64
	SortOrder     int
	TaxType       *string  // nil = "excluded" (default)
	TaxRate       *float64 // nil = 0.10 (default)
}

type ConsultationService interface {
	List(ctx context.Context, clinicID uint64) ([]model.Consultation, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Consultation, error)
	Create(ctx context.Context, clinicID uint64, input *CreateConsultationInput) (*model.Consultation, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateConsultationInput) (*model.Consultation, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type consultationService struct {
	repo repository.ConsultationRepository
}

func NewConsultationService(repo repository.ConsultationRepository) ConsultationService {
	return &consultationService{repo: repo}
}

func (s *consultationService) List(ctx context.Context, clinicID uint64) ([]model.Consultation, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list consultations")
	}
	return items, nil
}
func (s *consultationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Consultation, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get consultation")
	}
	return result, nil
}
func (s *consultationService) Create(ctx context.Context, clinicID uint64, input *CreateConsultationInput) (*model.Consultation, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, err
	}
	taxType := model.TaxTypeExcluded
	if input.TaxType != nil && *input.TaxType != "" {
		taxType = model.TaxType(*input.TaxType)
	}
	taxRate := 0.10
	if input.TaxRate != nil {
		taxRate = *input.TaxRate
	}
	consultation := &model.Consultation{
		ClinicID:      clinicID,
		Name:          input.Name,
		Price:         input.Price,
		IsActive:      input.IsActive,
		Description:   input.Description,
		TimeCondition: input.TimeCondition,
		Duration:      input.Duration,
		ParentID:      input.ParentID,
		SortOrder:     input.SortOrder,
		TaxType:       taxType,
		TaxRate:       taxRate,
	}
	if err := s.repo.Create(ctx, consultation); err != nil {
		return nil, apperrors.Wrap(err, "failed to create consultation")
	}
	slog.InfoContext(ctx, "consultation created", slog.Uint64("clinic_id", clinicID), slog.Uint64("consultation_id", consultation.ID))
	return consultation, nil
}
func (s *consultationService) Update(ctx context.Context, clinicID, id uint64, input *UpdateConsultationInput) (*model.Consultation, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
	}
	if err := validateOptionalName(input.Name); err != nil {
		return nil, err
	}
	fields := buildConsultationUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
	}
	consultation, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update consultation")
	}
	slog.InfoContext(ctx, "consultation updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("consultation_id", id))
	return consultation, nil
}
func (s *consultationService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to get consultation")
	}
	count, err := s.repo.CountUsageByConsultationID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check consultation dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("この診察項目は診療記録で使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete consultation")
	}
	slog.InfoContext(ctx, "consultation deleted", slog.Uint64("clinic_id", clinicID), slog.Uint64("consultation_id", id))
	return nil
}

func (s *consultationService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(ErrMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder consultations")
	}
	slog.InfoContext(ctx, "consultations reordered",
		slog.Uint64("clinic_id", clinicID),
		slog.Int("count", len(ids)))
	return nil
}

// UpdateConsultationInput は診察料金更新のサービス入力 DTO
type UpdateConsultationInput struct {
	Name          *string
	Price         *int64
	IsActive      *bool
	Description   *string
	TimeCondition *string
	Duration      *int
	ParentID      *uint64
	ClearParentID bool
	SortOrder     *int
	TaxType       *string
	TaxRate       *float64
}

func buildConsultationUpdateFields(input *UpdateConsultationInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colConsultationName] = *input.Name
	}
	if input.Price != nil {
		fields[colConsultationPrice] = *input.Price
	}
	if input.IsActive != nil {
		fields[colConsultationIsActive] = *input.IsActive
	}
	if input.Description != nil {
		fields[colConsultationDescription] = *input.Description
	}
	if input.TimeCondition != nil {
		fields[colConsultationTimeCondition] = *input.TimeCondition
	}
	if input.Duration != nil {
		fields[colConsultationDuration] = *input.Duration
	}
	if input.ClearParentID {
		fields[colConsultationParentID] = nil
	} else if input.ParentID != nil {
		fields[colConsultationParentID] = *input.ParentID
	}
	if input.SortOrder != nil {
		fields[colConsultationSortOrder] = *input.SortOrder
	}
	if input.TaxType != nil {
		fields[colConsultationTaxType] = model.TaxType(*input.TaxType)
	}
	if input.TaxRate != nil {
		fields[colConsultationTaxRate] = *input.TaxRate
	}
	return fields
}
