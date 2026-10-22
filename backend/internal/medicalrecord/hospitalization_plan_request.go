package medicalrecord

import (
	"encoding/json"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

const errMsgHospitalizationPlanEnumNotString = "対象体格・料金単位には文字列または null を指定してください"

// hospitalizationPlanNullableEnumField は PATCH の body_size / billing_unit の wire 型（EMR-528）。
// encoding/json は *string に null を落とすと省略と区別できないため、専有の型で
// 「未送信（保持）」「明示 null（SQL NULL へ戻す）」「値（設定）」を区別する。
// 値の enum 検査は service が行い、"" を null の代用にしない。
type hospitalizationPlanNullableEnumField struct {
	set   bool
	value *string
}

func (f *hospitalizationPlanNullableEnumField) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		f.set = true
		f.value = nil
		return nil
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return apperrors.WrapInvalidInput(errMsgHospitalizationPlanEnumNotString)
	}
	f.set = true
	f.value = &value
	return nil
}

// toServiceInput は nil=省略、&nil=明示 null、&&v=値 を返す。
func (f hospitalizationPlanNullableEnumField) toServiceInput() **string {
	if !f.set {
		return nil
	}
	value := f.value
	return &value
}

// createHospitalizationPlanRequest は入院プラン作成リクエスト。
type createHospitalizationPlanRequest struct {
	Name        string   `json:"name"         binding:"required"`
	Price       *int64   `json:"price"`
	IsActive    bool     `json:"is_active"`
	Description string   `json:"description"`
	BodySize    string   `json:"body_size"    binding:"omitempty,oneof=small medium large"`
	BillingUnit string   `json:"billing_unit" binding:"omitempty,oneof=per_day per_night"`
	SortOrder   int      `json:"sort_order"`
	TaxType     string   `json:"tax_type"     binding:"omitempty,oneof=included excluded exempt"`
	TaxRate     *float64 `json:"tax_rate"`
}

func (r *createHospitalizationPlanRequest) toServiceInput() *CreateHospitalizationPlanInput {
	return &CreateHospitalizationPlanInput{
		Name:        r.Name,
		Price:       r.Price,
		IsActive:    r.IsActive,
		Description: r.Description,
		SortOrder:   r.SortOrder,
		TaxType:     r.TaxType,
		TaxRate:     r.TaxRate,
		BodySize:    r.BodySize,
		BillingUnit: r.BillingUnit,
	}
}

// updateHospitalizationPlanRequest は入院プラン更新リクエスト。
type updateHospitalizationPlanRequest struct {
	Name        *string  `json:"name"`
	Price       *int64   `json:"price"`
	IsActive    *bool    `json:"is_active"`
	Description *string  `json:"description"`
	SortOrder   *int     `json:"sort_order"`
	TaxType     *string  `json:"tax_type"`
	TaxRate     *float64 `json:"tax_rate"`

	// EMR-528: 省略=保持、null=SQL NULL、enum 値=設定。"" と enum 外は 400。
	BodySize    hospitalizationPlanNullableEnumField `json:"body_size"`
	BillingUnit hospitalizationPlanNullableEnumField `json:"billing_unit"`
}

func (r updateHospitalizationPlanRequest) toServiceInput() *UpdateHospitalizationPlanInput {
	return &UpdateHospitalizationPlanInput{
		Name:        r.Name,
		Price:       r.Price,
		IsActive:    r.IsActive,
		Description: r.Description,
		BodySize:    r.BodySize.toServiceInput(),
		BillingUnit: r.BillingUnit.toServiceInput(),
		SortOrder:   r.SortOrder,
		TaxType:     r.TaxType,
		TaxRate:     r.TaxRate,
	}
}
