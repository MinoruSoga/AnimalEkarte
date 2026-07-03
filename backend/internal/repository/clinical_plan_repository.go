package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ClinicalPlanRepository は診察所見・診断・治療方針のデータアクセスインターフェース
type ClinicalPlanRepository interface {
	FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) (*model.ClinicalPlan, error)
	Create(ctx context.Context, plan *model.ClinicalPlan) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type clinicalPlanRepository struct {
	db *gorm.DB
}

// NewClinicalPlanRepository はClinicalPlanRepositoryを初期化して返す
func NewClinicalPlanRepository(db *gorm.DB) ClinicalPlanRepository {
	return &clinicalPlanRepository{db: db}
}

func (r *clinicalPlanRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) (*model.ClinicalPlan, error) {
	var plan model.ClinicalPlan
	err := r.db.WithContext(ctx).
		Preload("DiagnosisType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("DiagnosisName", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Joins("JOIN medical_records ON medical_records.id = clinical_plans.medical_record_id AND medical_records.deleted_at IS NULL").
		Where("medical_records.clinic_id = ? AND clinical_plans.medical_record_id = ?", clinicID, medicalRecordID).
		First(&plan).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "clinical_plan", fmt.Sprintf("%d", medicalRecordID))
	}
	return &plan, nil
}

func (r *clinicalPlanRepository) Create(ctx context.Context, plan *model.ClinicalPlan) error {
	err := r.db.WithContext(ctx).Create(plan).Error
	if err != nil {
		return apperrors.FromGORM(err, "clinical_plan", "")
	}
	return nil
}

func (r *clinicalPlanRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.ClinicalPlan{}).
		Where("clinical_plans.id = ? AND clinical_plans.medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ? AND deleted_at IS NULL)", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "clinical_plan", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("clinical_plan", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *clinicalPlanRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	// NOTE: GORM does not propagate Joins() into the generated DELETE statement's SQL
	// (it is a SELECT-only clause), so a WHERE referencing the joined table fails with
	// "missing FROM-clause entry". Use the same subquery form as Update above.
	result := r.db.WithContext(ctx).
		Where("clinical_plans.id = ? AND clinical_plans.medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ? AND deleted_at IS NULL)", id, clinicID).
		Delete(&model.ClinicalPlan{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "clinical_plan", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("clinical_plan", fmt.Sprintf("%d", id))
	}
	return nil
}
