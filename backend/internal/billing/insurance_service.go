// Package billing provides insurance master use cases.
package billing

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

const (
	colInsuranceName         = "name"
	colInsuranceIsActive     = "is_active"
	colInsuranceDescription  = "description"
	colInsuranceCoverageRate = "coverage_rate"
	colInsuranceContactPhone = "contact_phone"
	colInsuranceSortOrder    = "sort_order"
)

func buildInsuranceUpdate(input *UpdateInsuranceInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colInsuranceName] = *input.Name
	}
	if input.IsActive != nil {
		fields[colInsuranceIsActive] = *input.IsActive
	}
	if input.Description != nil {
		fields[colInsuranceDescription] = *input.Description
	}
	if input.CoverageRate != nil {
		fields[colInsuranceCoverageRate] = *input.CoverageRate
	}
	if input.ContactPhone != nil {
		fields[colInsuranceContactPhone] = *input.ContactPhone
	}
	if input.SortOrder != nil {
		fields[colInsuranceSortOrder] = *input.SortOrder
	}
	return fields
}

// ---- InsuranceService ----

// CreateInsuranceInput は保険作成の入力DTO
type CreateInsuranceInput struct {
	Name         string
	IsActive     bool
	Description  string
	CoverageRate *int // nil = 0 (default), ハンドラから nil をそのまま受け取る (BUG-379)
	ContactPhone string
	SortOrder    int
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

type InsuranceService interface {
	List(ctx context.Context, clinicID uint64) ([]model.Insurance, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Insurance, error)
	Create(ctx context.Context, clinicID uint64, input *CreateInsuranceInput) (*model.Insurance, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateInsuranceInput) (*model.Insurance, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type insuranceService struct {
	repo InsuranceRepository
}

func NewInsuranceService(repo InsuranceRepository) InsuranceService {
	return &insuranceService{repo: repo}
}

func (s *insuranceService) List(ctx context.Context, clinicID uint64) ([]model.Insurance, error) {
	result, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list insurance")
	}
	return result, nil
}
func (s *insuranceService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Insurance, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get insurance")
	}
	return result, nil
}
func (s *insuranceService) Create(ctx context.Context, clinicID uint64, input *CreateInsuranceInput) (*model.Insurance, error) {
	if err := sharedkernel.ValidateRequiredName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	// CoverageRate デフォルト値はサービス層で設定する (BUG-379)
	coverageRate := 0
	if input.CoverageRate != nil {
		coverageRate = *input.CoverageRate
	}
	if err := ValidateCoverageRate(coverageRate); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate coverage rate")
	}
	insurance := &model.Insurance{
		ClinicID:     clinicID,
		Name:         input.Name,
		IsActive:     input.IsActive,
		Description:  input.Description,
		CoverageRate: coverageRate,
		ContactPhone: input.ContactPhone,
		SortOrder:    input.SortOrder,
	}
	if err := s.repo.Create(ctx, insurance); err != nil {
		return nil, apperrors.Wrap(err, "failed to create insurance")
	}
	slog.InfoContext(ctx, "insurance created", slog.Uint64("clinic_id", clinicID), slog.Uint64("insurance_id", insurance.ID))
	return insurance, nil
}
func (s *insuranceService) Update(ctx context.Context, clinicID, id uint64, input *UpdateInsuranceInput) (*model.Insurance, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgInputNotNil)
	}
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to get insurance")
	}
	if err := sharedkernel.ValidateOptionalName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate optional name")
	}
	if err := ValidateOptionalCoverageRate(input.CoverageRate); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate optional coverage rate")
	}
	fields := buildInsuranceUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgAtLeastOneField)
	}
	insurance, err := s.repo.Update(ctx, clinicID, id, *input)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update insurance")
	}
	slog.InfoContext(ctx, "insurance updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("insurance_id", id))
	return insurance, nil
}
func (s *insuranceService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to get insurance")
	}
	count, err := s.repo.CountUsageByInsuranceID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check insurance dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("この保険はペット情報で使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete insurance")
	}
	slog.InfoContext(ctx, "insurance deleted", slog.Uint64("clinic_id", clinicID), slog.Uint64("insurance_id", id))
	return nil
}

func (s *insuranceService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(sharedkernel.ErrMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder insurances")
	}
	slog.InfoContext(ctx, "insurances reordered",
		slog.Uint64("clinic_id", clinicID),
		slog.Int("count", len(ids)))
	return nil
}
