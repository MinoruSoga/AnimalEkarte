package handler

// createMerchandiseItemRequest は物販品作成のバインド struct
type createMerchandiseItemRequest struct {
	Name      string  `json:"name"      binding:"required"`
	Category  string  `json:"category"  binding:"required,oneof=food goods other"`
	UnitPrice int64   `json:"unit_price" binding:"min=0"`
	TaxType   string  `json:"tax_type"  binding:"required,oneof=included excluded exempt"`
	TaxRate   float64 `json:"tax_rate"  binding:"min=0,max=1"`
	IsActive  bool    `json:"is_active"`
	SortOrder int     `json:"sort_order"`
}

// updateMerchandiseItemRequest は物販品更新のバインド struct（全フィールドポインタ型）
type updateMerchandiseItemRequest struct {
	Name      *string  `json:"name"`
	Category  *string  `json:"category"   binding:"omitempty,oneof=food goods other"`
	UnitPrice *int64   `json:"unit_price" binding:"omitempty,min=0"`
	TaxType   *string  `json:"tax_type"   binding:"omitempty,oneof=included excluded exempt"`
	TaxRate   *float64 `json:"tax_rate"   binding:"omitempty,min=0,max=1"`
	IsActive  *bool    `json:"is_active"`
	SortOrder *int     `json:"sort_order" binding:"omitempty,min=0"`
}
