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

// GetCategoryPaymentAllocationData は部門別・支払方法別の配賦（#247 DEC-16⑥）の入力を期間で読み出していた。held for EMR-509。
// 旧実装は completed 会計の payment_splits と billing_refunds（refunded_at）から支払方法別の実額を読んでいたが、
// いずれも廃止された（001_init.sql:10889,10919-10920）。後継は受領 (payment_funds)・配賦 (fund_consumption_events)・
// 返金の記録を実際の受領日で読む reader で、未接続のため、空の配賦を返さず固定エラー（422）で拒否する。
// 月次レポートと締めプレビューが共有する。部門別の重み付け（loadCategoryWeightsForBillings）と純粋な配賦関数は残す。
func (*accountingRepository) GetCategoryPaymentAllocationData(_ context.Context, _ uint64, _, _ time.Time) (*CategoryPaymentAllocationData, error) {
	return nil, apperrors.WrapFinanceConsumerPending("部門別・支払方法別の配賦は、受領・配賦・返金の記録の読取が未接続のため利用できません")
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
