package medicalrecord

// Moved from internal/repository (BE9-2D ⑥ Batch A). 旧 package-private helper は repohelpers
// 同等物へ置換（同一述語/ambient-tx参加）。外部は internal/repository の facade alias 経由で不変。

import (
	"context"
	"strconv"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

type TreatmentPlanRepository interface {
	FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.TreatmentPlan, error)
	FindByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.TreatmentPlan, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.TreatmentPlan, error)
	// LockByIDForUpdate serializes plan writes for discount recheck (SEC-CS-F10).
	// Fail-closed without an ambient transaction.
	LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.TreatmentPlan, error)
	Create(ctx context.Context, plan *model.TreatmentPlan) error
	// medicalRecordID / hospitalizationID optionally bind write to URL parent (MRD-03).
	Update(ctx context.Context, clinicID, id uint64, medicalRecordID, hospitalizationID *uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64, medicalRecordID, hospitalizationID *uint64) error
}

type treatmentPlanRepository struct{ db *gorm.DB }

func NewTreatmentPlanRepository(db *gorm.DB) TreatmentPlanRepository {
	return &treatmentPlanRepository{db: db}
}

func (r *treatmentPlanRepository) clinicScopeQuery(ctx context.Context, clinicID uint64) *gorm.DB {
	return persistence.DBOrTx(ctx, r.db).
		Model(&model.TreatmentPlan{}).
		Scopes(persistence.ClinicScope(clinicID))
}

func treatmentPlanParentClinicScope(db *gorm.DB) *gorm.DB {
	return db.Where(`
		(treatment_plans.medical_record_id IS NULL OR EXISTS (
			SELECT 1
			FROM medical_records
			WHERE medical_records.id = treatment_plans.medical_record_id
			  AND medical_records.clinic_id = treatment_plans.clinic_id
		))
		AND
		(treatment_plans.hospitalization_id IS NULL OR EXISTS (
			SELECT 1
			FROM hospitalizations
			WHERE hospitalizations.id = treatment_plans.hospitalization_id
			  AND hospitalizations.clinic_id = treatment_plans.clinic_id
		))
	`)
}

func applyTreatmentPlanParentBind(q *gorm.DB, medicalRecordID, hospitalizationID *uint64) *gorm.DB {
	if medicalRecordID != nil {
		q = q.Where("treatment_plans.medical_record_id = ?", *medicalRecordID)
	}
	if hospitalizationID != nil {
		q = q.Where("treatment_plans.hospitalization_id = ?", *hospitalizationID)
	}
	return q
}

func (r *treatmentPlanRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.TreatmentPlan, error) {
	plans := make([]model.TreatmentPlan, 0)
	if err := r.clinicScopeQuery(ctx, clinicID).
		Scopes(treatmentPlanParentClinicScope).
		Where("treatment_plans.medical_record_id = ?", medicalRecordID).
		Order("treatment_plans.sort_order ASC").
		Find(&plans).Error; err != nil {
		return nil, apperrors.FromGORM(err, "treatment_plan", "")
	}
	return plans, nil
}

func (r *treatmentPlanRepository) FindByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.TreatmentPlan, error) {
	plans := make([]model.TreatmentPlan, 0)
	if err := r.clinicScopeQuery(ctx, clinicID).
		Scopes(treatmentPlanParentClinicScope).
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
		Scopes(treatmentPlanParentClinicScope).
		Where("treatment_plans.id = ?", id).
		First(&plan).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "treatment_plan", strconv.FormatUint(id, 10))
	}
	return &plan, nil
}

// LockByIDForUpdate は clinic-scoped FOR UPDATE で treatment_plan 行を固定する（SEC-CS-F10）。
func (r *treatmentPlanRepository) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.TreatmentPlan, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("treatment plan lock requires an ambient transaction")
	}
	var plan model.TreatmentPlan
	// Call DBOrTx directly (not only via clinicScopeQuery) so lintscan inventory
	// recognizes ambient-tx participation for this lock helper.
	err := persistence.DBOrTx(ctx, r.db).
		Model(&model.TreatmentPlan{}).
		Scopes(persistence.ClinicScope(clinicID), treatmentPlanParentClinicScope).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("treatment_plans.id = ?", id).
		First(&plan).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "treatment_plan", strconv.FormatUint(id, 10))
	}
	return &plan, nil
}

func (r *treatmentPlanRepository) Create(ctx context.Context, plan *model.TreatmentPlan) error {
	if err := persistence.DBOrTx(ctx, r.db).Create(plan).Error; err != nil {
		return apperrors.FromGORM(err, "treatment_plan", "")
	}
	return nil
}

func (r *treatmentPlanRepository) Update(ctx context.Context, clinicID, id uint64, medicalRecordID, hospitalizationID *uint64, fields map[string]any) error {
	q := r.clinicScopeQuery(ctx, clinicID).
		Model(&model.TreatmentPlan{}).
		Where("treatment_plans.id = ?", id)
	q = applyTreatmentPlanParentBind(q, medicalRecordID, hospitalizationID)
	result := q.Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "treatment_plan", strconv.FormatUint(id, 10))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("treatment_plan", strconv.FormatUint(id, 10))
	}
	return nil
}

func (r *treatmentPlanRepository) Delete(ctx context.Context, clinicID, id uint64, medicalRecordID, hospitalizationID *uint64) error {
	q := r.clinicScopeQuery(ctx, clinicID).
		Where("treatment_plans.id = ?", id)
	q = applyTreatmentPlanParentBind(q, medicalRecordID, hospitalizationID)
	result := q.Delete(&model.TreatmentPlan{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "treatment_plan", strconv.FormatUint(id, 10))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("treatment_plan", strconv.FormatUint(id, 10))
	}
	return nil
}
