package medicalrecord

import (
	"encoding/json"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// EMR-456: is_non_insurance の wire 契約（null 契約の統一・チケット回答 2026-10-20）。
// POST: 省略 → service が新規初期値（保険対象外=true）を明示保存 / 明示 null → 400 / boolean → 保存。
// PATCH: 省略 → 変更なし / 明示 null → 400 / boolean → 保存。旧 NULL（未保存）読取の保持とは区別する。
// encoding/json は *bool/**bool に null を落とすと省略と区別できないため、専有のフィールド型で
// 「未送信」「明示 null」「boolean」を区別する（将来 454 の共有語彙へ集約する場合は 454 側で置き換える）。

const (
	errMsgCheckupTypeInsuranceNull       = "is_non_insurance に null は指定できません。省略するか true/false を指定してください"
	errMsgCheckupTypeInsuranceNotBoolean = "is_non_insurance には true/false を指定してください"
)

type checkupTypeInsuranceField struct {
	set   bool
	value bool
}

func (f *checkupTypeInsuranceField) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return apperrors.WrapInvalidInput(errMsgCheckupTypeInsuranceNull)
	}
	var value bool
	if err := json.Unmarshal(data, &value); err != nil {
		return apperrors.WrapInvalidInput(errMsgCheckupTypeInsuranceNotBoolean)
	}
	f.set = true
	f.value = value
	return nil
}

// toServiceInput は未送信を nil に変換する（POST=service が初期値 true を明示保存 / PATCH=変更なし）。
func (f checkupTypeInsuranceField) toServiceInput() *bool {
	if !f.set {
		return nil
	}
	value := f.value
	return &value
}

type createCheckupTypeRequest struct {
	Name        string  `json:"name"        binding:"required"`
	Price       *int64  `json:"price"`
	IsActive    bool    `json:"is_active"`
	Description string  `json:"description"`
	Interval    string  `json:"interval"`
	TargetAge   string  `json:"target_age"`
	ParentID    *uint64 `json:"parent_id"`
	SortOrder   int     `json:"sort_order"`
	// EMR-456: 省略 → service が新規初期値（保険対象外=true）を明示保存。明示 null は binding 層で 400。
	IsNonInsurance checkupTypeInsuranceField `json:"is_non_insurance"`
	// EMR-447 A01: 省略 → service が新規初期値 false を明示保存。明示 null / 非 boolean は 400。
	IsBillDiscountExcluded billDiscountExcludedField `json:"is_bill_discount_excluded"`
}

func (r *createCheckupTypeRequest) toServiceInput() *CreateCheckupTypeInput {
	return &CreateCheckupTypeInput{
		Name:           r.Name,
		Price:          r.Price,
		IsActive:       r.IsActive,
		Description:    r.Description,
		Interval:       r.Interval,
		TargetAge:      r.TargetAge,
		ParentID:       r.ParentID,
		SortOrder:      r.SortOrder,
		IsNonInsurance: r.IsNonInsurance.toServiceInput(),

		IsBillDiscountExcluded: r.IsBillDiscountExcluded.toCreateInput(),
	}
}

type updateCheckupTypeRequest struct {
	Name          *string `json:"name"`
	Price         *int64  `json:"price"`
	IsActive      *bool   `json:"is_active"`
	Description   *string `json:"description"`
	Interval      *string `json:"interval"`
	TargetAge     *string `json:"target_age"`
	ParentID      *uint64 `json:"parent_id"`
	ClearParentID bool    `json:"clear_parent_id"`
	SortOrder     *int    `json:"sort_order"`
	// EMR-456: 省略 → 変更なし。明示 null は binding 層で 400。NULL（未保存）へ戻す経路は持たない。
	IsNonInsurance checkupTypeInsuranceField `json:"is_non_insurance"`
	// EMR-447 A01: 省略 → 変更なし（NULL/false/true 保持）。明示 null / 非 boolean は 400。
	IsBillDiscountExcluded billDiscountExcludedField `json:"is_bill_discount_excluded"`
	// EMR-456: 楽観ロックの期待版。更新要求では必須（未指定は 400）。不一致は 409・書込なし。
	Version *int64 `json:"version" binding:"required,min=1"`
}

func (r updateCheckupTypeRequest) toServiceInput() *UpdateCheckupTypeInput {
	return &UpdateCheckupTypeInput{
		Name:           r.Name,
		Price:          r.Price,
		IsActive:       r.IsActive,
		Description:    r.Description,
		Interval:       r.Interval,
		TargetAge:      r.TargetAge,
		ParentID:       r.ParentID,
		ClearParentID:  r.ClearParentID,
		SortOrder:      r.SortOrder,
		IsNonInsurance: r.IsNonInsurance.toServiceInput(),

		IsBillDiscountExcluded: r.IsBillDiscountExcluded.toUpdateInput(),
		Version:                r.Version,
	}
}
