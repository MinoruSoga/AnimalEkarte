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
	FindAll(ctx context.Context, clinicID uint64) ([]model.Procedure, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Procedure, error)
	Create(ctx context.Context, procedure *model.Procedure) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Procedure, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByProcedureID(ctx context.Context, clinicID, procedureID uint64) (int64, error)
	CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error)
}

type procedureRepository struct{ db *gorm.DB }

func NewProcedureRepository(db *gorm.DB) ProcedureRepository { return &procedureRepository{db: db} }

func (r *procedureRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Procedure, error) {
	procedures := make([]model.Procedure, 0)
	if err := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Order("sort_order ASC, name ASC").Limit(10000).Find(&procedures).Error; err != nil {
		return nil, apperrors.FromGORM(err, "procedure", "")
	}
	return procedures, nil
}

func (r *procedureRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Procedure, error) {
	var procedure model.Procedure
	err := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).First(&procedure).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "procedure", fmt.Sprintf("%d", id))
	}
	return &procedure, nil
}

func (r *procedureRepository) Create(ctx context.Context, procedure *model.Procedure) error {
	if err := r.db.WithContext(ctx).Create(procedure).Error; err != nil {
		return apperrors.FromGORM(err, "procedure", "")
	}
	return nil
}

func (r *procedureRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Procedure, error) {
	if err := updateScopedByID(ctx, r.db, &model.Procedure{}, "procedure", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *procedureRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return deleteScopedByID(ctx, r.db, &model.Procedure{}, "procedure", clinicID, id)
}

// CountUsageByProcedureID は treatments と care_plan_items で参照されている件数の合計を返す（BUG-107）
// treatments/care_plan_items は直接 clinic_id を持たないため JOIN でテナント分離する
func (r *procedureRepository) CountUsageByProcedureID(ctx context.Context, clinicID, procedureID uint64) (int64, error) {
	var treatmentCount, carePlanCount int64
	if err := r.db.WithContext(ctx).
		Model(&model.Treatment{}).
		Joins("JOIN medical_records ON medical_records.id = treatments.medical_record_id AND medical_records.clinic_id = ? AND medical_records.deleted_at IS NULL", clinicID).
		Where("treatments.procedure_id = ? AND treatments.deleted_at IS NULL", procedureID).
		Count(&treatmentCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "treatment", "")
	}
	if err := r.db.WithContext(ctx).
		Model(&model.CarePlanItem{}).
		Joins("JOIN hospitalizations ON hospitalizations.id = care_plan_items.hospitalization_id AND hospitalizations.clinic_id = ? AND hospitalizations.deleted_at IS NULL", clinicID).
		Where("care_plan_items.procedure_id = ? AND care_plan_items.deleted_at IS NULL", procedureID).
		Count(&carePlanCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "care_plan_item", "")
	}
	return treatmentCount + carePlanCount, nil
}

func (r *procedureRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return reorderByClinicID(ctx, r.db, &model.Procedure{}, "procedure", clinicID, ids)
}

// CountChildrenByParentID は指定した処置の子処置数を返す (BUG-390)
func (r *procedureRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Procedure{}).
		Scopes(clinicScope(clinicID)).
		Where("parent_id = ? AND deleted_at IS NULL", parentID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "procedure", fmt.Sprintf("%d", parentID))
	}
	return count, nil
}
