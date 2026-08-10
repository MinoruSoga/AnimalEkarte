package billing

import (
	"github.com/animal-ekarte/backend/internal/model"
)

// unpaidAmountSQL は未収残高（円）の SQL 式。
// BUG-007: status=waiting の全額に加え、クレジット訂正などで「患者請求額 > 実支払額」になった
// completed 会計の差額も未収として扱う。
//
// 患者請求額 (patient_due) = payments.total_amount - insurance_amount - discount_amount
// （complete 時の billing_amount と同式。クレジット訂正後も medical/保険/割引は不変）。
// 実支払額 = payments.billing_amount（訂正後は split 合計へ再定義される）。
//
// payment 行が無い waiting は従来どおり billings.total_amount 全額を未収とする。
const unpaidAmountSQL = `CASE
	WHEN billings.status = 'waiting' AND payments.id IS NULL THEN billings.total_amount
	WHEN billings.status IN ('waiting', 'completed') AND payments.id IS NOT NULL THEN
		GREATEST(
			0,
			COALESCE(payments.total_amount, 0)
				- COALESCE(payments.insurance_amount, 0)
				- COALESCE(payments.discount_amount, 0)
				- COALESCE(payments.billing_amount, 0)
		)
	WHEN billings.status = 'waiting' THEN billings.total_amount
	ELSE 0
END`

// leftJoinPaymentsSQL は billings に対する payments の LEFT JOIN（soft-delete 除外）。
const leftJoinPaymentsSQL = `LEFT JOIN payments ON payments.billing_id = billings.id AND payments.deleted_at IS NULL`

// OutstandingAmount は会計 1 件の未収残高（円）を返す。SQL unpaidAmountSQL と同定義。
// Payments が Preload 済みなら先頭 payment を用いる。
func OutstandingAmount(b *model.Billing) int64 {
	if b == nil {
		return 0
	}
	switch b.Status {
	case model.BillingStatusWaiting:
		if len(b.Payments) == 0 {
			return b.TotalAmount
		}
		return patientOutstanding(&b.Payments[0])
	case model.BillingStatusCompleted:
		if len(b.Payments) == 0 {
			return 0
		}
		return patientOutstanding(&b.Payments[0])
	default:
		return 0
	}
}

func patientOutstanding(p *model.Payment) int64 {
	if p == nil {
		return 0
	}
	due := p.TotalAmount - p.InsuranceAmount - p.DiscountAmount
	residual := due - p.BillingAmount
	if residual < 0 {
		return 0
	}
	return residual
}
