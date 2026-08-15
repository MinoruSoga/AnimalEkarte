package billing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLargestRemainderDistribute_Conservation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		total   int64
		weights []int64
		want    []int64
	}{
		{
			name:    "1:2 端数 — 100 を 1:2 で配分",
			total:   100,
			weights: []int64{1, 2},
			want:    []int64{33, 67}, // 33+67=100 (floor 33,33 then +1 to larger rem of second? 100/3=33 rem1 each; both rem equal → idx0 first gets +1 → 34,66?)
		},
		{
			name:    "equal weights",
			total:   10,
			weights: []int64{1, 1},
			want:    []int64{5, 5},
		},
		{
			name:    "single weight takes all",
			total:   999,
			weights: []int64{5},
			want:    []int64{999},
		},
		{
			name:    "zero total",
			total:   0,
			weights: []int64{1, 2},
			want:    []int64{0, 0},
		},
		{
			name:    "all zero weights → index 0",
			total:   50,
			weights: []int64{0, 0, 0},
			want:    []int64{50, 0, 0},
		},
		{
			name:    "negative total (refund)",
			total:   -100,
			weights: []int64{1, 1},
			want:    []int64{-50, -50},
		},
		{
			name:    "three-way 1:1:1 of 10",
			total:   10,
			weights: []int64{1, 1, 1},
			// 10/3 = 3 rem1 each; leftover 1 → first largest rem (all equal) idx0
			want: []int64{4, 3, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := LargestRemainderDistribute(tt.total, tt.weights)
			require.Equal(t, len(tt.want), len(got))
			var sum int64
			for i := range got {
				sum += got[i]
			}
			assert.Equal(t, tt.total, sum, "conservation: sum must equal total")
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLargestRemainderDistribute_1to2FractionIsDeterministic(t *testing.T) {
	t.Parallel()
	// Classic bug case: period-wide integer truncation lost yen.
	// 1:2 of 100 must conserve 100.
	got := LargestRemainderDistribute(100, []int64{1, 2})
	assert.Equal(t, int64(100), got[0]+got[1])
	// floor shares: 33 + 66 = 99, leftover 1 → larger remainder is weight 2 (200-66*3=2 vs 100-33*3=1)
	assert.Equal(t, []int64{33, 67}, got)
}

func TestAllocateBillingPayments_ColumnAndGrandConservation(t *testing.T) {
	t.Parallel()

	in := AllocationBillingInput{
		CategoryWeights: map[string]int64{
			"examination": 1000,
			"goods":       2000,
		},
		Payments: []AllocationPayment{
			{MethodKey: "cash", Amount: 1000},
			{MethodKey: "credit_card", Amount: 2000},
		},
	}
	got := AllocateBillingPayments(in)

	// Column conservation
	cols := MatrixColumnTotals(got)
	assert.Equal(t, int64(1000), cols["cash"])
	assert.Equal(t, int64(2000), cols["credit_card"])

	// Grand total = payment net
	assert.Equal(t, int64(3000), MatrixGrandTotal(got))

	// Row totals proportional (1:2 of 3000 → 1000:2000)
	rows := MatrixRowTotals(got)
	assert.Equal(t, int64(1000), rows["examination"])
	assert.Equal(t, int64(2000), rows["goods"])
}

func TestAllocateBillingPayments_DiscountShrinksPaymentNet(t *testing.T) {
	t.Parallel()
	// Item weights 3000 but actual payment after billing discount is 2700.
	// Matrix must equal payment net, not gross item sum.
	in := AllocationBillingInput{
		CategoryWeights: map[string]int64{
			"examination": 1000,
			"goods":       2000,
		},
		Payments: []AllocationPayment{
			{MethodKey: "cash", Amount: 2700},
		},
	}
	got := AllocateBillingPayments(in)
	assert.Equal(t, int64(2700), MatrixGrandTotal(got))
	assert.Equal(t, int64(2700), MatrixColumnTotals(got)["cash"])
	rows := MatrixRowTotals(got)
	assert.Equal(t, int64(900), rows["examination"])  // 2700 * 1/3
	assert.Equal(t, int64(1800), rows["goods"])       // 2700 * 2/3
}

func TestAllocateBillingPayments_InsuranceAndTaxFoldedIntoPaymentNet(t *testing.T) {
	t.Parallel()
	// Tax/insurance are not separate cells: payment amount is already net receipt.
	in := AllocationBillingInput{
		CategoryWeights: map[string]int64{"examination": 11000}, // tax-included weight
		Payments: []AllocationPayment{
			{MethodKey: "cash", Amount: 8000}, // after insurance 3000
		},
	}
	got := AllocateBillingPayments(in)
	assert.Equal(t, int64(8000), MatrixGrandTotal(got))
	assert.Equal(t, int64(8000), got["examination"]["cash"])
}

func TestAllocateBillingPayments_RefundNegative(t *testing.T) {
	t.Parallel()
	in := AllocationBillingInput{
		CategoryWeights: map[string]int64{
			"examination": 1,
			"goods":       1,
		},
		Payments: []AllocationPayment{
			{MethodKey: "cash", Amount: -100},
		},
	}
	got := AllocateBillingPayments(in)
	assert.Equal(t, int64(-100), MatrixGrandTotal(got))
	assert.Equal(t, int64(-100), MatrixColumnTotals(got)["cash"])
}

func TestAllocateBillingPayments_ZeroWeightsFallbackOther(t *testing.T) {
	t.Parallel()
	in := AllocationBillingInput{
		CategoryWeights: map[string]int64{},
		Payments: []AllocationPayment{
			{MethodKey: "cash", Amount: 500},
		},
	}
	got := AllocateBillingPayments(in)
	assert.Equal(t, int64(500), got["other"]["cash"])
	assert.Equal(t, int64(500), MatrixGrandTotal(got))
}

func TestAllocateBillingPayments_NilLegacyCashKey(t *testing.T) {
	t.Parallel()
	in := AllocationBillingInput{
		CategoryWeights: map[string]int64{"examination": 1000},
		Payments: []AllocationPayment{
			{MethodKey: "cash", Amount: 1000}, // NULL id already mapped to cash by caller
		},
	}
	got := AllocateBillingPayments(in)
	assert.Equal(t, int64(1000), got["examination"]["cash"])
}

func TestAggregateCategoryPaymentMatrix_MultiBillingConservation(t *testing.T) {
	t.Parallel()
	billings := []AllocationBillingInput{
		{
			CategoryWeights: map[string]int64{"examination": 100, "goods": 200},
			Payments: []AllocationPayment{
				{MethodKey: "cash", Amount: 100},
				{MethodKey: "credit_card", Amount: 200},
			},
		},
		{
			CategoryWeights: map[string]int64{"examination": 50},
			Payments: []AllocationPayment{
				{MethodKey: "cash", Amount: 50},
			},
		},
		{
			// refund on second billing's categories
			CategoryWeights: map[string]int64{"examination": 50},
			Payments: []AllocationPayment{
				{MethodKey: "cash", Amount: -30},
			},
		},
	}
	got := AggregateCategoryPaymentMatrix(billings)
	// payments: 300 + 50 - 30 = 320
	assert.Equal(t, int64(320), MatrixGrandTotal(got))
	cols := MatrixColumnTotals(got)
	assert.Equal(t, int64(120), cols["cash"])        // 100+50-30
	assert.Equal(t, int64(200), cols["credit_card"]) // 200
}

func TestAggregateCategoryPaymentMatrix_FourMethodsInactiveUnknown(t *testing.T) {
	t.Parallel()
	billings := []AllocationBillingInput{
		{
			CategoryWeights: map[string]int64{"examination": 4000},
			Payments: []AllocationPayment{
				{MethodKey: "cash", Amount: 1000},
				{MethodKey: "credit_card", Amount: 1000},
				{MethodKey: "electronic_money", Amount: 1000},
				{MethodKey: "bank_transfer", Amount: 500},
				{MethodKey: "method_99", Amount: 500}, // unknown / deleted
			},
		},
	}
	got := AggregateCategoryPaymentMatrix(billings)
	assert.Equal(t, int64(4000), MatrixGrandTotal(got))
	cols := MatrixColumnTotals(got)
	assert.Equal(t, int64(1000), cols["cash"])
	assert.Equal(t, int64(1000), cols["credit_card"])
	assert.Equal(t, int64(1000), cols["electronic_money"])
	assert.Equal(t, int64(500), cols["bank_transfer"])
	assert.Equal(t, int64(500), cols["method_99"])
}

func TestRemapMatrixMethodKeys(t *testing.T) {
	t.Parallel()
	matrix := map[string]map[string]int64{
		"examination": {"1": 100, "null": 50},
	}
	got := RemapMatrixMethodKeys(matrix, func(k string) string {
		if k == "null" {
			return "現金"
		}
		if k == "1" {
			return "カード"
		}
		return k
	})
	assert.Equal(t, int64(100), got["examination"]["カード"])
	assert.Equal(t, int64(50), got["examination"]["現金"])
	assert.Equal(t, int64(150), MatrixGrandTotal(got))
}

func TestPaymentMethodKeyFromID(t *testing.T) {
	t.Parallel()
	id := uint64(5)
	idToKey := map[uint64]string{5: "credit_card"}
	assert.Equal(t, "cash", PaymentMethodKeyFromID(nil, idToKey, "cash"))
	assert.Equal(t, "credit_card", PaymentMethodKeyFromID(&id, idToKey, "cash"))
	unknown := uint64(99)
	assert.Equal(t, "method_99", PaymentMethodKeyFromID(&unknown, idToKey, "cash"))
}
