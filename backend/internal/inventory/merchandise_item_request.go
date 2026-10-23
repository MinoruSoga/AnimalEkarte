package inventory

import (
	"encoding/json"
	"net/url"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

type listMerchandiseItemsQuery struct {
	Category string
}

func newListMerchandiseItemsQuery(values url.Values) listMerchandiseItemsQuery {
	return listMerchandiseItemsQuery{Category: values.Get("category")}
}

// merchandiseVersionField は PATCH の省略可能な編集開始版（EMR-449 A.4）。
// 省略 = CAS なし（新設定を含まない旧 PATCH の従来互換）。明示送信は必ず検査し無視しない:
// null / 非整数 / 0 以下は 400。is_bill_discount_excluded 送信時は service が version 必須を検査する。
type merchandiseVersionField struct {
	set   bool
	value int64
}

func (f *merchandiseVersionField) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return apperrors.WrapInvalidInput("version に null は指定できません。省略するか1以上の整数を指定してください")
	}
	var value int64
	if err := json.Unmarshal(data, &value); err != nil {
		return apperrors.WrapInvalidInput(errMsgMerchandiseVersionInvalid)
	}
	if value < 1 {
		return apperrors.WrapInvalidInput(errMsgMerchandiseVersionInvalid)
	}
	f.set = true
	f.value = value
	return nil
}

// toServiceInput は未送信を nil（CAS なし）に変換する。
func (f merchandiseVersionField) toServiceInput() *int64 {
	if !f.set {
		return nil
	}
	value := f.value
	return &value
}

// createMerchandiseItemRequest は物販品作成のバインド struct
// tax_rate は EMR-530 の exact decimal 文字列契約。省略/null は service 層の既定 0.10。
// is_bill_discount_excluded は EMR-449 A01。省略は service が明示 false を保存する。
type createMerchandiseItemRequest struct {
	Name      string                `json:"name"       binding:"required,max=255"`
	Category  string                `json:"category"   binding:"required,oneof=food goods other"`
	UnitPrice int64                 `json:"unit_price" binding:"min=0"`
	TaxType   string                `json:"tax_type"   binding:"required,oneof=included excluded exempt"`
	TaxRate   merchandiseTaxRateDTO `json:"tax_rate"`
	IsActive  bool                  `json:"is_active"`
	SortOrder int                   `json:"sort_order"`

	// EMR-449: 省略 → service が明示 false を保存。明示 null は binding 層で 400。
	IsBillDiscountExcluded billDiscountExcludedField `json:"is_bill_discount_excluded"`
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

		IsBillDiscountExcluded: r.IsBillDiscountExcluded.toCreateInput(),
	}
}

// updateMerchandiseItemRequest は物販品更新のバインド struct（全フィールドポインタ型）
// tax_rate は省略/null で保持（*merchandiseTaxRateDTO は null で nil になるため省略と同じ扱い）。
// is_bill_discount_excluded は省略=変更なし、null=400、true/false=明示更新。
// version は省略=CAS なし、送信時は必ず照合（EMR-449 A.4）。
type updateMerchandiseItemRequest struct {
	Name      *string                `json:"name" binding:"omitempty,max=255"`
	Category  *string                `json:"category"   binding:"omitempty,oneof=food goods other"`
	UnitPrice *int64                 `json:"unit_price" binding:"omitempty,min=0"`
	TaxType   *string                `json:"tax_type"   binding:"omitempty,oneof=included excluded exempt"`
	TaxRate   *merchandiseTaxRateDTO `json:"tax_rate"`
	IsActive  *bool                  `json:"is_active"`
	SortOrder *int                   `json:"sort_order" binding:"omitempty,min=0"`

	// EMR-449: 省略 → 変更なし。明示 null は binding 層で 400。
	IsBillDiscountExcluded billDiscountExcludedField `json:"is_bill_discount_excluded"`
	// EMR-449: 省略可能な編集開始版。is_bill_discount_excluded 送信時は service が必須化する。
	Version merchandiseVersionField `json:"version"`
}

func (r updateMerchandiseItemRequest) toServiceInput() *UpdateMerchandiseItemInput {
	input := &UpdateMerchandiseItemInput{
		Name:      r.Name,
		Category:  r.Category,
		UnitPrice: r.UnitPrice,
		TaxType:   r.TaxType,
		IsActive:  r.IsActive,
		SortOrder: r.SortOrder,

		IsBillDiscountExcluded: r.IsBillDiscountExcluded.toUpdateInput(),
		Version:                r.Version.toServiceInput(),
	}
	if r.TaxRate != nil {
		v := r.TaxRate.value
		input.TaxRate = &v
	}
	return input
}
