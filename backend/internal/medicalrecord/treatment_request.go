package medicalrecord

import (
	"github.com/animal-ekarte/backend/internal/model"
)

type createTreatmentRequest struct {
	ItemType       string  `json:"item_type"       binding:"required,oneof=consultation procedure medicine other"`
	ConsultationID *uint64 `json:"consultation_id"`
	ProcedureID    *uint64 `json:"procedure_id"`
	MedicineID     *uint64 `json:"medicine_id"`
	InventoryID    *uint64 `json:"inventory_id"`
	UnitPrice      int64   `json:"unit_price"`
	Quantity       float64 `json:"quantity"`
	IsSelected     bool    `json:"is_selected"`
	Status         string  `json:"status"          binding:"omitempty,oneof=pending completed not_applicable"`
	Content        string  `json:"content"`
	Memo           string  `json:"memo"`
	AdminRoute     string  `json:"admin_route"`
	IsInsurance    bool    `json:"is_insurance"`
	DiscountRate   float64 `json:"discount_rate"`
	DiscountAmount int64   `json:"discount_amount"`
	SortOrder      int     `json:"sort_order"`
	// DoseDeviationReason は TASK-377: 上限内の著しい乖離/下限割れで必須。service が最終権限で再評価する。
	DoseDeviationReason string `json:"dose_deviation_reason"`
}

func (r *createTreatmentRequest) toServiceInput() *CreateTreatmentInput {
	return &CreateTreatmentInput{
		ItemType:            model.TreatmentItemType(r.ItemType),
		ConsultationID:      r.ConsultationID,
		ProcedureID:         r.ProcedureID,
		MedicineID:          r.MedicineID,
		InventoryID:         r.InventoryID,
		UnitPrice:           r.UnitPrice,
		Quantity:            r.Quantity,
		IsSelected:          r.IsSelected,
		Status:              r.Status,
		Content:             r.Content,
		Memo:                r.Memo,
		AdminRoute:          r.AdminRoute,
		IsInsurance:         r.IsInsurance,
		DiscountRate:        r.DiscountRate,
		DiscountAmount:      r.DiscountAmount,
		SortOrder:           r.SortOrder,
		DoseDeviationReason: r.DoseDeviationReason,
	}
}

type updateTreatmentRequest struct {
	ItemType       *string  `json:"item_type"       binding:"omitempty,oneof=consultation procedure medicine other"`
	ConsultationID *uint64  `json:"consultation_id"`
	ProcedureID    *uint64  `json:"procedure_id"`
	MedicineID     *uint64  `json:"medicine_id"`
	InventoryID    *uint64  `json:"inventory_id"`
	UnitPrice      *int64   `json:"unit_price"`
	Quantity       *float64 `json:"quantity"`
	IsSelected     *bool    `json:"is_selected"`
	Status         *string  `json:"status"          binding:"omitempty,oneof=pending completed not_applicable"`
	Content        *string  `json:"content"`
	Memo           *string  `json:"memo"`
	AdminRoute     *string  `json:"admin_route"`
	IsInsurance    *bool    `json:"is_insurance"`
	DiscountRate   *float64 `json:"discount_rate"`
	DiscountAmount *int64   `json:"discount_amount"`
	SortOrder      *int     `json:"sort_order"`
	// DoseDeviationReason は TASK-377: reason-required 時に必須。
	DoseDeviationReason *string `json:"dose_deviation_reason"`
}

func (r *updateTreatmentRequest) toServiceInput() *UpdateTreatmentInput {
	input := &UpdateTreatmentInput{
		ConsultationID:      r.ConsultationID,
		ProcedureID:         r.ProcedureID,
		MedicineID:          r.MedicineID,
		InventoryID:         r.InventoryID,
		UnitPrice:           r.UnitPrice,
		Quantity:            r.Quantity,
		IsSelected:          r.IsSelected,
		Status:              r.Status,
		Content:             r.Content,
		Memo:                r.Memo,
		AdminRoute:          r.AdminRoute,
		IsInsurance:         r.IsInsurance,
		DiscountRate:        r.DiscountRate,
		DiscountAmount:      r.DiscountAmount,
		SortOrder:           r.SortOrder,
		DoseDeviationReason: r.DoseDeviationReason,
	}
	if r.ItemType != nil {
		itemType := model.TreatmentItemType(*r.ItemType)
		input.ItemType = &itemType
	}
	return input
}

type bulkUpdateTreatmentsRequest struct {
	Treatments []bulkTreatmentItem `json:"treatments" binding:"required"`
}

func (r bulkUpdateTreatmentsRequest) toServiceInput() *BulkUpdateTreatmentsInput {
	items := make([]BulkTreatmentItem, 0, len(r.Treatments))
	for _, treatment := range r.Treatments {
		items = append(items, BulkTreatmentItem(treatment))
	}
	return &BulkUpdateTreatmentsInput{Treatments: items}
}

type bulkTreatmentItem struct {
	ID        uint64 `json:"id"         binding:"required"`
	SortOrder int    `json:"sort_order"`
}
