package medicalrecord

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// createVitalRequest はバイタル作成のバインド struct
type createVitalRequest struct {
	RecordedAt      time.Time `json:"recorded_at"       binding:"required"`
	StaffID         *uint64   `json:"staff_id"`
	Temperature     *float64  `json:"temperature"`
	HeartRate       *int      `json:"heart_rate"`
	RespirationRate *int      `json:"respiration_rate"`
	Weight          *float64  `json:"weight"`
	WeightUnit      *string   `json:"weight_unit"`
	Notes           string    `json:"notes"`
}

func (r *createVitalRequest) toServiceInput(clinicID, petID uint64) *CreateVitalInput {
	return &CreateVitalInput{
		ClinicID:        clinicID,
		PetID:           petID,
		RecordedAt:      r.RecordedAt,
		StaffID:         r.StaffID,
		Temperature:     r.Temperature,
		HeartRate:       r.HeartRate,
		RespirationRate: r.RespirationRate,
		Weight:          r.Weight,
		WeightUnit:      toBodyWeightUnit(r.WeightUnit),
		Notes:           r.Notes,
	}
}

// updateVitalRequest はバイタル更新のバインド struct（全フィールドがオプション）
type updateVitalRequest struct {
	RecordedAt      *time.Time `json:"recorded_at"`
	StaffID         *uint64    `json:"staff_id"`
	Temperature     *float64   `json:"temperature"`
	HeartRate       *int       `json:"heart_rate"`
	RespirationRate *int       `json:"respiration_rate"`
	Weight          *float64   `json:"weight"`
	WeightUnit      *string    `json:"weight_unit"`
	Notes           *string    `json:"notes"`
}

func (r updateVitalRequest) toServiceInput(actorID uint64) *UpdateVitalInput {
	return &UpdateVitalInput{
		RecordedAt:      r.RecordedAt,
		StaffID:         r.StaffID,
		Temperature:     r.Temperature,
		HeartRate:       r.HeartRate,
		RespirationRate: r.RespirationRate,
		Weight:          r.Weight,
		WeightUnit:      toBodyWeightUnit(r.WeightUnit),
		Notes:           r.Notes,
		ActorID:         &actorID,
	}
}

// toBodyWeightUnit は文字列ポインタを *model.BodyWeightUnit に変換するヘルパー。
// nil の場合は nil を返し、サービス層でデフォルト値（Kg）が適用される。
func toBodyWeightUnit(s *string) *model.BodyWeightUnit {
	if s == nil {
		return nil
	}
	u := model.BodyWeightUnit(*s)
	return &u
}
