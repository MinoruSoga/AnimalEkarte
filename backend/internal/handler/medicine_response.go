package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type medicineResponse struct {
	ID              uint64    `json:"id"`
	ClinicID        uint64    `json:"clinic_id"`
	Name            string    `json:"name"`
	ParentID        *uint64   `json:"parent_id,omitempty"`
	Price           *int64    `json:"price,omitempty"`
	IsActive        bool      `json:"is_active"`
	Description     string    `json:"description"`
	DosageForm      *string   `json:"dosage_form,omitempty"`
	MedicineUnit    *string   `json:"medicine_unit,omitempty"`
	InventoryID     *uint64   `json:"inventory_id,omitempty"`
	DefaultQuantity float64   `json:"default_quantity"`
	TaxType         string    `json:"tax_type"`
	TaxRate         float64   `json:"tax_rate"`
	SortOrder       int       `json:"sort_order"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func toMedicineResponse(m *model.Medicine) medicineResponse {
	var dosageForm *string
	if m.DosageForm != nil {
		s := string(*m.DosageForm)
		dosageForm = &s
	}
	var medicineUnit *string
	if m.MedicineUnit != nil {
		s := string(*m.MedicineUnit)
		medicineUnit = &s
	}
	return medicineResponse{
		ID:              m.ID,
		ClinicID:        m.ClinicID,
		Name:            m.Name,
		ParentID:        m.ParentID,
		Price:           m.Price,
		IsActive:        m.IsActive,
		Description:     m.Description,
		DosageForm:      dosageForm,
		MedicineUnit:    medicineUnit,
		InventoryID:     m.InventoryID,
		DefaultQuantity: m.DefaultQuantity,
		TaxType:         string(m.TaxType),
		TaxRate:         m.TaxRate,
		SortOrder:       m.SortOrder,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

func toMedicineResponseList(medicines []model.Medicine) []medicineResponse {
	list := make([]medicineResponse, 0, len(medicines))
	for i := range medicines {
		list = append(list, toMedicineResponse(&medicines[i]))
	}
	return list
}
