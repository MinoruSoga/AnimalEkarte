package repository

import (
	"context"
	"strconv"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type TreatmentPlanRepository interface {
	ListByMedicalRecordID(ctx context.Context, medicalRecordID uint64) ([]model.TreatmentPlan, error)
	ListByHospitalizationID(ctx context.Context, hospitalizationID uint64) ([]model.TreatmentPlan, error)
	FindByID(ctx context.Context, id uint64) (*model.TreatmentPlan, error)
	Create(ctx context.Context, plan *model.TreatmentPlan) error
	Update(ctx context.Context, id uint64, fields map[string]any) error
	Delete(ctx context.Context, id uint64) error
}

type treatmentPlanRepository struct{ db *gorm.DB }

func NewTreatmentPlanRepository(db *gorm.DB) TreatmentPlanRepository {
	return &treatmentPlanRepository{db: db}
}

func (r *treatmentPlanRepository) ListByMedicalRecordID(ctx context.Context, medicalRecordID uint64) ([]model.TreatmentPlan, error) {
	plans := make([]model.TreatmentPlan, 0)
	if err := r.db.WithContext(ctx).
		Where("medical_record_id = ?", medicalRecordID).
		Order("sort_order ASC").
		Find(&plans).Error; err != nil {
		return nil, apperrors.FromGORM(err, "treatment_plan", "")
	}
	return plans, nil
}

func (r *treatmentPlanRepository) ListByHospitalizationID(ctx context.Context, hospitalizationID uint64) ([]model.TreatmentPlan, error) {
	plans := make([]model.TreatmentPlan, 0)
	if err := r.db.WithContext(ctx).
		Where("hospitalization_id = ?", hospitalizationID).
		Order("sort_order ASC").
		Find(&plans).Error; err != nil {
		return nil, apperrors.FromGORM(err, "treatment_plan", "")
	}
	return plans, nil
}

func (r *treatmentPlanRepository) FindByID(ctx context.Context, id uint64) (*model.TreatmentPlan, error) {
	var plan model.TreatmentPlan
	err := r.db.WithContext(ctx).First(&plan, id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "treatment_plan", strconv.FormatUint(id, 10))
	}
	return &plan, nil
}

func (r *treatmentPlanRepository) Create(ctx context.Context, plan *model.TreatmentPlan) error {
	if err := r.db.WithContext(ctx).Create(plan).Error; err != nil {
		return apperrors.FromGORM(err, "treatment_plan", "")
	}
	return nil
}

func (r *treatmentPlanRepository) Update(ctx context.Context, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.TreatmentPlan{}).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "treatment_plan", strconv.FormatUint(id, 10))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("treatment_plan", strconv.FormatUint(id, 10))
	}
	return nil
}

func (r *treatmentPlanRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.TreatmentPlan{}, id)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "treatment_plan", strconv.FormatUint(id, 10))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("treatment_plan", strconv.FormatUint(id, 10))
	}
	return nil
}
