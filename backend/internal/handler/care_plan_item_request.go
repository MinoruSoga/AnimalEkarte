package handler

// createCarePlanItemRequest はケアプランアイテム作成のバインド struct
type createCarePlanItemRequest struct {
	Type                  string   `json:"type"                    binding:"required"`
	Name                  string   `json:"name"                    binding:"required"`
	Description           string   `json:"description"`
	Timing                []string `json:"timing"`
	Status                string   `json:"status"`
	Notes                 string   `json:"notes"`
	MedicineID            *uint64  `json:"medicine_id"`
	ProcedureID           *uint64  `json:"procedure_id"`
	HospitalizationPlanID *uint64  `json:"hospitalization_plan_id"`
	UnitPrice             float64  `json:"unit_price"`
	Category              string   `json:"category"`
	SortOrder             int      `json:"sort_order"`
}

// updateCarePlanItemRequest はケアプランアイテム更新のバインド struct
type updateCarePlanItemRequest struct {
	Type                  *string  `json:"type"`
	Name                  *string  `json:"name"`
	Description           *string  `json:"description"`
	Timing                []string `json:"timing"` // no omitempty - empty array should clear
	Status                *string  `json:"status"`
	Notes                 *string  `json:"notes"`
	MedicineID            *uint64  `json:"medicine_id"`
	ProcedureID           *uint64  `json:"procedure_id"`
	HospitalizationPlanID *uint64  `json:"hospitalization_plan_id"`
	UnitPrice             *float64 `json:"unit_price"`
	Category              *string  `json:"category"`
	SortOrder             *int     `json:"sort_order"`
}
