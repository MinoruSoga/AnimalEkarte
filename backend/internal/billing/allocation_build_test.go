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
			"現金":       100,
			"クレジット":    200,
			"旧ポイント":    50,
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
