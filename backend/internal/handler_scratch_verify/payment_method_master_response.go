package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type paymentMethodResponse struct {
	ID           uint64    `json:"id"`
	ClinicID     uint64    `json:"clinic_id"`
	Name         string    `json:"name"`
	DisplayOrder int       `json:"display_order"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func toPaymentMethodResponse(m *model.PaymentMethodMaster) paymentMethodResponse {
	return paymentMethodResponse{
		ID:           m.ID,
		ClinicID:     m.ClinicID,
		Name:         m.Name,
		DisplayOrder: m.DisplayOrder,
		IsActive:     m.IsActive,
		CreatedAt:    localTime(m.CreatedAt),
		UpdatedAt:    localTime(m.UpdatedAt),
	}
}
