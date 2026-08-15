package medicalrecord

import (
	"math"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
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

func (r *createVitalRequest) validate() error {
	return validateVitalWeightFields(r.Weight, r.WeightUnit)
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

func (r updateVitalRequest) validate() error {
	return validateVitalWeightFields(r.Weight, r.WeightUnit)
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

// validateVitalWeightFields は weight / weight_unit の構造検証（finite・正数・unit enum）。
// 臨床的な上限/下限は承認値が無いため導入しない（BUG-015）。
func validateVitalWeightFields(weight *float64, weightUnit *string) error {
	if weight != nil {
		if math.IsNaN(*weight) || math.IsInf(*weight, 0) {
			return apperrors.WrapInvalidInput("weight must be a finite number")
		}
		if *weight <= 0 {
			return apperrors.WrapInvalidInput("weight must be greater than zero")
		}
	}
	if weightUnit != nil {
		switch model.BodyWeightUnit(*weightUnit) {
		case model.BodyWeightUnitKg, model.BodyWeightUnitG:
			// ok
		default:
			return apperrors.WrapInvalidInput("weight_unit must be Kg or g")
		}
	}
	return nil
}

// toBodyWeightUnit は文字列ポインタを *model.BodyWeightUnit に変換するヘルパー。
// nil の場合は nil を返し、サービス層でデフォルト値（Kg）が適用される。
// 呼び出し前に validateVitalWeightFields で enum を検証すること。
func toBodyWeightUnit(s *string) *model.BodyWeightUnit {
	if s == nil {
		return nil
	}
	u := model.BodyWeightUnit(*s)
	return &u
}
