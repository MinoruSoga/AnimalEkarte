// Package repository provides data access implementations for Procedure entity.
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Procedure ----

type ProcedureRepository interface {
	FindAll(ctx context.Context) ([]model.Procedure, error)
	FindByID(ctx context.Context, id uint64) (*model.Procedure, error)
	Create(ctx context.Context, procedure *model.Procedure) error
	UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Procedure, error)
	Delete(ctx context.Context, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByProcedureID(ctx context.Context, procedureID uint64) (int64, error)
}

type procedureRepository struct{ db *gorm.DB }

func NewProcedureRepository(db *gorm.DB) ProcedureRepository { return &procedureRepository{db: db} }

func (r *procedureRepository) FindAll(ctx context.Context) ([]model.Procedure, error) {
	procedures := make([]model.Procedure, 0)
	if err := r.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&procedures).Error; err != nil {
		return nil, apperrors.Wrap(err, "find procedures")
	}
	return procedures, nil
}

func (r *procedureRepository) FindByID(ctx context.Context, id uint64) (*model.Procedure, error) {
	var procedure model.Procedure
	err := r.db.WithContext(ctx).First(&procedure, "id = ?", id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "procedure", fmt.Sprintf("%d", id))
	}
	return &procedure, nil
}

func (r *procedureRepository) Create(ctx context.Context, procedure *model.Procedure) error {
	if err := r.db.WithContext(ctx).Create(procedure).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("procedure", procedure.Name)
		}
		return apperrors.Wrap(err, "create procedure")
	}
	return nil
}

func (r *procedureRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Procedure, error) {
	result := r.db.WithContext(ctx).
		Model(&model.Procedure{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.Wrap(result.Error, "update procedure")
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("procedure", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, id)
}

func (r *procedureRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.Procedure{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete procedure")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("procedure", fmt.Sprintf("%d", id))
	}
	return nil
}

// CountUsageByProcedureID は treatments と care_plan_items で参照されている件数の合計を返す（BUG-107）
func (r *procedureRepository) CountUsageByProcedureID(ctx context.Context, procedureID uint64) (int64, error) {
	var treatmentCount, carePlanCount int64
	if err := r.db.WithContext(ctx).
		Model(&model.Treatment{}).
		Where("procedure_id = ?", procedureID).
		Count(&treatmentCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "treatment", "")
	}
	if err := r.db.WithContext(ctx).
		Model(&model.CarePlanItem{}).
		Where("procedure_id = ?", procedureID).
		Count(&carePlanCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "care_plan_item", "")
	}
	return treatmentCount + carePlanCount, nil
}

func (r *procedureRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&model.Procedure{}).
				Where("id = ? AND clinic_id = ?", id, clinicID).
				Update("sort_order", i+1)
			if result.Error != nil {
				return apperrors.Wrap(result.Error, "reorder procedure")
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("procedure id %d not found in this clinic", id))
			}
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "reorder procedure")
	}
	return nil
}
