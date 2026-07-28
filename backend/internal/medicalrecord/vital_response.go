package medicalrecord

import (
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type vitalResponse struct {
	ID              string               `json:"id"`
	MedicalRecordID *string              `json:"medical_record_id,omitempty"`
	RecordedAt      time.Time            `json:"recorded_at"`
	StaffID         *string              `json:"staff_id,omitempty"`
	Temperature     *float64             `json:"temperature,omitempty"`
	HeartRate       *int                 `json:"heart_rate,omitempty"`
	RespirationRate *int                 `json:"respiration_rate,omitempty"`
	Weight          *float64             `json:"weight,omitempty"`
	WeightUnit      model.BodyWeightUnit `json:"weight_unit"`
	Notes           string               `json:"notes"`
	CreatedAt       time.Time            `json:"created_at"`
}

func toVitalResponse(v *model.VitalRecord) vitalResponse {
	r := vitalResponse{
		ID:              strconv.FormatUint(v.ID, 10),
		RecordedAt:      httpapi.LocalTime(v.RecordedAt),
		Temperature:     v.Temperature,
		HeartRate:       v.HeartRate,
		RespirationRate: v.RespirationRate,
		Weight:          v.Weight,
		WeightUnit:      v.WeightUnit,
		Notes:           v.Notes,
		CreatedAt:       httpapi.LocalTime(v.CreatedAt),
	}
	if v.MedicalRecordID != nil {
		s := strconv.FormatUint(*v.MedicalRecordID, 10)
		r.MedicalRecordID = &s
	}
	if v.StaffID != nil {
		s := strconv.FormatUint(*v.StaffID, 10)
		r.StaffID = &s
	}
	return r
}
