package medicalrecord

import (
	"github.com/animal-ekarte/backend/internal/model"
)

// upsertMedicineDoseParamRequest は種別 dose パラメータの upsert バインド struct（#201 B-2c）。
// species は URL path (:species) から受け取り body には含めない（path/body 不一致を構造的に排除）。
// PUT セマンティクス: 全フィールドを置換する。省略 optional は NULL クリアされる。
type upsertMedicineDoseParamRequest struct {
	DoseBasis       string   `json:"dose_basis"        binding:"omitempty,oneof=per_administration per_day"`
	DosePerKg       float64  `json:"dose_per_kg"       binding:"required,gt=0"`
	MinMgPerKg      *float64 `json:"min_mg_per_kg"     binding:"omitempty,gt=0"`
	MaxMgPerKg      *float64 `json:"max_mg_per_kg"     binding:"omitempty,gt=0"`
	AbsoluteMaxDose *float64 `json:"absolute_max_dose" binding:"omitempty,gt=0"`
	RoundingStep    *float64 `json:"rounding_step"     binding:"omitempty,gt=0"`
	RoundingMode    *string  `json:"rounding_mode"     binding:"omitempty,oneof=up down nearest"`
	Notes           string   `json:"notes"`
}

func (r *upsertMedicineDoseParamRequest) toServiceInput(species model.MedicineDoseSpecies) *MedicineDoseParamInput {
	doseBasis := model.MedicineDoseBasisPerAdministration
	if r.DoseBasis != "" {
		doseBasis = model.MedicineDoseBasis(r.DoseBasis)
	}
	var roundingMode *model.MedicineRoundingMode
	if r.RoundingMode != nil && *r.RoundingMode != "" {
		rm := model.MedicineRoundingMode(*r.RoundingMode)
		roundingMode = &rm
	}
	return &MedicineDoseParamInput{
		Species:         species,
		DoseBasis:       doseBasis,
		DosePerKg:       r.DosePerKg,
		MinMgPerKg:      r.MinMgPerKg,
		MaxMgPerKg:      r.MaxMgPerKg,
		AbsoluteMaxDose: r.AbsoluteMaxDose,
		RoundingStep:    r.RoundingStep,
		RoundingMode:    roundingMode,
		Notes:           r.Notes,
	}
}
