package medicalrecord

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- VaccineService ----

// CreateVaccineInput はワクチン作成の入力DTO
type CreateVaccineInput struct {
	Name        string
	Price       *int64
	IsActive    bool
	Description string
	Species     *string
	Interval    string
	ParentID    *uint64
	SortOrder   int
}

// UpdateVaccineInput はワクチン更新のサービス入力 DTO
type UpdateVaccineInput struct {
	Name          *string
	Price         *int64
	IsActive      *bool
	Description   *string
	Species       *string
	Interval      *string
	ParentID      *uint64
	ClearParentID bool
	SortOrder     *int
}

const (
	colVaccineName        = "name"
	colVaccinePrice       = "price"
	colVaccineIsActive    = "is_active"
	colVaccineDescription = "description"
	colVaccineSpecies     = "species"
	colVaccineInterval    = "interval"
	colVaccineParentID    = "parent_id"
	colVaccineSortOrder   = "sort_order"
)

func buildVaccineUpdate(input *UpdateVaccineInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colVaccineName] = *input.Name
	}
	if input.Price != nil {
		fields[colVaccinePrice] = *input.Price
	}
	if input.IsActive != nil {
		fields[colVaccineIsActive] = *input.IsActive
	}
	if input.Description != nil {
		fields[colVaccineDescription] = *input.Description
	}
	if input.Species != nil {
		fields[colVaccineSpecies] = model.VaccineSpecies(*input.Species)
	}
	if input.Interval != nil {
		fields[colVaccineInterval] = *input.Interval
	}
	setNullableUint64Field(fields, colVaccineParentID, input.ClearParentID, input.ParentID)
	if input.SortOrder != nil {
		fields[colVaccineSortOrder] = *input.SortOrder
	}
	return fields
}

type VaccineService interface {
	List(ctx context.Context, clinicID uint64, species *string) ([]model.Vaccine, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Vaccine, error)
	Create(ctx context.Context, clinicID uint64, input *CreateVaccineInput) (*model.Vaccine, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateVaccineInput) (*model.Vaccine, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type vaccineService struct{ repo VaccineRepository }

func NewVaccineService(repo VaccineRepository) VaccineService {
	return &vaccineService{repo: repo}
}

func (s *vaccineService) List(ctx context.Context, clinicID uint64, species *string) ([]model.Vaccine, error) {
	items, err := s.repo.FindAll(ctx, clinicID, species)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list vaccines")
	}
	return items, nil
}
func (s *vaccineService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Vaccine, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get vaccine")
	}
	return result, nil
}
func (s *vaccineService) Create(ctx context.Context, clinicID uint64, input *CreateVaccineInput) (*model.Vaccine, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	if err := validateNonNegativePrice(input.Price); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate non negative price")
	}
	if input.Species != nil {
		if err := validateVaccineSpecies(*input.Species); err != nil {
			return nil, apperrors.Wrap(err, "failed to validate vaccine species")
		}
	}
	if err := s.validateParentOwnership(ctx, clinicID, input.ParentID); err != nil {
		return nil, err
	}
	var species *model.VaccineSpecies
	if input.Species != nil {
		s := model.VaccineSpecies(*input.Species)
		species = &s
	}
	vaccine := &model.Vaccine{
		ClinicID:    clinicID,
		Name:        input.Name,
		Price:       input.Price,
		IsActive:    input.IsActive,
		Description: input.Description,
		Species:     species,
		Interval:    input.Interval,
		ParentID:    input.ParentID,
		SortOrder:   input.SortOrder,
	}
	if err := s.repo.Create(ctx, vaccine); err != nil {
		if conflict := apperrors.AsNameUniqueConflict(
			err,
			input.Name,
			apperrors.ConstraintVaccineName,
			apperrors.CodeVaccineNameConflict,
		); conflict != nil {
			return nil, conflict
		}
		return nil, apperrors.Wrap(err, "failed to create vaccine")
	}
	slog.InfoContext(ctx, "vaccine created", slog.Uint64("clinic_id", clinicID), slog.Uint64("vaccine_id", vaccine.ID))
	return vaccine, nil
}
func (s *vaccineService) Update(ctx context.Context, clinicID, id uint64, input *UpdateVaccineInput) (*model.Vaccine, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(errMsgInputNotNil)
	}
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to get vaccine")
	}
	if err := s.validateParentOwnership(ctx, clinicID, input.ParentID); err != nil {
		return nil, err
	}
	if err := validateOptionalName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate optional name")
	}
	if err := validateNonNegativePrice(input.Price); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate non negative price")
	}
	if input.Species != nil {
		if err := validateVaccineSpecies(*input.Species); err != nil {
			return nil, apperrors.Wrap(err, "failed to validate vaccine species")
		}
	}
	fields := buildVaccineUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(errMsgAtLeastOneField)
	}
	vaccine, err := s.repo.Update(ctx, clinicID, id, *input)
	if err != nil {
		nameForConflict := ""
		if input.Name != nil {
			nameForConflict = *input.Name
		}
		if conflict := apperrors.AsNameUniqueConflict(
			err,
			nameForConflict,
			apperrors.ConstraintVaccineName,
			apperrors.CodeVaccineNameConflict,
		); conflict != nil {
			return nil, conflict
		}
		return nil, apperrors.Wrap(err, "failed to update vaccine")
	}
	slog.InfoContext(ctx, "vaccine updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("vaccine_id", id))
	return vaccine, nil
}

func (s *vaccineService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to get vaccine")
	}
	childCount, err := s.repo.CountChildrenByParentID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to count vaccine children")
	}
	if childCount > 0 {
		return apperrors.WrapConflict("このワクチンは子ワクチンが存在するため削除できません")
	}
	count, err := s.repo.CountUsageByVaccineID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check vaccine dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("このワクチンはワクチン接種記録で使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete vaccine")
	}
	slog.InfoContext(ctx, "vaccine deleted", slog.Uint64("clinic_id", clinicID), slog.Uint64("vaccine_id", id))
	return nil
}

func (s *vaccineService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(errMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder vaccines")
	}
	slog.InfoContext(ctx, "vaccines reordered", slog.Uint64("clinic_id", clinicID), slog.Int("count", len(ids)))
	return nil
}

// validateParentOwnership verifies a request-supplied parent_id belongs to the caller's
// clinic before it is persisted (X-14 self-ref master FK guard).
func (s *vaccineService) validateParentOwnership(ctx context.Context, clinicID uint64, parentID *uint64) error {
	return validateOwnedMasterFK(ctx, "parent vaccine", clinicID, parentID,
		func(actx context.Context, cid, mid uint64) error {
			_, err := s.repo.FindByID(actx, cid, mid)
			return err
		})
}
