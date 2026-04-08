// Package service provides business logic implementations for Insurance entity.
package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- InsuranceService ----

type InsuranceService interface {
	List(ctx context.Context, clinicID uint64) ([]model.Insurance, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Insurance, error)
	Create(ctx context.Context, insurance *model.Insurance) error
	Update(ctx context.Context, clinicID, id uint64, input UpdateInsuranceInput) (*model.Insurance, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type insuranceService struct {
	repo repository.InsuranceRepository
}

func NewInsuranceService(repo repository.InsuranceRepository) InsuranceService {
	return &insuranceService{repo: repo}
}

func (s *insuranceService) List(ctx context.Context, clinicID uint64) ([]model.Insurance, error) {
	return s.repo.FindAll(ctx, clinicID)
}
func (s *insuranceService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Insurance, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}
func (s *insuranceService) Create(ctx context.Context, insurance *model.Insurance) error {
	return s.repo.Create(ctx, insurance)
}
func (s *insuranceService) Update(ctx context.Context, clinicID, id uint64, input UpdateInsuranceInput) (*model.Insurance, error) {
	fields := buildInsuranceUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	insurance, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update insurance")
	}
	slog.InfoContext(ctx, "insurance updated", slog.Uint64("insurance_id", id))
	return insurance, nil
}
func (s *insuranceService) Delete(ctx context.Context, clinicID, id uint64) error {
	count, err := s.repo.CountPetsByInsuranceID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check insurance dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("この保険はペット情報で使用中のため削除できません")
	}
	slog.InfoContext(ctx, "insurance deleted", slog.Uint64("insurance_id", id))
	return s.repo.Delete(ctx, clinicID, id)
}

func (s *insuranceService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return s.repo.Reorder(ctx, clinicID, ids)
}

// UpdateInsuranceInput は保険更新のサービス入力 DTO
type UpdateInsuranceInput struct {
	Name         *string
	IsActive     *bool
	Description  *string
	CoverageRate *int
	ContactPhone *string
	SortOrder    *int
}

func buildInsuranceUpdateFields(input UpdateInsuranceInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields["name"] = *input.Name
	}
	if input.IsActive != nil {
		fields["is_active"] = *input.IsActive
	}
	if input.Description != nil {
		fields["description"] = *input.Description
	}
	if input.CoverageRate != nil {
		fields["coverage_rate"] = *input.CoverageRate
	}
	if input.ContactPhone != nil {
		fields["contact_phone"] = *input.ContactPhone
	}
	if input.SortOrder != nil {
		fields["sort_order"] = *input.SortOrder
	}
	return fields
}
