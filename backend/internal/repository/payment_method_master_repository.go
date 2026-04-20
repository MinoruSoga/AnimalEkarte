package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// PaymentMethodMasterRepository は支払方法マスタのデータアクセスインターフェース
type PaymentMethodMasterRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.PaymentMethodMaster, error)
	Create(ctx context.Context, m *model.PaymentMethodMaster) (*model.PaymentMethodMaster, error)
	UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.PaymentMethodMaster, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	CountUsageByID(ctx context.Context, id uint64) (int64, error)
}

type paymentMethodMasterRepository struct{ db *gorm.DB }

// NewPaymentMethodMasterRepository は PaymentMethodMasterRepository を初期化して返す
func NewPaymentMethodMasterRepository(db *gorm.DB) PaymentMethodMasterRepository {
	return &paymentMethodMasterRepository{db: db}
}

func (r *paymentMethodMasterRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error) {
	var ms []model.PaymentMethodMaster
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Order("display_order ASC, id ASC").
		Find(&ms).Error
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find payment methods")
	}
	return ms, nil
}

func (r *paymentMethodMasterRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.PaymentMethodMaster, error) {
	var m model.PaymentMethodMaster
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		First(&m, id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "payment_method", fmt.Sprintf("%d", id))
	}
	return &m, nil
}

func (r *paymentMethodMasterRepository) Create(ctx context.Context, m *model.PaymentMethodMaster) (*model.PaymentMethodMaster, error) {
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, apperrors.FromGORM(err, "payment_method", "")
	}
	return m, nil
}

func (r *paymentMethodMasterRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.PaymentMethodMaster, error) {
	result := r.db.WithContext(ctx).
		Model(&model.PaymentMethodMaster{}).
		Scopes(clinicScope(clinicID)).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "payment_method", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("payment_method", fmt.Sprintf("%d", id))
	}
	var m model.PaymentMethodMaster
	if err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		First(&m, id).Error; err != nil {
		return nil, apperrors.FromGORM(err, "payment_method", fmt.Sprintf("%d", id))
	}
	return &m, nil
}

func (r *paymentMethodMasterRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Delete(&model.PaymentMethodMaster{}, id)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "payment_method", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("payment_method", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *paymentMethodMasterRepository) CountUsageByID(ctx context.Context, id uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Payment{}).
		Where("payment_method_id = ?", id).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.Wrap(err, "failed to count payment method usage")
	}
	return count, nil
}
