package billing

// accounting_repository_reports_daily_test.go — AccountingRepository.GetDailySummary の
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

// setupDailySummaryTestDB は GetDailySummary 系テスト用の DB を用意する。
func setupDailySummaryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	return setupAccountingIsolationTestDB(t)
}

// makeDailyBilling は完了済み会計を明細1件+支払い内訳1件付きで作成して返す。
func makeDailyBilling(t *testing.T, db *gorm.DB, clinicID uint64, completedAt time.Time, category model.ItemCategory, unitPrice int64, method model.PaymentMethod, splitAmount int64) *model.Billing {
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
		TaxRate:   0.10,
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

func TestAccountingRepository_GetDailySummary_AggregatesPaymentAndCategoryTotals(t *testing.T) {
	db := setupDailySummaryTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	makeDailyBilling(t, db, clinicID, midJune(10), model.ItemCategoryExamination, 1000, model.PaymentMethodCash, 1000)
	makeDailyBilling(t, db, clinicID, midJune(10), model.ItemCategoryMedicine, 2000, model.PaymentMethodCreditCard, 2000)

	result, err := repo.GetDailySummary(ctx, clinicID, midJune(10))
	require.NoError(t, err)

	assert.Equal(t, int64(2), result.BillingCount)
	assert.Equal(t, int64(3000), result.GrandTotal)

	byMethod := make(map[string]int64, len(result.PaymentTotals))
	for _, p := range result.PaymentTotals {
		byMethod[p.Method] = p.Total
	}
	assert.Equal(t, int64(1000), byMethod[string(model.PaymentMethodCash)])
	assert.Equal(t, int64(2000), byMethod[string(model.PaymentMethodCreditCard)])

	byCategory := make(map[string]int64, len(result.CategoryTotals))
	for _, c := range result.CategoryTotals {
		byCategory[c.Category] = c.Total
	}
	assert.Equal(t, int64(1000), byCategory[string(model.ItemCategoryExamination)])
	assert.Equal(t, int64(2000), byCategory[string(model.ItemCategoryMedicine)])
}

func TestAccountingRepository_GetDailySummary_ExcludesOtherDaysAndNonCompleted(t *testing.T) {
	db := setupDailySummaryTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	// 前日・翌日（除外）
	makeDailyBilling(t, db, clinicID, midJune(9), model.ItemCategoryExamination, 500, model.PaymentMethodCash, 500)
	makeDailyBilling(t, db, clinicID, midJune(11), model.ItemCategoryExamination, 700, model.PaymentMethodCash, 700)
	// waiting（除外）
	waiting := &model.Billing{ClinicID: clinicID, Status: model.BillingStatusWaiting, ScheduledDate: midJune(10)}
	require.NoError(t, db.WithContext(ctx).Create(waiting).Error)

	result, err := repo.GetDailySummary(ctx, clinicID, midJune(10))
	require.NoError(t, err)

	assert.Equal(t, int64(0), result.BillingCount)
	assert.Equal(t, int64(0), result.GrandTotal)
	assert.Empty(t, result.PaymentTotals)
	assert.Empty(t, result.CategoryTotals)
}

func TestAccountingRepository_GetDailySummary_ClinicIsolation(t *testing.T) {
	db := setupDailySummaryTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	makeDailyBilling(t, db, clinicA, midJune(10), model.ItemCategoryExamination, 1000, model.PaymentMethodCash, 1000)
	makeDailyBilling(t, db, clinicB, midJune(10), model.ItemCategoryExamination, 9999, model.PaymentMethodCash, 9999)

	resultA, err := repo.GetDailySummary(ctx, clinicA, midJune(10))
	require.NoError(t, err)
	assert.Equal(t, int64(1000), resultA.GrandTotal, "clinic B の会計が混入してはならない")

	resultB, err := repo.GetDailySummary(ctx, clinicB, midJune(10))
	require.NoError(t, err)
	assert.Equal(t, int64(9999), resultB.GrandTotal)
}

func TestAccountingRepository_GetDailySummary_EmptyDayReturnsZeroValues(t *testing.T) {
	db := setupDailySummaryTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	result, err := repo.GetDailySummary(ctx, clinicID, midJune(20))
	require.NoError(t, err)

	assert.Equal(t, int64(0), result.BillingCount)
	assert.Equal(t, int64(0), result.GrandTotal)
	assert.Empty(t, result.PaymentTotals)
	assert.Empty(t, result.CategoryTotals)
}

func TestAccountingRepository_GetDailySummary_MultiplePaymentMethodsSameBilling(t *testing.T) {
	db := setupDailySummaryTestDB(t)
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

	result, err := repo.GetDailySummary(ctx, clinicID, midJune(10))
	require.NoError(t, err)

	// billing_count は COUNT(DISTINCT billings.id) のため混在支払いでも1件
	assert.Equal(t, int64(1), result.BillingCount, "混在支払いでも billing_count は DISTINCT で1件")
	assert.Equal(t, int64(3000), result.GrandTotal)
}
