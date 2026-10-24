package medicalrecord

// Moved from internal/repository (BE9-2D ⑦ Batch A). 旧 package-private helper は repohelpers
// 同等物へ置換（同一述語/ambient-tx参加）。外部は internal/repository の facade alias 経由で不変。

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

type MedicalRecordAddendumRepository interface {
	Create(ctx context.Context, addendum *model.MedicalRecordAddendum) error
	FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecordAddendum, error)
	FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]*model.MedicalRecordAddendum, error)
	// LatestConsultationEventKind は同 clinic・同カルテの最大 revision の診療終了イベント種別を返す
	// （EMR-495 Q2: 訂正追記は最新イベントが ended のときだけ許可する）。履歴なしは found=false
	// （過去未終了と断定せず、ended とも扱わない）。ambient tx に参加する。
	LatestConsultationEventKind(ctx context.Context, clinicID, medicalRecordID uint64) (kind string, found bool, err error)
}

type medicalRecordAddendumRepository struct {
	db *gorm.DB
}

func NewMedicalRecordAddendumRepository(db *gorm.DB) MedicalRecordAddendumRepository {
	return &medicalRecordAddendumRepository{db: db}
}

func (r *medicalRecordAddendumRepository) Create(ctx context.Context, addendum *model.MedicalRecordAddendum) error {
	if err := persistence.DBOrTx(ctx, r.db).Create(addendum).Error; err != nil {
		return apperrors.FromGORM(err, "medical_record_addendum", "")
	}
	return nil
}

func (r *medicalRecordAddendumRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecordAddendum, error) {
	var addendum model.MedicalRecordAddendum
	err := persistence.DBOrTx(ctx, r.db).
		Joins("JOIN medical_records ON medical_records.id = medical_record_addenda.medical_record_id AND medical_records.clinic_id = medical_record_addenda.clinic_id AND medical_records.deleted_at IS NULL").
		Where("medical_record_addenda.clinic_id = ? AND medical_record_addenda.id = ?", clinicID, id).
		First(&addendum).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "medical_record_addendum", fmt.Sprintf("%d", id))
	}
	return &addendum, nil
}

func (r *medicalRecordAddendumRepository) LatestConsultationEventKind(ctx context.Context, clinicID, medicalRecordID uint64) (string, bool, error) {
	var event model.MedicalRecordConsultationEvent
	err := persistence.DBOrTx(ctx, r.db).
		Select("event_kind").
		Where("clinic_id = ? AND medical_record_id = ?", clinicID, medicalRecordID).
		Order("revision DESC").
		Take(&event).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", false, nil
		}
		return "", false, apperrors.FromGORM(err, "medical_record_consultation_event", fmt.Sprintf("medical_record_id=%d", medicalRecordID))
	}
	return event.EventKind, true, nil
}

func (r *medicalRecordAddendumRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]*model.MedicalRecordAddendum, error) {
	var addenda []*model.MedicalRecordAddendum
	err := persistence.DBOrTx(ctx, r.db).
		Preload("AuthorStaff", medicalRecordAddendumAuthorStaffPreload(clinicID)).
		Joins("JOIN medical_records ON medical_records.id = medical_record_addenda.medical_record_id AND medical_records.clinic_id = medical_record_addenda.clinic_id AND medical_records.deleted_at IS NULL").
		Where("medical_record_addenda.clinic_id = ? AND medical_record_addenda.medical_record_id = ?", clinicID, medicalRecordID).
		Order("medical_record_addenda.created_at ASC").
		Find(&addenda).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "medical_record_addendum", fmt.Sprintf("medical_record_id=%d", medicalRecordID))
	}
	return addenda, nil
}

// medicalRecordAddendumAuthorStaffPreload は追記者を「親カルテの clinic_id」スコープで
// 解決するバッチ preload（行ごとのクエリは発行しない）。一覧の全追記は親カルテと同じ
// clinic_id を持つため、引数 clinicID がそのまま親カルテの医院を表す。
//   - staffs.id = medical_record_addenda.author_user_id（foreignKey 既定）
//   - staffs.deleted_at IS NULL
//   - 有効な staff_clinic_assignments 行が親カルテの clinic_id に存在する
//
// 医院所属の正本は staff_clinic_assignments（認証側の current_access_staff_reader と
// scopedBillingStaffPreload と同じ所属相関 predicate）。主医院が他院でも兼務所属が
// あれば著者として解決され、削除・異動（所属解除）・他医院のみの追記者は解決されず
// AuthorStaff=nil → 応答側で author_user_id フォールバック（他医院スタッフ名は
// 絶対に漏らさない）。
// is_active は掛けない: 退職後も監査上の帰属名は残す（scopedBillingStaffPreload と同じ
// 判断。clinic 非スコープの medicalRecordEnteredByStaffPreload は真似しない）。
func medicalRecordAddendumAuthorStaffPreload(clinicID uint64) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.
			Where("staffs.deleted_at IS NULL").
			Where(`EXISTS (
				SELECT 1
				FROM staff_clinic_assignments AS addendum_author_assignment
				WHERE addendum_author_assignment.staff_id = staffs.id
				  AND addendum_author_assignment.clinic_id = ?
				  AND addendum_author_assignment.deleted_at IS NULL
			)`, clinicID)
	}
}
