package repository

import (
	"context"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// InquiryRepository は医療記録問診の永続化インターフェース
type InquiryRepository interface {
	UpsertByMedicalRecordID(ctx context.Context, inquiry *model.Inquiry) (*model.Inquiry, error)
	CountByChiefComplaintCategoryID(ctx context.Context, categoryID uint64) (int64, error)
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
//
// BUG-079 修正: FirstOrCreate+Assign に同一ポインタを渡すと既存レコード取得後の
// Assign が無効になるため、FirstOrCreate でレコードを確保してから
// map[string]any で明示的に Updates する 2 ステップ方式に変更。
func (r *inquiryRepository) UpsertByMedicalRecordID(ctx context.Context, inquiry *model.Inquiry) (*model.Inquiry, error) {
	// Step 1: medical_record_id で既存レコードを取得または新規作成
	var existing model.Inquiry
	if err := r.db.WithContext(ctx).
		Where(model.Inquiry{MedicalRecordID: inquiry.MedicalRecordID}).
		FirstOrCreate(&existing).Error; err != nil {
		return nil, apperrors.Wrap(err, "upsert inquiry: first or create")
	}

	// Step 2: 更新フィールドを map[string]any で明示的に Updates（GORM ゼロ値問題を回避）
	updates := map[string]any{
		"chief_complaint":             inquiry.ChiefComplaint,
		"notes":                       inquiry.Notes,
		"history":                     inquiry.History,
		"current_medications":         inquiry.CurrentMedications,
		"allergy_info":                inquiry.AllergyInfo,
		"last_meal":                   inquiry.LastMeal,
		"last_defecation":             inquiry.LastDefecation,
		"last_urination":              inquiry.LastUrination,
		"owner_observations":          inquiry.OwnerObservations,
		"chief_complaint_category_id": inquiry.ChiefComplaintCategoryID,
		"appetite":                    inquiry.Appetite,
		"water_intake":                inquiry.WaterIntake,
		"staff_id":                    inquiry.StaffID,
	}
	if err := r.db.WithContext(ctx).
		Model(&existing).
		Updates(updates).Error; err != nil {
		return nil, apperrors.Wrap(err, "upsert inquiry: update fields")
	}

	// 最新状態を返す（フィールドをローカル変数に反映）
	existing.ChiefComplaint = inquiry.ChiefComplaint
	existing.Notes = inquiry.Notes
	existing.History = inquiry.History
	existing.CurrentMedications = inquiry.CurrentMedications
	existing.AllergyInfo = inquiry.AllergyInfo
	existing.LastMeal = inquiry.LastMeal
	existing.LastDefecation = inquiry.LastDefecation
	existing.LastUrination = inquiry.LastUrination
	existing.OwnerObservations = inquiry.OwnerObservations
	existing.ChiefComplaintCategoryID = inquiry.ChiefComplaintCategoryID
	existing.Appetite = inquiry.Appetite
	existing.WaterIntake = inquiry.WaterIntake
	existing.StaffID = inquiry.StaffID

	return &existing, nil
}

// CountByChiefComplaintCategoryID は指定カテゴリIDを参照するInquiryの件数を返す。
// Delete の FK チェックに使用する。
func (r *inquiryRepository) CountByChiefComplaintCategoryID(ctx context.Context, categoryID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Inquiry{}).
		Where("chief_complaint_category_id = ?", categoryID).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.Wrap(err, "count inquiries by chief_complaint_category_id")
	}
	return count, nil
}
