// Package service provides business logic implementations for Vaccine entity.
package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
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

type VaccineService interface {
	List(ctx context.Context, clinicID uint64, species *string) ([]model.Vaccine, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Vaccine, error)
	Create(ctx context.Context, clinicID uint64, input *CreateVaccineInput) (*model.Vaccine, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateVaccineInput) (*model.Vaccine, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type vaccineService struct{ repo repository.VaccineRepository }

func NewVaccineService(repo repository.VaccineRepository) VaccineService {
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
		return nil, err
	}
	if err := validateNonNegativePrice(input.Price, "金額"); err != nil {
		return nil, err
	}
	if input.Species != nil {
		if err := validateVaccineSpecies(*input.Species); err != nil {
			return nil, err
		}
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
		return nil, apperrors.Wrap(err, "failed to create vaccine")
	}
	slog.InfoContext(ctx, "vaccine created", slog.Uint64("clinic_id", clinicID), slog.Uint64("vaccine_id", vaccine.ID))
	return vaccine, nil
}
func (s *vaccineService) Update(ctx context.Context, clinicID, id uint64, input *UpdateVaccineInput) (*model.Vaccine, error) {
	if err := validateOptionalName(input.Name); err != nil {
		return nil, err
	}
	if err := validateNonNegativePrice(input.Price, "金額"); err != nil {
		return nil, err
	}
	if input.Species != nil {
		if err := validateVaccineSpecies(*input.Species); err != nil {
			return nil, err
		}
	}
	fields := buildVaccineUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
	}
	vaccine, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update vaccine")
	}
	slog.InfoContext(ctx, "vaccine updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("vaccine_id", id))
	return vaccine, nil
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

func buildVaccineUpdateFields(input *UpdateVaccineInput) map[string]any {
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
	if input.ClearParentID {
		fields[colVaccineParentID] = nil
	} else if input.ParentID != nil {
		fields[colVaccineParentID] = *input.ParentID
	}
	if input.SortOrder != nil {
		fields[colVaccineSortOrder] = *input.SortOrder
	}
	return fields
}
func (s *vaccineService) Delete(ctx context.Context, clinicID, id uint64) error {
	// 子ワクチンの存在チェック (BUG-390)
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
		return apperrors.WrapInvalidInput(ErrMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder vaccines")
	}
	slog.InfoContext(ctx, "vaccines reordered", slog.Uint64("clinic_id", clinicID), slog.Int("count", len(ids)))
	return nil
}
