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
		Preload("Diagnosis2Type", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Diagnosis2Name", "clinic_id = ? AND deleted_at IS NULL", clinicID).
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

// existsInClinic は id が clinicID 配下（親カルテの draft/finalized を問わず）に存在するかを返す。
// Update/Delete の atomic WHERE が 0 行だった理由を「存在しない」と「親カルテが確定済み」で
// 区別するために使う。
func (r *clinicalPlanRepository) existsInClinic(ctx context.Context, clinicID, id uint64) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.ClinicalPlan{}).
		Where("clinical_plans.id = ? AND clinical_plans.medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ? AND deleted_at IS NULL)", id, clinicID).
		Count(&count).Error; err != nil {
		return false, apperrors.FromGORM(err, "clinical_plan", fmt.Sprintf("%d", id))
	}
	return count > 0, nil
}

// Update は親カルテが draft のときのみ更新する（BE-refactor.md X-11 の確定済みカルテ書込ガード）。
// clinicalPlanService は Transactor を持たない設計上の制約（cross_tenant_master_fk_write_test.go の
// 既存呼び出しが repository.Transactor 無しの旧コンストラクタ引数に依存しており、変更できない）のため
// LockByIDForUpdate による行ロックは行わない。代わりに medical_records の status='draft' 条件を
// WHERE 部分に含め、UPDATE 自体を原子的に（サービス層の事前チェックとのレース窓を残さず）拒否する。
// RowsAffected==0 は「存在しない」と「親が確定済み」の両方を意味しうるため existsInClinic で
// 再照会し正しいエラー種別（NotFound / Conflict）に正規化する
// （estimate_repository.go の normalizeDeleteIfNotLockedMiss と同型）。
func (r *clinicalPlanRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.ClinicalPlan{}).
		Where("clinical_plans.id = ? AND clinical_plans.medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ? AND status = ? AND deleted_at IS NULL)",
			id, clinicID, model.MedicalRecordStatusDraft).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "clinical_plan", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		exists, err := r.existsInClinic(ctx, clinicID, id)
		if err != nil {
			return err
		}
		if !exists {
			return apperrors.WrapNotFound("clinical_plan", fmt.Sprintf("%d", id))
		}
		return apperrors.WrapConflict("確定済みカルテの所見・診断は編集できません")
	}
	return nil
}

// Delete も Update と同じ理由で親カルテ draft 条件を課す。
func (r *clinicalPlanRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	// NOTE: GORM does not propagate Joins() into the generated DELETE statement's SQL
	// (it is a SELECT-only clause), so a WHERE referencing the joined table fails with
	// "missing FROM-clause entry". Use the same subquery form as Update above.
	result := r.db.WithContext(ctx).
		Where("clinical_plans.id = ? AND clinical_plans.medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ? AND status = ? AND deleted_at IS NULL)",
			id, clinicID, model.MedicalRecordStatusDraft).
		Delete(&model.ClinicalPlan{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "clinical_plan", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		exists, err := r.existsInClinic(ctx, clinicID, id)
		if err != nil {
			return err
		}
		if !exists {
			return apperrors.WrapNotFound("clinical_plan", fmt.Sprintf("%d", id))
		}
		return apperrors.WrapConflict("確定済みカルテの所見・診断は削除できません")
	}
	return nil
}
