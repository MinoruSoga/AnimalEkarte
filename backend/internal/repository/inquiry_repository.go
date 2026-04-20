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
	UpsertByMedicalRecordID(ctx context.Context, clinicID uint64, inquiry *model.Inquiry) (*model.Inquiry, error)
	CountByChiefComplaintTypeID(ctx context.Context, clinicID, categoryID uint64) (int64, error)
}

type inquiryRepository struct {
	db *gorm.DB
}

// NewInquiryRepository は InquiryRepository を生成する
func NewInquiryRepository(db *gorm.DB) InquiryRepository {
	return &inquiryRepository{db: db}
}

// UpsertByMedicalRecordID は medical_record_id に対応する Inquiry を upsert する。
// レコードが存在しない場合は INSERT、存在する場合は UPDATE する。
// clinicID により clinic 境界を検証する。
//
// BUG-079 修正: FirstOrCreate+Assign に同一ポインタを渡すと既存レコード取得後の
// Assign が無効になるため、FirstOrCreate でレコードを確保してから
// map[string]any で明示的に Updates する 2 ステップ方式に変更。
func (r *inquiryRepository) UpsertByMedicalRecordID(ctx context.Context, clinicID uint64, inquiry *model.Inquiry) (*model.Inquiry, error) {
	// Verify the medical_record belongs to this clinic before upserting
	var mrCount int64
	if err := r.db.WithContext(ctx).
		Table("medical_records").
		Scopes(clinicScope(clinicID)).Where("id = ?", inquiry.MedicalRecordID).
		Count(&mrCount).Error; err != nil {
		return nil, apperrors.FromGORM(err, "inquiry", "")
	}
	if mrCount == 0 {
		return nil, apperrors.WrapNotFound("medical_record", fmt.Sprintf("%d", inquiry.MedicalRecordID))
	}

	// Step 1: medical_record_id で既存レコードを取得または新規作成
	var existing model.Inquiry
	if err := r.db.WithContext(ctx).
		Where(model.Inquiry{MedicalRecordID: inquiry.MedicalRecordID}).
		FirstOrCreate(&existing).Error; err != nil {
		return nil, apperrors.FromGORM(err, "inquiry", "")
	}

	// Step 2: 更新フィールドを map[string]any で明示的に Updates（GORM ゼロ値問題を回避）
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
	if err := r.db.WithContext(ctx).
		Model(&existing).
		Updates(updates).Error; err != nil {
		return nil, apperrors.FromGORM(err, "inquiry", "")
	}

	// 最新状態を DB から取得（updated_at 等の DB 管理フィールドも正確に反映）
	var refreshed model.Inquiry
	if err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Where("id = ?", existing.ID).
		First(&refreshed).Error; err != nil {
		return nil, apperrors.FromGORM(err, "inquiry", fmt.Sprintf("%d", existing.ID))
	}
	return &refreshed, nil
}

// CountByChiefComplaintTypeID は指定クリニック・カテゴリIDを参照するInquiryの件数を返す。
// Delete の FK チェックに使用する。clinic_id フィルタにより他クリニックのレコードを誤検知しない。
func (r *inquiryRepository) CountByChiefComplaintTypeID(ctx context.Context, clinicID, categoryID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Inquiry{}).
		Scopes(clinicScope(clinicID)).
		Where("chief_complaint_type_id = ?", categoryID).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "inquiry", "")
	}
	return count, nil
}
