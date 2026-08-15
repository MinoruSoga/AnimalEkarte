package medicalrecord

// createCarePlanItemRequest はケアプランアイテム作成のバインド struct
type createCarePlanItemRequest struct {
	Type                  string   `json:"type"                    binding:"required,oneof=food medicine treatment instruction item"`
	Name                  string   `json:"name"                    binding:"required"`
	Description           string   `json:"description"`
	Timing                []string `json:"timing"`
	Status                string   `json:"status"                  binding:"omitempty,oneof=active completed discontinued"`
	Notes                 string   `json:"notes"`
	MedicineID            *uint64  `json:"medicine_id"`
	ProcedureID           *uint64  `json:"procedure_id"`
	HospitalizationPlanID *uint64  `json:"hospitalization_plan_id"`
	UnitPrice             int64    `json:"unit_price"`
	Category              string   `json:"category"`
	SortOrder             int      `json:"sort_order"`
}

func (r *createCarePlanItemRequest) toServiceInput() *CreateCarePlanItemInput {
	return &CreateCarePlanItemInput{
		Type:                  r.Type,
		Name:                  r.Name,
		Description:           r.Description,
		Timing:                r.Timing,
		Status:                r.Status,
		Notes:                 r.Notes,
		MedicineID:            r.MedicineID,
		ProcedureID:           r.ProcedureID,
		HospitalizationPlanID: r.HospitalizationPlanID,
		UnitPrice:             r.UnitPrice,
		Category:              r.Category,
		SortOrder:             r.SortOrder,
	}
}

// updateCarePlanItemRequest はケアプランアイテム更新のバインド struct
type updateCarePlanItemRequest struct {
	Type                  *string  `json:"type"                    binding:"omitempty,oneof=food medicine treatment instruction item"`
	Name                  *string  `json:"name"`
	Description           *string  `json:"description"`
	Timing                []string `json:"timing"` // no omitempty - empty array should clear
	Status                *string  `json:"status"                  binding:"omitempty,oneof=active completed discontinued"`
	Notes                 *string  `json:"notes"`
	MedicineID            *uint64  `json:"medicine_id"`
	ProcedureID           *uint64  `json:"procedure_id"`
	HospitalizationPlanID *uint64  `json:"hospitalization_plan_id"`
	UnitPrice             *int64   `json:"unit_price"`
	Category              *string  `json:"category"`
	SortOrder             *int     `json:"sort_order"`
}

func (r *updateCarePlanItemRequest) toServiceInput() *UpdateCarePlanItemInput {
	return &UpdateCarePlanItemInput{
		Type:                  r.Type,
		Name:                  r.Name,
		Description:           r.Description,
		Timing:                r.Timing,
		Status:                r.Status,
		Notes:                 r.Notes,
		MedicineID:            r.MedicineID,
		ProcedureID:           r.ProcedureID,
		HospitalizationPlanID: r.HospitalizationPlanID,
		UnitPrice:             r.UnitPrice,
		Category:              r.Category,
		SortOrder:             r.SortOrder,
	}
}
