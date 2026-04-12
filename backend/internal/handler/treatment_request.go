package handler

type createTreatmentRequest struct {
	ItemType       string  `json:"item_type"       binding:"required"`
	ConsultationID *uint64 `json:"consultation_id"`
	ProcedureID    *uint64 `json:"procedure_id"`
	MedicineID     *uint64 `json:"medicine_id"`
	InventoryID    *uint64 `json:"inventory_id"`
	UnitPrice      int64   `json:"unit_price"`
	Quantity       float64 `json:"quantity"`
	IsSelected     bool    `json:"is_selected"`
	Status         string  `json:"status"`
	Content        string  `json:"content"`
	Memo           string  `json:"memo"`
	AdminRoute     string  `json:"admin_route"`
	IsInsurance    bool    `json:"is_insurance"`
	DiscountRate   float64 `json:"discount_rate"`
	DiscountAmount int64   `json:"discount_amount"`
	SortOrder      int     `json:"sort_order"`
}

type updateTreatmentRequest struct {
	ItemType       *string  `json:"item_type"`
	ConsultationID *uint64  `json:"consultation_id"`
	ProcedureID    *uint64  `json:"procedure_id"`
	MedicineID     *uint64  `json:"medicine_id"`
	InventoryID    *uint64  `json:"inventory_id"`
	UnitPrice      *int64   `json:"unit_price"`
	Quantity       *float64 `json:"quantity"`
	IsSelected     *bool    `json:"is_selected"`
	Status         *string  `json:"status"`
	Content        *string  `json:"content"`
	Memo           *string  `json:"memo"`
	AdminRoute     *string  `json:"admin_route"`
	IsInsurance    *bool    `json:"is_insurance"`
	DiscountRate   *float64 `json:"discount_rate"`
	DiscountAmount *int64   `json:"discount_amount"`
	SortOrder      *int     `json:"sort_order"`
}

type bulkUpdateTreatmentsRequest struct {
	Treatments []bulkTreatmentItem `json:"treatments" binding:"required"`
}

type bulkTreatmentItem struct {
	ID        uint64 `json:"id"         binding:"required"`
	SortOrder int    `json:"sort_order"`
}
