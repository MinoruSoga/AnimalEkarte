package medicalrecord

import (
	"github.com/animal-ekarte/backend/internal/httpapi"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type medicineDoseParamResponse struct {
	ID              uint64   `json:"id"`
	ClinicID        uint64   `json:"clinic_id"`
	MedicineID      uint64   `json:"medicine_id"`
	Species         string   `json:"species"`
	DoseBasis       string   `json:"dose_basis"`
	DosePerKg       float64  `json:"dose_per_kg"`
	MinMgPerKg      *float64 `json:"min_mg_per_kg,omitempty"`
	MaxMgPerKg      *float64 `json:"max_mg_per_kg,omitempty"`
	AbsoluteMaxDose *float64 `json:"absolute_max_dose,omitempty"`
	RoundingStep    *float64 `json:"rounding_step,omitempty"`
	RoundingMode    *string  `json:"rounding_mode,omitempty"`
	Notes           string   `json:"notes"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toMedicineDoseParamResponse(p *model.MedicineDoseParam) medicineDoseParamResponse {
	var roundingMode *string
	if p.RoundingMode != nil {
		v := string(*p.RoundingMode)
		roundingMode = &v
	}
	return medicineDoseParamResponse{
		ID:              p.ID,
		ClinicID:        p.ClinicID,
		MedicineID:      p.MedicineID,
		Species:         string(p.Species),
		DoseBasis:       string(p.DoseBasis),
		DosePerKg:       p.DosePerKg,
		MinMgPerKg:      p.MinMgPerKg,
		MaxMgPerKg:      p.MaxMgPerKg,
		AbsoluteMaxDose: p.AbsoluteMaxDose,
		RoundingStep:    p.RoundingStep,
		RoundingMode:    roundingMode,
		Notes:           p.Notes,
		CreatedAt:       httpapi.LocalTime(p.CreatedAt),
		UpdatedAt:       httpapi.LocalTime(p.UpdatedAt),
	}
}

func toMedicineDoseParamListResponse(ps []model.MedicineDoseParam) []medicineDoseParamResponse {
	return httpapi.MapSlice(ps, toMedicineDoseParamResponse)
}
