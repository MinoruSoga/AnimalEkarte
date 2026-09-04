package medicalrecord

// Moved from internal/repository (BE9-2D ⑦ Batch A). 旧 package-private helper は repohelpers
// 同等物へ置換（同一述語/ambient-tx参加）。外部は internal/repository の facade alias 経由で不変。

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

type ExaminationRepository interface {
	FindAll(ctx context.Context, clinicID uint64, petID, ownerID, medicalRecordID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Examination, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Examination, error)
	LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Examination, error)
	// FindByJobID は clinic_id + job_id で絞り込んだ exams を日付降順で返す（Phase 4B.2）。
	FindByJobID(ctx context.Context, clinicID uint64, jobID uuid.UUID) ([]*model.Examination, error)
	Create(ctx context.Context, exam *model.Examination) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Examination, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	CountItemsByExamID(ctx context.Context, clinicID, examID uint64) (int64, error)
	FindAllItemsByExamID(ctx context.Context, clinicID, examID uint64) ([]model.ExamResult, error)
	// ReplaceItemsByExamID は exam_results を一括置換する。第 2 戻り値は実際に削除された行数
	// （DELETE の RowsAffected）。サービス層が「削除が実際に起きたか」を監査ゲートに使う
	// （BE-refactor.md R1-2・#211 checkup_field_result と同方針）。
	ReplaceItemsByExamID(ctx context.Context, clinicID, examID uint64, items []model.ExamResult) ([]model.ExamResult, int64, error)
}

type examinationRepository struct {
	db *gorm.DB
}

func NewExaminationRepository(db *gorm.DB) ExaminationRepository {
	return &examinationRepository{db: db}
}

func (r *examinationRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID, medicalRecordID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Examination, int64, error) {
	buildBase := func() *gorm.DB {
		q := persistence.DBOrTx(ctx, r.db).Model(&model.Examination{}).
			Where("exams.clinic_id = ?", clinicID).
			Scopes(examinationPatientRelationsScope(clinicID))
		switch {
		case petID != nil && medicalRecordID != nil:
			// カルテ検査タブの契約: そのカルテへ取り込んだ検査に加え、同じペットの
			// 未取り込み検査（機器受信 attach 直後は medical_record_id IS NULL）も返す。
			// record 単独条件にすると機器由来の結果が診察中に見えなくなる。
			q = q.Where("exams.pet_id = ?", *petID).
				Where("(exams.medical_record_id = ? OR exams.medical_record_id IS NULL)", *medicalRecordID)
		case petID != nil:
			q = q.Where("exams.pet_id = ?", *petID)
		case medicalRecordID != nil:
			q = q.Where("exams.medical_record_id = ?", *medicalRecordID)
		}
		if ownerID != nil {
			q = q.Joins(
				"JOIN pets ON pets.id = exams.pet_id AND pets.clinic_id = ? AND pets.deleted_at IS NULL",
				clinicID,
			).Where("pets.owner_id = ?", *ownerID)
		}
		if status != nil {
			q = q.Where("exams.status = ?", *status)
		}
		if startDate != nil {
			q = q.Where("exams.date >= ?", *startDate)
		}
		if endDate != nil {
			q = q.Where("exams.date <= ?", *endDate)
		}
		return q
	}

	var total int64
	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "exam", "")
	}

	exams := make([]model.Examination, 0)
	if err := buildBase().Scopes(examinationReadPreloads(clinicID)).
		Scopes(paginate(page, limit)).Order("exams.date DESC, exams.created_at DESC").
		Find(&exams).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "exam", "")
	}
	return exams, total, nil
}

func (r *examinationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Examination, error) {
	var exam model.Examination
	err := persistence.DBOrTx(ctx, r.db).
		Where("exams.id = ? AND exams.clinic_id = ?", id, clinicID).
		Scopes(
			examinationPatientRelationsScope(clinicID),
			examinationReadPreloads(clinicID),
		).
		First(&exam).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "exam", fmt.Sprintf("%d", id))
	}
	return &exam, nil
}

// LockByIDForUpdate serializes examination status, parent movement, result replacement,
// and deletion. The lock is meaningful only while an ambient transaction remains open.
func (r *examinationRepository) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Examination, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("examination lock requires an ambient transaction")
	}
	var exam model.Examination
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).
		First(&exam).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "exam", fmt.Sprintf("%d", id))
	}
	return &exam, nil
}

// FindByJobID は clinic_id + job_id scope で exams を日付降順・作成日時降順で返す。
// exam_results と exam_type を Preload する（report DTO 組み立てに必要）。
// clinic scope は WHERE exams.clinic_id = ? で保証する（exam_results は clinic_id を持たないため JOIN で隔離）。
func (r *examinationRepository) FindByJobID(ctx context.Context, clinicID uint64, jobID uuid.UUID) ([]*model.Examination, error) {
	var exams []*model.Examination
	err := persistence.DBOrTx(ctx, r.db).
		Where("exams.clinic_id = ? AND exams.job_id = ?", clinicID, jobID).
		Scopes(examinationPatientRelationsScope(clinicID)).
		Preload("ExaminationType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Items").
		Order("exams.date DESC, exams.created_at DESC").
		Find(&exams).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "exam", fmt.Sprintf("job_id=%s", jobID))
	}
	return exams, nil
}

// Create は persistence.DBOrTx(ctx, r.db) で ambient tx に参加する（BE-refactor.md X-11）。
// LockByIDForUpdate の行ロック保持 tx 内から呼ばれた場合、別コネクションで INSERT すると
// examinations.medical_record_id の FK 制約チェックが同一行への FOR UPDATE ロックと
// デッドロックする（FK チェックは FOR KEY SHARE を要求し FOR UPDATE と競合するため）。
func (r *examinationRepository) Create(ctx context.Context, exam *model.Examination) error {
	err := persistence.DBOrTx(ctx, r.db).Create(exam).Error
	if err != nil {
		return apperrors.FromGORM(err, "exam", "")
	}
	return nil
}

// Update は persistence.DBOrTx(ctx, r.db) で ambient tx に参加する（Create と同じ理由、BE-refactor.md X-11）。
func (r *examinationRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Examination, error) {
	if err := persistence.UpdateScopedByID(ctx, persistence.DBOrTx(ctx, r.db), &model.Examination{}, "exam", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

// Delete は persistence.DBOrTx(ctx, r.db) で ambient tx に参加する（Create/Update と同じ理由、BE-refactor.md H-8d）。
func (r *examinationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).
		Delete(&model.Examination{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "exam", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("exam", fmt.Sprintf("%d", id))
	}
	return nil
}

func examinationReadPreloads(clinicID uint64) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.
			Preload("ExaminationType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
			Preload("Pet", "clinic_id = ? AND deleted_at IS NULL", clinicID).
			Preload("Pet.Owner", "clinic_id = ? AND deleted_at IS NULL", clinicID).
			Preload("Doctor", "deleted_at IS NULL AND is_active = TRUE AND staff_type = ?", model.StaffTypeDoctor).
			Preload("Items")
	}
}

// examinationPatientRelationsScope excludes polluted rows before their raw foreign IDs
// reach an HTTP response. Required master data and every non-nil patient relation must
// resolve inside the examination clinic, and a linked record must describe the same pet.
func examinationPatientRelationsScope(clinicID uint64) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(`
			EXISTS (
				SELECT 1 FROM exam_types scoped_exam_type
				WHERE scoped_exam_type.id = exams.exam_type_id
				  AND scoped_exam_type.clinic_id = ?
				  AND scoped_exam_type.deleted_at IS NULL
			)
			AND (
				exams.pet_id IS NULL OR EXISTS (
					SELECT 1 FROM pets scoped_pet
					JOIN owners scoped_pet_owner
					  ON scoped_pet_owner.id = scoped_pet.owner_id
					 AND scoped_pet_owner.clinic_id = scoped_pet.clinic_id
					 AND scoped_pet_owner.deleted_at IS NULL
					WHERE scoped_pet.id = exams.pet_id
					  AND scoped_pet.clinic_id = ?
					  AND scoped_pet.deleted_at IS NULL
				)
			)
			AND (
				exams.medical_record_id IS NULL OR EXISTS (
					SELECT 1 FROM medical_records scoped_record
					WHERE scoped_record.id = exams.medical_record_id
					  AND scoped_record.clinic_id = ?
					  AND scoped_record.deleted_at IS NULL
					  AND (
						scoped_record.owner_id IS NULL OR EXISTS (
							SELECT 1 FROM owners scoped_record_owner
							WHERE scoped_record_owner.id = scoped_record.owner_id
							  AND scoped_record_owner.clinic_id = ?
							  AND scoped_record_owner.deleted_at IS NULL
						)
					  )
					  AND (
						scoped_record.pet_id IS NULL OR EXISTS (
							SELECT 1 FROM pets scoped_record_pet
							JOIN owners scoped_record_pet_owner
							  ON scoped_record_pet_owner.id = scoped_record_pet.owner_id
							 AND scoped_record_pet_owner.clinic_id = scoped_record_pet.clinic_id
							 AND scoped_record_pet_owner.deleted_at IS NULL
							WHERE scoped_record_pet.id = scoped_record.pet_id
							  AND scoped_record_pet.clinic_id = ?
							  AND scoped_record_pet.deleted_at IS NULL
						)
					  )
					  AND (
						exams.pet_id IS NULL OR
						scoped_record.pet_id = exams.pet_id
					  )
				)
			)
			AND (
				exams.doctor_id IS NULL OR EXISTS (
					SELECT 1 FROM staff_clinic_assignments scoped_doctor_assignment
					JOIN staffs scoped_doctor
					  ON scoped_doctor.id = scoped_doctor_assignment.staff_id
					 AND scoped_doctor.deleted_at IS NULL
					 AND scoped_doctor.is_active = TRUE
					 AND scoped_doctor.staff_type = ?
					WHERE scoped_doctor_assignment.staff_id = exams.doctor_id
					  AND scoped_doctor_assignment.clinic_id = ?
					  AND scoped_doctor_assignment.deleted_at IS NULL
				)
			)
		`, clinicID, clinicID, clinicID, clinicID, clinicID, model.StaffTypeDoctor, clinicID)
	}
}

func (r *examinationRepository) CountItemsByExamID(ctx context.Context, clinicID, examID uint64) (int64, error) {
	var count int64
	err := persistence.DBOrTx(ctx, r.db).
		Model(&model.ExamResult{}).
		Joins("JOIN exams ON exam_results.exam_id = exams.id AND exams.deleted_at IS NULL").
		Where("exams.clinic_id = ? AND exam_results.exam_id = ? ", clinicID, examID).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "exam_item", fmt.Sprintf("exam_id=%d", examID))
	}
	return count, nil
}

// FindAllItemsByExamID は exam_results を sort_order ASC で取得する。
// clinic_id 隔離は exams を JOIN して保証する（exam_results 自体は clinic_id を持たない）。
// dbOrTx: ambient tx（Transactor.WithTx）内から呼ばれた場合は同一 tx で読む（BE-refactor.md R1-2:
// ReplaceItemsByExamID 置換直後の read-your-writes に必須。tx 外呼び出し時は base db＝従来挙動）。
func (r *examinationRepository) FindAllItemsByExamID(ctx context.Context, clinicID, examID uint64) ([]model.ExamResult, error) {
	items := make([]model.ExamResult, 0)
	err := persistence.DBOrTx(ctx, r.db).
		Model(&model.ExamResult{}).
		Joins("JOIN exams ON exam_results.exam_id = exams.id AND exams.deleted_at IS NULL").
		Where("exams.clinic_id = ? AND exam_results.exam_id = ?", clinicID, examID).
		Order("exam_results.sort_order ASC, exam_results.id ASC").
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "exam_item", fmt.Sprintf("exam_id=%d", examID))
	}
	return items, nil
}

// ReplaceItemsByExamID は exam_results を一括置換する（既存全削除→一括挿入をトランザクション内で実行）。
// 親 exam は置換トランザクション内で clinic_id を再検証し、FOR UPDATE で固定する。
// items の ExamID は本メソッド内で examID に強制上書きする（呼び出し側の指定ミスを防ぐ）。
//
// dbOrTx: ambient tx（Transactor.WithTx）内から呼ばれた場合は同一 tx に join し、.Transaction は
// savepoint（ネスト）として実行される。これにより削除+挿入が caller の tx に入り、後続の監査書込が
// 失敗して caller が tx を rollback すると削除も巻き戻る（BE-refactor.md R1-2・#211 fail-closed
// 原子性と同方針）。ambient tx が無い場合は base db で独立トランザクションを開く（従来挙動＝後方互換）。
// 第 2 戻り値 deletedCount は DELETE の RowsAffected — サービス層の監査ゲートに使う。
func (r *examinationRepository) ReplaceItemsByExamID(ctx context.Context, clinicID, examID uint64, items []model.ExamResult) ([]model.ExamResult, int64, error) {
	var deletedCount int64
	err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		// サービスを経由しない取込経路もあるため、親 exam を行ロックして
		// status/move/delete と結果置換を直列化する。
		var exam model.Examination
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", examID, clinicID).
			First(&exam).Error; err != nil {
			return apperrors.FromGORM(err, "exam", fmt.Sprintf("%d", examID))
		}
		if exam.Status == model.ExaminationStatusConfirmed {
			return apperrors.WrapConflict("確定済みの検査は編集できません")
		}

		// 既存 items を全削除（exam_results は CASCADE では消えないため明示削除）
		del := tx.Where("exam_id = ?", examID).Delete(&model.ExamResult{})
		if del.Error != nil {
			return apperrors.FromGORM(del.Error, "exam_item", fmt.Sprintf("exam_id=%d", examID))
		}
		deletedCount = del.RowsAffected

		if len(items) == 0 {
			return nil
		}

		// ExamID を強制上書きして一括挿入
		for i := range items {
			items[i].ExamID = examID
			items[i].ID = 0 // 新規挿入のため auto increment に任せる
		}
		if err := tx.Create(&items).Error; err != nil {
			return apperrors.FromGORM(err, "exam_item", fmt.Sprintf("exam_id=%d", examID))
		}
		return nil
	})
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to replace exam items")
	}
	// 置換後の最終状態を読み直す。ctx は ambient tx を保持するため dbOrTx で同一 tx から読み、
	// 削除+挿入後の結果を read-your-writes で返す（tx 外呼び出し時は base db＝従来挙動）。
	saved, err := r.FindAllItemsByExamID(ctx, clinicID, examID)
	if err != nil {
		return nil, 0, err
	}
	return saved, deletedCount, nil
}
