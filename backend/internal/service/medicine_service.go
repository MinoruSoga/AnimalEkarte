// Package service provides business logic implementations for Medicine entity.
package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// --- DB column constants ---

const (
	colMedicineName            = "name"
	colMedicineDrugCategory    = "drug_category"
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
	DrugCategory    *string
	Price           *float64
	IsActive        bool
	Description     string
	DosageForm      string // "" means not set
	MedicineUnit    string // "" means not set
	InventoryID     *uint64
	DefaultQuantity int
	SortOrder       int
}

// UpdateMedicineInput は薬剤更新の入力DTO（nil = 未指定）
type UpdateMedicineInput struct {
	Name            *string
	DrugCategory    *string  // nil = 未指定, "" = NULL クリア, "抗生剤" = 値セット
	Price           *float64
	IsActive        *bool
	Description     *string
	DosageForm      *string  // nil = 未指定, "" = NULL クリア, "tablet" = 値セット
	MedicineUnit    *string  // nil = 未指定, "" = NULL クリア, "per_ml" = 値セット
	InventoryID     **uint64 // nil = 未指定, &nil = NULL クリア, &&val = 値セット
	DefaultQuantity *int
	SortOrder       *int
}

// buildMedicineUpdateFields は UpdateMedicineInput から map[string]any を構築する。
// GORM のゼロ値スキップ問題（bool false が無視される等）を回避するために使用する。
func buildMedicineUpdateFields(input *UpdateMedicineInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colMedicineName] = *input.Name
	}
	if input.DrugCategory != nil {
		if *input.DrugCategory == "" {
			fields[colMedicineDrugCategory] = nil
		} else {
			fields[colMedicineDrugCategory] = *input.DrugCategory
		}
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
	List(ctx context.Context, clinicID uint64) ([]model.Medicine, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Medicine, error)
	Create(ctx context.Context, clinicID uint64, input *CreateMedicineInput) (*model.Medicine, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateMedicineInput) (*model.Medicine, error)
	Delete(ctx context.Context, clinicID, id uint64) error
}

type medicineService struct {
	repo   repository.MedicineRepository
	logger *slog.Logger
}

func NewMedicineService(repo repository.MedicineRepository, logger *slog.Logger) MedicineService {
	return &medicineService{repo: repo, logger: logger}
}

func (s *medicineService) List(ctx context.Context, clinicID uint64) ([]model.Medicine, error) {
	return s.repo.FindAll(ctx, clinicID)
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
		DrugCategory:    input.DrugCategory,
		Price:           input.Price,
		IsActive:        input.IsActive,
		Description:     input.Description,
		InventoryID:     input.InventoryID,
		DefaultQuantity: input.DefaultQuantity,
		SortOrder:       input.SortOrder,
	}
	if input.DosageForm != "" {
		df := model.DosageForm(input.DosageForm)
		medicine.DosageForm = &df
	}
	if input.MedicineUnit != "" {
		mu := model.MedicineUnit(input.MedicineUnit)
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

func (s *medicineService) Delete(ctx context.Context, clinicID, id uint64) error {
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return err
	}
	s.logger.InfoContext(ctx, "medicine deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("medicine_id", id),
	)
	return nil
}
