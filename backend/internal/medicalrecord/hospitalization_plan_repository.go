// Package medicalrecord provides hospitalization plan persistence.
package medicalrecord

// Moved from internal/repository (BE9-2D ⑤ Batch A). 旧 package-private helper は repohelpers 同等物へ
// 置換（同一述語/ambient-tx参加）。外部呼び出しは internal/repository の facade alias 経由で不変。

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
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
	err := r.db.WithContext(ctx).Scopes(persistence.ClinicScope(clinicID)).Order("sort_order ASC, name ASC").Find(&plans).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "hospitalization_plan", "")
	}
	return plans, nil
}

func (r *hospitalizationPlanRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.HospitalizationPlan, error) {
	return persistence.FindByIDScoped[model.HospitalizationPlan](ctx, r.db, "hospitalization_plan", clinicID, id)
}

func (r *hospitalizationPlanRepository) Create(ctx context.Context, plan *model.HospitalizationPlan) error {
	db := r.db.WithContext(ctx)
	wantActive := plan.IsActive
	if err := db.Create(plan).Error; err != nil {
		return apperrors.FromGORM(err, "hospitalization_plan", "")
	}
	if !wantActive {
		if err := db.Model(plan).Update("is_active", false).Error; err != nil {
			return apperrors.FromGORM(err, "hospitalization_plan", fmt.Sprintf("%d", plan.ID))
		}
		plan.IsActive = false
	}
	return nil
}

func (r *hospitalizationPlanRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.HospitalizationPlan, error) {
	if err := persistence.UpdateScopedByID(ctx, r.db, &model.HospitalizationPlan{}, "hospitalization_plan", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *hospitalizationPlanRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return persistence.DeleteScopedByID(ctx, r.db, &model.HospitalizationPlan{}, "hospitalization_plan", clinicID, id)
}

// CountUsageByHospitalizationPlanID は指定入院プランを参照する care_plan_items の件数を返す（BUG-105）。
// care_plan_items は直接 clinic_id を持たないため、hospitalization_plans を JOIN して
// clinic 境界を明示する（CODE-QUALITY-229）。
func (r *hospitalizationPlanRepository) CountUsageByHospitalizationPlanID(ctx context.Context, clinicID, planID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.CarePlanItem{}).
		Joins("JOIN hospitalization_plans hp ON hp.id = care_plan_items.hospitalization_plan_id AND hp.clinic_id = ? AND hp.deleted_at IS NULL", clinicID).
		Where("care_plan_items.hospitalization_plan_id = ?", planID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "care_plan_item", "")
	}
	return count, nil
}

func (r *hospitalizationPlanRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return persistence.ReorderByClinicID(ctx, r.db, &model.HospitalizationPlan{}, "hospitalization_plan", clinicID, ids, "sort_order")
}
