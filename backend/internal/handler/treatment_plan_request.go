package handler

type createTreatmentPlanRequest struct {
	TreatmentContent string  `json:"treatment_content" binding:"required"`
	Memo             string  `json:"memo"`
	Insurance        bool    `json:"insurance"`
	UnitPrice        float64 `json:"unit_price"`
	Quantity         float64 `json:"quantity"`
	DiscountRate     float64 `json:"discount_rate"`
	DiscountAmount   float64 `json:"discount_amount"`
	Subtotal         float64 `json:"subtotal"`
	SortOrder        int     `json:"sort_order"`
}

type updateTreatmentPlanRequest struct {
	TreatmentContent *string  `json:"treatment_content"`
	Memo             *string  `json:"memo"`
	Insurance        *bool    `json:"insurance"`
	UnitPrice        *float64 `json:"unit_price"`
	Quantity         *float64 `json:"quantity"`
	DiscountRate     *float64 `json:"discount_rate"`
	DiscountAmount   *float64 `json:"discount_amount"`
	Subtotal         *float64 `json:"subtotal"`
	SortOrder        *int     `json:"sort_order"`
}
