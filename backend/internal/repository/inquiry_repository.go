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
func (r *inquiryRepository) UpsertByMedicalRecordID(ctx context.Context, inquiry *model.Inquiry) (*model.Inquiry, error) {
	result := r.db.WithContext(ctx).
		Where(model.Inquiry{MedicalRecordID: inquiry.MedicalRecordID}).
		Assign(inquiry).
		FirstOrCreate(inquiry)
	if result.Error != nil {
		return nil, apperrors.Wrap(result.Error, "upsert inquiry")
	}
	return inquiry, nil
}
