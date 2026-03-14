package handler

// confirmBillingReviewRequest は会計医師確認のバインド struct
type confirmBillingReviewRequest struct {
	ConfirmedBy uint64 `json:"confirmed_by" binding:"required"`
	Memo        string `json:"memo"`
}

// returnBillingReviewRequest は会計差し戻しのバインド struct
type returnBillingReviewRequest struct {
	ReturnedBy   uint64 `json:"returned_by"   binding:"required"`
	ReturnReason string `json:"return_reason" binding:"required"`
	Memo         string `json:"memo"`
}
