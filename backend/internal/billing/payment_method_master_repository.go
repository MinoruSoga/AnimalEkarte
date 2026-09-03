package billing

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
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
	err := persistence.DBOrTx(ctx, r.db).
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
	// clinic_id + id + 非システム + usage 不在を同一 DELETE で要求し、
	// CountUsage→Delete 間の参照追加 TOCTOU を原子的に塞ぐ。
	result := persistence.DBOrTx(ctx, r.db).
		Scopes(clinicScope(clinicID)).
		Where("id = ?", id).
		Where("system_key IS NULL OR system_key = ''").
		Where(`NOT EXISTS (
			SELECT 1 FROM payments
			JOIN billings ON billings.id = payments.billing_id
			  AND billings.clinic_id = ?
			  AND billings.deleted_at IS NULL
			WHERE payments.payment_method_id = payment_methods.id
			  AND payments.deleted_at IS NULL
		)`, clinicID).
		Delete(&model.PaymentMethodMaster{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "payment_method", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return r.normalizeDeleteIfUnusedMiss(ctx, clinicID, id)
	}
	return nil
}

// normalizeDeleteIfUnusedMiss は原子 DELETE が 0 行だった理由を再読取で区別する。
func (r *paymentMethodMasterRepository) normalizeDeleteIfUnusedMiss(ctx context.Context, clinicID, id uint64) error {
	existing, err := r.FindByID(ctx, clinicID, id)
	if err != nil {
		return err
	}
	if isSystemPaymentMethod(existing) {
		return apperrors.WrapConflict("システム標準の支払方法は削除できません")
	}
	count, err := r.CountUsageByPaymentMethodID(ctx, clinicID, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return apperrors.WrapConflict("この支払方法は使用中のため削除できません")
	}
	return apperrors.WrapConflict("この支払方法は使用中のため削除できません")
}

// CountUsageByPaymentMethodID は指定した支払方法を参照している payments の件数を返す。
// payments テーブルに直接 clinic_id がないため billings を JOIN してテナント分離する。
func (r *paymentMethodMasterRepository) CountUsageByPaymentMethodID(ctx context.Context, clinicID, id uint64) (int64, error) {
	var count int64
	err := persistence.DBOrTx(ctx, r.db).
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
