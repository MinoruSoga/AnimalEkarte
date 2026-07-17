package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// InquiryRepository は医療記録問診の永続化インターフェース
type InquiryRepository interface {
	SaveByMedicalRecordID(ctx context.Context, clinicID uint64, inquiry *model.Inquiry) (*model.Inquiry, error)
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
// SD-2 系ガード監査で発見された欠落: 親カルテが確定済みの場合は問診の作成/更新を拒否する
// (BE-refactor.md X-11 と同種の不変条件)。inquiryService は Transactor を持たない設計上の制約
// (cross_tenant_master_fk_write_test.go の既存呼び出しが repository.Transactor 無しの
// 旧コンストラクタシグネチャに依存しており変更できない) のため、examination/vital 等が使う
// LockByIDForUpdate による行ロックではなく、本メソッド内の事前ステータス確認 + Updates 自体への
// status='draft' WHERE 条件（原子的レース対策）の二段構えで守る。
func (r *inquiryRepository) SaveByMedicalRecordID(ctx context.Context, clinicID uint64, inquiry *model.Inquiry) (*model.Inquiry, error) {
	// Verify the medical_record belongs to this clinic and is still draft before upserting.
	var mr model.MedicalRecord
	if err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Where("id = ?", inquiry.MedicalRecordID).
		First(&mr).Error; err != nil {
		return nil, apperrors.FromGORM(err, "inquiry", fmt.Sprintf("medical_record_id=%d", inquiry.MedicalRecordID))
	}
	if mr.Status == model.MedicalRecordStatusFinalized {
		return nil, apperrors.WrapConflict("確定済みカルテの問診は編集できません")
	}

	// Step 1: medical_record_id で既存レコードを取得または新規作成
	var existing model.Inquiry
	if err := r.db.WithContext(ctx).
		Where(model.Inquiry{MedicalRecordID: inquiry.MedicalRecordID}).
		FirstOrCreate(&existing).Error; err != nil {
		return nil, apperrors.FromGORM(err, "inquiry", "")
	}

	// Step 2: 更新フィールドを map[string]any で明示的に Updates（GORM ゼロ値問題を回避）。
	// medical_records の status='draft' 条件を Where に含め、直前の事前チェックと本 UPDATE の
	// 間で親カルテが確定した場合でも原子的に 0 行更新（Conflict）とする。
	updates := map[string]any{
		"chief_complaint":         inquiry.ChiefComplaint,
		"notes":                   inquiry.Notes,
		"history":                 inquiry.History,
		"current_medications":     inquiry.CurrentMedications,
		"allergy_info":            inquiry.AllergyInfo,
		"last_meal":               inquiry.LastMeal,
		"last_defecation":         inquiry.LastDefecation,
		"last_urination":          inquiry.LastUrination,
		"owner_observations":      inquiry.OwnerObservations,
		"chief_complaint_type_id": inquiry.ChiefComplaintTypeID,
		"appetite":                inquiry.Appetite,
		"water_intake":            inquiry.WaterIntake,
		"staff_id":                inquiry.StaffID,
	}
	result := r.db.WithContext(ctx).
		Model(&existing).
		Where("medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ? AND status = ? AND deleted_at IS NULL)",
			clinicID, model.MedicalRecordStatusDraft).
		Updates(updates)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "inquiry", "")
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapConflict("確定済みカルテの問診は編集できません")
	}

	// 最新状態を DB から取得（updated_at 等の DB 管理フィールドも正確に反映）
	var refreshed model.Inquiry
	if err := r.db.WithContext(ctx).
		Where("id = ?", existing.ID).
		First(&refreshed).Error; err != nil {
		return nil, apperrors.FromGORM(err, "inquiry", fmt.Sprintf("%d", existing.ID))
	}
	return &refreshed, nil
}
