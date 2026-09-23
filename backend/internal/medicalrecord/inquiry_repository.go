package medicalrecord

// inquiry_repository.go — InquiryRepository の実装。
// Moved from internal/repository/inquiry_repository.go — BE9-2D roll-up. Type/constructor names
// are unchanged; the internal/repository facade re-exports them as aliases so no caller changes.
// Package-private clinicScope is swapped for persistence.ClinicScope (this package must not
// import internal/repository — that would be an import cycle via the facade).

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// InquiryUpsertFields は SaveByMedicalRecordID への部分更新入力。
// nil ポインタのフィールドは「未送信」を意味し既存列値を保持する（PATCH 意味論）。
// ChiefComplaintTypeID は **uint64: nil=未送信(保持) / &nil=明示 JSON null(NULL 書込) / &&v=値セット。
// EMR-87: 以前は *model.Inquiry を受け全列を無条件に書き込んでいたため、
// 未送信フィールドがゼロ値で既存値を上書きしていた（区分のみ解除すると主訴本文が消える不具合）。
type InquiryUpsertFields struct {
	MedicalRecordID      uint64
	ChiefComplaint       *string
	Notes                *string
	ChiefComplaintTypeID **uint64
}

// InquiryRepository は医療記録問診の永続化インターフェース
type InquiryRepository interface {
	SaveByMedicalRecordID(ctx context.Context, clinicID uint64, fields InquiryUpsertFields) (*model.Inquiry, error)
}

type inquiryRepository struct {
	db *gorm.DB
}

// NewInquiryRepository は InquiryRepository を生成する
func NewInquiryRepository(db *gorm.DB) InquiryRepository {
	return &inquiryRepository{db: db}
}

// SaveByMedicalRecordID は medical_record_id に対応する Inquiry を upsert する。
// レコードが存在しない場合は INSERT、存在する場合は UPDATE する。
// clinicID により clinic 境界を検証する。
//
// BUG-079 修正: FirstOrCreate+Assign に同一ポインタを渡すと既存レコード取得後の
// Assign が無効になるため、FirstOrCreate でレコードを確保してから
// map[string]any で明示的に Updates する 2 ステップ方式に変更。
//
// BE-refactor.md MRC-12 / phase2.html:195 / X-06:
// 親カルテの FOR UPDATE ロック、FirstOrCreate、Updates を単一 Transaction に収める。
// Conflict（確定済み）時は FirstOrCreate で作った空行も rollback され残らない。
// 本メソッドは明示 Transaction を開くため ambient DBOrTx 参加者にはしない
// （呼び出し側 Transactor と二重接続にならない）。
func (r *inquiryRepository) SaveByMedicalRecordID(ctx context.Context, clinicID uint64, fields InquiryUpsertFields) (*model.Inquiry, error) {
	var refreshed model.Inquiry
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 親カルテを FOR UPDATE で固定し finalize と直列化する（vital/examination と同型）。
		var mr model.MedicalRecord
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Scopes(persistence.ClinicScope(clinicID)).
			Where("id = ?", fields.MedicalRecordID).
			First(&mr).Error; err != nil {
			return apperrors.FromGORM(err, "inquiry", fmt.Sprintf("medical_record_id=%d", fields.MedicalRecordID))
		}
		if mr.Status == model.MedicalRecordStatusFinalized {
			return apperrors.WrapConflict("確定済みカルテの問診は編集できません")
		}

		// Step 1: medical_record_id で既存レコードを取得または新規作成
		var existing model.Inquiry
		if err := tx.
			Where(model.Inquiry{MedicalRecordID: fields.MedicalRecordID}).
			FirstOrCreate(&existing).Error; err != nil {
			return apperrors.FromGORM(err, "inquiry", "")
		}

		// Step 2: 送信されたフィールドのみ map[string]any で明示的に Updates（GORM ゼロ値問題を回避）。
		// 未送信フィールドはキー自体を載せないため既存値を保持する（EMR-87）。
		// medical_records の status='draft' 条件は defense-in-depth。
		updates := map[string]any{}
		if fields.ChiefComplaint != nil {
			updates["chief_complaint"] = *fields.ChiefComplaint
		}
		if fields.Notes != nil {
			updates["notes"] = *fields.Notes
		}
		if fields.ChiefComplaintTypeID != nil {
			updates["chief_complaint_type_id"] = *fields.ChiefComplaintTypeID
		}
		if len(updates) > 0 {
			result := tx.
				Model(&existing).
				Where("medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ? AND status = ? AND deleted_at IS NULL)",
					clinicID, model.MedicalRecordStatusDraft).
				Updates(updates)
			if result.Error != nil {
				return apperrors.FromGORM(result.Error, "inquiry", "")
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapConflict("確定済みカルテの問診は編集できません")
			}
		}

		// 最新状態を同一 tx から取得（updated_at 等の DB 管理フィールドも正確に反映）
		if err := tx.
			Where("id = ?", existing.ID).
			First(&refreshed).Error; err != nil {
			return apperrors.FromGORM(err, "inquiry", fmt.Sprintf("%d", existing.ID))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &refreshed, nil
}
