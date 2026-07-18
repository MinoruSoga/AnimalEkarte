package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// BillingConfirmationRepository は会計医師確認のデータアクセスインターフェース
type BillingConfirmationRepository interface {
	FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) (*model.BillingConfirmation, error)
	Create(ctx context.Context, review *model.BillingConfirmation) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
}

type billingConfirmationRepository struct {
	db *gorm.DB
}

// NewBillingConfirmationRepository はBillingConfirmationRepositoryを初期化して返す
func NewBillingConfirmationRepository(db *gorm.DB) BillingConfirmationRepository {
	return &billingConfirmationRepository{db: db}
}

func (r *billingConfirmationRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) (*model.BillingConfirmation, error) {
	var review model.BillingConfirmation
	err := r.db.WithContext(ctx).
		Joins("JOIN medical_records ON medical_records.id = billing_confirmations.medical_record_id AND medical_records.deleted_at IS NULL").
		Where("medical_records.clinic_id = ? AND billing_confirmations.medical_record_id = ?", clinicID, medicalRecordID).
		First(&review).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "billing_confirmation", fmt.Sprintf("medical_record_id=%d", medicalRecordID))
	}
	return &review, nil
}

func (r *billingConfirmationRepository) Create(ctx context.Context, review *model.BillingConfirmation) error {
	if err := r.db.WithContext(ctx).Create(review).Error; err != nil {
		return apperrors.FromGORM(err, "billing_confirmation", "")
	}
	return nil
}

// Update は dbOrTx(ctx, r.db) で ambient tx に参加する（billingConfirmationService.Confirm/Return が
// SD-2/X-11 と同種の確定済みカルテガードのため LockByIDForUpdate の行ロックと同一 tx に束ねる）。
func (r *billingConfirmationRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	// Restrict update to rows belonging to this clinic via subquery on medical_records
	result := dbOrTx(ctx, r.db).
		Model(&model.BillingConfirmation{}).
		Where("id = ? AND medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ?)", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "billing_confirmation", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("billing_confirmation", fmt.Sprintf("%d", id))
	}
	return nil
}
