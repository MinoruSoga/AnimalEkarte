package medicalrecord

import (
	"encoding/json"

	"github.com/animal-ekarte/backend/internal/model"
)

// createMedicineRequest は薬剤作成のバインド struct。
// default_quantity / tax_rate / strength は EMR-447 の exact decimal 文字列（medicine_decimal.go）で受ける。
type createMedicineRequest struct {
	Name            string          `json:"name"             binding:"required"`
	ParentID        *uint64         `json:"parent_id"`
	Price           *int64          `json:"price"`
	IsActive        bool            `json:"is_active"`
	Description     string          `json:"description"`
	DosageForm      *string         `json:"dosage_form"      binding:"omitempty,oneof=tablet liquid injection topical powder"`
	MedicineUnit    *string         `json:"medicine_unit"    binding:"omitempty,oneof=per_tablet per_ml per_dose per_gram"`
	InventoryID     *uint64         `json:"inventory_id"`
	DefaultQuantity json.RawMessage `json:"default_quantity"`
	SortOrder       int             `json:"sort_order"`
	TaxType         *string         `json:"tax_type"         binding:"omitempty,oneof=included excluded exempt"`
	TaxRate         json.RawMessage `json:"tax_rate"`
	IsNonInsurance  bool            `json:"is_non_insurance"`

	// #201 投与量計算（製品軸）
	CalculationType     *string         `json:"calculation_type"      binding:"omitempty,oneof=none per_weight"`
	Strength            json.RawMessage `json:"strength"`
	FrequencyPerDay     *int            `json:"frequency_per_day"     binding:"omitempty,gt=0"`
	DefaultDurationDays *int            `json:"default_duration_days" binding:"omitempty,gt=0"`
}

// medicineDecimals は default_quantity / tax_rate / strength の wire を厳密に解析する（EMR-447）。
func medicineDecimals(defaultQuantity, taxRate, strength json.RawMessage) (dq, tr, st *model.ExactNumeric, err error) {
	if dq, err = parseMedicineDecimal(defaultQuantity, medicineDefaultQuantitySpec); err != nil {
		return nil, nil, nil, err
	}
	if tr, err = parseMedicineDecimal(taxRate, medicineTaxRateSpec); err != nil {
		return nil, nil, nil, err
	}
	if st, err = parseMedicineDecimal(strength, medicineStrengthSpec); err != nil {
		return nil, nil, nil, err
	}
	return dq, tr, st, nil
}

func (r *createMedicineRequest) toServiceInput() (*CreateMedicineInput, error) {
	defaultQuantity, taxRate, strength, err := medicineDecimals(r.DefaultQuantity, r.TaxRate, r.Strength)
	if err != nil {
		return nil, err
	}
	return &CreateMedicineInput{
		Name:                r.Name,
		ParentID:            r.ParentID,
		Price:               r.Price,
		IsActive:            r.IsActive,
		Description:         r.Description,
		DosageForm:          r.DosageForm,
		MedicineUnit:        r.MedicineUnit,
		InventoryID:         r.InventoryID,
		DefaultQuantity:     defaultQuantity,
		SortOrder:           r.SortOrder,
		TaxType:             r.TaxType,
		TaxRate:             taxRate,
		IsNonInsurance:      r.IsNonInsurance,
		CalculationType:     r.CalculationType,
		Strength:            strength,
		FrequencyPerDay:     r.FrequencyPerDay,
		DefaultDurationDays: r.DefaultDurationDays,
	}, nil
}

// updateMedicineRequest は薬剤更新のバインド struct（全フィールドポインタ型）
// DosageForm/MedicineUnit: nil = 未指定, "" = NULL クリア, "value" = 値セット
// ParentID: nil = 未指定, clear_parent_id = true = NULL クリア, non-nil = 値セット
// InventoryID: nil = 未指定, non-nil = 値セット
// DefaultQuantity/TaxRate/Strength: 省略・null = 変更なし, decimal 文字列 = 値セット（EMR-447）。strength の NULL クリアは clear_strength
type updateMedicineRequest struct {
	Name            *string         `json:"name"`
	ParentID        *uint64         `json:"parent_id"`
	ClearParentID   bool            `json:"clear_parent_id"`
	Price           *int64          `json:"price"`
	IsActive        *bool           `json:"is_active"`
	Description     *string         `json:"description"`
	DosageForm      *string         `json:"dosage_form"      binding:"omitempty,oneof=tablet liquid injection topical powder"`
	MedicineUnit    *string         `json:"medicine_unit"    binding:"omitempty,oneof=per_tablet per_ml per_dose per_gram"`
	InventoryID     *uint64         `json:"inventory_id"`
	DefaultQuantity json.RawMessage `json:"default_quantity"`
	SortOrder       *int            `json:"sort_order"`
	TaxType         *string         `json:"tax_type"         binding:"omitempty,oneof=included excluded exempt"`
	TaxRate         json.RawMessage `json:"tax_rate"`
	IsNonInsurance  *bool           `json:"is_non_insurance"`

	// #201 投与量計算（製品軸）
	CalculationType     *string         `json:"calculation_type"      binding:"omitempty,oneof=none per_weight"`
	Strength            json.RawMessage `json:"strength"`
	ClearStrength       bool            `json:"clear_strength"`
	FrequencyPerDay     *int            `json:"frequency_per_day"     binding:"omitempty,gt=0"`
	DefaultDurationDays *int            `json:"default_duration_days" binding:"omitempty,gt=0"`
}

func (r *updateMedicineRequest) toServiceInput() (*UpdateMedicineInput, error) {
	defaultQuantity, taxRate, strength, err := medicineDecimals(r.DefaultQuantity, r.TaxRate, r.Strength)
	if err != nil {
		return nil, err
	}
	return &UpdateMedicineInput{
		Name:                r.Name,
		ParentID:            r.ParentID,
		ClearParentID:       r.ClearParentID,
		Price:               r.Price,
		IsActive:            r.IsActive,
		Description:         r.Description,
		DosageForm:          r.DosageForm,
		MedicineUnit:        r.MedicineUnit,
		InventoryID:         r.InventoryID,
		DefaultQuantity:     defaultQuantity,
		SortOrder:           r.SortOrder,
		TaxType:             r.TaxType,
		TaxRate:             taxRate,
		IsNonInsurance:      r.IsNonInsurance,
		CalculationType:     r.CalculationType,
		Strength:            strength,
		ClearStrength:       r.ClearStrength,
		FrequencyPerDay:     r.FrequencyPerDay,
		DefaultDurationDays: r.DefaultDurationDays,
	}, nil
}
