package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// BillingReviewRepository は会計医師確認のデータアクセスインターフェース
type BillingReviewRepository interface {
	FindByMedicalRecordID(ctx context.Context, medicalRecordID uint64) (*model.BillingReview, error)
	Create(ctx context.Context, review *model.BillingReview) error
	Update(ctx context.Context, id uint64, fields map[string]any) error
}

type billingReviewRepository struct {
	db *gorm.DB
}

// NewBillingReviewRepository はBillingReviewRepositoryを初期化して返す
func NewBillingReviewRepository(db *gorm.DB) BillingReviewRepository {
	return &billingReviewRepository{db: db}
}

func (r *billingReviewRepository) FindByMedicalRecordID(ctx context.Context, medicalRecordID uint64) (*model.BillingReview, error) {
	var review model.BillingReview
	if err := r.db.WithContext(ctx).
		Where("medical_record_id = ?", medicalRecordID).
		First(&review).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("billing_review", fmt.Sprintf("medical_record_id=%d", medicalRecordID))
		}
		return nil, apperrors.Wrap(err, "find billing_review by medical_record_id")
	}
	return &review, nil
}

func (r *billingReviewRepository) Create(ctx context.Context, review *model.BillingReview) error {
	if err := r.db.WithContext(ctx).Create(review).Error; err != nil {
		return apperrors.Wrap(err, "create billing_review")
	}
	return nil
}

func (r *billingReviewRepository) Update(ctx context.Context, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.BillingReview{}).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update billing_review")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("billing_review", fmt.Sprintf("%d", id))
	}
	return nil
}
