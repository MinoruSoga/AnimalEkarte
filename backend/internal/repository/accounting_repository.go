package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type AccountingRepository interface {
	FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error)
	Create(ctx context.Context, clinicID uint64, accounting *model.Billing) error
	Update(ctx context.Context, clinicID, billingID uint64, fields map[string]any) (*model.Billing, error)
	SavePayment(ctx context.Context, payment *model.Payment) error
	// SavePaymentSplits は billing の payment_splits を delete-then-recreate で保存する。
	SavePaymentSplits(ctx context.Context, splits []model.PaymentSplit) error
	// CompleteAccountingAppointments は会計完了に伴い、同日同一ペットの会計待ち予約を完了へ進める。
	CompleteAccountingAppointments(ctx context.Context, clinicID uint64, ownerID, petID *uint64, scheduledDate time.Time) (int64, error)
	// BUG-370: 月末未納者一覧
	FindUnpaidByBilling(ctx context.Context, clinicID uint64, baseDate string, page, limit int) ([]model.Billing, int64, error)
	FindUnpaidByOwner(ctx context.Context, clinicID uint64, baseDate string, page, limit int) ([]UnpaidOwnerAggregate, int64, UnpaidSummary, error)
	// BUG-368: レジ締め日次集計
	GetDailySummary(ctx context.Context, clinicID uint64, date time.Time) (*DailySummaryResult, error)
	// FEAT-368: 集計・締め機能
	GetCloseAggregate(ctx context.Context, input GetCloseAggregateInput) (*CloseAggregateResult, error)
	GetMonthlyReport(ctx context.Context, clinicID uint64, year, month int) (*MonthlyReportResult, error)
	// SumPaidByOwner は飼い主の支払済み請求合計（LTV）を返す（Lステップタグ同期用）。
	SumPaidByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error)
	// MaxSingleVisitAmountByOwner は飼い主の1回来院最大支払額を返す（CPMスポット判定用）。
	MaxSingleVisitAmountByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error)
	// FindOwnersByAnnualRevenue は直近365日の完了済み請求額合計を飼い主ごとに集計し、降順で返す（LTV上位％判定用）。
	FindOwnersByAnnualRevenue(ctx context.Context, clinicID uint64) ([]OwnerAnnualRevenue, error)
}

type accountingRepository struct {
	db *gorm.DB
}

func NewAccountingRepository(db *gorm.DB) AccountingRepository {
	return &accountingRepository{db: db}
}

func (r *accountingRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error) {
	billings := make([]model.Billing, 0)
	var total int64

	q := r.db.WithContext(ctx).Model(&model.Billing{}).Scopes(clinicScope(clinicID))
	if petID != nil {
		q = q.Where("pet_id = ?", *petID)
	}
	if ownerID != nil {
		q = q.Where("owner_id = ?", *ownerID)
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if startDate != nil {
		q = q.Where("scheduled_date >= ?", *startDate)
	}
	if endDate != nil {
		q = q.Where("scheduled_date <= ?", *endDate)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "billing", "")
	}
	if err := q.Preload("Owner", "deleted_at IS NULL").Preload("Pet", "deleted_at IS NULL").Preload("Payments", "deleted_at IS NULL").Preload("Payments.PaidByStaff", "deleted_at IS NULL").Preload("Items", "deleted_at IS NULL").Preload("PaymentSplits").
		Offset((page - 1) * limit).Limit(limit).Order("scheduled_date DESC, created_at DESC").Find(&billings).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "billing", "")
	}

	// 返金合計をサブクエリで一括取得
	if len(billings) > 0 {
		ids := make([]uint64, 0, len(billings))
		for i := range billings {
			ids = append(ids, billings[i].ID)
		}
		type refundSum struct {
			BillingID uint64
			Total     int64
		}
		var sums []refundSum
		if err := r.db.WithContext(ctx).
			Model(&model.BillingRefund{}).
			Unscoped().
			Select("billing_id, COALESCE(SUM(amount), 0) AS total").
			Where("billing_id IN ?", ids).
			Group("billing_id").
			Scan(&sums).Error; err != nil {
			return nil, 0, apperrors.FromGORM(err, "billing_refund", "")
		}
		sumMap := make(map[uint64]int64, len(sums))
		for _, s := range sums {
			sumMap[s.BillingID] = s.Total
		}
		for i := range billings {
			billings[i].TotalRefundedAmount = sumMap[billings[i].ID]
		}
	}
	return billings, total, nil
}

func (r *accountingRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error) {
	var billing model.Billing
	err := r.db.WithContext(ctx).
		Preload("Items", "deleted_at IS NULL").
		Preload("Payments", "deleted_at IS NULL").
		Preload("Payments.PaidByStaff", "deleted_at IS NULL").
		Preload("Refunds").
		Preload("Refunds.RefundedByStaff").
		Preload("Owner", "deleted_at IS NULL").
		Preload("Pet", "deleted_at IS NULL").
		Preload("PaymentSplits").
		Scopes(clinicScope(clinicID)).Where("id = ?", id).First(&billing).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "billing", fmt.Sprintf("%d", id))
	}
	// Preload した Refunds から TotalRefundedAmount を計算（FindAll と同じ算出ロジック）
	var total int64
	for i := range billing.Refunds {
		total += billing.Refunds[i].Amount
	}
	billing.TotalRefundedAmount = total
	return &billing, nil
}

func (r *accountingRepository) Create(ctx context.Context, clinicID uint64, accounting *model.Billing) error {
	accounting.ClinicID = clinicID
	if err := r.db.WithContext(ctx).Create(accounting).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("billing", accounting.ScheduledDate.String())
		}
		return apperrors.FromGORM(err, "billing", "")
	}
	return nil
}

// Update は指定フィールドのみを更新し、更新後のレコードを返す。
// map[string]any を使うことで GORM のゼロ値スキップ問題を回避する。
func (r *accountingRepository) Update(ctx context.Context, clinicID, billingID uint64, fields map[string]any) (*model.Billing, error) {
	result := r.db.WithContext(ctx).
		Model(&model.Billing{}).
		Scopes(clinicScope(clinicID)).
		Where("id = ?", billingID).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "billing", fmt.Sprintf("%d", billingID))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("billing", fmt.Sprintf("%d", billingID))
	}
	var billing model.Billing
	if err := r.db.WithContext(ctx).
		Preload("Items", "deleted_at IS NULL").Preload("Payments", "deleted_at IS NULL").Preload("Payments.PaidByStaff", "deleted_at IS NULL").Preload("Refunds").Preload("Refunds.RefundedByStaff").Preload("Owner", "deleted_at IS NULL").Preload("Pet", "deleted_at IS NULL").Preload("PaymentSplits").
		Scopes(clinicScope(clinicID)).
		First(&billing, "id = ?", billingID).Error; err != nil {
		return nil, apperrors.FromGORM(err, "billing", fmt.Sprintf("%d", billingID))
	}
	return &billing, nil
}

func (r *accountingRepository) SavePayment(ctx context.Context, payment *model.Payment) error {
	// map[string]any を使用してゼロ値（Subtotal=0 等）も確実に更新する。
	// struct の Assign では GORM がゼロ値フィールドをスキップする問題がある。
	fields := map[string]any{
		"subtotal":         payment.Subtotal,
		"tax_total":        payment.TaxTotal,
		"total_amount":     payment.TotalAmount,
		"insurance_name":   payment.InsuranceName,
		"insurance_ratio":  payment.InsuranceRatio,
		"insurance_amount": payment.InsuranceAmount,
		"discount_amount":  payment.DiscountAmount,
		"billing_amount":   payment.BillingAmount,
		"received_amount":  payment.ReceivedAmount,
		"change_amount":    payment.ChangeAmount,
		"method":           payment.Method,
		"paid_by":          payment.PaidBy,
	}

	var existing model.Payment
	err := r.db.WithContext(ctx).
		Where("billing_id = ?", payment.BillingID).
		First(&existing).Error

	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			// DB エラー → 変換して返す
			return apperrors.FromGORM(err, "payment", fmt.Sprintf("billing_id=%d", payment.BillingID))
		}
		// レコードなし → 新規作成
		if err := r.db.WithContext(ctx).Create(payment).Error; err != nil {
			return apperrors.FromGORM(err, "payment", fmt.Sprintf("billing_id=%d", payment.BillingID))
		}
		return nil
	}

	// 既存レコード → map で更新（ゼロ値も反映）
	if err := r.db.WithContext(ctx).
		Model(&model.Payment{}).
		Where("billing_id = ?", payment.BillingID).
		Updates(fields).Error; err != nil {
		return apperrors.FromGORM(err, "payment", fmt.Sprintf("billing_id=%d", payment.BillingID))
	}
	payment.ID = existing.ID
	return nil
}

// SavePaymentSplits は billing の payment_splits を delete-then-recreate で保存する。
// splits が空の場合は既存レコードを削除のみ行う。
// P4 clinicScope は直接 clinic_id カラムで保証する（payment_splits 自体に clinic_id を持つ）。
func (r *accountingRepository) SavePaymentSplits(ctx context.Context, splits []model.PaymentSplit) error {
	if len(splits) == 0 {
		return nil
	}
	billingID := splits[0].BillingID
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("billing_id = ?", billingID).Delete(&model.PaymentSplit{}).Error; err != nil {
			return apperrors.FromGORM(err, "payment_split", fmt.Sprintf("billing_id=%d", billingID))
		}
		if err := tx.Create(&splits).Error; err != nil {
			return apperrors.FromGORM(err, "payment_split", fmt.Sprintf("billing_id=%d", billingID))
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to save payment splits")
	}
	return nil
}

func (r *accountingRepository) CompleteAccountingAppointments(ctx context.Context, clinicID uint64, ownerID, petID *uint64, scheduledDate time.Time) (int64, error) {
	if ownerID == nil || petID == nil || scheduledDate.IsZero() {
		return 0, nil
	}
	result := r.db.WithContext(ctx).
		Model(&model.Reservation{}).
		Where("clinic_id = ? AND owner_id = ? AND pet_id = ? AND status = ? AND deleted_at IS NULL",
			clinicID, *ownerID, *petID, model.ReservationStatusAccounting).
		Where("DATE(start_time AT TIME ZONE 'Asia/Tokyo') = DATE(? AT TIME ZONE 'Asia/Tokyo')", scheduledDate).
		Update("status", model.ReservationStatusCompleted)
	if result.Error != nil {
		return 0, apperrors.FromGORM(result.Error, "reservation", fmt.Sprintf("clinic=%d owner=%d pet=%d scheduled_date=%s", clinicID, *ownerID, *petID, scheduledDate.Format("2006-01-02")))
	}
	return result.RowsAffected, nil
}
