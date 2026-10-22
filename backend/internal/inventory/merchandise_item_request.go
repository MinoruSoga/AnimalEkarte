package inventory

import "net/url"

type listMerchandiseItemsQuery struct {
	Category string
}

func newListMerchandiseItemsQuery(values url.Values) listMerchandiseItemsQuery {
	return listMerchandiseItemsQuery{Category: values.Get("category")}
}

// createMerchandiseItemRequest は物販品作成のバインド struct
// tax_rate は EMR-530 の exact decimal 文字列契約。省略/null は service 層の既定 0.10。
type createMerchandiseItemRequest struct {
	Name      string                `json:"name"       binding:"required,max=255"`
	Category  string                `json:"category"   binding:"required,oneof=food goods other"`
	UnitPrice int64                 `json:"unit_price" binding:"min=0"`
	TaxType   string                `json:"tax_type"   binding:"required,oneof=included excluded exempt"`
	TaxRate   merchandiseTaxRateDTO `json:"tax_rate"`
	IsActive  bool                  `json:"is_active"`
	SortOrder int                   `json:"sort_order"`
}

func (r *createMerchandiseItemRequest) toServiceInput() *CreateMerchandiseItemInput {
	return &CreateMerchandiseItemInput{
		Name:      r.Name,
		Category:  r.Category,
		UnitPrice: r.UnitPrice,
		TaxType:   r.TaxType,
		TaxRate:   r.TaxRate.optional(),
		IsActive:  r.IsActive,
		SortOrder: r.SortOrder,
	}
}

// updateMerchandiseItemRequest は物販品更新のバインド struct（全フィールドポインタ型）
// tax_rate は省略/null で保持（*merchandiseTaxRateDTO は null で nil になるため省略と同じ扱い）。
type updateMerchandiseItemRequest struct {
	Name      *string                 `json:"name" binding:"omitempty,max=255"`
	Category  *string                 `json:"category"   binding:"omitempty,oneof=food goods other"`
	UnitPrice *int64                  `json:"unit_price" binding:"omitempty,min=0"`
	TaxType   *string                 `json:"tax_type"   binding:"omitempty,oneof=included excluded exempt"`
	TaxRate   *merchandiseTaxRateDTO  `json:"tax_rate"`
	IsActive  *bool                   `json:"is_active"`
	SortOrder *int                    `json:"sort_order" binding:"omitempty,min=0"`
}

func (r updateMerchandiseItemRequest) toServiceInput() *UpdateMerchandiseItemInput {
	input := &UpdateMerchandiseItemInput{
		Name:      r.Name,
		Category:  r.Category,
		UnitPrice: r.UnitPrice,
		TaxType:   r.TaxType,
		IsActive:  r.IsActive,
		SortOrder: r.SortOrder,
	}
	if r.TaxRate != nil {
		v := r.TaxRate.value
		input.TaxRate = &v
	}
	return input
}
