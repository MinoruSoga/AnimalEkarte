package handler

import "github.com/animal-ekarte/backend/internal/service"

// createMedicineRequest は薬剤作成のバインド struct
type createMedicineRequest struct {
	Name            string   `json:"name"             binding:"required"`
	ParentID        *uint64  `json:"parent_id"`
	Price           *int64   `json:"price"`
	IsActive        bool     `json:"is_active"`
	Description     string   `json:"description"`
	DosageForm      *string  `json:"dosage_form"      binding:"omitempty,oneof=tablet liquid injection topical powder"`
	MedicineUnit    *string  `json:"medicine_unit"`
	InventoryID     *uint64  `json:"inventory_id"`
	DefaultQuantity float64  `json:"default_quantity"`
	SortOrder       int      `json:"sort_order"`
	TaxType         *string  `json:"tax_type"         binding:"omitempty,oneof=included excluded exempt"`
	TaxRate         *float64 `json:"tax_rate"         binding:"omitempty,min=0,max=1"`
	IsNonInsurance  bool     `json:"is_non_insurance"`
}

func (r createMedicineRequest) toServiceInput() *service.CreateMedicineInput {
	return &service.CreateMedicineInput{
		Name:            r.Name,
		ParentID:        r.ParentID,
		Price:           r.Price,
		IsActive:        r.IsActive,
		Description:     r.Description,
		DosageForm:      r.DosageForm,
		MedicineUnit:    r.MedicineUnit,
		InventoryID:     r.InventoryID,
		DefaultQuantity: r.DefaultQuantity,
		SortOrder:       r.SortOrder,
		TaxType:         r.TaxType,
		TaxRate:         r.TaxRate,
		IsNonInsurance:  r.IsNonInsurance,
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
	MedicineUnit    *string  `json:"medicine_unit"`
	InventoryID     *uint64  `json:"inventory_id"`
	DefaultQuantity *float64 `json:"default_quantity"`
	SortOrder       *int     `json:"sort_order"`
	TaxType         *string  `json:"tax_type"         binding:"omitempty,oneof=included excluded exempt"`
	TaxRate         *float64 `json:"tax_rate"         binding:"omitempty,min=0,max=1"`
	IsNonInsurance  *bool    `json:"is_non_insurance"`
}

func (r updateMedicineRequest) toServiceInput() *service.UpdateMedicineInput {
	return &service.UpdateMedicineInput{
		Name:            r.Name,
		ParentID:        r.ParentID,
		ClearParentID:   r.ClearParentID,
		Price:           r.Price,
		IsActive:        r.IsActive,
		Description:     r.Description,
		DosageForm:      r.DosageForm,
		MedicineUnit:    r.MedicineUnit,
		InventoryID:     r.InventoryID,
		DefaultQuantity: r.DefaultQuantity,
		SortOrder:       r.SortOrder,
		TaxType:         r.TaxType,
		TaxRate:         r.TaxRate,
		IsNonInsurance:  r.IsNonInsurance,
	}
}
