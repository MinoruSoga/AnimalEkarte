package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type inventoryResponse struct {
	ID            uint64     `json:"id"`
	ClinicID      uint64     `json:"clinic_id"`
	Name          string     `json:"name"`
	Category      string     `json:"category"`
	Quantity      int        `json:"quantity"`
	Unit          string     `json:"unit"`
	MinStockLevel int        `json:"min_stock_level"`
	Location      string     `json:"location"`
	ExpiryDate    *time.Time `json:"expiry_date,omitempty"`
	Supplier      string     `json:"supplier"`
	LastRestocked *time.Time `json:"last_restocked,omitempty"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func toInventoryResponse(item *model.InventoryItem) inventoryResponse {
	return inventoryResponse{
		ID:            item.ID,
		ClinicID:      item.ClinicID,
		Name:          item.Name,
		Category:      string(item.Category),
		Quantity:      item.Quantity,
		Unit:          item.Unit,
		MinStockLevel: item.MinStockLevel,
		Location:      item.Location,
		ExpiryDate:    item.ExpiryDate,
		Supplier:      item.Supplier,
		LastRestocked: item.LastRestocked,
		Status:        string(item.Status),
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
	}
}
