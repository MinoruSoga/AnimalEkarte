package inventory

import "net/url"

type listMerchandiseItemsQuery struct {
	Category string
}

func newListMerchandiseItemsQuery(values url.Values) listMerchandiseItemsQuery {
	return listMerchandiseItemsQuery{Category: values.Get("category")}
}

// createMerchandiseItemRequest は物販品作成のバインド struct
type createMerchandiseItemRequest struct {
	Name      string   `json:"name"       binding:"required"`
	Category  string   `json:"category"   binding:"required,oneof=food goods other"`
	UnitPrice int64    `json:"unit_price" binding:"min=0"`
	TaxType   string   `json:"tax_type"   binding:"required,oneof=included excluded exempt"`
	TaxRate   *float64 `json:"tax_rate"   binding:"omitempty,min=0,max=1"`
	IsActive  bool     `json:"is_active"`
	SortOrder int      `json:"sort_order"`
}

func (r *createMerchandiseItemRequest) toServiceInput() *CreateMerchandiseItemInput {
	return &CreateMerchandiseItemInput{
		Name:      r.Name,
		Category:  r.Category,
		UnitPrice: r.UnitPrice,
		TaxType:   r.TaxType,
		TaxRate:   r.TaxRate,
		IsActive:  r.IsActive,
		SortOrder: r.SortOrder,
	}
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

func (r updateMerchandiseItemRequest) toServiceInput() *UpdateMerchandiseItemInput {
	return &UpdateMerchandiseItemInput{
		Name:      r.Name,
		Category:  r.Category,
		UnitPrice: r.UnitPrice,
		TaxType:   r.TaxType,
		TaxRate:   r.TaxRate,
		IsActive:  r.IsActive,
		SortOrder: r.SortOrder,
	}
}
