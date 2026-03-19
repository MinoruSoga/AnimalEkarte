package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type merchandiseItemResponse struct {
	ID        uint64    `json:"id"`
	ClinicID  uint64    `json:"clinic_id"`
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	UnitPrice int64     `json:"unit_price"`
	TaxRate   float64   `json:"tax_rate"`
	IsActive  bool      `json:"is_active"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toMerchandiseItemResponse(m *model.MerchandiseItem) merchandiseItemResponse {
	return merchandiseItemResponse{
		ID:        m.ID,
		ClinicID:  m.ClinicID,
		Name:      m.Name,
		Category:  string(m.Category),
		UnitPrice: m.UnitPrice,
		TaxRate:   m.TaxRate,
		IsActive:  m.IsActive,
		SortOrder: m.SortOrder,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func toMerchandiseItemResponseList(items []model.MerchandiseItem) []merchandiseItemResponse {
	list := make([]merchandiseItemResponse, 0, len(items))
	for i := range items {
		list = append(list, toMerchandiseItemResponse(&items[i]))
	}
	return list
}
