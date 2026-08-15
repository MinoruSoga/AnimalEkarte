package medicalrecord

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// ClinicalPlanRepository は診察所見・診断・治療方針のデータアクセスインターフェース。
// Moved from internal/repository (BE9-2D sub-batch④a).
// BUG-010 residual: write/read paths use persistence.DBOrTx so service Update can keep
// business write + audit in one ambient transaction.
type ClinicalPlanRepository interface {
	FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) (*model.ClinicalPlan, error)
	Create(ctx context.Context, plan *model.ClinicalPlan) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any, expectedVersion *int) error
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
	err := persistence.DBOrTx(ctx, r.db).
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
	err := persistence.DBOrTx(ctx, r.db).Create(plan).Error
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
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.ClinicalPlan{}).
		Where("clinical_plans.id = ? AND clinical_plans.medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ? AND deleted_at IS NULL)", id, clinicID).
		Count(&count).Error; err != nil {
		return false, apperrors.FromGORM(err, "clinical_plan", fmt.Sprintf("%d", id))
	}
	return count > 0, nil
}

// parentStillDraft は id の clinical_plan の親カルテが（clinicID 配下で）draft のままかを返す。
// Update の RowsAffected==0 を「バージョン不一致（親は draft のまま）」と「親が確定済み」で
// 区別するために使う（existsInClinic と同じサブクエリ形状に status='draft' を追加したもの）。
func (r *clinicalPlanRepository) parentStillDraft(ctx context.Context, clinicID, id uint64) (bool, error) {
	var count int64
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.ClinicalPlan{}).
		Where("clinical_plans.id = ? AND clinical_plans.medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ? AND status = ? AND deleted_at IS NULL)",
			id, clinicID, model.MedicalRecordStatusDraft).
		Count(&count).Error; err != nil {
		return false, apperrors.FromGORM(err, "clinical_plan", fmt.Sprintf("%d", id))
	}
	return count > 0, nil
}

// Update は親カルテが draft のときのみ更新する（BE-refactor.md X-11 の確定済みカルテ書込ガード）。
// LockByIDForUpdate は取らない。medical_records の status='draft' 条件を WHERE に含め、
// UPDATE 自体を原子的に拒否する。BUG-010 residual: persistence.DBOrTx で ambient tx に参加し、
// service の audit と同一 tx に載せる。
// expectedVersion が非 nil の場合、WHERE に version 述語を追加し（BUG-416③: 楽観ロックの原子化、
// medical_record_repository.go の Update と同型）、RowsAffected==0 を「バージョン不一致（親は draft
// のまま）」と「親が確定済み」に区別する。expectedVersion が nil の場合は従来どおり version 述語
// なし（照合スキップ、medical_record_subrecords.go の best-effort 呼び出し等）。
// RowsAffected==0 は「存在しない」「親が確定済み」「バージョン不一致」のいずれかを意味しうるため
// existsInClinic / parentStillDraft で再照会し正しいエラー種別に正規化する
// （estimate_repository.go の normalizeDeleteIfNotLockedMiss と同型）。
func (r *clinicalPlanRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any, expectedVersion *int) error {
	q := persistence.DBOrTx(ctx, r.db).
		Model(&model.ClinicalPlan{}).
		Where("clinical_plans.id = ? AND clinical_plans.medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ? AND status = ? AND deleted_at IS NULL)",
			id, clinicID, model.MedicalRecordStatusDraft)
	if expectedVersion != nil {
		q = q.Where("clinical_plans.version = ?", *expectedVersion)
	}
	result := q.Updates(fields)
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
		if expectedVersion != nil {
			// 再照会失敗（親が確定済み等）はエラーを出さず従来の Conflict にフォールバックする
			// （情報を出し過ぎない。medical_record_repository.go の Update と同じ方針）。
			if stillDraft, draftErr := r.parentStillDraft(ctx, clinicID, id); draftErr == nil && stillDraft {
				return apperrors.WrapConflict("他のユーザーがこの所見・診断を変更しました。再読み込みしてください")
			}
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
	result := persistence.DBOrTx(ctx, r.db).
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
