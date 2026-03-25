// Package repository provides data access implementations for HospitalizationPlan entity.
package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- HospitalizationPlan ----

type HospitalizationPlanRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.HospitalizationPlan, error)
	FindByID(ctx context.Context, id uint64) (*model.HospitalizationPlan, error)
	Create(ctx context.Context, plan *model.HospitalizationPlan) error
	Update(ctx context.Context, plan *model.HospitalizationPlan) error
	Delete(ctx context.Context, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type hospitalizationPlanRepository struct{ db *gorm.DB }

func NewHospitalizationPlanRepository(db *gorm.DB) HospitalizationPlanRepository {
	return &hospitalizationPlanRepository{db: db}
}

func (r *hospitalizationPlanRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.HospitalizationPlan, error) {
	plans := make([]model.HospitalizationPlan, 0)
	if err := r.db.WithContext(ctx).Where("clinic_id = ?", clinicID).Order("sort_order ASC, name ASC").Find(&plans).Error; err != nil {
		return nil, apperrors.Wrap(err, "find hospitalization plans")
	}
	return plans, nil
}

func (r *hospitalizationPlanRepository) FindByID(ctx context.Context, id uint64) (*model.HospitalizationPlan, error) {
	var plan model.HospitalizationPlan
	if err := r.db.WithContext(ctx).First(&plan, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("hospitalization_plan", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.Wrap(err, "find hospitalization plan by id")
	}
	return &plan, nil
}

func (r *hospitalizationPlanRepository) Create(ctx context.Context, plan *model.HospitalizationPlan) error {
	if err := r.db.WithContext(ctx).Create(plan).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("hospitalization_plan", plan.Name)
		}
		return apperrors.Wrap(err, "create hospitalization plan")
	}
	return nil
}

func (r *hospitalizationPlanRepository) Update(ctx context.Context, plan *model.HospitalizationPlan) error {
	result := r.db.WithContext(ctx).
		Model(&model.HospitalizationPlan{}).
		Where("id = ? AND clinic_id = ?", plan.ID, plan.ClinicID).
		Updates(plan)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update hospitalization plan")
	}
	if result.RowsAffected == 0 {
		return apperrors.Wrap(apperrors.ErrNotFound, "update hospitalization plan")
	}
	return nil
}

func (r *hospitalizationPlanRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.HospitalizationPlan{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete hospitalization plan")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("hospitalization_plan", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *hospitalizationPlanRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&model.HospitalizationPlan{}).
				Where("id = ? AND clinic_id = ?", id, clinicID).
				Update("sort_order", i+1)
			if result.Error != nil {
				return apperrors.Wrap(result.Error, "reorder hospitalization plan")
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("hospitalization_plan id %d not found in this clinic", id))
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("reorder hospitalization plan: %w", err)
	}
	return nil
}
