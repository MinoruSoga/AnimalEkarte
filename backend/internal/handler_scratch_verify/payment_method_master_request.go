package handler

import "github.com/animal-ekarte/backend/internal/service"

type createPaymentMethodRequest struct {
	Name         string `json:"name"          binding:"required"`
	DisplayOrder int    `json:"display_order"`
}

func (r createPaymentMethodRequest) toServiceInput() *service.CreatePaymentMethodMasterInput {
	return &service.CreatePaymentMethodMasterInput{
		Name:         r.Name,
		DisplayOrder: r.DisplayOrder,
	}
}

type updatePaymentMethodRequest struct {
	Name         *string `json:"name"`
	DisplayOrder *int    `json:"display_order"`
	IsActive     *bool   `json:"is_active"`
}

func (r updatePaymentMethodRequest) toServiceInput() *service.UpdatePaymentMethodMasterInput {
	return &service.UpdatePaymentMethodMasterInput{
		Name:         r.Name,
		DisplayOrder: r.DisplayOrder,
		IsActive:     r.IsActive,
	}
}
