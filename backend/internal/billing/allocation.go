package billing

import (
	"fmt"
	"sort"
)

// allocation.go — #247 DEC-16⑥ pure allocation helper.
//
// Contract (DEC-16⑥ / DEC-40 Q2):
//   - matrix 総額 = 支払実額基準（割引適用後・締め合計と一致）
//   - 割引は明細金額比例で配賦（category weight = 明細金額、payment を配分）
//   - 返金は発生日の負値として payment 側に負額を載せる（pre-net 禁止・行単位配賦）
//   - 端数は最大剰余法で行合計=列合計=総計を円単位保存
//   - 「件数」= 会計 distinct（本 helper は金額のみ。件数は呼び出し側）
//
// Dual amount definitions (do not collapse):
//   - 締め合計 / matrix grand = Σpayments(completed_at∈period) − Σrefunds(refunded_at∈period)
//   - KPI NetAmount may attach refunds to parent completed_at period instead; that figure is
//     NOT the matrix total. See docs/spec/screens/29-closing-aggregation.md §DEC-16⑥.
//
// Do NOT use period-wide payment ratio (old buildCategoryBreakdown bug).

// AllocationPayment is one payment (or refund as negative) attributed to a method key.
type AllocationPayment struct {
	// MethodKey is a stable key for the payment method column
	// (system_key, "method_<id>", or cash key for NULL legacy).
	MethodKey string
	Amount    int64
}

// AllocationBillingInput is one billing's category weights and payments for allocation.
type AllocationBillingInput struct {
	// CategoryWeights maps item category → 明細金額 weight (after item discount, yen).
	// Zero/negative weights are ignored. If all weights are 0 but payments remain,
	// the full payment lands on FallbackCategory (default "other").
	CategoryWeights map[string]int64
	Payments        []AllocationPayment
}

// LargestRemainderDistribute distributes total yen across weights so sum(result) == total.
// Weights must be non-negative; zero-weight slots receive 0 unless every weight is 0,
// in which case the entire total is assigned to index 0 (when len(weights) > 0).
// Negative totals (refunds) are supported: floors and remainders track the sign of total.
func LargestRemainderDistribute(total int64, weights []int64) []int64 {
	n := len(weights)
	out := make([]int64, n)
	if n == 0 || total == 0 {
		return out
	}

	var weightSum int64
	for _, w := range weights {
		if w > 0 {
			weightSum += w
		}
	}
	if weightSum == 0 {
		out[0] = total
		return out
	}

	type remainder struct {
		idx int
		rem int64 // absolute remainder for ranking
	}
	remainders := make([]remainder, 0, n)
	var assigned int64

	// Use integer arithmetic: floor = total * w / weightSum (trunc toward zero for positive;
	// for negative total, Go truncates toward zero as well — then we fix via remainder pass).
	for i, w := range weights {
		if w <= 0 {
			continue
		}
		share := total * w / weightSum
		out[i] = share
		assigned += share
		// remainder magnitude for largest-remainder ranking:
		// |total|*w - |share|*weightSum  (works for both signs when share = total*w/weightSum)
		absTotal := total
		if absTotal < 0 {
			absTotal = -absTotal
		}
		absShare := share
		if absShare < 0 {
			absShare = -absShare
		}
		rem := absTotal*w - absShare*weightSum
		remainders = append(remainders, remainder{idx: i, rem: rem})
	}

	leftover := total - assigned
	if leftover == 0 {
		return out
	}

	// Sort by remainder desc, then index asc for determinism.
	sort.SliceStable(remainders, func(i, j int) bool {
		if remainders[i].rem != remainders[j].rem {
			return remainders[i].rem > remainders[j].rem
		}
		return remainders[i].idx < remainders[j].idx
	})

	step := int64(1)
	if leftover < 0 {
		step = -1
		leftover = -leftover
	}
	for k := int64(0); k < leftover && int(k) < len(remainders); k++ {
		out[remainders[k].idx] += step
	}
	return out
}

// AllocateBillingPayments allocates one billing's payments onto categories by weight
// using largest remainder per payment method. Returns map[category]map[methodKey]amount.
//
// Conservation (per billing):
//   - column sum for method M == payment amount for M
//   - grand sum of cells == sum of payment amounts
func AllocateBillingPayments(in AllocationBillingInput) map[string]map[string]int64 {
	result := make(map[string]map[string]int64)
	if len(in.Payments) == 0 {
		return result
	}

	cats := make([]string, 0, len(in.CategoryWeights))
	weights := make([]int64, 0, len(in.CategoryWeights))
	for cat, w := range in.CategoryWeights {
		if w > 0 {
			cats = append(cats, cat)
			weights = append(weights, w)
		}
	}
	// Deterministic category order for largest-remainder tie-breaks.
	sort.Strings(cats)
	// Rebuild weights in sorted category order.
	if len(cats) > 0 {
		weights = weights[:0]
		for _, cat := range cats {
			weights = append(weights, in.CategoryWeights[cat])
		}
	}

	fallback := fallbackAllocationCategory()
	if len(cats) == 0 {
		cats = []string{fallback}
		weights = []int64{1}
	}

	for _, pay := range in.Payments {
		if pay.MethodKey == "" || pay.Amount == 0 {
			continue
		}
		shares := LargestRemainderDistribute(pay.Amount, weights)
		for i, cat := range cats {
			if shares[i] == 0 {
				continue
			}
			byMethod, ok := result[cat]
			if !ok {
				byMethod = make(map[string]int64)
				result[cat] = byMethod
			}
			byMethod[pay.MethodKey] += shares[i]
		}
	}
	return result
}

// AggregateCategoryPaymentMatrix allocates each billing and sums cells.
// Conservation across billings: grand total == sum of all payment amounts
// (including negative refunds folded into Payments).
func AggregateCategoryPaymentMatrix(billings []AllocationBillingInput) map[string]map[string]int64 {
	agg := make(map[string]map[string]int64)
	for _, b := range billings {
		part := AllocateBillingPayments(b)
		for cat, byMethod := range part {
			dst, ok := agg[cat]
			if !ok {
				dst = make(map[string]int64)
				agg[cat] = dst
			}
			for method, amount := range byMethod {
				dst[method] += amount
			}
		}
	}
	return agg
}

// MatrixGrandTotal returns the sum of all cells.
func MatrixGrandTotal(matrix map[string]map[string]int64) int64 {
	var sum int64
	for _, byMethod := range matrix {
		for _, amount := range byMethod {
			sum += amount
		}
	}
	return sum
}

// MatrixColumnTotals returns methodKey → column sum.
func MatrixColumnTotals(matrix map[string]map[string]int64) map[string]int64 {
	cols := make(map[string]int64)
	for _, byMethod := range matrix {
		for method, amount := range byMethod {
			cols[method] += amount
		}
	}
	return cols
}

// MatrixRowTotals returns category → row sum.
func MatrixRowTotals(matrix map[string]map[string]int64) map[string]int64 {
	rows := make(map[string]int64)
	for cat, byMethod := range matrix {
		var sum int64
		for _, amount := range byMethod {
			sum += amount
		}
		rows[cat] = sum
	}
	return rows
}

// RemapMatrixMethodKeys rebuilds a matrix with transformed method keys.
// transform must return the new key; empty string drops the column contribution
// (should not happen in production paths).
func RemapMatrixMethodKeys(matrix map[string]map[string]int64, transform func(methodKey string) string) map[string]map[string]int64 {
	out := make(map[string]map[string]int64, len(matrix))
	for cat, byMethod := range matrix {
		dst := make(map[string]int64, len(byMethod))
		for key, amount := range byMethod {
			newKey := transform(key)
			if newKey == "" {
				continue
			}
			dst[newKey] += amount
		}
		if len(dst) > 0 {
			out[cat] = dst
		}
	}
	return out
}

// PaymentMethodKeyFromID resolves a payment_method_id to the snapshot/system key used
// by cash-register category_breakdown (system_key, method_N, or cash for NULL).
func PaymentMethodKeyFromID(id *uint64, idToKey map[uint64]string, cashKey string) string {
	if id == nil {
		return cashKey
	}
	if key, ok := idToKey[*id]; ok {
		return key
	}
	return fmt.Sprintf("method_%d", *id)
}

// PaymentMethodIDKey is a neutral allocation key based on payment_method_id
// ("null" for legacy NULL cash). Used while aggregating before display remapping.
func PaymentMethodIDKey(id *uint64) string {
	if id == nil {
		return "null"
	}
	return fmt.Sprintf("%d", *id)
}

func fallbackAllocationCategory() string {
	// model.ItemCategoryOther as string without importing cycles — "other" is stable enum.
	return "other"
}
