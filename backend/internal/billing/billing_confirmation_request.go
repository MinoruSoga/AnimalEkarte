package billing

// confirmBillingConfirmationRequest は会計医師確認のバインド struct
type confirmBillingConfirmationRequest struct {
	ConfirmedBy uint64 `json:"confirmed_by" binding:"required"`
	Memo        string `json:"memo"`
}

func (r confirmBillingConfirmationRequest) toServiceInput() *ConfirmBillingConfirmationInput {
	return &ConfirmBillingConfirmationInput{
		ConfirmedBy: r.ConfirmedBy,
		Memo:        r.Memo,
	}
}

// returnBillingConfirmationRequest は会計差し戻しのバインド struct
type returnBillingConfirmationRequest struct {
	ReturnedBy   uint64 `json:"returned_by"   binding:"required"`
	ReturnReason string `json:"return_reason" binding:"required"`
	Memo         string `json:"memo"`
}

func (r returnBillingConfirmationRequest) toServiceInput() *ReturnBillingConfirmationInput {
	return &ReturnBillingConfirmationInput{
		ReturnedBy:   r.ReturnedBy,
		ReturnReason: r.ReturnReason,
		Memo:         r.Memo,
	}
}
