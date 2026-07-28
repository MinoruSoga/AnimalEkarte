package medicalrecord

// createMedicineRequest は薬剤作成のバインド struct
type createMedicineRequest struct {
	Name            string   `json:"name"             binding:"required"`
	ParentID        *uint64  `json:"parent_id"`
	Price           *int64   `json:"price"`
	IsActive        bool     `json:"is_active"`
	Description     string   `json:"description"`
	DosageForm      *string  `json:"dosage_form"      binding:"omitempty,oneof=tablet liquid injection topical powder"`
	MedicineUnit    *string  `json:"medicine_unit"    binding:"omitempty,oneof=per_tablet per_ml per_dose per_gram"`
	InventoryID     *uint64  `json:"inventory_id"`
	DefaultQuantity float64  `json:"default_quantity"`
	SortOrder       int      `json:"sort_order"`
	TaxType         *string  `json:"tax_type"         binding:"omitempty,oneof=included excluded exempt"`
	TaxRate         *float64 `json:"tax_rate"         binding:"omitempty,min=0,max=1"`
	IsNonInsurance  bool     `json:"is_non_insurance"`

	// #201 投与量計算（製品軸）
	CalculationType     *string  `json:"calculation_type"      binding:"omitempty,oneof=none per_weight"`
	Strength            *float64 `json:"strength"              binding:"omitempty,gt=0"`
	FrequencyPerDay     *int     `json:"frequency_per_day"     binding:"omitempty,gt=0"`
	DefaultDurationDays *int     `json:"default_duration_days" binding:"omitempty,gt=0"`
}

func (r *createMedicineRequest) toServiceInput() *CreateMedicineInput {
	return &CreateMedicineInput{
		Name:                r.Name,
		ParentID:            r.ParentID,
		Price:               r.Price,
		IsActive:            r.IsActive,
		Description:         r.Description,
		DosageForm:          r.DosageForm,
		MedicineUnit:        r.MedicineUnit,
		InventoryID:         r.InventoryID,
		DefaultQuantity:     r.DefaultQuantity,
		SortOrder:           r.SortOrder,
		TaxType:             r.TaxType,
		TaxRate:             r.TaxRate,
		IsNonInsurance:      r.IsNonInsurance,
		CalculationType:     r.CalculationType,
		Strength:            r.Strength,
		FrequencyPerDay:     r.FrequencyPerDay,
		DefaultDurationDays: r.DefaultDurationDays,
	}
}

// updateMedicineRequest は薬剤更新のバインド struct（全フィールドポインタ型）
// DosageForm/MedicineUnit: nil = 未指定, "" = NULL クリア, "value" = 値セット
// ParentID: nil = 未指定, clear_parent_id = true = NULL クリア, non-nil = 値セット
// InventoryID: nil = 未指定, non-nil = 値セット
type updateMedicineRequest struct {
	Name            *string  `json:"name"`
	ParentID        *uint64  `json:"parent_id"`
	ClearParentID   bool     `json:"clear_parent_id"`
	Price           *int64   `json:"price"`
	IsActive        *bool    `json:"is_active"`
	Description     *string  `json:"description"`
	DosageForm      *string  `json:"dosage_form"      binding:"omitempty,oneof=tablet liquid injection topical powder"`
	MedicineUnit    *string  `json:"medicine_unit"    binding:"omitempty,oneof=per_tablet per_ml per_dose per_gram"`
	InventoryID     *uint64  `json:"inventory_id"`
	DefaultQuantity *float64 `json:"default_quantity"`
	SortOrder       *int     `json:"sort_order"`
	TaxType         *string  `json:"tax_type"         binding:"omitempty,oneof=included excluded exempt"`
	TaxRate         *float64 `json:"tax_rate"         binding:"omitempty,min=0,max=1"`
	IsNonInsurance  *bool    `json:"is_non_insurance"`

	// #201 投与量計算（製品軸）
	CalculationType     *string  `json:"calculation_type"      binding:"omitempty,oneof=none per_weight"`
	Strength            *float64 `json:"strength"              binding:"omitempty,gt=0"`
	ClearStrength       bool     `json:"clear_strength"`
	FrequencyPerDay     *int     `json:"frequency_per_day"     binding:"omitempty,gt=0"`
	DefaultDurationDays *int     `json:"default_duration_days" binding:"omitempty,gt=0"`
}

func (r *updateMedicineRequest) toServiceInput() *UpdateMedicineInput {
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
		DefaultQuantity:     r.DefaultQuantity,
		SortOrder:           r.SortOrder,
		TaxType:             r.TaxType,
		TaxRate:             r.TaxRate,
		IsNonInsurance:      r.IsNonInsurance,
		CalculationType:     r.CalculationType,
		Strength:            r.Strength,
		ClearStrength:       r.ClearStrength,
		FrequencyPerDay:     r.FrequencyPerDay,
		DefaultDurationDays: r.DefaultDurationDays,
	}
}
