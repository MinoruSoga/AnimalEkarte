package handler

// createMedicineRequest は薬剤作成のバインド struct
type createMedicineRequest struct {
	Name            string   `json:"name"             binding:"required"`
	DrugCategory    *string  `json:"drug_category"`
	Price           *float64 `json:"price"`
	IsActive        bool     `json:"is_active"`
	Description     string   `json:"description"`
	DosageForm      string   `json:"dosage_form"`
	MedicineUnit    string   `json:"medicine_unit"`
	InventoryID     *uint64  `json:"inventory_id"`
	DefaultQuantity int      `json:"default_quantity"`
	SortOrder       int      `json:"sort_order"`
}

// updateMedicineRequest は薬剤更新のバインド struct（全フィールドポインタ型）
// DrugCategory/DosageForm/MedicineUnit: nil = 未指定, "" = NULL クリア, "value" = 値セット
// InventoryID: nil = 未指定, &nil = NULL クリア, &&val = 値セット
type updateMedicineRequest struct {
	Name            *string  `json:"name"`
	DrugCategory    *string  `json:"drug_category"`
	Price           *float64 `json:"price"`
	IsActive        *bool    `json:"is_active"`
	Description     *string  `json:"description"`
	DosageForm      *string  `json:"dosage_form"`
	MedicineUnit    *string  `json:"medicine_unit"`
	InventoryID     **uint64 `json:"inventory_id"`
	DefaultQuantity *int     `json:"default_quantity"`
	SortOrder       *int     `json:"sort_order"`
}
