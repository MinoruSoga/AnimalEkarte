package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ClinicalPlanRepository は診察所見・診断・治療方針のデータアクセスインターフェース
type ClinicalPlanRepository interface {
	FindByMedicalRecordID(ctx context.Context, medicalRecordID uint64) (*model.ClinicalPlan, error)
	Create(ctx context.Context, plan *model.ClinicalPlan) error
	Update(ctx context.Context, id uint64, fields map[string]any) error
	Delete(ctx context.Context, id uint64) error
}

type clinicalPlanRepository struct {
	db *gorm.DB
}

// NewClinicalPlanRepository はClinicalPlanRepositoryを初期化して返す
func NewClinicalPlanRepository(db *gorm.DB) ClinicalPlanRepository {
	return &clinicalPlanRepository{db: db}
}

func (r *clinicalPlanRepository) FindByMedicalRecordID(ctx context.Context, medicalRecordID uint64) (*model.ClinicalPlan, error) {
	var plan model.ClinicalPlan
	err := r.db.WithContext(ctx).
		Preload("DiagnosisCategory").
		Preload("DiagnosisName").
		Where("medical_record_id = ?", medicalRecordID).
		First(&plan).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("clinical_plan", fmt.Sprintf("%d", medicalRecordID))
		}
		return nil, apperrors.Wrap(err, "find clinical_plan by medical_record_id")
	}
	return &plan, nil
}

func (r *clinicalPlanRepository) Create(ctx context.Context, plan *model.ClinicalPlan) error {
	if err := r.db.WithContext(ctx).Create(plan).Error; err != nil {
		return apperrors.Wrap(err, "create clinical_plan")
	}
	return nil
}

func (r *clinicalPlanRepository) Update(ctx context.Context, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.ClinicalPlan{}).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update clinical_plan")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("clinical_plan", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *clinicalPlanRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&model.ClinicalPlan{})
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete clinical_plan")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("clinical_plan", fmt.Sprintf("%d", id))
	}
	return nil
}
