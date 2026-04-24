package handler

// createInventoryRequest は在庫アイテム作成リクエスト。
type createInventoryRequest struct {
	Name          string  `json:"name"            binding:"required"`
	Category      string  `json:"category"        binding:"required,oneof=medicine consumable food other"`
	Quantity      int     `json:"quantity"`
	Unit          string  `json:"unit"            binding:"required"`
	MinStockLevel int     `json:"min_stock_level"`
	Location      string  `json:"location"`
	ExpiryDate    *string `json:"expiry_date"`
	Supplier      string  `json:"supplier"`
	LastRestocked *string `json:"last_restocked"`
	Status        string  `json:"status"          binding:"omitempty,oneof=sufficient low out_of_stock"`
}

// updateInventoryRequest は在庫アイテム更新リクエスト。
type updateInventoryRequest struct {
	Name          *string `json:"name"`
	Category      *string `json:"category"        binding:"omitempty,oneof=medicine consumable food other"`
	Quantity      *int    `json:"quantity"`
	Unit          *string `json:"unit"`
	MinStockLevel *int    `json:"min_stock_level"`
	Location      *string `json:"location"`
	ExpiryDate    *string `json:"expiry_date"`
	Supplier      *string `json:"supplier"`
	LastRestocked *string `json:"last_restocked"`
	Status        *string `json:"status"          binding:"omitempty,oneof=sufficient low out_of_stock"`
}
