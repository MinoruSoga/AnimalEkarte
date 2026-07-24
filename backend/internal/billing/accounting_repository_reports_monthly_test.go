package billing

// accounting_repository_reports_monthly_test.go — AccountingRepository.GetMonthlyReport /
// GetMonthlyReportByPeriod の統合テスト（実 Postgres テスト DB）。
//
// setupAccountingIsolationTestDB / makeBillingRet は accounting_repository_clinic_isolation_test.go
// で定義済みのためここでは再利用する（重複宣言禁止）。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupMonthlyReportTestDB は GetMonthlyReport 系テスト用に cash_register_closes を追加整備する。
func setupMonthlyReportTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupAccountingIsolationTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.CashRegisterClose{}))
	db.Exec("TRUNCATE TABLE cash_register_closes CASCADE")
	return db
}

// makeMonthlyBilling は完了済み会計を明細1件+支払い内訳1件付きで作成して返す。
func makeMonthlyBilling(t *testing.T, db *gorm.DB, clinicID uint64, completedAt time.Time, category model.ItemCategory, unitPrice int64, taxRate float64, method model.PaymentMethod, splitAmount int64) *model.Billing {
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

// midJune / midMay / midJuly は AT TIME ZONE 'Asia/Tokyo' 変換・time.Local 双方で
// 日付境界の揺れに影響されない「安全な月中」の時刻（UTC 03:00 = JST 12:00）。
func midJune(day int) time.Time { return time.Date(2026, 6, day, 3, 0, 0, 0, time.UTC) }
func midMay(day int) time.Time  { return time.Date(2026, 5, day, 3, 0, 0, 0, time.UTC) }
func midJuly(day int) time.Time { return time.Date(2026, 7, day, 3, 0, 0, 0, time.UTC) }

func TestAccountingRepository_GetMonthlyReport_AggregatesPaymentsCategoriesAndRefunds(t *testing.T) {
	db := setupMonthlyReportTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	b1 := makeMonthlyBilling(t, db, clinicID, midJune(10), model.ItemCategoryExamination, 1000, 0.10, model.PaymentMethodCash, 1000)
	makeMonthlyBilling(t, db, clinicID, midJune(10), model.ItemCategoryMedicine, 2000, 0.08, model.PaymentMethodCreditCard, 2000)

	// 返金 500 円 (b1 に対して)
	refund := &model.BillingRefund{ClinicID: clinicID, BillingID: b1.ID, Amount: 500, RefundedAt: midJune(11)}
	require.NoError(t, db.WithContext(ctx).Create(refund).Error)

	result, err := repo.GetMonthlyReport(ctx, clinicID, 2026, 6)
	require.NoError(t, err)

	assert.Equal(t, int64(2), result.BillingCount, "6月完了分は2件")
	assert.Equal(t, int64(500), result.TotalRefund)
	assert.Equal(t, int64(3000-500), result.GrandTotal, "支払い内訳合計(3000) - 返金(500)")

	var paymentTotal int64
	for _, p := range result.PaymentRows {
		paymentTotal += p.Amount
	}
	assert.Equal(t, int64(3000), paymentTotal, "支払い内訳の合計は3000円")

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

	dateKey := "2026-06-10"
	assert.Equal(t, int64(2), result.DailyBillingCount[dateKey])

	var tax10, tax8 int64
	for _, tr := range result.TaxBreakdown {
		switch tr.TaxRate {
		case 10:
			tax10 = tr.TaxAmount
		case 8:
			tax8 = tr.TaxAmount
		}
	}
	assert.Equal(t, int64(100), tax10, "1000円×10%=100")
	assert.Equal(t, int64(160), tax8, "2000円×8%=160")
}

func TestAccountingRepository_GetMonthlyReport_ExcludesOutsideMonthAndNonCompleted(t *testing.T) {
	db := setupMonthlyReportTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	// 5月分（除外）
	makeMonthlyBilling(t, db, clinicID, midMay(20), model.ItemCategoryExamination, 500, 0.10, model.PaymentMethodCash, 500)
	// 7月分（除外）
	makeMonthlyBilling(t, db, clinicID, midJuly(1), model.ItemCategoryExamination, 700, 0.10, model.PaymentMethodCash, 700)
	// waiting ステータス（除外・completed_at なし）
	waiting := &model.Billing{ClinicID: clinicID, Status: model.BillingStatusWaiting, ScheduledDate: midJune(5)}
	require.NoError(t, db.WithContext(ctx).Create(waiting).Error)

	result, err := repo.GetMonthlyReport(ctx, clinicID, 2026, 6)
	require.NoError(t, err)

	assert.Equal(t, int64(0), result.BillingCount, "6月完了分は0件")
	assert.Equal(t, int64(0), result.GrandTotal)
	assert.Empty(t, result.PaymentRows)
	assert.Empty(t, result.CategoryRows)
}

func TestAccountingRepository_GetMonthlyReport_ClosedAMPMMap(t *testing.T) {
	db := setupMonthlyReportTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	amClose := &model.CashRegisterClose{ClinicID: clinicID, CloseDate: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC), Period: "am", CategoryBreakdown: []byte("{}")}
	pmClose := &model.CashRegisterClose{ClinicID: clinicID, CloseDate: time.Date(2026, 6, 12, 0, 0, 0, 0, time.UTC), Period: "pm", CategoryBreakdown: []byte("{}")}
	require.NoError(t, db.WithContext(ctx).Create(amClose).Error)
	require.NoError(t, db.WithContext(ctx).Create(pmClose).Error)

	result, err := repo.GetMonthlyReport(ctx, clinicID, 2026, 6)
	require.NoError(t, err)

	assert.True(t, result.ClosedAM["2026-06-10"], "6/10 は AM 締め済み")
	assert.False(t, result.ClosedPM["2026-06-10"], "6/10 は PM 未締め")
	assert.True(t, result.ClosedPM["2026-06-12"], "6/12 は PM 締め済み")
	assert.False(t, result.ClosedAM["2026-06-15"], "未締め日は false")
}

func TestAccountingRepository_GetMonthlyReport_ClinicIsolation(t *testing.T) {
	db := setupMonthlyReportTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	makeMonthlyBilling(t, db, clinicA, midJune(10), model.ItemCategoryExamination, 1000, 0.10, model.PaymentMethodCash, 1000)
	makeMonthlyBilling(t, db, clinicB, midJune(10), model.ItemCategoryExamination, 9999, 0.10, model.PaymentMethodCash, 9999)

	resultA, err := repo.GetMonthlyReport(ctx, clinicA, 2026, 6)
	require.NoError(t, err)
	assert.Equal(t, int64(1), resultA.BillingCount)
	assert.Equal(t, int64(1000), resultA.GrandTotal, "clinic B の会計が混入してはならない")

	resultB, err := repo.GetMonthlyReport(ctx, clinicB, 2026, 6)
	require.NoError(t, err)
	assert.Equal(t, int64(1), resultB.BillingCount)
	assert.Equal(t, int64(9999), resultB.GrandTotal)
}

func TestAccountingRepository_GetMonthlyReportByPeriod_CustomRangeAcrossMonths(t *testing.T) {
	db := setupMonthlyReportTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	makeMonthlyBilling(t, db, clinicID, midJune(28), model.ItemCategoryExamination, 1500, 0.10, model.PaymentMethodCash, 1500)
	makeMonthlyBilling(t, db, clinicID, midJuly(2), model.ItemCategoryExamination, 2500, 0.10, model.PaymentMethodCash, 2500)

	// 6/25 〜 7/5 のカスタム期間（月をまたぐ）
	start := time.Date(2026, 6, 25, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)

	result, err := repo.GetMonthlyReportByPeriod(ctx, clinicID, start, end)
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.BillingCount, "月をまたぐ期間の両方が含まれる")
	assert.Equal(t, int64(4000), result.GrandTotal)
}

func TestAccountingRepository_GetMonthlyReport_EmptyMonthReturnsZeroValues(t *testing.T) {
	db := setupMonthlyReportTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	result, err := repo.GetMonthlyReport(ctx, clinicID, 2026, 6)
	require.NoError(t, err)

	assert.Equal(t, int64(0), result.BillingCount)
	assert.Equal(t, int64(0), result.GrandTotal)
	assert.Equal(t, int64(0), result.TotalRefund)
	assert.Empty(t, result.PaymentRows)
	assert.Empty(t, result.CategoryRows)
	assert.Empty(t, result.DailyBillingCount)
	assert.Empty(t, result.TaxBreakdown)
}
