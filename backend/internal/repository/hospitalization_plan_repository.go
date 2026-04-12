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
	UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.HospitalizationPlan, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountCarePlanItemsByPlanID(ctx context.Context, planID uint64) (int64, error)
}

type hospitalizationPlanRepository struct{ db *gorm.DB }

func NewHospitalizationPlanRepository(db *gorm.DB) HospitalizationPlanRepository {
	return &hospitalizationPlanRepository{db: db}
}

func (r *hospitalizationPlanRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.HospitalizationPlan, error) {
	plans := make([]model.HospitalizationPlan, 0)
	err := r.db.WithContext(ctx).Where("clinic_id = ?", clinicID).Order("sort_order ASC, name ASC").Find(&plans).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "hospitalization_plan", "")
	}
	return plans, nil
}

func (r *hospitalizationPlanRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.HospitalizationPlan, error) {
	var plan model.HospitalizationPlan
	err := r.db.WithContext(ctx).First(&plan, "id = ? AND clinic_id = ?", id, clinicID).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "hospitalization_plan", fmt.Sprintf("%d", id))
	}
	return &plan, nil
}

func (r *hospitalizationPlanRepository) Create(ctx context.Context, plan *model.HospitalizationPlan) error {
	err := r.db.WithContext(ctx).Create(plan).Error
	if err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapConflict("同じ名称が既に登録されています")
		}
		return apperrors.FromGORM(err, "hospitalization_plan", "")
	}
	return nil
}

func (r *hospitalizationPlanRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.HospitalizationPlan, error) {
	result := r.db.WithContext(ctx).
		Model(&model.HospitalizationPlan{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "hospitalization_plan", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("hospitalization_plan", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *hospitalizationPlanRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.HospitalizationPlan{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "hospitalization_plan", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("hospitalization_plan", fmt.Sprintf("%d", id))
	}
	return nil
}

// CountCarePlanItemsByPlanID は指定入院プランを参照する care_plan_items の件数を返す（BUG-105）
func (r *hospitalizationPlanRepository) CountCarePlanItemsByPlanID(ctx context.Context, planID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.CarePlanItem{}).
		Where("hospitalization_plan_id = ?", planID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "care_plan_item", "")
	}
	return count, nil
}

func (r *hospitalizationPlanRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return reorderByClinicID(ctx, r.db, &model.HospitalizationPlan{}, "hospitalization_plan", clinicID, ids)
}
