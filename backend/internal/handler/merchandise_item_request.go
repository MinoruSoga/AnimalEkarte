package handler

// createMerchandiseItemRequest は物販品作成のバインド struct
type createMerchandiseItemRequest struct {
	Name      string  `json:"name"      binding:"required"`
	Category  string  `json:"category"  binding:"required,oneof=food goods other"`
	UnitPrice int64   `json:"unit_price"`
	TaxType   string  `json:"tax_type"`
	TaxRate   float64 `json:"tax_rate"`
	IsActive  bool    `json:"is_active"`
	SortOrder int     `json:"sort_order"`
}

// reorderMerchandiseItemRequest は物販品並び替えのバインド struct
type reorderMerchandiseItemRequest struct {
	IDs []uint64 `json:"ids" binding:"required,min=1"`
}

// updateMerchandiseItemRequest は物販品更新のバインド struct（全フィールドポインタ型）
type updateMerchandiseItemRequest struct {
	Name      *string  `json:"name"`
	Category  *string  `json:"category"`
	UnitPrice *int64   `json:"unit_price"`
	TaxType   *string  `json:"tax_type"`
	TaxRate   *float64 `json:"tax_rate"`
	IsActive  *bool    `json:"is_active"`
	SortOrder *int     `json:"sort_order"`
}
