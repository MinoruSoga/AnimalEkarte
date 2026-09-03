package billing

type createPaymentMethodRequest struct {
	Name         string `json:"name"          binding:"required,max=255"`
	DisplayOrder int    `json:"display_order"`
}

func (r createPaymentMethodRequest) toServiceInput() *CreatePaymentMethodMasterInput {
	return &CreatePaymentMethodMasterInput{
		Name:         r.Name,
		DisplayOrder: r.DisplayOrder,
	}
}

type updatePaymentMethodRequest struct {
	Name         *string `json:"name" binding:"omitempty,max=255"`
	DisplayOrder *int    `json:"display_order"`
	IsActive     *bool   `json:"is_active"`
}

func (r updatePaymentMethodRequest) toServiceInput() *UpdatePaymentMethodMasterInput {
	return &UpdatePaymentMethodMasterInput{
		Name:         r.Name,
		DisplayOrder: r.DisplayOrder,
		IsActive:     r.IsActive,
	}
}
