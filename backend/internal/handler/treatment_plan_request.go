package handler

type createTreatmentPlanRequest struct {
	TreatmentContent string  `json:"treatment_content" binding:"required"`
	Memo             string  `json:"memo"`
	IsInsurance       bool    `json:"is_insurance"`
	UnitPrice        int64   `json:"unit_price"`
	Quantity         float64 `json:"quantity"`
	DiscountRate     float64 `json:"discount_rate"`
	DiscountAmount   int64   `json:"discount_amount"`
	Subtotal         int64   `json:"subtotal"`
	SortOrder        int     `json:"sort_order"`
}

type updateTreatmentPlanRequest struct {
	TreatmentContent *string  `json:"treatment_content"`
	Memo             *string  `json:"memo"`
	IsInsurance       *bool    `json:"is_insurance"`
	UnitPrice        *int64   `json:"unit_price"`
	Quantity         *float64 `json:"quantity"`
	DiscountRate     *float64 `json:"discount_rate"`
	DiscountAmount   *int64   `json:"discount_amount"`
	Subtotal         *int64   `json:"subtotal"`
	SortOrder        *int     `json:"sort_order"`
}
