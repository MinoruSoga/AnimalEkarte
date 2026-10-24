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
	Update(ctx context.Context, clinicID, id uint64, cmd UpdatePaymentMethodMasterInput) (*model.PaymentMethodMaster, error)
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

func (r *paymentMethodMasterRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdatePaymentMethodMasterInput) (*model.PaymentMethodMaster, error) {
	var loaded *model.PaymentMethodMaster
	err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		if err := r.update(txCtx, clinicID, id, buildPaymentMethodUpdate(&cmd)); err != nil {
			return err
		}
		reloaded, err := r.FindByID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "reload payment method after update")
		}
		loaded = reloaded
		return nil
	})
	if err != nil {
		return nil, err
	}
	return loaded, nil
}

func (r *paymentMethodMasterRepository) update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return updateScopedByID(ctx, persistence.DBOrTx(ctx, r.db), &model.PaymentMethodMaster{}, "payment_method", clinicID, id, fields)
}

// paymentMethodInUseSQL は支払方法マスタ（payment_methods）を参照する保存先の存在判定（5 FK 参照元）。
// 受領 (payment_funds)、実返金 (fund_consumption_events)、受領記録の訂正 (fund_record_corrections)、
// 預り金 (owner_deposit_entries)、財務 source event (finance_source_events) のいずれかが参照していれば使用中。
// いずれも payments への RESTRICT 参照だが、マスタの削除は論理削除で RESTRICT を迂回するため、
// 1 つでも参照があれば削除を拒否する。参照元の clinic_id を親行に相関させる。held for EMR-506。
const paymentMethodInUseSQL = `EXISTS (SELECT 1 FROM payment_funds r WHERE r.clinic_id = payment_methods.clinic_id AND r.payment_method_id = payment_methods.id)
	OR EXISTS (SELECT 1 FROM fund_consumption_events r WHERE r.clinic_id = payment_methods.clinic_id AND r.actual_refund_method_id = payment_methods.id)
	OR EXISTS (SELECT 1 FROM fund_record_corrections r WHERE r.clinic_id = payment_methods.clinic_id AND r.corrected_method_id = payment_methods.id)
	OR EXISTS (SELECT 1 FROM owner_deposit_entries r WHERE r.clinic_id = payment_methods.clinic_id AND r.payment_method_id = payment_methods.id)
	OR EXISTS (SELECT 1 FROM finance_source_events r WHERE r.clinic_id = payment_methods.clinic_id AND r.actual_method_id = payment_methods.id)`

// countPaymentMethodUsageSQL は paymentMethodInUseSQL と同じ 5 参照元の件数合計（引数は clinic_id, 支払方法 id の組 × 5）。
const countPaymentMethodUsageSQL = `SELECT
	(SELECT count(*) FROM payment_funds WHERE clinic_id = ? AND payment_method_id = ?)
	+ (SELECT count(*) FROM fund_consumption_events WHERE clinic_id = ? AND actual_refund_method_id = ?)
	+ (SELECT count(*) FROM fund_record_corrections WHERE clinic_id = ? AND corrected_method_id = ?)
	+ (SELECT count(*) FROM owner_deposit_entries WHERE clinic_id = ? AND payment_method_id = ?)
	+ (SELECT count(*) FROM finance_source_events WHERE clinic_id = ? AND actual_method_id = ?)`

func (r *paymentMethodMasterRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	// clinic_id + id + 非システム + usage 不在を同一 DELETE で要求し、
	// CountUsage→Delete 間の参照追加 TOCTOU を原子的に塞ぐ。
	result := persistence.DBOrTx(ctx, r.db).
		Scopes(clinicScope(clinicID)).
		Where("id = ?", id).
		Where("system_key IS NULL OR system_key = ''").
		Where("NOT (" + paymentMethodInUseSQL + ")").
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

// CountUsageByPaymentMethodID は指定した支払方法を参照している保存先（5 FK 参照元）の件数合計を返す。
// いずれの参照元にも clinic_id があるため、テナント分離は clinic_id の一致で行う。
func (r *paymentMethodMasterRepository) CountUsageByPaymentMethodID(ctx context.Context, clinicID, id uint64) (int64, error) {
	var count int64
	err := persistence.DBOrTx(ctx, r.db).
		Raw(countPaymentMethodUsageSQL, clinicID, id, clinicID, id, clinicID, id, clinicID, id, clinicID, id).
		Scan(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "payment_method", fmt.Sprintf("%d", id))
	}
	return count, nil
}

func (r *paymentMethodMasterRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return reorderByClinicID(ctx, r.db, &model.PaymentMethodMaster{}, "payment_method", clinicID, ids, "display_order")
}
