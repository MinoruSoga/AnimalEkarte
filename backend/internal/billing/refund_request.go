package billing

// refund_request.go — BE9-2C B③: accounting_request/response.go から返金系のみ分離移動。
// 返金の応答 DTO（RefundResponse）は、返金行の保存先（billing_refunds）の退役で返せる行が無くなったため削除した。held for EMR-506。

// createRefundRequest は返金登録リクエスト。
type createRefundRequest struct {
	Amount int64  `json:"amount" binding:"required,min=1"`
	Reason string `json:"reason" binding:"max=500"`
	// PaymentMethod は返金先の支払手段（任意・ENUM・#60）。会計 payment_splits.method と同体系。
	// 混在支払いの方法別返金上限(Phase 2)に使う。
	PaymentMethod *string `json:"payment_method" binding:"omitempty,oneof=cash credit_card electronic_money bank_transfer"`
}
