package handler

type createPaymentMethodRequest struct {
	Name         string `json:"name"          binding:"required"`
	DisplayOrder int    `json:"display_order"`
}

type updatePaymentMethodRequest struct {
	Name         *string `json:"name"`
	DisplayOrder *int    `json:"display_order"`
	IsActive     *bool   `json:"is_active"`
}
