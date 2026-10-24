package inventory

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type inventoryResponse struct {
	ID            uint64     `json:"id"`
	ClinicID      uint64     `json:"clinic_id"`
	Name          string     `json:"name"`
	Category      string     `json:"category"`
	Quantity      *int64     `json:"quantity"`
	Unit          string     `json:"unit"`
	MinStockLevel *int       `json:"min_stock_level"`
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
		// 日付フィールドは canonical 規約 (BE-refactor.md R2-1) に従い localTimePtr で
		// time.Local へ変換する。pet.BirthDate/NeuteredDate/LastVisit と同一パターン。
		ExpiryDate:    httpapi.LocalTimePtr(item.ExpiryDate),
		Supplier:      item.Supplier,
		LastRestocked: httpapi.LocalTimePtr(item.LastRestocked),
		// SD-4 決裁A（q&a.html SD-4）: status は保存値を信頼せず、読み取りのたびに
		// quantity/min_stock_level から導出する（保存列は DROP 済み）。二重管理を避けるため
		// 唯一の判定ロジックは model.DeriveInventoryStatus に一本化する（inventoryStatusText 経由）。
		// quantity / min_stock_level が NULL（未記録）の品目は導出できないため空文字（unknown）にする。
		Status:    inventoryStatusText(item),
		CreatedAt: httpapi.LocalTime(item.CreatedAt),
		UpdatedAt: httpapi.LocalTime(item.UpdatedAt),
	}
}

// inventoryStatusText は quantity / min_stock_level から表示 status を導出する（SD-4: 唯一の判定は model.DeriveInventoryStatus）。
// どちらかが NULL（未記録）の品目は導出できないため空文字（unknown）にし、既定値で補完しない。
func inventoryStatusText(item *model.InventoryItem) string {
	if item.Quantity == nil || item.MinStockLevel == nil {
		return ""
	}
	return string(model.DeriveInventoryStatus(int(*item.Quantity), *item.MinStockLevel)) //nolint:gosec // int は 64bit のため bigint を保持できる
}
