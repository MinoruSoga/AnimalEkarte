package medicalrecord

import (
	"math"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// buildVitalUpdate はnilでないフィールドのみmap[string]anyに変換する
func buildVitalUpdate(input *UpdateVitalInput) map[string]any {
	fields := map[string]any{}
	if input.RecordedAt != nil {
		fields["recorded_at"] = *input.RecordedAt
	}
	if input.StaffID != nil {
		fields["staff_id"] = *input.StaffID
	}
	if input.Temperature != nil {
		fields["temperature"] = *input.Temperature
	}
	if input.HeartRate != nil {
		fields["heart_rate"] = *input.HeartRate
	}
	if input.RespirationRate != nil {
		fields["respiration_rate"] = *input.RespirationRate
	}
	if input.Weight != nil {
		fields["weight"] = *input.Weight
	}
	if input.WeightUnit != nil {
		fields["weight_unit"] = *input.WeightUnit
	}
	if input.Notes != nil {
		fields["notes"] = *input.Notes
	}
	return fields
}

func validateUpdatedVitalRelation(
	vital *model.VitalRecord,
	clinicID, medicalRecordID, vitalID, petID uint64,
) error {
	if vital == nil ||
		vital.ID != vitalID ||
		vital.ClinicID != clinicID ||
		vital.PetID != petID ||
		vital.MedicalRecordID == nil ||
		*vital.MedicalRecordID != medicalRecordID {
		return apperrors.WrapNotFound("vital", "not found in medical record")
	}
	return nil
}

func validateVitalMedicalRecordRelation(
	parent *model.MedicalRecord,
	clinicID, medicalRecordID, petID uint64,
) error {
	if parent == nil ||
		parent.ID != medicalRecordID ||
		parent.ClinicID != clinicID ||
		parent.PetID == nil ||
		*parent.PetID != petID {
		return apperrors.WrapNotFound("medical_record", "relation")
	}
	return nil
}

// validateVitalWeight は service 層での weight 構造検証（request 境界と二重）。
func validateVitalWeight(weight *float64, unit *model.BodyWeightUnit) error {
	if weight != nil {
		if math.IsNaN(*weight) || math.IsInf(*weight, 0) {
			return apperrors.WrapInvalidInput("weight must be a finite number")
		}
		if *weight <= 0 {
			return apperrors.WrapInvalidInput("weight must be greater than zero")
		}
	}
	if unit != nil {
		switch *unit {
		case model.BodyWeightUnitKg, model.BodyWeightUnitG:
			// ok
		default:
			return apperrors.WrapInvalidInput("weight_unit must be Kg or g")
		}
	}
	return nil
}

// weightUnitOrDefault は nil の場合に BodyWeightUnitKg を返すヘルパー。
func weightUnitOrDefault(u *model.BodyWeightUnit) model.BodyWeightUnit {
	if u != nil {
		return *u
	}
	return model.BodyWeightUnitKg
}
