package billing

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type PaymentMethodResponse struct {
	ID           uint64    `json:"id"`
	ClinicID     uint64    `json:"clinic_id"`
	Name         string    `json:"name"`
	DisplayOrder int       `json:"display_order"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func ToPaymentMethodResponse(m *model.PaymentMethodMaster) PaymentMethodResponse {
	return PaymentMethodResponse{
		ID:           m.ID,
		ClinicID:     m.ClinicID,
		Name:         m.Name,
		DisplayOrder: m.DisplayOrder,
		IsActive:     m.IsActive,
		CreatedAt:    httpapi.LocalTime(m.CreatedAt),
		UpdatedAt:    httpapi.LocalTime(m.UpdatedAt),
	}
}
