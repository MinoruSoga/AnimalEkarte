package medicalrecord

import (
	"github.com/animal-ekarte/backend/internal/httpapi"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type medicineResponse struct {
	ID              uint64  `json:"id"`
	ClinicID        uint64  `json:"clinic_id"`
	Name            string  `json:"name"`
	ParentID        *uint64 `json:"parent_id,omitempty"`
	Price           *int64  `json:"price,omitempty"`
	IsActive        bool    `json:"is_active"`
	Description     string  `json:"description"`
	DosageForm      *string `json:"dosage_form,omitempty"`
	MedicineUnit    *string `json:"medicine_unit,omitempty"`
	InventoryID     *uint64 `json:"inventory_id,omitempty"`
	DefaultQuantity float64 `json:"default_quantity"`
	TaxType         string  `json:"tax_type"`
	TaxRate         float64 `json:"tax_rate"`
	SortOrder       int     `json:"sort_order"`
	IsNonInsurance  bool    `json:"is_non_insurance"`

	// #201 投与量計算（製品軸）
	CalculationType     string   `json:"calculation_type"`
	Strength            *float64 `json:"strength,omitempty"`
	FrequencyPerDay     *int     `json:"frequency_per_day,omitempty"`
	DefaultDurationDays *int     `json:"default_duration_days,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toMedicineResponse(m *model.Medicine) medicineResponse {
	var dosageForm *string
	if m.DosageForm != nil {
		v := string(*m.DosageForm)
		dosageForm = &v
	}
	var medicineUnit *string
	if m.MedicineUnit != nil {
		v := string(*m.MedicineUnit)
		medicineUnit = &v
	}
	return medicineResponse{
		ID:                  m.ID,
		ClinicID:            m.ClinicID,
		Name:                m.Name,
		ParentID:            m.ParentID,
		Price:               m.Price,
		IsActive:            m.IsActive,
		Description:         m.Description,
		DosageForm:          dosageForm,
		MedicineUnit:        medicineUnit,
		InventoryID:         m.InventoryID,
		DefaultQuantity:     m.DefaultQuantity,
		TaxType:             string(m.TaxType),
		TaxRate:             m.TaxRate,
		SortOrder:           m.SortOrder,
		IsNonInsurance:      m.IsNonInsurance,
		CalculationType:     string(m.CalculationType),
		Strength:            m.Strength,
		FrequencyPerDay:     m.FrequencyPerDay,
		DefaultDurationDays: m.DefaultDurationDays,
		CreatedAt:           httpapi.LocalTime(m.CreatedAt),
		UpdatedAt:           httpapi.LocalTime(m.UpdatedAt),
	}
}
