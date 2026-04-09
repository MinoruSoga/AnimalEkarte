// Package service provides business logic implementations for Consultation entity.
package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- ConsultationService ----

type ConsultationService interface {
	List(ctx context.Context, clinicID uint64) ([]model.Consultation, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Consultation, error)
	Create(ctx context.Context, consultation *model.Consultation) error
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
	return s.repo.FindAll(ctx, clinicID)
}
func (s *consultationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Consultation, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}
func (s *consultationService) Create(ctx context.Context, consultation *model.Consultation) error {
	if err := s.repo.Create(ctx, consultation); err != nil {
		return apperrors.Wrap(err, "failed to create consultation")
	}
	slog.InfoContext(ctx, "consultation created", slog.Uint64("consultation_id", consultation.ID))
	return nil
}
func (s *consultationService) Update(ctx context.Context, clinicID, id uint64, input *UpdateConsultationInput) (*model.Consultation, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input must not be nil")
	}
	fields := buildConsultationUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	consultation, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update consultation")
	}
	slog.InfoContext(ctx, "consultation updated", slog.Uint64("consultation_id", id))
	return consultation, nil
}
func (s *consultationService) Delete(ctx context.Context, clinicID, id uint64) error {
	count, err := s.repo.CountUsageByConsultationID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check consultation dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("この診察項目は診療記録で使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete consultation")
	}
	slog.InfoContext(ctx, "consultation deleted", slog.Uint64("consultation_id", id), slog.Uint64("clinic_id", clinicID))
	return nil
}

func (s *consultationService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	return s.repo.Reorder(ctx, clinicID, ids)
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
	TaxType       *model.TaxType
	TaxRate       *float64
}

func buildConsultationUpdateFields(input *UpdateConsultationInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields["name"] = *input.Name
	}
	if input.Price != nil {
		fields["price"] = *input.Price
	}
	if input.IsActive != nil {
		fields["is_active"] = *input.IsActive
	}
	if input.Description != nil {
		fields["description"] = *input.Description
	}
	if input.TimeCondition != nil {
		fields["time_condition"] = *input.TimeCondition
	}
	if input.Duration != nil {
		fields["duration"] = *input.Duration
	}
	if input.ClearParentID {
		fields["parent_id"] = nil
	} else if input.ParentID != nil {
		fields["parent_id"] = *input.ParentID
	}
	if input.SortOrder != nil {
		fields["sort_order"] = *input.SortOrder
	}
	if input.TaxType != nil {
		fields["tax_type"] = *input.TaxType
	}
	if input.TaxRate != nil {
		fields["tax_rate"] = *input.TaxRate
	}
	return fields
}
