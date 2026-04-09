package repository

import (
	"context"
	"strconv"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type TreatmentPlanRepository interface {
	ListByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.TreatmentPlan, error)
	ListByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.TreatmentPlan, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.TreatmentPlan, error)
	Create(ctx context.Context, plan *model.TreatmentPlan) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type treatmentPlanRepository struct{ db *gorm.DB }

func NewTreatmentPlanRepository(db *gorm.DB) TreatmentPlanRepository {
	return &treatmentPlanRepository{db: db}
}

// clinicScopeQuery はテナント境界を適用したクエリビルダーを返す。
// treatment_plans は clinic_id を直接持たないため、親テーブル（medical_records / hospitalizations）経由で検証する。
func (r *treatmentPlanRepository) clinicScopeQuery(ctx context.Context, clinicID uint64) *gorm.DB {
	return r.db.WithContext(ctx).
		Joins("LEFT JOIN medical_records ON medical_records.id = treatment_plans.medical_record_id AND medical_records.deleted_at IS NULL").
		Joins("LEFT JOIN hospitalizations ON hospitalizations.id = treatment_plans.hospitalization_id AND hospitalizations.deleted_at IS NULL").
		Where("(medical_records.clinic_id = ? OR hospitalizations.clinic_id = ?)", clinicID, clinicID)
}

func (r *treatmentPlanRepository) ListByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.TreatmentPlan, error) {
	plans := make([]model.TreatmentPlan, 0)
	if err := r.clinicScopeQuery(ctx, clinicID).
		Where("treatment_plans.medical_record_id = ?", medicalRecordID).
		Order("treatment_plans.sort_order ASC").
		Find(&plans).Error; err != nil {
		return nil, apperrors.FromGORM(err, "treatment_plan", "")
	}
	return plans, nil
}

func (r *treatmentPlanRepository) ListByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.TreatmentPlan, error) {
	plans := make([]model.TreatmentPlan, 0)
	if err := r.clinicScopeQuery(ctx, clinicID).
		Where("treatment_plans.hospitalization_id = ?", hospitalizationID).
		Order("treatment_plans.sort_order ASC").
		Find(&plans).Error; err != nil {
		return nil, apperrors.FromGORM(err, "treatment_plan", "")
	}
	return plans, nil
}

func (r *treatmentPlanRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.TreatmentPlan, error) {
	var plan model.TreatmentPlan
	err := r.clinicScopeQuery(ctx, clinicID).
		Where("treatment_plans.id = ?", id).
		First(&plan).Error
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

func (r *treatmentPlanRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.clinicScopeQuery(ctx, clinicID).
		Model(&model.TreatmentPlan{}).
		Where("treatment_plans.id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "treatment_plan", strconv.FormatUint(id, 10))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("treatment_plan", strconv.FormatUint(id, 10))
	}
	return nil
}

func (r *treatmentPlanRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.clinicScopeQuery(ctx, clinicID).
		Where("treatment_plans.id = ?", id).
		Delete(&model.TreatmentPlan{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "treatment_plan", strconv.FormatUint(id, 10))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("treatment_plan", strconv.FormatUint(id, 10))
	}
	return nil
}
