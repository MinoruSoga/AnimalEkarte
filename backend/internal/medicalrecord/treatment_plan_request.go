package medicalrecord

type createTreatmentPlanRequest struct {
	TreatmentContent string  `json:"treatment_content" binding:"required,max=1000"`
	Memo             string  `json:"memo"              binding:"max=1000"`
	IsInsurance      bool    `json:"is_insurance"`
	UnitPrice        int64   `json:"unit_price" binding:"min=0"`
	Quantity         float64 `json:"quantity" binding:"required,gt=0"`
	DiscountRate     float64 `json:"discount_rate" binding:"min=0,max=100"`
	DiscountAmount   int64   `json:"discount_amount" binding:"min=0"`
	// Subtotal is accepted for backward-compatible JSON but ignored server-side (MRD-04).
	Subtotal  int64 `json:"subtotal"`
	SortOrder int   `json:"sort_order"`
}

func (r *createTreatmentPlanRequest) toServiceInput() *CreateTreatmentPlanInput {
	return &CreateTreatmentPlanInput{
		TreatmentContent: r.TreatmentContent,
		Memo:             r.Memo,
		IsInsurance:      r.IsInsurance,
		UnitPrice:        r.UnitPrice,
		Quantity:         r.Quantity,
		DiscountRate:     r.DiscountRate,
		DiscountAmount:   r.DiscountAmount,
		Subtotal:         r.Subtotal,
		SortOrder:        r.SortOrder,
	}
}

type updateTreatmentPlanRequest struct {
	TreatmentContent *string  `json:"treatment_content" binding:"omitempty,max=1000"`
	Memo             *string  `json:"memo"              binding:"omitempty,max=1000"`
	IsInsurance      *bool    `json:"is_insurance"`
	UnitPrice        *int64   `json:"unit_price" binding:"omitempty,min=0"`
	Quantity         *float64 `json:"quantity" binding:"omitempty,gt=0"`
	DiscountRate     *float64 `json:"discount_rate" binding:"omitempty,min=0,max=100"`
	DiscountAmount   *int64   `json:"discount_amount" binding:"omitempty,min=0"`
	// Subtotal is accepted for backward-compatible JSON but ignored server-side (MRD-04).
	Subtotal  *int64 `json:"subtotal"`
	SortOrder *int   `json:"sort_order"`
}

func (r updateTreatmentPlanRequest) toServiceInput() *UpdateTreatmentPlanInput {
	return &UpdateTreatmentPlanInput{
		TreatmentContent: r.TreatmentContent,
		Memo:             r.Memo,
		IsInsurance:      r.IsInsurance,
		UnitPrice:        r.UnitPrice,
		Quantity:         r.Quantity,
		DiscountRate:     r.DiscountRate,
		DiscountAmount:   r.DiscountAmount,
		Subtotal:         r.Subtotal,
		SortOrder:        r.SortOrder,
	}
}
