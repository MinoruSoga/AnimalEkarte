package handler

import "github.com/animal-ekarte/backend/internal/service"

type createTreatmentPlanRequest struct {
	TreatmentContent string  `json:"treatment_content" binding:"required"`
	Memo             string  `json:"memo"`
	IsInsurance      bool    `json:"is_insurance"`
	UnitPrice        int64   `json:"unit_price"`
	Quantity         float64 `json:"quantity"`
	DiscountRate     float64 `json:"discount_rate"`
	DiscountAmount   int64   `json:"discount_amount"`
	Subtotal         int64   `json:"subtotal"`
	SortOrder        int     `json:"sort_order"`
}

func (r *createTreatmentPlanRequest) toServiceInput() *service.CreateTreatmentPlanInput {
	return &service.CreateTreatmentPlanInput{
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
	TreatmentContent *string  `json:"treatment_content"`
	Memo             *string  `json:"memo"`
	IsInsurance      *bool    `json:"is_insurance"`
	UnitPrice        *int64   `json:"unit_price"`
	Quantity         *float64 `json:"quantity"`
	DiscountRate     *float64 `json:"discount_rate"`
	DiscountAmount   *int64   `json:"discount_amount"`
	Subtotal         *int64   `json:"subtotal"`
	SortOrder        *int     `json:"sort_order"`
}

func (r updateTreatmentPlanRequest) toServiceInput() *service.UpdateTreatmentPlanInput {
	return &service.UpdateTreatmentPlanInput{
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
