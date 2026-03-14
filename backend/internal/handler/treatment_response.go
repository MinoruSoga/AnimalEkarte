package handler

import (
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type treatmentResponse struct {
	ID              string    `json:"id"`
	MedicalRecordID string    `json:"medical_record_id"`
	ItemType        string    `json:"item_type"`
	ConsultationID  *string   `json:"consultation_id,omitempty"`
	ProcedureID     *string   `json:"procedure_id,omitempty"`
	MedicineID      *string   `json:"medicine_id,omitempty"`
	InventoryID     *string   `json:"inventory_id,omitempty"`
	UnitPrice       float64   `json:"unit_price"`
	Quantity        int       `json:"quantity"`
	Selected        bool      `json:"selected"`
	Status          string    `json:"status"`
	Content         string    `json:"content"`
	Memo            string    `json:"memo"`
	Insurance       bool      `json:"insurance"`
	DiscountRate    float64   `json:"discount_rate"`
	DiscountAmount  float64   `json:"discount_amount"`
	SortOrder       int       `json:"sort_order"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func toTreatmentResponse(t *model.Treatment) treatmentResponse {
	return treatmentResponse{
		ID:              strconv.FormatUint(t.ID, 10),
		MedicalRecordID: strconv.FormatUint(t.MedicalRecordID, 10),
		ItemType:        string(t.ItemType),
		ConsultationID:  uint64PtrToStringPtr(t.ConsultationID),
		ProcedureID:     uint64PtrToStringPtr(t.ProcedureID),
		MedicineID:      uint64PtrToStringPtr(t.MedicineID),
		InventoryID:     uint64PtrToStringPtr(t.InventoryID),
		UnitPrice:       t.UnitPrice,
		Quantity:        t.Quantity,
		Selected:        t.Selected,
		Status:          string(t.Status),
		Content:         t.Content,
		Memo:            t.Memo,
		Insurance:       t.Insurance,
		DiscountRate:    t.DiscountRate,
		DiscountAmount:  t.DiscountAmount,
		SortOrder:       t.SortOrder,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
	}
}

// uint64PtrToStringPtr は *uint64 を *string に変換するヘルパー。
// nilの場合はnilを返す。
func uint64PtrToStringPtr(v *uint64) *string {
	if v == nil {
		return nil
	}
	s := strconv.FormatUint(*v, 10)
	return &s
}
