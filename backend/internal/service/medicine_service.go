// Package service provides business logic implementations for Medicine entity.
package service

import (
	"context"
	"fmt"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// --- DB column constants ---

const (
	colMedicineName            = "name"
	colMedicineParentID        = "parent_id"
	colMedicinePrice           = "price"
	colMedicineIsActive        = "is_active"
	colMedicineDescription     = "description"
	colMedicineDosageForm      = "dosage_form"
	colMedicineMedicineUnit    = "medicine_unit"
	colMedicineInventoryID     = "inventory_id"
	colMedicineDefaultQuantity = "default_quantity"
	colMedicineSortOrder       = "sort_order"
)

// --- Input DTOs ---

// CreateMedicineInput は薬剤作成の入力DTO
type CreateMedicineInput struct {
	Name            string
	ParentID        *uint64
	Price           *float64
	IsActive        bool
	Description     string
	DosageForm      *string // nil = 未指定, "tablet" 等 = 値セット
	MedicineUnit    *string // nil = 未指定, "per_ml" 等 = 値セット
	InventoryID     *uint64
	DefaultQuantity float64
	SortOrder       int
}

// UpdateMedicineInput は薬剤更新の入力DTO（nil = 未指定）
type UpdateMedicineInput struct {
	Name            *string
	ParentID        *uint64 // nil = 未指定（ClearParentID=false 時）
	ClearParentID   bool    // true = parent_id を NULL にクリア
	Price           *float64
	IsActive        *bool
	Description     *string
	DosageForm      *string  // nil = 未指定, "" = NULL クリア, "tablet" = 値セット
	MedicineUnit    *string  // nil = 未指定, "" = NULL クリア, "per_ml" = 値セット
	InventoryID     **uint64 // nil = 未指定, &nil = NULL クリア, &&val = 値セット
	DefaultQuantity *float64
	SortOrder       *int
}

// buildMedicineUpdateFields は UpdateMedicineInput から map[string]any を構築する。
// GORM のゼロ値スキップ問題（bool false が無視される等）を回避するために使用する。
func buildMedicineUpdateFields(input *UpdateMedicineInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colMedicineName] = *input.Name
	}
	if input.ClearParentID {
		fields[colMedicineParentID] = nil
	} else if input.ParentID != nil {
		fields[colMedicineParentID] = *input.ParentID
	}
	if input.Price != nil {
		fields[colMedicinePrice] = *input.Price
	}
	if input.IsActive != nil {
		fields[colMedicineIsActive] = *input.IsActive
	}
	if input.Description != nil {
		fields[colMedicineDescription] = *input.Description
	}
	if input.DosageForm != nil {
		if *input.DosageForm == "" {
			fields[colMedicineDosageForm] = nil
		} else {
			fields[colMedicineDosageForm] = *input.DosageForm
		}
	}
	if input.MedicineUnit != nil {
		if *input.MedicineUnit == "" {
			fields[colMedicineMedicineUnit] = nil
		} else {
			fields[colMedicineMedicineUnit] = *input.MedicineUnit
		}
	}
	if input.InventoryID != nil {
		fields[colMedicineInventoryID] = *input.InventoryID // *uint64 (nil = NULL)
	}
	if input.DefaultQuantity != nil {
		fields[colMedicineDefaultQuantity] = *input.DefaultQuantity
	}
	if input.SortOrder != nil {
		fields[colMedicineSortOrder] = *input.SortOrder
	}
	return fields
}

// ---- MedicineService ----

type MedicineService interface {
	List(ctx context.Context, clinicID uint64, page, limit int) ([]model.Medicine, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Medicine, error)
	Create(ctx context.Context, clinicID uint64, input *CreateMedicineInput) (*model.Medicine, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateMedicineInput) (*model.Medicine, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type medicineService struct {
	repo   repository.MedicineRepository
	logger *slog.Logger
}

func NewMedicineService(repo repository.MedicineRepository, logger *slog.Logger) MedicineService {
	return &medicineService{repo: repo, logger: logger}
}

func (s *medicineService) List(ctx context.Context, clinicID uint64, page, limit int) ([]model.Medicine, int64, error) {
	return s.repo.FindAll(ctx, clinicID, page, limit)
}

func (s *medicineService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Medicine, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *medicineService) Create(ctx context.Context, clinicID uint64, input *CreateMedicineInput) (*model.Medicine, error) {
	if input.Name == "" {
		return nil, apperrors.WrapInvalidInput("name is required")
	}

	medicine := &model.Medicine{
		ClinicID:        clinicID,
		Name:            input.Name,
		ParentID:        input.ParentID,
		Price:           input.Price,
		IsActive:        input.IsActive,
		Description:     input.Description,
		InventoryID:     input.InventoryID,
		DefaultQuantity: input.DefaultQuantity,
		SortOrder:       input.SortOrder,
	}
	if input.DosageForm != nil && *input.DosageForm != "" {
		df := model.DosageForm(*input.DosageForm)
		medicine.DosageForm = &df
	}
	if input.MedicineUnit != nil && *input.MedicineUnit != "" {
		mu := model.MedicineUnit(*input.MedicineUnit)
		medicine.MedicineUnit = &mu
	}

	if err := s.repo.Create(ctx, medicine); err != nil {
		return nil, err
	}

	s.logger.InfoContext(ctx, "medicine created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("medicine_id", medicine.ID),
		slog.String("name", medicine.Name),
	)
	return medicine, nil
}

func (s *medicineService) Update(ctx context.Context, clinicID, id uint64, input *UpdateMedicineInput) (*model.Medicine, error) {
	fields := buildMedicineUpdateFields(input)
	if len(fields) == 0 {
		return s.repo.FindByID(ctx, clinicID, id)
	}

	if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		return nil, err
	}

	s.logger.InfoContext(ctx, "medicine updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("medicine_id", id),
	)
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *medicineService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	return s.repo.Reorder(ctx, clinicID, ids)
}

func (s *medicineService) Delete(ctx context.Context, clinicID, id uint64) error {
	m, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return err
	}

	// カテゴリ（parent_id = NULL）の場合、子アイテムが存在すれば削除を拒否する
	if m.ParentID == nil {
		count, err := s.repo.CountChildren(ctx, clinicID, id)
		if err != nil {
			return err
		}
		if count > 0 {
			return apperrors.WrapInvalidInput(
				fmt.Sprintf("このカテゴリには%d件の薬剤が含まれています。先に薬剤を移動または削除してください", count),
			)
		}
	}

	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return err
	}
	s.logger.InfoContext(ctx, "medicine deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("medicine_id", id),
	)
	return nil
}
