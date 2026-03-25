package repository

import (
	"context"
	"errors"
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
		return nil, apperrors.Wrap(err, "list treatment plans by medical record id")
	}
	return plans, nil
}

func (r *treatmentPlanRepository) ListByHospitalizationID(ctx context.Context, hospitalizationID uint64) ([]model.TreatmentPlan, error) {
	plans := make([]model.TreatmentPlan, 0)
	if err := r.db.WithContext(ctx).
		Where("hospitalization_id = ?", hospitalizationID).
		Order("sort_order ASC").
		Find(&plans).Error; err != nil {
		return nil, apperrors.Wrap(err, "list treatment plans by hospitalization id")
	}
	return plans, nil
}

func (r *treatmentPlanRepository) FindByID(ctx context.Context, id uint64) (*model.TreatmentPlan, error) {
	var plan model.TreatmentPlan
	if err := r.db.WithContext(ctx).First(&plan, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("treatment_plan", strconv.FormatUint(id, 10))
		}
		return nil, apperrors.Wrap(err, "find treatment plan by id")
	}
	return &plan, nil
}

func (r *treatmentPlanRepository) Create(ctx context.Context, plan *model.TreatmentPlan) error {
	if err := r.db.WithContext(ctx).Create(plan).Error; err != nil {
		return apperrors.Wrap(err, "create treatment plan")
	}
	return nil
}

func (r *treatmentPlanRepository) Update(ctx context.Context, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.TreatmentPlan{}).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update treatment plan")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("treatment_plan", strconv.FormatUint(id, 10))
	}
	return nil
}

func (r *treatmentPlanRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.TreatmentPlan{}, id)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete treatment plan")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("treatment_plan", strconv.FormatUint(id, 10))
	}
	return nil
}
