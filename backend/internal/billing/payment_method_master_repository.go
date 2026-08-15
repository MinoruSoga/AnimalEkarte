package billing

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// PaymentMethodMasterRepository は支払方法マスタのデータアクセスインターフェース
type PaymentMethodMasterRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.PaymentMethodMaster, error)
	Create(ctx context.Context, m *model.PaymentMethodMaster) (*model.PaymentMethodMaster, error)
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.PaymentMethodMaster, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	CountUsageByPaymentMethodID(ctx context.Context, clinicID, id uint64) (int64, error)
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type paymentMethodMasterRepository struct{ db *gorm.DB }

// New は PaymentMethodMasterRepository を初期化して返す
func NewPaymentMethodMasterRepository(db *gorm.DB) PaymentMethodMasterRepository {
	return &paymentMethodMasterRepository{db: db}
}

func (r *paymentMethodMasterRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error) {
	var ms []model.PaymentMethodMaster
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Order("display_order ASC, name ASC").
		Find(&ms).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "payment_method", "")
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

func (r *paymentMethodMasterRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.PaymentMethodMaster, error) {
	if err := updateScopedByID(ctx, r.db, &model.PaymentMethodMaster{}, "payment_method", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *paymentMethodMasterRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return deleteScopedByID(ctx, r.db, &model.PaymentMethodMaster{}, "payment_method", clinicID, id)
}

// CountUsageByPaymentMethodID は指定した支払方法を参照している payments の件数を返す。
// payments テーブルに直接 clinic_id がないため billings を JOIN してテナント分離する。
func (r *paymentMethodMasterRepository) CountUsageByPaymentMethodID(ctx context.Context, clinicID, id uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Payment{}).
		Joins("JOIN billings ON billings.id = payments.billing_id AND billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
		Where("payments.payment_method_id = ? AND payments.deleted_at IS NULL", id).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "payment_method", fmt.Sprintf("%d", id))
	}
	return count, nil
}

func (r *paymentMethodMasterRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return reorderByClinicID(ctx, r.db, &model.PaymentMethodMaster{}, "payment_method", clinicID, ids, "display_order")
}
