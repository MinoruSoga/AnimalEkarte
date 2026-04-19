package repository

import (
	"context"
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
	UpdateFields(ctx context.Context, clinicID, billingID uint64, fields map[string]any) (*model.Billing, error)
	UpsertPayment(ctx context.Context, payment *model.Payment) error
	// BUG-370: 月末未納者一覧
	FindUnpaidByBilling(ctx context.Context, clinicID uint64, baseDate string, page, limit int) ([]model.Billing, int64, error)
	FindUnpaidByOwner(ctx context.Context, clinicID uint64, baseDate string, page, limit int) ([]UnpaidOwnerAggregate, int64, UnpaidSummary, error)
	// BUG-368: レジ締め日次集計
	GetDailySummary(ctx context.Context, clinicID uint64, date time.Time) (*DailySummaryResult, error)
}

// PaymentMethodTotal は支払方法別の売上合計。BUG-368
type PaymentMethodTotal struct {
	Method string `json:"method"`
	Total  int64  `json:"total"`
}

// CategoryTotal は診療区分別の売上合計。BUG-368
type CategoryTotal struct {
	Category string `json:"category"`
	Total    int64  `json:"total"`
}

// DailySummaryResult はレジ締め日次集計の結果。BUG-368
type DailySummaryResult struct {
	PaymentTotals  []PaymentMethodTotal `json:"payment_totals"`
	CategoryTotals []CategoryTotal      `json:"category_totals"`
	BillingCount   int64                `json:"billing_count"`
	GrandTotal     int64                `json:"grand_total"`
}

// UnpaidOwnerAggregate は飼主単位の未納集約結果
// BUG-370
type UnpaidOwnerAggregate struct {
	OwnerID         uint64 `json:"owner_id"`
	OwnerName       string `json:"owner_name"`
	Count           int64  `json:"count"`
	TotalAmount     int64  `json:"total_amount"`
	OldestScheduled string `json:"oldest_scheduled"`
	LatestScheduled string `json:"latest_scheduled"`
}

// UnpaidSummary は未納者一覧のサマリー情報（売掛金総額）
// BUG-370
type UnpaidSummary struct {
	TotalAmount  int64 `json:"total_amount"`
	BillingCount int64 `json:"billing_count"`
	OwnerCount   int64 `json:"owner_count"`
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
	if err := q.Preload("Owner").Preload("Pet").Preload("Payments").Preload("Payments.PaidByStaff").Preload("Items").
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
		Preload("Items").
		Preload("Payments").
		Preload("Payments.PaidByStaff").
		Preload("Refunds").
		Preload("Refunds.RefundedByStaff").
		Preload("Owner").
		Preload("Pet").
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

// UpdateFields は指定フィールドのみを更新し、更新後のレコードを返す。
// map[string]any を使うことで GORM のゼロ値スキップ問題を回避する。
func (r *accountingRepository) UpdateFields(ctx context.Context, clinicID, billingID uint64, fields map[string]any) (*model.Billing, error) {
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
		Preload("Items").Preload("Payments").Preload("Payments.PaidByStaff").Preload("Refunds").Preload("Refunds.RefundedByStaff").Preload("Owner").Preload("Pet").
		Scopes(clinicScope(clinicID)).
		First(&billing, "id = ?", billingID).Error; err != nil {
		return nil, apperrors.FromGORM(err, "billing", fmt.Sprintf("%d", billingID))
	}
	return &billing, nil
}

func (r *accountingRepository) UpsertPayment(ctx context.Context, payment *model.Payment) error {
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

// FindUnpaidByBilling は未納 (status=waiting かつ scheduled_date < baseDate) の billings を
// 会計単位で返す。BUG-370 AC-5
func (r *accountingRepository) FindUnpaidByBilling(ctx context.Context, clinicID uint64, baseDate string, page, limit int) ([]model.Billing, int64, error) {
	billings := make([]model.Billing, 0)
	var total int64

	q := r.db.WithContext(ctx).Model(&model.Billing{}).
		Scopes(clinicScope(clinicID)).
		Where("status = ?", model.BillingStatusWaiting).
		Where("scheduled_date < ?", baseDate)

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "billing", "")
	}
	if err := q.Preload("Owner").Preload("Pet").Preload("Items").
		Offset((page - 1) * limit).Limit(limit).
		Order("scheduled_date ASC, created_at ASC").
		Find(&billings).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "billing", "")
	}
	return billings, total, nil
}

// FindUnpaidByOwner は未納を飼主単位で集約する。BUG-370 AC-4
// GROUP BY owner_id で 1 クエリで取得（N+1 回避）。
func (r *accountingRepository) FindUnpaidByOwner(ctx context.Context, clinicID uint64, baseDate string, page, limit int) ([]UnpaidOwnerAggregate, int64, UnpaidSummary, error) {
	aggregates := make([]UnpaidOwnerAggregate, 0)
	var totalOwners int64
	var summary UnpaidSummary

	base := r.db.WithContext(ctx).
		Table("billings").
		Joins("JOIN owners ON owners.id = billings.owner_id AND owners.deleted_at IS NULL").
		Where("billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
		Where("billings.status = ?", model.BillingStatusWaiting).
		Where("billings.scheduled_date < ?", baseDate)

	// サマリー取得（売掛金総額・件数・飼主数）
	if err := base.Session(&gorm.Session{}).
		Select("COALESCE(SUM(billings.total_amount), 0) AS total_amount, COUNT(billings.id) AS billing_count, COUNT(DISTINCT billings.owner_id) AS owner_count").
		Scan(&summary).Error; err != nil {
		return nil, 0, summary, apperrors.FromGORM(err, "billing", "")
	}
	totalOwners = summary.OwnerCount

	// 飼主単位集約
	if err := base.Session(&gorm.Session{}).
		Select("billings.owner_id AS owner_id, owners.name AS owner_name, COUNT(billings.id) AS count, COALESCE(SUM(billings.total_amount), 0) AS total_amount, MIN(billings.scheduled_date)::text AS oldest_scheduled, MAX(billings.scheduled_date)::text AS latest_scheduled").
		Group("billings.owner_id, owners.name").
		Order("oldest_scheduled ASC").
		Offset((page - 1) * limit).
		Limit(limit).
		Scan(&aggregates).Error; err != nil {
		return nil, 0, summary, apperrors.FromGORM(err, "billing", "")
	}
	return aggregates, totalOwners, summary, nil
}

// GetDailySummary は指定日（JST）の会計完了分を集計する。BUG-368
func (r *accountingRepository) GetDailySummary(ctx context.Context, clinicID uint64, date time.Time) (*DailySummaryResult, error) {
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	jstDate := date.In(jst).Format("2006-01-02")

	// 合計件数・売上合計
	var base struct {
		BillingCount int64
		GrandTotal   int64
	}
	if err := r.db.WithContext(ctx).
		Table("billings").
		Joins("JOIN payments ON payments.billing_id = billings.id").
		Where("billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
		Where("billings.status = ?", model.BillingStatusCompleted).
		Where("DATE(billings.completed_at AT TIME ZONE 'Asia/Tokyo') = ?", jstDate).
		Select("COUNT(DISTINCT billings.id) AS billing_count, COALESCE(SUM(payments.billing_amount), 0) AS grand_total").
		Scan(&base).Error; err != nil {
		return nil, apperrors.FromGORM(err, "billing", "")
	}

	// 支払方法別合計
	paymentTotals := make([]PaymentMethodTotal, 0)
	if err := r.db.WithContext(ctx).
		Table("billings").
		Joins("JOIN payments ON payments.billing_id = billings.id").
		Where("billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
		Where("billings.status = ?", model.BillingStatusCompleted).
		Where("DATE(billings.completed_at AT TIME ZONE 'Asia/Tokyo') = ?", jstDate).
		Select("payments.method AS method, COALESCE(SUM(payments.billing_amount), 0) AS total").
		Group("payments.method").
		Scan(&paymentTotals).Error; err != nil {
		return nil, apperrors.FromGORM(err, "billing", "")
	}

	// 診療区分別合計
	categoryTotals := make([]CategoryTotal, 0)
	if err := r.db.WithContext(ctx).
		Table("billings").
		Joins("JOIN billing_items ON billing_items.billing_id = billings.id AND billing_items.deleted_at IS NULL").
		Where("billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
		Where("billings.status = ?", model.BillingStatusCompleted).
		Where("DATE(billings.completed_at AT TIME ZONE 'Asia/Tokyo') = ?", jstDate).
		Select("billing_items.category::text AS category, COALESCE(SUM(ROUND(billing_items.unit_price * billing_items.quantity::numeric)), 0) AS total").
		Group("billing_items.category").
		Scan(&categoryTotals).Error; err != nil {
		return nil, apperrors.FromGORM(err, "billing_item", "")
	}

	return &DailySummaryResult{
		PaymentTotals:  paymentTotals,
		CategoryTotals: categoryTotals,
		BillingCount:   base.BillingCount,
		GrandTotal:     base.GrandTotal,
	}, nil
}
