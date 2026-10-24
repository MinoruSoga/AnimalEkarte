package medicalrecord

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// ClinicalPlanRepository は診察所見・診断・治療方針のデータアクセスインターフェース。
// Moved from internal/repository (BE9-2D sub-batch④a).
// BUG-010 residual: write/read paths use persistence.DBOrTx so service Update can keep
// business write + audit in one ambient transaction.
// EMR-495 A: 診断は clinical_plan_diagnoses の行（ListDiagnoses / ReplaceDiagnoses / DeleteDiagnoses）で保持する。
type ClinicalPlanRepository interface {
	FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) (*model.ClinicalPlan, error)
	Create(ctx context.Context, plan *model.ClinicalPlan) error
	Update(ctx context.Context, clinicID, id uint64, cmd UpdateClinicalPlanInput, expectedVersion *int) error
	Delete(ctx context.Context, clinicID, id uint64) error
	// ListDiagnoses は plan の生存中の診断行を slot, sort_order, id 昇順で返す。
	ListDiagnoses(ctx context.Context, clinicID, planID uint64) ([]model.ClinicalPlanDiagnosis, error)
	// ReplaceDiagnoses は plan の診断行集合を rows に置き換える（ambient tx に参加。無ければ自前の tx）。
	// 生存行と rows が 6 値すべてで一致する場合は何も書かない。それ以外は生存行を論理削除（deleted_at）し、
	// rows を新規行として挿入する。version はストレージが管理するため設定しない。
	ReplaceDiagnoses(ctx context.Context, clinicID, planID uint64, rows []ClinicalPlanDiagnosisInput) error
	// DeleteDiagnoses は plan の生存中の診断行を論理削除する（plan 削除と同じ tx で呼ぶ）。
	DeleteDiagnoses(ctx context.Context, clinicID, planID uint64) error
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
		Joins("JOIN medical_records ON medical_records.id = clinical_plans.medical_record_id AND medical_records.deleted_at IS NULL").
		Where("medical_records.clinic_id = ? AND clinical_plans.medical_record_id = ?", clinicID, medicalRecordID).
		First(&plan).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "clinical_plan", fmt.Sprintf("%d", medicalRecordID))
	}
	return &plan, nil
}

func (r *clinicalPlanRepository) Create(ctx context.Context, plan *model.ClinicalPlan) error {
	// clinical_plans.clinic_id は NOT NULL（001）。呼び出し元が trusted clinic を設定すること。
	if plan.ClinicID == 0 {
		return apperrors.WrapInternalServerError("clinical plan clinic_id is required")
	}
	err := persistence.DBOrTx(ctx, r.db).Create(plan).Error
	if err != nil {
		return apperrors.FromGORM(err, "clinical_plan", "")
	}
	return nil
}

func (r *clinicalPlanRepository) ListDiagnoses(ctx context.Context, clinicID, planID uint64) ([]model.ClinicalPlanDiagnosis, error) {
	return listLiveDiagnoses(persistence.DBOrTx(ctx, r.db), clinicID, planID, false)
}

func listLiveDiagnoses(db *gorm.DB, clinicID, planID uint64, forUpdate bool) ([]model.ClinicalPlanDiagnosis, error) {
	rows := make([]model.ClinicalPlanDiagnosis, 0)
	q := db.
		Where("clinic_id = ? AND clinical_plan_id = ? AND deleted_at IS NULL", clinicID, planID).
		Order("diagnosis_slot ASC, sort_order ASC, id ASC")
	if forUpdate {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, apperrors.FromGORM(err, "clinical_plan_diagnosis", fmt.Sprintf("clinical_plan_id=%d", planID))
	}
	return rows, nil
}

func (r *clinicalPlanRepository) ReplaceDiagnoses(ctx context.Context, clinicID, planID uint64, rows []ClinicalPlanDiagnosisInput) error {
	return persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		live, err := listLiveDiagnoses(tx, clinicID, planID, true)
		if err != nil {
			return err
		}
		if diagnosisRowsEqual(live, rows) {
			return nil
		}
		if err := softDeleteDiagnoses(tx, clinicID, planID, live); err != nil {
			return err
		}
		for _, row := range sortedDiagnosisInputs(rows) {
			created := &model.ClinicalPlanDiagnosis{
				ClinicID:              clinicID,
				ClinicalPlanID:        planID,
				DiagnosisSlot:         row.Slot,
				DiagnosisTypeID:       row.DiagnosisTypeID,
				DiagnosisNameID:       row.DiagnosisNameID,
				DiagnosisNameSnapshot: row.DiagnosisNameSnapshot,
				CategoryNameSnapshot:  row.CategoryNameSnapshot,
				SortOrder:             row.SortOrder,
			}
			if err := tx.Create(created).Error; err != nil {
				return apperrors.FromGORM(err, "clinical_plan_diagnosis", "")
			}
		}
		return nil
	})
}

func (r *clinicalPlanRepository) DeleteDiagnoses(ctx context.Context, clinicID, planID uint64) error {
	return persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		live, err := listLiveDiagnoses(tx, clinicID, planID, true)
		if err != nil {
			return err
		}
		return softDeleteDiagnoses(tx, clinicID, planID, live)
	})
}

// softDeleteDiagnoses は行の deleted_at を設定して論理削除する。model の DeletedAt は *time.Time で
// gorm の soft delete 対象ではないため Delete() は使わない（物理削除になる）。version は触らない
// （非 version 列の変更でストレージが自動的に +1 する）。
func softDeleteDiagnoses(tx *gorm.DB, clinicID, planID uint64, live []model.ClinicalPlanDiagnosis) error {
	if len(live) == 0 {
		return nil
	}
	ids := make([]uint64, 0, len(live))
	for _, row := range live {
		ids = append(ids, row.ID)
	}
	if err := tx.Model(&model.ClinicalPlanDiagnosis{}).
		Where("clinic_id = ? AND clinical_plan_id = ? AND id IN ? AND deleted_at IS NULL", clinicID, planID, ids).
		Update("deleted_at", gorm.Expr("clock_timestamp()")).Error; err != nil {
		return apperrors.FromGORM(err, "clinical_plan_diagnosis", fmt.Sprintf("clinical_plan_id=%d", planID))
	}
	return nil
}

// existsInClinic は id が clinicID 配下に存在するかを返す。
// Update/Delete の atomic WHERE が 0 行だった理由を「存在しない」と「バージョン不一致」で区別するために使う。
func (r *clinicalPlanRepository) existsInClinic(ctx context.Context, clinicID, id uint64) (bool, error) {
	var count int64
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.ClinicalPlan{}).
		Where("clinical_plans.id = ? AND clinical_plans.clinic_id = ? AND clinical_plans.medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ? AND deleted_at IS NULL)", id, clinicID, clinicID).
		Count(&count).Error; err != nil {
		return false, apperrors.FromGORM(err, "clinical_plan", fmt.Sprintf("%d", id))
	}
	return count > 0, nil
}

func (r *clinicalPlanRepository) conflictAfterZeroClinicalPlanRows(
	ctx context.Context,
	clinicID, id uint64,
	expectedVersion *int,
) error {
	exists, err := r.existsInClinic(ctx, clinicID, id)
	if err != nil {
		return err
	}
	if !exists || expectedVersion == nil {
		// version 述語なしで 0 行になるのは、行が存在しない（競合して消えた）場合だけ。
		return apperrors.WrapNotFound("clinical_plan", fmt.Sprintf("%d", id))
	}
	return apperrors.WrapConflict("他のユーザーがこの所見・診断を変更しました。再読み込みしてください")
}

// Update は診療計画の本文列とバージョンを更新する（BE-refactor.md X-11 / BUG-416③）。
// EMR-495: 旧 親カルテ status=draft の述語と「確定済み」Conflict は廃止した（status 列は存在しない）。
// persistence.DBOrTx で ambient tx に参加し、service の audit と同一 tx に載せる（BUG-010 residual）。
// expectedVersion が非 nil の場合、WHERE に version 述語を追加して楽観ロックを UPDATE と原子化する。
// RowsAffected==0 は「存在しない」または「バージョン不一致」を意味しうるため existsInClinic で再照会し
// 正しいエラー種別に正規化する（estimate_repository.go の normalizeDeleteIfNotLockedMiss と同型）。
func (r *clinicalPlanRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateClinicalPlanInput, expectedVersion *int) error {
	return r.update(ctx, clinicID, id, buildClinicalPlanUpdate(&cmd), expectedVersion)
}

func (r *clinicalPlanRepository) update(ctx context.Context, clinicID, id uint64, fields map[string]any, expectedVersion *int) error {
	q := persistence.DBOrTx(ctx, r.db).
		Model(&model.ClinicalPlan{}).
		Where("clinical_plans.id = ? AND clinical_plans.clinic_id = ?", id, clinicID)
	if expectedVersion != nil {
		q = q.Where("clinical_plans.version = ?", *expectedVersion)
	}
	result := q.Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "clinical_plan", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return r.conflictAfterZeroClinicalPlanRows(ctx, clinicID, id, expectedVersion)
	}
	return nil
}

// Delete は clinic スコープで診療計画を論理削除する。
func (r *clinicalPlanRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Where("clinical_plans.id = ? AND clinical_plans.clinic_id = ?", id, clinicID).
		Delete(&model.ClinicalPlan{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "clinical_plan", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("clinical_plan", fmt.Sprintf("%d", id))
	}
	return nil
}
