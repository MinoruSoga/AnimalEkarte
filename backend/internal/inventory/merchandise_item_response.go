package inventory

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type merchandiseItemResponse struct {
	ID        uint64             `json:"id"`
	ClinicID  uint64             `json:"clinic_id"`
	Name      string             `json:"name"`
	Category  model.ItemCategory `json:"category"`
	UnitPrice int64              `json:"unit_price"`
	TaxType   model.TaxType      `json:"tax_type"`
	TaxRate   float64            `json:"tax_rate"`
	IsActive  bool               `json:"is_active"`
	SortOrder int                `json:"sort_order"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

func toMerchandiseItemResponse(m *model.MerchandiseItem) merchandiseItemResponse {
	return merchandiseItemResponse{
		ID:        m.ID,
		ClinicID:  m.ClinicID,
		Name:      m.Name,
		Category:  m.Category,
		UnitPrice: m.UnitPrice,
		TaxType:   m.TaxType,
		TaxRate:   m.TaxRate,
		IsActive:  m.IsActive,
		SortOrder: m.SortOrder,
		CreatedAt: httpapi.LocalTime(m.CreatedAt),
		UpdatedAt: httpapi.LocalTime(m.UpdatedAt),
	}
}
