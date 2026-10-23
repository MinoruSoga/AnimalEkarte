package medicalrecord

import (
	"encoding/json"
	"net/url"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

type listVaccinesQuery struct {
	Species string
}

func newListVaccinesQuery(values url.Values) listVaccinesQuery {
	return listVaccinesQuery{Species: values.Get("species")}
}

func (q listVaccinesQuery) toServiceFilter() *string {
	return optionalStringQueryFilter(q.Species)
}

type createVaccineRequest struct {
	Name        string  `json:"name"        binding:"required"`
	Price       *int64  `json:"price"`
	IsActive    bool    `json:"is_active"`
	Description string  `json:"description"`
	Species     string  `json:"species"     binding:"omitempty,oneof=dog cat both"`
	Interval    string  `json:"interval"`
	ParentID    *uint64 `json:"parent_id"`
	SortOrder   int     `json:"sort_order"`
	// IsNonInsurance は品目の保険対象外設定（EMR-455）。未送信=新規初期値 true（対象外）、
	// 明示 null は 400（省略と null を区別するため *bool は使わない）、true/false=明示値。
	IsNonInsurance vaccineInsuranceSettingField `json:"is_non_insurance"`

	// EMR-447 A01: 全体割引除外。未送信=新規初期値 false を service が明示保存、
	// 明示 null / 非 boolean は 400。
	IsBillDiscountExcluded billDiscountExcludedField `json:"is_bill_discount_excluded"`
}

func (r *createVaccineRequest) toServiceInput() *CreateVaccineInput {
	return &CreateVaccineInput{
		Name:           r.Name,
		Price:          r.Price,
		IsActive:       r.IsActive,
		Description:    r.Description,
		Species:        nilIfEmpty(r.Species),
		Interval:       r.Interval,
		ParentID:       r.ParentID,
		SortOrder:      r.SortOrder,
		IsNonInsurance: r.IsNonInsurance.value,

		IsBillDiscountExcluded: r.IsBillDiscountExcluded.toCreateInput(),
	}
}

type updateVaccineRequest struct {
	Version       int64   `json:"version" binding:"required,gt=0"`
	Name          *string `json:"name"`
	Price         *int64  `json:"price"`
	IsActive      *bool   `json:"is_active"`
	Description   *string `json:"description"`
	Species       *string `json:"species"      binding:"omitempty,oneof=dog cat both"`
	Interval      *string `json:"interval"`
	ParentID      *uint64 `json:"parent_id"`
	ClearParentID bool    `json:"clear_parent_id"`
	SortOrder     *int    `json:"sort_order"`
	// IsNonInsurance は未送信=変更なし、true/false=明示保存。null は明示値を欠落へ戻すため拒否する。
	IsNonInsurance vaccineInsuranceSettingField `json:"is_non_insurance"`

	// EMR-447 A01: 未送信=変更なし（NULL/false/true 保持）、null/非 boolean=400、true/false=明示 SET。
	IsBillDiscountExcluded billDiscountExcludedField `json:"is_bill_discount_excluded"`
}

func (r updateVaccineRequest) toServiceInput() *UpdateVaccineInput {
	return &UpdateVaccineInput{
		Version:        r.Version,
		Name:           r.Name,
		Price:          r.Price,
		IsActive:       r.IsActive,
		Description:    r.Description,
		Species:        r.Species,
		Interval:       r.Interval,
		ParentID:       r.ParentID,
		ClearParentID:  r.ClearParentID,
		SortOrder:      r.SortOrder,
		IsNonInsurance: r.IsNonInsurance.value,

		IsBillDiscountExcluded: r.IsBillDiscountExcluded.toUpdateInput(),
	}
}

// vaccineInsuranceSettingField は create/PATCH の is_non_insurance で「未送信」と「JSON null」を区別する。
// *bool だけでは両方 nil になり、create では省略と同一視されて新規初期値が適用され、
// PATCH では保存済み明示値を黙って維持/欠落扱いにしてしまう。いずれも明示 null は 400 で拒否する。
// UnmarshalJSON は値が存在する場合だけ呼ばれるので、未送信は value=nil のまま残る。
type vaccineInsuranceSettingField struct {
	value *bool
}

func (f *vaccineInsuranceSettingField) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return apperrors.WrapInvalidInput("is_non_insurance に null は指定できません")
	}
	var value bool
	if err := json.Unmarshal(data, &value); err != nil {
		return apperrors.WrapInvalidInput("is_non_insurance は真偽値で指定してください")
	}
	f.value = &value
	return nil
}
