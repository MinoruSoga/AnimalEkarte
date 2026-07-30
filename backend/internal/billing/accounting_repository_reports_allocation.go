package billing

import (
	"context"
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// allocationCategoryWeightRow is one billing×category weight row from SQL.
type allocationCategoryWeightRow struct {
	BillingID uint64
	Category  string
	Amount    int64
}

// allocationPaymentRow is one billing×payment_method amount.
type allocationPaymentRow struct {
	BillingID       uint64
	PaymentMethodID *uint64
	Amount          int64
}

// allocationRefundRow is a refund in the period (by refunded_at).
type allocationRefundRow struct {
	BillingID     uint64
	PaymentMethod *string // ENUM text; nullable
	Amount        int64
}

// allocationCategoryCountRow is accounting-distinct count per category.
type allocationCategoryCountRow struct {
	Category string
	Count    int64
}

// CategoryPaymentAllocationData is raw per-billing inputs for the pure allocation helper (#247).
type CategoryPaymentAllocationData struct {
	Weights             []allocationCategoryWeightRow
	Payments            []allocationPaymentRow
	Refunds             []allocationRefundRow
	RefundParentWeights []allocationCategoryWeightRow
	CategoryCounts      map[string]int64
}

// itemWeightSQL is DEC-16⑥ 明細金額: max(0, ROUND(unit_price*qty) - discount_amount).
// Tax/insurance/billing-level discount shrink payment net (the allocated amount), not weights.
const itemWeightSQL = `GREATEST(0, ROUND(bi.unit_price * bi.quantity::numeric) - bi.discount_amount)::bigint`

// GetCategoryPaymentAllocationData loads per-billing weights/payments/refunds/counts for a period.
// Shared by cash-register close and monthly category×payment matrix (#247 DEC-16⑥).
func (r *accountingRepository) GetCategoryPaymentAllocationData(
	ctx context.Context,
	clinicID uint64,
	periodStart, periodEnd time.Time,
) (*CategoryPaymentAllocationData, error) {
	cArgs := []any{clinicID, model.BillingStatusCompleted, periodStart.In(time.Local), periodEnd.In(time.Local)}
	completedCTE := completedBillingsCTE("id, clinic_id")

	var weights []allocationCategoryWeightRow
	if err := r.db.WithContext(ctx).Raw(
		completedCTE+`
		SELECT bi.billing_id,
		       bi.category::text AS category,
		       COALESCE(SUM(`+itemWeightSQL+`), 0) AS amount
		FROM billing_items bi
		WHERE bi.billing_id IN (SELECT id FROM completed_billings)
		  AND bi.deleted_at IS NULL
		GROUP BY bi.billing_id, bi.category
		`, cArgs...).Scan(&weights).Error; err != nil {
		return nil, apperrors.Wrap(err, "failed to load allocation category weights")
	}

	var payments []allocationPaymentRow
	if err := r.db.WithContext(ctx).Raw(
		completedCTE+`
		SELECT ps.billing_id,
		       ps.payment_method_id,
		       COALESCE(SUM(ps.amount), 0) AS amount
		FROM payment_splits ps
		JOIN completed_billings cb
		  ON cb.id = ps.billing_id
		 AND cb.clinic_id = ps.clinic_id
		GROUP BY ps.billing_id, ps.payment_method_id
		`, cArgs...).Scan(&payments).Error; err != nil {
		return nil, apperrors.Wrap(err, "failed to load allocation payments")
	}

	// Refunds by occurrence date (refunded_at) — DEC-16⑥ 発生日の負値.
	var refunds []allocationRefundRow
	if err := r.db.WithContext(ctx).Raw(`
		SELECT br.billing_id,
		       br.payment_method::text AS payment_method,
		       br.amount
		FROM billing_refunds br
		WHERE br.clinic_id = ?
		  AND br.refunded_at >= ?
		  AND br.refunded_at < ?
		`, clinicID, periodStart.In(time.Local), periodEnd.In(time.Local)).
		Scan(&refunds).Error; err != nil {
		return nil, apperrors.Wrap(err, "failed to load allocation refunds")
	}

	// Parent category weights for refunds whose billing is outside the completed period set.
	refundParentIDs := make([]uint64, 0)
	seenParent := make(map[uint64]struct{})
	completedSet := make(map[uint64]struct{}, len(weights)+len(payments))
	for _, w := range weights {
		completedSet[w.BillingID] = struct{}{}
	}
	for _, p := range payments {
		completedSet[p.BillingID] = struct{}{}
	}
	for _, ref := range refunds {
		if _, ok := completedSet[ref.BillingID]; ok {
			continue
		}
		if _, ok := seenParent[ref.BillingID]; ok {
			continue
		}
		seenParent[ref.BillingID] = struct{}{}
		refundParentIDs = append(refundParentIDs, ref.BillingID)
	}

	var refundParentWeights []allocationCategoryWeightRow
	if len(refundParentIDs) > 0 {
		var err error
		refundParentWeights, err = r.loadCategoryWeightsForBillings(ctx, clinicID, refundParentIDs)
		if err != nil {
			return nil, err
		}
	}

	var countRows []allocationCategoryCountRow
	if err := r.db.WithContext(ctx).Raw(
		completedCTE+`
		SELECT bi.category::text AS category,
		       COUNT(DISTINCT bi.billing_id) AS count
		FROM billing_items bi
		WHERE bi.billing_id IN (SELECT id FROM completed_billings)
		  AND bi.deleted_at IS NULL
		GROUP BY bi.category
		`, cArgs...).Scan(&countRows).Error; err != nil {
		return nil, apperrors.Wrap(err, "failed to load allocation category counts")
	}

	return &CategoryPaymentAllocationData{
		Weights:             weights,
		Payments:            payments,
		Refunds:             refunds,
		RefundParentWeights: refundParentWeights,
		CategoryCounts:      categoryCountMap(countRows),
	}, nil
}

func (r *accountingRepository) loadCategoryWeightsForBillings(
	ctx context.Context,
	clinicID uint64,
	billingIDs []uint64,
) ([]allocationCategoryWeightRow, error) {
	if len(billingIDs) == 0 {
		return nil, nil
	}
	var rows []allocationCategoryWeightRow
	if err := r.db.WithContext(ctx).Raw(`
		SELECT bi.billing_id,
		       bi.category::text AS category,
		       COALESCE(SUM(`+itemWeightSQL+`), 0) AS amount
		FROM billing_items bi
		JOIN billings b ON b.id = bi.billing_id AND b.clinic_id = ? AND b.deleted_at IS NULL
		WHERE bi.billing_id IN ?
		  AND bi.deleted_at IS NULL
		GROUP BY bi.billing_id, bi.category
		`, clinicID, billingIDs).Scan(&rows).Error; err != nil {
		return nil, apperrors.Wrap(err, "failed to load refund parent category weights")
	}
	return rows, nil
}

// BuildAllocationBillings assembles pure-helper inputs from SQL rows.
// methodKeyFn maps payment_method_id → allocation method key.
// refundMethodKeyFn maps refund payment_method ENUM text → method key.
func BuildAllocationBillings(
	data *CategoryPaymentAllocationData,
	methodKeyFn func(id *uint64) string,
	refundMethodKeyFn func(enum *string) string,
) []AllocationBillingInput {
	if data == nil {
		return nil
	}

	weightByBilling := make(map[uint64]map[string]int64)
	hasPeriodWeights := make(map[uint64]struct{})
	addWeight := func(billingID uint64, category string, amount int64) {
		if amount <= 0 {
			return
		}
		m, ok := weightByBilling[billingID]
		if !ok {
			m = make(map[string]int64)
			weightByBilling[billingID] = m
		}
		m[category] += amount
	}
	for _, w := range data.Weights {
		addWeight(w.BillingID, w.Category, w.Amount)
		hasPeriodWeights[w.BillingID] = struct{}{}
	}
	// Refund parents outside period completed set only (all categories for that billing).
	for _, w := range data.RefundParentWeights {
		if _, ok := hasPeriodWeights[w.BillingID]; ok {
			continue
		}
		addWeight(w.BillingID, w.Category, w.Amount)
	}

	payByBilling := make(map[uint64][]AllocationPayment)
	for _, p := range data.Payments {
		if p.Amount == 0 {
			continue
		}
		payByBilling[p.BillingID] = append(payByBilling[p.BillingID], AllocationPayment{
			MethodKey: methodKeyFn(p.PaymentMethodID),
			Amount:    p.Amount,
		})
	}
	for _, ref := range data.Refunds {
		if ref.Amount == 0 {
			continue
		}
		payByBilling[ref.BillingID] = append(payByBilling[ref.BillingID], AllocationPayment{
			MethodKey: refundMethodKeyFn(ref.PaymentMethod),
			Amount:    -ref.Amount,
		})
	}

	out := make([]AllocationBillingInput, 0, len(payByBilling))
	for id, pays := range payByBilling {
		srcW := weightByBilling[id]
		wCopy := make(map[string]int64, len(srcW))
		for k, v := range srcW {
			wCopy[k] = v
		}
		out = append(out, AllocationBillingInput{
			CategoryWeights: wCopy,
			Payments:        pays,
		})
	}
	return out
}

func categoryCountMap(rows []allocationCategoryCountRow) map[string]int64 {
	out := make(map[string]int64, len(rows))
	for _, r := range rows {
		out[r.Category] = r.Count
	}
	return out
}

// BuildSystemKeyMethodResolvers builds method key resolvers for snapshot/system_key consumers
// (cash-register category_breakdown).
func BuildSystemKeyMethodResolvers(payMethods []model.PaymentMethodMaster) (
	methodKeyFn func(*uint64) string,
	refundMethodKeyFn func(*string) string,
) {
	idToKey := make(map[uint64]string, len(payMethods))
	systemKeyToID := make(map[string]uint64, len(payMethods))
	for i := range payMethods {
		m := &payMethods[i]
		if m.SystemKey != nil {
			idToKey[m.ID] = *m.SystemKey
			systemKeyToID[*m.SystemKey] = m.ID
		} else {
			idToKey[m.ID] = "method_" + strconv.FormatUint(m.ID, 10)
		}
	}
	cashKey := PaymentMethodSystemKeys[model.PaymentMethodCash]

	methodKeyFn = func(id *uint64) string {
		return PaymentMethodKeyFromID(id, idToKey, cashKey)
	}
	refundMethodKeyFn = func(enum *string) string {
		if enum == nil || *enum == "" {
			return cashKey
		}
		if sk, ok := PaymentMethodSystemKeys[model.PaymentMethod(*enum)]; ok {
			return sk
		}
		// Unknown enum → keep as-is for visibility
		return *enum
	}
	return methodKeyFn, refundMethodKeyFn
}

// BuildNameMethodResolvers builds method key resolvers that use display names
// (preview UI / monthly matrix columns).
func BuildNameMethodResolvers(payMethods []model.PaymentMethodMaster) (
	methodKeyFn func(*uint64) string,
	refundMethodKeyFn func(*string) string,
) {
	idToName := make(map[uint64]string, len(payMethods))
	systemKeyToName := make(map[string]string, len(payMethods))
	for i := range payMethods {
		m := &payMethods[i]
		idToName[m.ID] = m.Name
		if m.SystemKey != nil {
			systemKeyToName[*m.SystemKey] = m.Name
		}
	}
	cashName := "現金"
	if name, ok := systemKeyToName[PaymentMethodSystemKeys[model.PaymentMethodCash]]; ok {
		cashName = name
	}

	methodKeyFn = func(id *uint64) string {
		if id == nil {
			return cashName
		}
		if name, ok := idToName[*id]; ok {
			return name
		}
		return "不明な支払方法(" + strconv.FormatUint(*id, 10) + ")"
	}
	refundMethodKeyFn = func(enum *string) string {
		if enum == nil || *enum == "" {
			return cashName
		}
		if sk, ok := PaymentMethodSystemKeys[model.PaymentMethod(*enum)]; ok {
			if name, ok := systemKeyToName[sk]; ok {
				return name
			}
		}
		switch model.PaymentMethod(*enum) {
		case model.PaymentMethodCash:
			return cashName
		case model.PaymentMethodCreditCard:
			return "クレジットカード"
		case model.PaymentMethodElectronicMoney:
			return "電子マネー"
		case model.PaymentMethodBankTransfer:
			return "銀行振込"
		default:
			return *enum
		}
	}
	return methodKeyFn, refundMethodKeyFn
}
