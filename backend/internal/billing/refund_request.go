package billing

// refund_request.go — BE9-2C B③: accounting_request/response.go から返金系のみ分離移動。

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// createRefundRequest は返金登録リクエスト。
type createRefundRequest struct {
	Amount int64  `json:"amount" binding:"required,min=1"`
	Reason string `json:"reason" binding:"max=500"`
	// PaymentMethod は返金先の支払手段（任意・ENUM・#60）。会計 payment_splits.method と同体系。
	// 混在支払いの方法別返金上限(Phase 2)に使う。
	PaymentMethod *string `json:"payment_method" binding:"omitempty,oneof=cash credit_card electronic_money bank_transfer"`
}

type RefundResponse struct {
	ID             uint64    `json:"id"`
	ClinicID       uint64    `json:"clinic_id"`
	BillingID      uint64    `json:"billing_id"`
	Amount         int64     `json:"amount"`
	Reason         string    `json:"reason"`
	RefundedBy     *uint64   `json:"refunded_by"`
	RefundedByName string    `json:"refunded_by_name"`
	PaymentMethod  *string   `json:"payment_method,omitempty"`
	RefundedAt     time.Time `json:"refunded_at"`
	CreatedAt      time.Time `json:"created_at"`
}

func ToRefundResponse(r *model.BillingRefund) RefundResponse {
	var staffName string
	if r.RefundedByStaff != nil {
		staffName = r.RefundedByStaff.Name
	}
	var paymentMethod *string
	if r.PaymentMethod != nil {
		pm := string(*r.PaymentMethod)
		paymentMethod = &pm
	}
	return RefundResponse{
		ID:             r.ID,
		ClinicID:       r.ClinicID,
		BillingID:      r.BillingID,
		Amount:         r.Amount,
		Reason:         r.Reason,
		RefundedBy:     r.RefundedBy,
		RefundedByName: staffName,
		PaymentMethod:  paymentMethod,
		RefundedAt:     httpapi.LocalTime(r.RefundedAt),
		CreatedAt:      httpapi.LocalTime(r.CreatedAt),
	}
}
