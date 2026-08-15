package billing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// TestBuildAllocationBillings_Conservation covers the shared path used by close + monthly:
// discount-shrunk payments, multi-category, multi-method, refunds, NULL legacy cash.
func TestBuildAllocationBillings_Conservation(t *testing.T) {
	t.Parallel()

	cashID := uint64(1)
	cardID := uint64(2)
	unknownID := uint64(99)
	cashEnum := string(model.PaymentMethodCash)

	data := &CategoryPaymentAllocationData{
		Weights: []allocationCategoryWeightRow{
			{BillingID: 10, Category: "examination", Amount: 1000}, // after item discount
			{BillingID: 10, Category: "goods", Amount: 2000},
			{BillingID: 11, Category: "examination", Amount: 500},
		},
		Payments: []allocationPaymentRow{
			// billing 10: payment net 2700 (billing discount 300 on 3000 items)
			{BillingID: 10, PaymentMethodID: &cashID, Amount: 900},
			{BillingID: 10, PaymentMethodID: &cardID, Amount: 1800},
			// billing 11: NULL legacy cash + unknown method
			{BillingID: 11, PaymentMethodID: nil, Amount: 200},
			{BillingID: 11, PaymentMethodID: &unknownID, Amount: 300},
		},
		Refunds: []allocationRefundRow{
			// refund on occurrence day, allocated to original categories of billing 10
			{BillingID: 10, PaymentMethod: &cashEnum, Amount: 100},
		},
		CategoryCounts: map[string]int64{
			"examination": 2,
			"goods":       1,
		},
	}

	payMethods := []model.PaymentMethodMaster{
		{ID: 1, Name: "現金", SystemKey: ptrString("cash"), IsActive: true, DisplayOrder: 1},
		{ID: 2, Name: "クレジット", SystemKey: ptrString("credit_card"), IsActive: true, DisplayOrder: 2},
	}
	sysFn, sysRefundFn := BuildSystemKeyMethodResolvers(payMethods)
	billings := BuildAllocationBillings(data, sysFn, sysRefundFn)
	matrix := AggregateCategoryPaymentMatrix(billings)

	// payments 900+1800+200+300 - refund 100 = 3100
	assert.Equal(t, int64(3100), MatrixGrandTotal(matrix), "grand total = payment net after refund")

	cols := MatrixColumnTotals(matrix)
	assert.Equal(t, int64(1000), cols["cash"])        // 900 + 200 - 100
	assert.Equal(t, int64(1800), cols["credit_card"]) // 1800
	assert.Equal(t, int64(300), cols["method_99"])    // unknown

	// Row sum == column sum == grand
	var rowSum int64
	for _, v := range MatrixRowTotals(matrix) {
		rowSum += v
	}
	var colSum int64
	for _, v := range cols {
		colSum += v
	}
	assert.Equal(t, rowSum, colSum)
	assert.Equal(t, rowSum, MatrixGrandTotal(matrix))

	// Name-keyed path for preview/monthly
	nameFn, nameRefundFn := BuildNameMethodResolvers(payMethods)
	nameMatrix := AggregateCategoryPaymentMatrix(BuildAllocationBillings(data, nameFn, nameRefundFn))
	assert.Equal(t, int64(3100), MatrixGrandTotal(nameMatrix))
	assert.Equal(t, int64(1000), MatrixColumnTotals(nameMatrix)["現金"])
	assert.Equal(t, int64(1800), MatrixColumnTotals(nameMatrix)["クレジット"])
	assert.Contains(t, MatrixColumnTotals(nameMatrix), "不明な支払方法(99)")
}

// TestBuildAllocationBillings_RefundsRemainSeparateNegativeRows locks DEC-16⑥ (a):
// refunds are occurrence-day negative rows — not pre-netted into payment amounts before allocate.
func TestBuildAllocationBillings_RefundsRemainSeparateNegativeRows(t *testing.T) {
	t.Parallel()

	cashID := uint64(1)
	cashEnum := string(model.PaymentMethodCash)
	data := &CategoryPaymentAllocationData{
		Weights: []allocationCategoryWeightRow{
			{BillingID: 42, Category: "examination", Amount: 600},
			{BillingID: 42, Category: "goods", Amount: 400},
		},
		Payments: []allocationPaymentRow{
			{BillingID: 42, PaymentMethodID: &cashID, Amount: 1000},
		},
		// Two refunds on occurrence day — must stay two negative rows (not -150 pre-sum, not 850 net payment).
		Refunds: []allocationRefundRow{
			{BillingID: 42, PaymentMethod: &cashEnum, Amount: 100},
			{BillingID: 42, PaymentMethod: &cashEnum, Amount: 50},
		},
	}

	payMethods := []model.PaymentMethodMaster{
		{ID: 1, Name: "現金", SystemKey: ptrString("cash"), IsActive: true, DisplayOrder: 1},
	}
	sysFn, sysRefundFn := BuildSystemKeyMethodResolvers(payMethods)
	billings := BuildAllocationBillings(data, sysFn, sysRefundFn)
	require.Len(t, billings, 1)

	pays := billings[0].Payments
	require.Len(t, pays, 3, "1 payment + 2 refund rows; must not collapse to net payment")
	var positive, negatives int
	var negSum int64
	for _, p := range pays {
		assert.Equal(t, "cash", p.MethodKey)
		switch {
		case p.Amount > 0:
			positive++
			assert.Equal(t, int64(1000), p.Amount)
		case p.Amount < 0:
			negatives++
			negSum += p.Amount
		}
	}
	assert.Equal(t, 1, positive)
	assert.Equal(t, 2, negatives)
	assert.Equal(t, int64(-150), negSum)

	matrix := AggregateCategoryPaymentMatrix(billings)
	// Conservation still holds on occurrence-net: 1000 - 100 - 50
	assert.Equal(t, int64(850), MatrixGrandTotal(matrix))
	assert.Equal(t, int64(850), MatrixColumnTotals(matrix)["cash"])
}

// TestBuildAllocationBillings_CrossPeriodRefundIsOccurrenceNegativeOnly locks DEC-16⑥:
// a refund whose parent completed outside the period still contributes a negative row only
// (no positive payment from that billing), using RefundParentWeights for category share.
func TestBuildAllocationBillings_CrossPeriodRefundIsOccurrenceNegativeOnly(t *testing.T) {
	t.Parallel()

	cashEnum := string(model.PaymentMethodCash)
	data := &CategoryPaymentAllocationData{
		// Period completed set has no billing 99 — only the refund occurs in period.
		Weights:  nil,
		Payments: nil,
		Refunds: []allocationRefundRow{
			{BillingID: 99, PaymentMethod: &cashEnum, Amount: 200},
		},
		RefundParentWeights: []allocationCategoryWeightRow{
			{BillingID: 99, Category: "examination", Amount: 1},
			{BillingID: 99, Category: "goods", Amount: 1},
		},
	}

	payMethods := []model.PaymentMethodMaster{
		{ID: 1, Name: "現金", SystemKey: ptrString("cash"), IsActive: true, DisplayOrder: 1},
	}
	sysFn, sysRefundFn := BuildSystemKeyMethodResolvers(payMethods)
	billings := BuildAllocationBillings(data, sysFn, sysRefundFn)
	require.Len(t, billings, 1)
	require.Len(t, billings[0].Payments, 1)
	assert.Equal(t, int64(-200), billings[0].Payments[0].Amount)
	assert.Equal(t, "cash", billings[0].Payments[0].MethodKey)

	matrix := AggregateCategoryPaymentMatrix(billings)
	assert.Equal(t, int64(-200), MatrixGrandTotal(matrix), "occurrence-day refund is the only matrix contribution")
	assert.Equal(t, int64(-200), MatrixColumnTotals(matrix)["cash"])
}

// TestBuildAllocationBillings_MatrixGrandEqualsOccurrenceNetNotCompletedAttachedKPI documents
// DEC-16⑥ (b) dual definition: matrix grand == 締め合計 (occurrence refund net), which can
// differ from KPI NetAmount that attaches refunds to the completed-billing period.
func TestBuildAllocationBillings_MatrixGrandEqualsOccurrenceNetNotCompletedAttachedKPI(t *testing.T) {
	t.Parallel()

	cashID := uint64(1)
	cashEnum := string(model.PaymentMethodCash)

	// Scenario (same calendar period window for matrix):
	//   - billing A completed in period: payment 3000 cash
	//   - billing B completed outside period, refund 400 cash with refunded_at in period
	// Matrix grand (DEC-16⑥ 締め合計) = 3000 - 400 = 2600
	// KPI NetAmount (completed-attached refunds only for billings completed in period) =
	//   3000 - 0 = 3000  (billing B's refund is NOT attached to a completed-in-period billing)
	data := &CategoryPaymentAllocationData{
		Weights: []allocationCategoryWeightRow{
			{BillingID: 1, Category: "examination", Amount: 3000},
		},
		Payments: []allocationPaymentRow{
			{BillingID: 1, PaymentMethodID: &cashID, Amount: 3000},
		},
		Refunds: []allocationRefundRow{
			{BillingID: 2, PaymentMethod: &cashEnum, Amount: 400},
		},
		RefundParentWeights: []allocationCategoryWeightRow{
			{BillingID: 2, Category: "goods", Amount: 1000},
		},
	}

	payMethods := []model.PaymentMethodMaster{
		{ID: 1, Name: "現金", SystemKey: ptrString("cash"), IsActive: true, DisplayOrder: 1},
	}
	sysFn, sysRefundFn := BuildSystemKeyMethodResolvers(payMethods)
	matrix := AggregateCategoryPaymentMatrix(BuildAllocationBillings(data, sysFn, sysRefundFn))

	matrixGrand := MatrixGrandTotal(matrix)
	assert.Equal(t, int64(2600), matrixGrand, "締め合計 = period payments − occurrence-date refunds")

	// KPI NetAmount path (documented dual definition — not used as matrix total):
	var paymentSum int64
	for _, p := range data.Payments {
		paymentSum += p.Amount
	}
	// completed-attached refunds for billings that have period payments/weights only
	completedInPeriod := map[uint64]struct{}{1: {}}
	var completedAttachedRefund int64
	for _, ref := range data.Refunds {
		if _, ok := completedInPeriod[ref.BillingID]; ok {
			completedAttachedRefund += ref.Amount
		}
	}
	kpiNet := paymentSum - completedAttachedRefund
	assert.Equal(t, int64(3000), kpiNet, "KPI NetAmount ignores occurrence-only cross-period refund")
	assert.NotEqual(t, matrixGrand, kpiNet, "dual definition: matrix grand ≠ KPI NetAmount when refund occurrence ≠ completion period")
}

func TestOrderPaymentMethodsForMatrix_InactiveAndUnknownAtEnd(t *testing.T) {
	t.Parallel()
	masters := []model.PaymentMethodMaster{
		{ID: 1, Name: "現金", IsActive: true, DisplayOrder: 1},
		{ID: 2, Name: "旧ポイント", IsActive: false, DisplayOrder: 9},
		{ID: 3, Name: "クレジット", IsActive: true, DisplayOrder: 2},
	}
	// FindAll already sorts; simulate that order
	masters = []model.PaymentMethodMaster{
		{ID: 1, Name: "現金", IsActive: true},
		{ID: 3, Name: "クレジット", IsActive: true},
		{ID: 2, Name: "旧ポイント", IsActive: false},
	}
	matrix := map[string]map[string]int64{
		"examination": {
			"現金":      100,
			"クレジット":   200,
			"旧ポイント":   50,
			"削除済みカード": 30,
		},
	}
	ordered := orderPaymentMethodsForMatrix(masters, matrix)
	require.GreaterOrEqual(t, len(ordered), 4)
	names := make([]string, len(ordered))
	for i, m := range ordered {
		names[i] = m.Name
	}
	// active first
	assert.Equal(t, "現金", names[0])
	assert.Equal(t, "クレジット", names[1])
	// inactive with data next
	assert.Equal(t, "旧ポイント", names[2])
	// unknown last
	assert.Equal(t, "削除済みカード", names[3])
}
