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
	// IsNonInsurance は品目の保険対象外設定（EMR-455）。未送信/null は新規初期値 true（対象外）。
	IsNonInsurance *bool `json:"is_non_insurance"`
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
		IsNonInsurance: r.IsNonInsurance,
	}
}

type updateVaccineRequest struct {
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
}

func (r updateVaccineRequest) toServiceInput() *UpdateVaccineInput {
	return &UpdateVaccineInput{
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
	}
}

// vaccineInsuranceSettingField は PATCH の is_non_insurance で「未送信」と「JSON null」を区別する。
// *bool だけでは両方 nil になるため、null を受理すると保存済み明示値を黙って維持/欠落扱いにしてしまう。
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
