package billing

// accounting_repository_reports_close_test.go — AccountingRepository.GetCloseAggregate の
// 統合テスト（実 Postgres テスト DB）。
//
// setupAccountingIsolationTestDB は accounting_repository_clinic_isolation_test.go で定義済みのため再利用する。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// setupCloseAggregateTestDB は GetCloseAggregate 系テスト用の DB を用意する。
func setupCloseAggregateTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	return setupAccountingIsolationTestDB(t)
}

// makeCloseBilling は完了済み会計を明細1件+支払い内訳1件付きで作成して返す。
func makeCloseBilling(t *testing.T, db *gorm.DB, clinicID uint64, completedAt time.Time, category model.ItemCategory, unitPrice int64, taxRate float64, method model.PaymentMethod, splitAmount int64) *model.Billing {
	t.Helper()
	ctx := context.Background()
	ca := completedAt
	b := &model.Billing{
		ClinicID:      clinicID,
		Status:        model.BillingStatusCompleted,
		CompletedAt:   &ca,
		ScheduledDate: time.Date(completedAt.Year(), completedAt.Month(), completedAt.Day(), 0, 0, 0, 0, time.UTC),
		TotalAmount:   splitAmount,
	}
	require.NoError(t, db.WithContext(ctx).Create(b).Error)

	item := &model.BillingItem{
		BillingID: b.ID,
		Category:  category,
		Name:      string(category),
		UnitPrice: unitPrice,
		Quantity:  1,
		TaxType:   model.TaxTypeExcluded,
		TaxRate:   taxRate,
	}
	require.NoError(t, db.WithContext(ctx).Create(item).Error)

	split := &model.PaymentSplit{
		ClinicID:  clinicID,
		BillingID: b.ID,
		Method:    method,
		Amount:    splitAmount,
	}
	require.NoError(t, db.WithContext(ctx).Create(split).Error)

	return b
}

func TestAccountingRepository_GetCloseAggregate_AggregatesPaymentsAndCategories(t *testing.T) {
	db := setupCloseAggregateTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	makeCloseBilling(t, db, clinicID, midJune(10), model.ItemCategoryExamination, 1000, 0.10, model.PaymentMethodCash, 1000)
	makeCloseBilling(t, db, clinicID, midJune(15), model.ItemCategoryMedicine, 2000, 0.08, model.PaymentMethodCreditCard, 2000)

	result, err := repo.GetCloseAggregate(ctx, GetCloseAggregateInput{
		ClinicID:    clinicID,
		PeriodStart: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	var paymentTotal int64
	for _, p := range result.PaymentRows {
		paymentTotal += p.Amount
	}
	assert.Equal(t, int64(3000), paymentTotal)

	var examTotal, medTotal int64
	for _, c := range result.CategoryRows {
		switch c.Category {
		case string(model.ItemCategoryExamination):
			examTotal += c.Amount
		case string(model.ItemCategoryMedicine):
			medTotal += c.Amount
		}
	}
	assert.Equal(t, int64(1000), examTotal)
	assert.Equal(t, int64(2000), medTotal)
	assert.Equal(t, int64(0), result.TotalRefund)

	var tax10, tax8 int64
	for _, tr := range result.TaxBreakdown {
		switch tr.TaxRate {
		case 10:
			tax10 = tr.TaxAmount
		case 8:
			tax8 = tr.TaxAmount
		}
	}
	assert.Equal(t, int64(100), tax10)
	assert.Equal(t, int64(160), tax8)
}

func TestAccountingRepository_GetCloseAggregate_RefundReducesNetAmountNotBillingAmount(t *testing.T) {
	db := setupCloseAggregateTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	b := makeCloseBilling(t, db, clinicID, midJune(10), model.ItemCategoryExamination, 5000, 0.10, model.PaymentMethodCash, 5000)
	refund := &model.BillingRefund{ClinicID: clinicID, BillingID: b.ID, Amount: 1200, RefundedAt: midJune(11)}
	require.NoError(t, db.WithContext(ctx).Create(refund).Error)

	result, err := repo.GetCloseAggregate(ctx, GetCloseAggregateInput{
		ClinicID:    clinicID,
		PeriodStart: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	assert.Equal(t, int64(1200), result.TotalRefund)
	require.Len(t, result.BillingDetails, 1)
	detail := result.BillingDetails[0]
	assert.Equal(t, b.ID, detail.BillingID)
	assert.Equal(t, int64(5000), detail.BillingAmount, "BillingAmount は返金前の会計額")
	assert.Equal(t, int64(1200), detail.RefundAmount)
	assert.Equal(t, int64(3800), detail.NetAmount, "NetAmount = BillingAmount - RefundAmount")
	assert.False(t, detail.IsHospitalization)
}

func TestAccountingRepository_GetCloseAggregate_MixedPaymentSplitsProduceMultipleDetailRows(t *testing.T) {
	db := setupCloseAggregateTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	ca := midJune(10)
	b := &model.Billing{
		ClinicID:      clinicID,
		Status:        model.BillingStatusCompleted,
		CompletedAt:   &ca,
		ScheduledDate: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
		TotalAmount:   3000,
	}
	require.NoError(t, db.WithContext(ctx).Create(b).Error)
	item := &model.BillingItem{BillingID: b.ID, Category: model.ItemCategoryExamination, Name: "diagnosis", UnitPrice: 3000, Quantity: 1, TaxType: model.TaxTypeExcluded, TaxRate: 0.10}
	require.NoError(t, db.WithContext(ctx).Create(item).Error)

	splitCash := &model.PaymentSplit{ClinicID: clinicID, BillingID: b.ID, Method: model.PaymentMethodCash, Amount: 1000}
	splitCard := &model.PaymentSplit{ClinicID: clinicID, BillingID: b.ID, Method: model.PaymentMethodCreditCard, Amount: 2000}
	require.NoError(t, db.WithContext(ctx).Create(splitCash).Error)
	require.NoError(t, db.WithContext(ctx).Create(splitCard).Error)

	result, err := repo.GetCloseAggregate(ctx, GetCloseAggregateInput{
		ClinicID:    clinicID,
		PeriodStart: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	// 混在会計は payment_splits ベースで複数行になる（1会計だが支払い方法2件で明細2行）
	require.Len(t, result.BillingDetails, 2, "混在支払いは billing_id が同じ2明細行になる")
	var sum int64
	for _, d := range result.BillingDetails {
		assert.Equal(t, b.ID, d.BillingID)
		sum += d.BillingAmount
	}
	assert.Equal(t, int64(3000), sum)
}

func TestAccountingRepository_GetCloseAggregate_ClinicIsolation(t *testing.T) {
	db := setupCloseAggregateTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	makeCloseBilling(t, db, clinicA, midJune(10), model.ItemCategoryExamination, 1000, 0.10, model.PaymentMethodCash, 1000)
	makeCloseBilling(t, db, clinicB, midJune(10), model.ItemCategoryExamination, 9999, 0.10, model.PaymentMethodCash, 9999)

	input := GetCloseAggregateInput{
		PeriodStart: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}
	input.ClinicID = clinicA
	resultA, err := repo.GetCloseAggregate(ctx, input)
	require.NoError(t, err)
	var totalA int64
	for _, p := range resultA.PaymentRows {
		totalA += p.Amount
	}
	assert.Equal(t, int64(1000), totalA, "clinic B の会計が混入してはならない")

	input.ClinicID = clinicB
	resultB, err := repo.GetCloseAggregate(ctx, input)
	require.NoError(t, err)
	var totalB int64
	for _, p := range resultB.PaymentRows {
		totalB += p.Amount
	}
	assert.Equal(t, int64(9999), totalB)
}

func TestAccountingRepository_GetCloseAggregate_EmptyPeriodReturnsZeroValues(t *testing.T) {
	db := setupCloseAggregateTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	result, err := repo.GetCloseAggregate(ctx, GetCloseAggregateInput{
		ClinicID:    clinicID,
		PeriodStart: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	assert.Empty(t, result.PaymentRows)
	assert.Empty(t, result.CategoryRows)
	assert.Equal(t, int64(0), result.TotalRefund)
	assert.Empty(t, result.BillingDetails)
	assert.Empty(t, result.TaxBreakdown)
}
