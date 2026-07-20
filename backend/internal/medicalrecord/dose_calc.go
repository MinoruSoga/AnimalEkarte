package medicalrecord

// dose_calc.go — #201 カルテ薬量（投与量）自動計算の純粋関数群。
//
// この計算ロジックは FE(TS) への移植が一致するよう副作用なし・I/O なしの純粋関数として隔離する。
// DB アクセス・slog・apperrors はここに置かない（呼び出し側 service が担う）。
//
// 安全設計の核（#201）:
//   - C1: 安全域(min/max)判定は「丸め後の実効用量(mg)」に対して行う。
//         切上で実効 mg が上限を越えても丸め前チェックでは検出できない → ExceedsMax で必ずフラグ化（silent 越え禁止）。
//   - 両上限: 体重連動上限(weight × max_mg/kg) と 体重非依存上限(absolute_max_dose mg/head) の小さい方で cap。
//   - fail-closed: per_weight 以外・含量不明・体重未記録・per_day で頻度欠落・dose<=0 は Eligible=false（手動入力へ）。
//   - species hard gate: NormalizeDoseSpecies はマップ不能種を fail-closed（犬→猫の silent フォールバック禁止）。

import (
	"math"
	"strings"

	"github.com/animal-ekarte/backend/internal/model"
)

// roundEpsilon は浮動小数の丸め境界での誤検出を防ぐ微小許容値。
const roundEpsilon = 1e-9

// safetyEpsilon は実効用量(mg)の安全域比較に使う許容値（numeric(12,6) 精度に整合）。
const safetyEpsilon = 1e-6

// doseSpeciesAliases は free-text animal_species.name → 正規化種のエイリアス表。
// 安全のため exact-match（trim + ASCII 小文字化後）のみ。未知語は fail-closed で manual に落とす。
// 犬・猫を取り違えるエイリアスは絶対に追加しないこと（猫への犬用量は致死的になり得る）。
var doseSpeciesAliases = map[string]model.MedicineDoseSpecies{
	"犬":      model.MedicineDoseSpeciesDog,
	"いぬ":     model.MedicineDoseSpeciesDog,
	"イヌ":     model.MedicineDoseSpeciesDog,
	"ドッグ":    model.MedicineDoseSpeciesDog,
	"dog":    model.MedicineDoseSpeciesDog,
	"canine": model.MedicineDoseSpeciesDog,
	"猫":      model.MedicineDoseSpeciesCat,
	"ねこ":     model.MedicineDoseSpeciesCat,
	"ネコ":     model.MedicineDoseSpeciesCat,
	"キャット":   model.MedicineDoseSpeciesCat,
	"cat":    model.MedicineDoseSpeciesCat,
	"feline": model.MedicineDoseSpeciesCat,
}

// NormalizeDoseSpecies は free-text の種名を計算用の正規化種へ写像する。
// マップ不能な場合は ("", false) を返す（fail-closed）。犬↔猫の取り違え・silent フォールバックは行わない。
func NormalizeDoseSpecies(freeText string) (model.MedicineDoseSpecies, bool) {
	key := strings.ToLower(strings.TrimSpace(freeText))
	if key == "" {
		return "", false
	}
	s, ok := doseSpeciesAliases[key]
	return s, ok
}

// NormalizeWeightToKg は体重を kg に正規化する。weight <= 0 は (0, false)（fail-closed）。
func NormalizeWeightToKg(weight float64, unit model.BodyWeightUnit) (float64, bool) {
	if weight <= 0 {
		return 0, false
	}
	if unit == model.BodyWeightUnitG {
		return weight / 1000.0, true
	}
	// 既定（Kg・未指定）は kg そのまま。
	return weight, true
}

// DoseCalcInput は per_weight 計算の入力。WeightKg は NormalizeWeightToKg 済み（kg・> 0 期待）。
type DoseCalcInput struct {
	CalculationType model.MedicineCalculationType
	MedicineUnit    *model.MedicineUnit // strength の分母。per_tablet/per_ml/per_gram のみ適格（per_dose/nil は不適格）
	Strength        *float64            // mg/単位
	DoseBasis       model.MedicineDoseBasis
	FrequencyPerDay *int     // per_day 基準時に必須（> 0）
	DosePerKg       float64  // mg/kg
	MinMgPerKg      *float64 // 安全域下限（任意）
	MaxMgPerKg      *float64 // 体重連動上限（任意）
	AbsoluteMaxDose *float64 // 体重非依存 mg/head 上限（任意）
	RoundingStep    *float64 // 丸め単位（nil = 丸めなし）
	RoundingMode    *model.MedicineRoundingMode
	WeightKg        float64
}

// DoseCalcResult は計算結果と安全フラグ。Eligible=false の場合は計算値は使わず手動入力へ。
type DoseCalcResult struct {
	Eligible         bool
	IneligibleReason string

	EffectiveRate   float64 // clamp(dose_per_kg, min, max)
	RawMg           float64 // 丸め・cap 前の 1回投与あたり mg
	CappedMg        float64 // 両上限適用後・丸め前
	Quantity        float64 // 丸め後のプリフィル数量（単位数）
	EffectiveDoseMg float64 // Quantity × Strength（丸め後の実効用量。C1 判定の基準）

	// 安全フラグ（C1: いずれも丸め後の EffectiveDoseMg に対して判定）
	BelowMin   bool // 実効用量が安全域下限を下回る
	ExceedsMax bool // 実効用量が両上限の小さい方を上回る（切上による silent 越えを必ず検出）
}

// eligibleMedicineUnitsForPerWeight は per_weight 計算で含量分母として解釈可能な単位。
// per_dose（包/vial）は含量定義が曖昧なため除外。nil medicine_unit も不適格（fail-closed）。
var eligibleMedicineUnitsForPerWeight = map[model.MedicineUnit]struct{}{
	model.MedicineUnitPerTablet: {},
	model.MedicineUnitPerML:     {},
	model.MedicineUnitPerGram:   {},
}

// CalculateDose は #201 の計算モデルに従い 1回投与あたりの数量と実効用量を算出する。
//
// 計算順序:
//
//	effective_rate = clamp(dose_per_kg, min, max)
//	basis_factor   = 1 (per_administration) | 1/frequency_per_day (per_day)
//	raw_mg         = effective_rate × weight_kg × basis_factor
//	capped_mg      = min(raw_mg, weight × max_mg/kg × basis_factor, absolute_cap)  ← 設定済みの上限のみ
//	quantity       = round(capped_mg ÷ strength, step, mode)
//	effective_mg   = quantity × strength
//	安全域判定      = effective_mg（C1: 丸め後）を上限/下限と比較
//
// 注（per_day の上限解釈）: #201 の素の式は weight×max を raw（1回量）と同次元で比較するが、per_day では
// 次元が一致しない。dimensional に安全側へ揃えるため per_day では両上限にも basis_factor を適用し
// （absolute_cap は 1日総量とみなし frequency で按分）、全て 1回投与あたりに正規化して比較する。
// 支配的な per_administration（basis_factor=1）では #201 の式と完全一致する。
func CalculateDose(in DoseCalcInput) DoseCalcResult {
	if reason, ok := doseEligibility(in); !ok {
		return DoseCalcResult{Eligible: false, IneligibleReason: reason}
	}
	strength := *in.Strength

	basisFactor := doseBasisFactor(in)

	effectiveRate := clampRate(in.DosePerKg, in.MinMgPerKg, in.MaxMgPerKg)
	rawMg := effectiveRate * in.WeightKg * basisFactor

	// 両上限を 1回投与あたりに正規化して算出（設定済みのみ）。
	cappedMg := rawMg
	upperCapMg, hasUpperCap := perAdministrationUpperCap(in, basisFactor)
	if hasUpperCap && upperCapMg < cappedMg {
		cappedMg = upperCapMg
	}

	quantity := roundQuantity(cappedMg/strength, in.RoundingStep, in.RoundingMode)
	effectiveDoseMg := quantity * strength

	res := DoseCalcResult{
		Eligible:        true,
		EffectiveRate:   effectiveRate,
		RawMg:           rawMg,
		CappedMg:        cappedMg,
		Quantity:        quantity,
		EffectiveDoseMg: effectiveDoseMg,
	}

	// C1: 安全域判定は丸め後の実効用量(mg)で行う。
	if hasUpperCap && effectiveDoseMg > upperCapMg+safetyEpsilon {
		res.ExceedsMax = true
	}
	if in.MinMgPerKg != nil {
		lowerMg := *in.MinMgPerKg * in.WeightKg * basisFactor
		if effectiveDoseMg < lowerMg-safetyEpsilon {
			res.BelowMin = true
		}
	}
	return res
}

// doseEligibility は per_weight 自動計算の適格性を判定する（fail-closed）。
func doseEligibility(in DoseCalcInput) (string, bool) {
	if in.CalculationType != model.MedicineCalculationTypePerWeight {
		return "calculation_type is not per_weight", false
	}
	if in.MedicineUnit == nil {
		return "medicine_unit is required for per_weight", false
	}
	if _, ok := eligibleMedicineUnitsForPerWeight[*in.MedicineUnit]; !ok {
		return "medicine_unit is not eligible for per_weight", false
	}
	if in.Strength == nil || *in.Strength <= 0 {
		return "strength is missing or non-positive", false
	}
	if in.DosePerKg <= 0 {
		return "dose_per_kg is non-positive", false
	}
	if in.WeightKg <= 0 {
		return "weight is not recorded", false
	}
	if in.DoseBasis == model.MedicineDoseBasisPerDay && (in.FrequencyPerDay == nil || *in.FrequencyPerDay <= 0) {
		return "frequency_per_day is required for per_day basis", false
	}
	return "", true
}

// perAdministrationUpperCap は 1回投与あたりの上限（両上限の小さい方）を返す。
// 設定済みの上限がなければ (0, false)。
func perAdministrationUpperCap(in DoseCalcInput, basisFactor float64) (float64, bool) {
	upper := math.Inf(1)
	has := false
	if in.MaxMgPerKg != nil {
		weightCap := *in.MaxMgPerKg * in.WeightKg * basisFactor
		if weightCap < upper {
			upper = weightCap
		}
		has = true
	}
	if in.AbsoluteMaxDose != nil {
		absCap := *in.AbsoluteMaxDose
		if in.DoseBasis == model.MedicineDoseBasisPerDay {
			// 1日総量とみなし 1回あたりへ按分（per_day で frequency は適格性で > 0 を保証済み）。
			absCap *= basisFactor
		}
		if absCap < upper {
			upper = absCap
		}
		has = true
	}
	if !has {
		return 0, false
	}
	return upper, true
}

// doseBasisFactor は dose_basis に応じた 1回投与あたり係数を返す。
// per_administration=1 / per_day=1/frequency_per_day。per_day で frequency が無効なら 1 を返す
// （呼び出し前に doseEligibility で frequency>0 を保証している前提）。
func doseBasisFactor(in DoseCalcInput) float64 {
	if in.DoseBasis == model.MedicineDoseBasisPerDay && in.FrequencyPerDay != nil && *in.FrequencyPerDay > 0 {
		return 1.0 / float64(*in.FrequencyPerDay)
	}
	return 1.0
}

// clampRate は dose_per_kg を [lo, hi] に収める（設定済みのみ）。
// 引数名は Go 組込み min/max のシャドーイングを避けるため lo/hi とする。
func clampRate(rate float64, lo, hi *float64) float64 {
	if lo != nil && rate < *lo {
		rate = *lo
	}
	if hi != nil && rate > *hi {
		rate = *hi
	}
	return rate
}

// roundQuantity は value を step 刻みで mode 方向に丸める。step/mode が nil の場合は丸めない。
func roundQuantity(value float64, step *float64, mode *model.MedicineRoundingMode) float64 {
	if step == nil || mode == nil || *step <= 0 {
		return value
	}
	n := value / *step
	switch *mode {
	case model.MedicineRoundingModeUp:
		n = math.Ceil(n - roundEpsilon)
	case model.MedicineRoundingModeDown:
		n = math.Floor(n + roundEpsilon)
	case model.MedicineRoundingModeNearest:
		n = math.Round(n)
	default:
		return value
	}
	return n * *step
}
