package billing

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

func makeItemlessCloseBilling(t *testing.T, db *gorm.DB, clinicID uint64, completedAt time.Time, amounts []int64) *model.Billing {
	t.Helper()
	var total int64
	for _, amount := range amounts {
		total += amount
	}
	b := &model.Billing{
		ClinicID: clinicID, Status: model.BillingStatusCompleted, CompletedAt: &completedAt,
		ScheduledDate: completedAt, TotalAmount: total,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(b).Error)
	for i, amount := range amounts {
		method := model.PaymentMethodCash
		if i > 0 {
			method = model.PaymentMethodCreditCard
		}
		require.NoError(t, db.WithContext(context.Background()).Create(&model.PaymentSplit{
			ClinicID: clinicID, BillingID: b.ID, Method: method, Amount: amount,
		}).Error)
	}
	return b
}

func TestAccountingRepository_GetCloseAggregate_ItemlessClassification(t *testing.T) {
	tests := []struct {
		name                string
		itemCategory        model.ItemCategory
		softDeleteItems     bool
		amounts             []int64
		refund              int64
		wantCategory        string
		wantCount           int64
		wantNet             []int64
		wantCategoryAmounts []map[string]int64
	}{
		{
			name: "paid_itemless", amounts: []int64{3000},
			wantCategory: "other", wantCount: 1, wantNet: []int64{3000},
			wantCategoryAmounts: []map[string]int64{{"other": 3000}},
		},
		{
			name: "all_items_soft_deleted", itemCategory: model.ItemCategoryExamination,
			softDeleteItems: true, amounts: []int64{3000},
			wantCategory: "other", wantCount: 1, wantNet: []int64{3000},
			wantCategoryAmounts: []map[string]int64{{"other": 3000}},
		},
		{
			name: "two_payment_splits_count_once", amounts: []int64{1000, 2000},
			wantCategory: "other", wantCount: 1, wantNet: []int64{1000, 2000},
			wantCategoryAmounts: []map[string]int64{{"other": 1000}, {"other": 2000}},
		},
		{
			name: "refund_preserves_gross", amounts: []int64{3000}, refund: 1200,
			wantCategory: "other", wantCount: 1, wantNet: []int64{1800},
			wantCategoryAmounts: []map[string]int64{{"other": 3000}},
		},
		{
			name: "categorized_non_other_unchanged", itemCategory: model.ItemCategoryExamination,
			amounts: []int64{3000}, wantCategory: "examination", wantCount: 0, wantNet: []int64{3000},
			wantCategoryAmounts: []map[string]int64{{"examination": 3000}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupCloseAggregateTestDB(t)
			ctx := context.Background()
			const clinicID = uint64(1)
			var b *model.Billing
			if tt.itemCategory == "" {
				b = makeItemlessCloseBilling(t, db, clinicID, midJune(10), tt.amounts)
			} else {
				b = makeCloseBilling(t, db, clinicID, midJune(10), tt.itemCategory, 3000, 0.10, model.PaymentMethodCash, 3000)
			}
			if tt.softDeleteItems {
				require.NoError(t, db.WithContext(ctx).Where("billing_id = ?", b.ID).Delete(&model.BillingItem{}).Error)
			}
			if tt.refund > 0 {
				require.NoError(t, db.WithContext(ctx).Create(&model.BillingRefund{
					ClinicID: clinicID, BillingID: b.ID, Amount: tt.refund, RefundedAt: midJune(11),
				}).Error)
			}

			result, err := NewAccountingRepository(db).GetCloseAggregate(ctx, GetCloseAggregateInput{
				ClinicID:    clinicID,
				PeriodStart: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
				PeriodEnd:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
			})
			require.NoError(t, err)
			assert.Equal(t, tt.wantCount, result.UnclassifiedOtherCount)
			assert.Equal(t, tt.refund, result.TotalRefund)
			require.Len(t, result.BillingDetails, len(tt.amounts))
			var gross, net []int64
			var categoryAmounts []map[string]int64
			for _, detail := range result.BillingDetails {
				assert.Equal(t, b.ID, detail.BillingID)
				assert.Equal(t, tt.wantCategory, detail.Category)
				assert.Equal(t, tt.refund, detail.RefundAmount)
				gross = append(gross, detail.BillingAmount)
				net = append(net, detail.NetAmount)
				categoryAmounts = append(categoryAmounts, detail.CategoryAmounts)
			}
			assert.ElementsMatch(t, tt.amounts, gross)
			assert.ElementsMatch(t, tt.wantNet, net)
			assert.ElementsMatch(t, tt.wantCategoryAmounts, categoryAmounts)
			var paid int64
			for _, row := range result.PaymentRows {
				paid += row.Amount
			}
			assert.Equal(t, int64(3000), paid)
			if tt.wantCategory == "other" {
				assert.Empty(t, result.CategoryRows)
				assert.Empty(t, result.TaxBreakdown)
			} else {
				assert.Equal(t, []CategoryAggregateRow{{Category: "examination", Amount: 3000}}, result.CategoryRows)
				assert.Equal(t, []TaxBreakdownRow{{TaxRate: 10, TaxableAmount: 3000, TaxAmount: 300}}, result.TaxBreakdown)
			}
		})
	}
}

func TestAccountingRepository_GetCloseAggregate_MixedOtherAndItemlessCount(t *testing.T) {
	db := setupCloseAggregateTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)
	mixed := makeCloseBilling(t, db, clinicID, midJune(10), model.ItemCategoryExamination, 3000, 0.10, model.PaymentMethodCash, 4000)
	require.NoError(t, db.WithContext(ctx).Create(&model.BillingItem{
		BillingID: mixed.ID, Category: model.ItemCategoryOther, Name: "other",
		UnitPrice: 1000, Quantity: 1, TaxType: model.TaxTypeExcluded, TaxRate: 0.10,
	}).Error)
	itemless := makeItemlessCloseBilling(t, db, clinicID, midJune(11), []int64{1000})

	result, err := NewAccountingRepository(db).GetCloseAggregate(ctx, GetCloseAggregateInput{
		ClinicID:    clinicID,
		PeriodStart: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.UnclassifiedOtherCount)
	assert.Equal(t, int64(0), result.TotalRefund)
	require.Len(t, result.BillingDetails, 2)
	byID := make(map[uint64]CloseBillingDetailRow)
	for _, detail := range result.BillingDetails {
		byID[detail.BillingID] = detail
	}
	require.Contains(t, byID, mixed.ID)
	require.Contains(t, byID, itemless.ID)
	assert.Equal(t, "examination", byID[mixed.ID].Category)
	assert.Equal(t, int64(4000), byID[mixed.ID].BillingAmount)
	assert.Equal(t, int64(4000), byID[mixed.ID].NetAmount)
	assert.Equal(t, map[string]int64{"examination": 3000, "other": 1000}, byID[mixed.ID].CategoryAmounts)
	assert.Equal(t, "other", byID[itemless.ID].Category)
	assert.Equal(t, int64(1000), byID[itemless.ID].BillingAmount)
	assert.Equal(t, int64(1000), byID[itemless.ID].NetAmount)
	assert.Equal(t, map[string]int64{"other": 1000}, byID[itemless.ID].CategoryAmounts)
	var paid int64
	for _, row := range result.PaymentRows {
		paid += row.Amount
	}
	assert.Equal(t, int64(5000), paid)
}

func TestAccountingRepository_GetCloseAggregate_ItemlessScope(t *testing.T) {
	tests := []struct {
		name        string
		clinicID    uint64
		completedAt time.Time
		status      model.BillingStatus
		deleted     bool
	}{
		{name: "foreign_clinic", clinicID: 2, completedAt: midJune(10), status: model.BillingStatusCompleted},
		{name: "not_completed", clinicID: 1, completedAt: midJune(10), status: model.BillingStatusWaiting},
		{name: "before_period", clinicID: 1, completedAt: time.Date(2026, 5, 31, 23, 59, 59, 0, time.UTC), status: model.BillingStatusCompleted},
		{name: "period_end_exclusive", clinicID: 1, completedAt: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC), status: model.BillingStatusCompleted},
		{name: "deleted_billing", clinicID: 1, completedAt: midJune(10), status: model.BillingStatusCompleted, deleted: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupCloseAggregateTestDB(t)
			ctx := context.Background()
			included := makeItemlessCloseBilling(t, db, 1, midJune(12), []int64{1000})
			excluded := makeItemlessCloseBilling(t, db, tt.clinicID, tt.completedAt, []int64{9999})
			require.NoError(t, db.WithContext(ctx).Model(excluded).Update("status", tt.status).Error)
			if tt.deleted {
				require.NoError(t, db.WithContext(ctx).Delete(excluded).Error)
			}
			require.NoError(t, db.WithContext(ctx).Create(&model.BillingRefund{
				ClinicID: tt.clinicID, BillingID: excluded.ID, Amount: 1234, RefundedAt: midJune(13),
			}).Error)

			result, err := NewAccountingRepository(db).GetCloseAggregate(ctx, GetCloseAggregateInput{
				ClinicID:    1,
				PeriodStart: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
				PeriodEnd:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
			})
			require.NoError(t, err)
			assert.Equal(t, int64(1), result.UnclassifiedOtherCount)
			assert.Equal(t, int64(0), result.TotalRefund)
			require.Len(t, result.BillingDetails, 1)
			assert.Equal(t, included.ID, result.BillingDetails[0].BillingID)
			assert.Equal(t, "other", result.BillingDetails[0].Category)
			assert.Equal(t, int64(1000), result.BillingDetails[0].BillingAmount)
			assert.Equal(t, int64(1000), result.BillingDetails[0].NetAmount)
			assert.Equal(t, int64(0), result.BillingDetails[0].RefundAmount)
			assert.Equal(t, map[string]int64{"other": 1000}, result.BillingDetails[0].CategoryAmounts)
			var paid int64
			for _, row := range result.PaymentRows {
				paid += row.Amount
			}
			assert.Equal(t, int64(1000), paid)
		})
	}
}
