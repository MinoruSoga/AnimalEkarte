package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// BillingReviewRepository は会計医師確認のデータアクセスインターフェース
type BillingReviewRepository interface {
	FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) (*model.BillingReview, error)
	Create(ctx context.Context, review *model.BillingReview) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
}

type billingReviewRepository struct {
	db *gorm.DB
}

// NewBillingReviewRepository はBillingReviewRepositoryを初期化して返す
func NewBillingReviewRepository(db *gorm.DB) BillingReviewRepository {
	return &billingReviewRepository{db: db}
}

func (r *billingReviewRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) (*model.BillingReview, error) {
	var review model.BillingReview
	err := r.db.WithContext(ctx).
		Joins("JOIN medical_records ON medical_records.id = billing_reviews.medical_record_id").
		Where("medical_records.clinic_id = ? AND billing_reviews.medical_record_id = ?", clinicID, medicalRecordID).
		First(&review).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "billing_review", fmt.Sprintf("medical_record_id=%d", medicalRecordID))
	}
	return &review, nil
}

func (r *billingReviewRepository) Create(ctx context.Context, review *model.BillingReview) error {
	if err := r.db.WithContext(ctx).Create(review).Error; err != nil {
		return apperrors.FromGORM(err, "billing_review", "")
	}
	return nil
}

func (r *billingReviewRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	// Restrict update to rows belonging to this clinic via subquery on medical_records
	result := r.db.WithContext(ctx).
		Model(&model.BillingReview{}).
		Where("id = ? AND medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ?)", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "billing_review", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("billing_review", fmt.Sprintf("%d", id))
	}
	return nil
}
