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

type VaccineService interface {
	List(ctx context.Context, clinicID uint64, species *string) ([]model.Vaccine, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Vaccine, error)
	Create(ctx context.Context, vaccine *model.Vaccine) error
	Update(ctx context.Context, clinicID, id uint64, input UpdateVaccineInput) (*model.Vaccine, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type vaccineService struct{ repo repository.VaccineRepository }

func NewVaccineService(repo repository.VaccineRepository) VaccineService {
	return &vaccineService{repo: repo}
}

func (s *vaccineService) List(ctx context.Context, clinicID uint64, species *string) ([]model.Vaccine, error) {
	return s.repo.FindAll(ctx, clinicID, species)
}
func (s *vaccineService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Vaccine, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}
func (s *vaccineService) Create(ctx context.Context, vaccine *model.Vaccine) error {
	return s.repo.Create(ctx, vaccine)
}
func (s *vaccineService) Update(ctx context.Context, clinicID, id uint64, input UpdateVaccineInput) (*model.Vaccine, error) {
	fields := buildVaccineUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	vaccine, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update vaccine")
	}
	slog.InfoContext(ctx, "vaccine updated", slog.Uint64("vaccine_id", id))
	return vaccine, nil
}

// UpdateVaccineInput はワクチン更新のサービス入力 DTO
type UpdateVaccineInput struct {
	Name          *string
	Price         *int64
	IsActive      *bool
	Description   *string
	Species       *model.VaccineSpecies
	Interval      *string
	ParentID      *uint64
	ClearParentID bool
	SortOrder     *int
}

func buildVaccineUpdateFields(input UpdateVaccineInput) map[string]any {
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
	if input.Species != nil {
		fields["species"] = *input.Species
	}
	if input.Interval != nil {
		fields["interval"] = *input.Interval
	}
	if input.ClearParentID {
		fields["parent_id"] = nil
	} else if input.ParentID != nil {
		fields["parent_id"] = *input.ParentID
	}
	if input.SortOrder != nil {
		fields["sort_order"] = *input.SortOrder
	}
	return fields
}
func (s *vaccineService) Delete(ctx context.Context, clinicID, id uint64) error {
	count, err := s.repo.CountUsageByVaccineID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check vaccine dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("このワクチンはワクチン接種記録で使用中のため削除できません")
	}
	return s.repo.Delete(ctx, clinicID, id)
}

func (s *vaccineService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	return s.repo.Reorder(ctx, clinicID, ids)
}
