package medicalrecord

import (
	"fmt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// validateMedicalImageType は診療画像種別がドメイン上有効かを検証する。
// BE9-2D sub-batch④a で medicalRecordImageService とともに移設したが、internal/service 側の
// 同名関数は validators_test.go が引き続き検証しているため *複製* とした（byte-for-byte 同一）。
func validateMedicalImageType(t string) error {
	if t == "" {
		return nil
	}
	switch model.MedicalImageType(t) {
	case model.MedicalImageTypeXray, model.MedicalImageTypeEcho, model.MedicalImageTypePhoto,
		model.MedicalImageTypeEndoscope, model.MedicalImageTypeCT, model.MedicalImageTypeMRI,
		model.MedicalImageTypeMicroscope, model.MedicalImageTypeOther:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("画像種別の値が不正です: %s", t))
	}
}

// validateVaccineSpecies は request の vaccine species 文字列がドメイン enum として有効かを検証する。
// BE9-2D で internal/service/validators_master.go から *移設* した（複製ではない）: 唯一の consumer が
// この package の vaccineService だったため、内部の validators kernel を分割せず本関数だけを移した。
func validateVaccineSpecies(species string) error {
	if species == "" {
		return nil
	}
	switch model.VaccineSpecies(species) {
	case model.VaccineSpeciesDog, model.VaccineSpeciesCat, model.VaccineSpeciesBoth:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("ワクチン対象種別の値が不正です: %s", species))
	}
}
