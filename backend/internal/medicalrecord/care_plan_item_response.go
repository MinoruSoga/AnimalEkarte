package medicalrecord

import (
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type carePlanItemResponse struct {
	ID                    string    `json:"id"`
	HospitalizationID     string    `json:"hospitalization_id"`
	Type                  string    `json:"type"`
	Name                  string    `json:"name"`
	Description           string    `json:"description"`
	Timing                []string  `json:"timing"`
	Status                string    `json:"status"`
	Notes                 string    `json:"notes"`
	MedicineID            *string   `json:"medicine_id,omitempty"`
	ProcedureID           *string   `json:"procedure_id,omitempty"`
	HospitalizationPlanID *string   `json:"hospitalization_plan_id,omitempty"`
	UnitPrice             int64     `json:"unit_price"`
	Category              string    `json:"category"`
	SortOrder             int       `json:"sort_order"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

func toCarePlanItemResponse(item *model.CarePlanItem) carePlanItemResponse {
	timing := []string{}
	if item.Timing != nil {
		timing = []string(item.Timing)
	}

	r := carePlanItemResponse{
		ID:                strconv.FormatUint(item.ID, 10),
		HospitalizationID: strconv.FormatUint(item.HospitalizationID, 10),
		Type:              string(item.Type),
		Name:              item.Name,
		Description:       item.Description,
		Timing:            timing,
		Status:            string(item.Status),
		Notes:             item.Notes,
		UnitPrice:         item.UnitPrice,
		Category:          item.Category,
		SortOrder:         item.SortOrder,
		CreatedAt:         httpapi.LocalTime(item.CreatedAt),
		UpdatedAt:         httpapi.LocalTime(item.UpdatedAt),
	}
	if item.MedicineID != nil {
		s := strconv.FormatUint(*item.MedicineID, 10)
		r.MedicineID = &s
	}
	if item.ProcedureID != nil {
		s := strconv.FormatUint(*item.ProcedureID, 10)
		r.ProcedureID = &s
	}
	if item.HospitalizationPlanID != nil {
		s := strconv.FormatUint(*item.HospitalizationPlanID, 10)
		r.HospitalizationPlanID = &s
	}
	return r
}
