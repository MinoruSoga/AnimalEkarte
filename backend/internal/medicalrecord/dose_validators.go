package medicalrecord

// dose_validators.go — #201 薬マスタ書込時の投与量計算パラメータ検証（純粋関数）。
// BE9-2D ④b: dose_calc.go の eligibleMedicineUnitsForPerWeight（per_weight 適格単位の安全マップ）を
// 共有するため dose kernel ごと medicalrecord へ移動（マップ複製は許可集合のドリフト源になるため不可。
// medicine の domain 帰属は ADR-006 論点#4 で medicalrecord 確定済み）。internal/service の
// medicine/medicine_dose_param service は medicalrecord. 修飾で引き続き消費する（service→medicalrecord
// は合法方向）。
//
// default-deny / strength 単位整合 / range の二重化（DB CHECK と service 検証）をマスタ書込時に強制する。
// 計算時だけでなく書込時に弾くことで、誤 per_weight 設定（IU/%/非mg・含量欠落）が DB に入るのを防ぐ。

import (
	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	// maxReasonableDosePerKg / maxReasonableStrength は桁間違い等のタイプミスを弾く緩い上限。
	// 臨床上限ではない（false-reject 回避のため意図的に緩い）。確定閾値は暫定確定(2026-07-15 PO)（#201 確認事項）。
	maxReasonableDosePerKg = 100000.0  // mg/kg
	maxReasonableStrength  = 1000000.0 // mg/単位
)

// isPerWeightEligibleUnit は per_weight で strength の分母として mg 換算可能な単位かを判定する。
// per_dose（包/vial）と nil は不適格（含量の mg 解釈が曖昧 → IU/%/非mg の誤 per_weight を構造的に拒否）。
func isPerWeightEligibleUnit(unit *model.MedicineUnit) bool {
	if unit == nil {
		return false
	}
	_, ok := eligibleMedicineUnitsForPerWeight[*unit]
	return ok
}

// ValidateMedicineDoseConfig は medicines の製品軸計算設定を検証する（マスタ書込時）。
// calculation_type=none の既存薬剤は strength 等が未設定でも通る（後方互換・default-deny）。
func ValidateMedicineDoseConfig(
	calcType model.MedicineCalculationType,
	unit *model.MedicineUnit,
	strength *float64,
	frequencyPerDay, defaultDurationDays *int,
) error {
	switch calcType {
	case model.MedicineCalculationTypeNone:
		// 手動。strength 等の per_weight 整合は不問（後方互換）。
		// frequency_per_day / default_duration_days の正値検証は calculation_type に依らず後段で行う。
	case model.MedicineCalculationTypePerWeight:
		if !isPerWeightEligibleUnit(unit) {
			return apperrors.WrapInvalidInput("per_weight 計算には mg 換算可能な単位（per_tablet/per_ml/per_gram）が必要です。per_dose・未指定は使用できません")
		}
		if strength == nil {
			return apperrors.WrapInvalidInput("per_weight 計算には strength（製品含量）が必要です")
		}
		if *strength <= 0 {
			return apperrors.WrapInvalidInput("strength は正の値である必要があります")
		}
		if *strength > maxReasonableStrength {
			return apperrors.WrapInvalidInput("strength が大きすぎます（入力誤りの可能性）")
		}
	default:
		return apperrors.WrapInvalidInput("calculation_type が不正です")
	}

	if frequencyPerDay != nil && *frequencyPerDay <= 0 {
		return apperrors.WrapInvalidInput("frequency_per_day は正の値である必要があります")
	}
	if defaultDurationDays != nil && *defaultDurationDays <= 0 {
		return apperrors.WrapInvalidInput("default_duration_days は正の値である必要があります")
	}
	return nil
}

// MedicineDoseParamInput は種軸パラメータの作成/更新入力 DTO。
type MedicineDoseParamInput struct {
	Species         model.MedicineDoseSpecies
	DoseBasis       model.MedicineDoseBasis
	DosePerKg       float64
	MinMgPerKg      *float64
	MaxMgPerKg      *float64
	AbsoluteMaxDose *float64
	RoundingStep    *float64
	RoundingMode    *model.MedicineRoundingMode
	Notes           string
}

// validDoseBasis は dose_basis の許可集合。
func isValidDoseBasis(b model.MedicineDoseBasis) bool {
	return b == model.MedicineDoseBasisPerAdministration || b == model.MedicineDoseBasisPerDay
}

// isValidRoundingMode は rounding_mode の許可集合。
func isValidRoundingMode(m model.MedicineRoundingMode) bool {
	switch m {
	case model.MedicineRoundingModeUp, model.MedicineRoundingModeDown, model.MedicineRoundingModeNearest:
		return true
	}
	return false
}

// ValidateMedicineDoseParamInput は種軸パラメータを検証する（マスタ書込時・DB CHECK と二重化）。
func ValidateMedicineDoseParamInput(in *MedicineDoseParamInput) error {
	if in == nil {
		return apperrors.WrapInvalidInput("dose param input is required")
	}
	if !model.ValidMedicineDoseSpecies(in.Species) {
		return apperrors.WrapInvalidInput("species は dog または cat である必要があります")
	}
	if !isValidDoseBasis(in.DoseBasis) {
		return apperrors.WrapInvalidInput("dose_basis が不正です")
	}
	if in.DosePerKg <= 0 {
		return apperrors.WrapInvalidInput("dose_per_kg は正の値である必要があります")
	}
	if in.DosePerKg > maxReasonableDosePerKg {
		return apperrors.WrapInvalidInput("dose_per_kg が大きすぎます（入力誤りの可能性）")
	}
	if in.MinMgPerKg != nil && *in.MinMgPerKg <= 0 {
		return apperrors.WrapInvalidInput("min_mg_per_kg は正の値である必要があります")
	}
	if in.MaxMgPerKg != nil && *in.MaxMgPerKg <= 0 {
		return apperrors.WrapInvalidInput("max_mg_per_kg は正の値である必要があります")
	}
	if in.AbsoluteMaxDose != nil && *in.AbsoluteMaxDose <= 0 {
		return apperrors.WrapInvalidInput("absolute_max_dose は正の値である必要があります")
	}
	if in.MinMgPerKg != nil && in.MaxMgPerKg != nil && *in.MinMgPerKg > *in.MaxMgPerKg {
		return apperrors.WrapInvalidInput("min_mg_per_kg は max_mg_per_kg 以下である必要があります")
	}
	// 患者安全: base dose は自身の安全域 [min, max] 内である必要がある（healthcare review MEDIUM）。
	// 範囲外を許すと calc 側の clamp が入力 typo（例 5→50）を silent に上限へ吸収し、
	// 過量にはならないが意図と異なる用量が投与される。書込時に fail-fast で弾く。
	if in.MinMgPerKg != nil && in.DosePerKg < *in.MinMgPerKg {
		return apperrors.WrapInvalidInput("dose_per_kg は min_mg_per_kg 以上である必要があります")
	}
	if in.MaxMgPerKg != nil && in.DosePerKg > *in.MaxMgPerKg {
		return apperrors.WrapInvalidInput("dose_per_kg は max_mg_per_kg 以下である必要があります")
	}
	// 患者安全: per_weight 自動計算には上限が必須（healthcare review HIGH-1）。
	// 上限が一切ないと丸め上げで実効用量が silent に過量化しても検出できない。
	// max_mg_per_kg（体重連動）または absolute_max_dose（mg/head）の少なくとも一方を要求する。
	if in.MaxMgPerKg == nil && in.AbsoluteMaxDose == nil {
		return apperrors.WrapInvalidInput("max_mg_per_kg または absolute_max_dose の少なくとも一方を設定してください（過量防止の上限が必須）")
	}
	// 丸めは step と mode をペアで設定/未設定（片方だけは不正）。
	if (in.RoundingStep == nil) != (in.RoundingMode == nil) {
		return apperrors.WrapInvalidInput("rounding_step と rounding_mode はペアで指定してください")
	}
	if in.RoundingStep != nil && *in.RoundingStep <= 0 {
		return apperrors.WrapInvalidInput("rounding_step は正の値である必要があります")
	}
	if in.RoundingMode != nil && !isValidRoundingMode(*in.RoundingMode) {
		return apperrors.WrapInvalidInput("rounding_mode が不正です")
	}
	return nil
}
