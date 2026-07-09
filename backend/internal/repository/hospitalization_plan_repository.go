// Package repository provides data access implementations for HospitalizationPlan entity.
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- HospitalizationPlan ----

type HospitalizationPlanRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.HospitalizationPlan, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.HospitalizationPlan, error)
	Create(ctx context.Context, plan *model.HospitalizationPlan) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.HospitalizationPlan, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByHospitalizationPlanID(ctx context.Context, clinicID, planID uint64) (int64, error)
}

type hospitalizationPlanRepository struct{ db *gorm.DB }

func NewHospitalizationPlanRepository(db *gorm.DB) HospitalizationPlanRepository {
	return &hospitalizationPlanRepository{db: db}
}

func (r *hospitalizationPlanRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.HospitalizationPlan, error) {
	plans := make([]model.HospitalizationPlan, 0)
	err := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Order("sort_order ASC, name ASC").Find(&plans).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "hospitalization_plan", "")
	}
	return plans, nil
}

func (r *hospitalizationPlanRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.HospitalizationPlan, error) {
	var plan model.HospitalizationPlan
	err := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).First(&plan).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "hospitalization_plan", fmt.Sprintf("%d", id))
	}
	return &plan, nil
}

func (r *hospitalizationPlanRepository) Create(ctx context.Context, plan *model.HospitalizationPlan) error {
	err := r.db.WithContext(ctx).Create(plan).Error
	if err != nil {
		return apperrors.FromGORM(err, "hospitalization_plan", "")
	}
	return nil
}

func (r *hospitalizationPlanRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.HospitalizationPlan, error) {
	if err := updateScopedByID(ctx, r.db, &model.HospitalizationPlan{}, "hospitalization_plan", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *hospitalizationPlanRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return deleteScopedByID(ctx, r.db, &model.HospitalizationPlan{}, "hospitalization_plan", clinicID, id)
}

// CountUsageByHospitalizationPlanID は指定入院プランを参照する care_plan_items の件数を返す（BUG-105）。
// care_plan_items は直接 clinic_id を持たないため、hospitalization_plans を JOIN して
// clinic 境界を明示する（CODE-QUALITY-229）。
func (r *hospitalizationPlanRepository) CountUsageByHospitalizationPlanID(ctx context.Context, clinicID, planID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.CarePlanItem{}).
		Joins("JOIN hospitalization_plans hp ON hp.id = care_plan_items.hospitalization_plan_id AND hp.clinic_id = ? AND hp.deleted_at IS NULL", clinicID).
		Where("care_plan_items.hospitalization_plan_id = ? AND care_plan_items.deleted_at IS NULL", planID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "care_plan_item", "")
	}
	return count, nil
}

func (r *hospitalizationPlanRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return reorderByClinicID(ctx, r.db, &model.HospitalizationPlan{}, "hospitalization_plan", clinicID, ids)
}
