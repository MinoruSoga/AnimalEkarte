package handler

type createMedicineRequest struct {
	Name         string   `json:"name"          binding:"required"`
	Price        *float64 `json:"price"`
	IsActive     bool     `json:"is_active"`
	Description  string   `json:"description"`
	DosageForm   *string  `json:"dosage_form"`
	MedicineUnit *string  `json:"medicine_unit"`
	SortOrder    int      `json:"sort_order"`
}

type updateMedicineRequest struct {
	Name         string   `json:"name"`
	Price        *float64 `json:"price"`
	IsActive     *bool    `json:"is_active"`
	Description  string   `json:"description"`
	DosageForm   *string  `json:"dosage_form"`
	MedicineUnit *string  `json:"medicine_unit"`
	SortOrder    int      `json:"sort_order"`
}
