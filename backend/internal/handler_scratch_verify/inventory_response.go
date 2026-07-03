package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type inventoryResponse struct {
	ID            uint64    `json:"id"`
	ClinicID      uint64    `json:"clinic_id"`
	Name          string    `json:"name"`
	Category      string    `json:"category"`
	Quantity      int       `json:"quantity"`
	Unit          string    `json:"unit"`
	MinStockLevel int       `json:"min_stock_level"`
	Location      string    `json:"location"`
	ExpiryDate    *string   `json:"expiry_date,omitempty"`
	Supplier      string    `json:"supplier"`
	LastRestocked *string   `json:"last_restocked,omitempty"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// BE-refactor.md R2-1 (D3): expiry_date/last_restocked は openapi で format:date（日付のみ）と
// 宣言されており、gorm 側も type:date（時刻を持たない）。従来は *time.Time をそのまま JSON 化しており
// RFC3339 datetime（`…T00:00:00Z`）として配信され、宣言と実体が乖離していた（6/30 39ddaf11 で
// localTimePtr は datetime のままのため不適合と判明・date-only 文字列化が正しい修正と特定済み）。
// checkup/shift 等の date-only フィールドと同じ `.In(time.Local).Format("2006-01-02")` 規約に統一する。
func toInventoryResponse(item *model.InventoryItem) inventoryResponse {
	resp := inventoryResponse{
		ID:            item.ID,
		ClinicID:      item.ClinicID,
		Name:          item.Name,
		Category:      string(item.Category),
		Quantity:      item.Quantity,
		Unit:          item.Unit,
		MinStockLevel: item.MinStockLevel,
		Location:      item.Location,
		Supplier:      item.Supplier,
		Status:        string(item.Status),
		CreatedAt:     localTime(item.CreatedAt),
		UpdatedAt:     localTime(item.UpdatedAt),
	}
	if item.ExpiryDate != nil {
		s := item.ExpiryDate.In(time.Local).Format("2006-01-02")
		resp.ExpiryDate = &s
	}
	if item.LastRestocked != nil {
		s := item.LastRestocked.In(time.Local).Format("2006-01-02")
		resp.LastRestocked = &s
	}
	return resp
}
