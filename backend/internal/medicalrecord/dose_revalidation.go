package medicalrecord

// dose_revalidation.go — #201 B-2: treatment 保存時の BE 再検証（患者安全の核・healthcare HIGH-3/HIGH-4）。
//
// FE 警告のみでは バグ・改変・別経路保存で素通りするため、保存時にサーバ側で計算を再実行し、
//   - C1: 丸め後の「実際に保存される実効用量(mg)」で安全域を再判定（プリフィル値でなく保存値）
//   - 両上限の小さい方で cap 判定
//   - submitted vs computed の乖離検出（獣医の上書き＝逸脱）
// を行い、計算根拠スナップショットを値で固定する。species は pet から正規化した値で param を引くため
// 犬↔猫 越境は構造的に発生しない（HIGH-4）。B-1 の純粋関数 CalculateDose を再利用する（DRY）。
//
// TASK-377: 上限内の BelowMinSaved / DeviatesFromComputed では free-text 逸脱理由を必須とし、
// reason-required 経路の snapshot marshal は fail-closed（best-effort を使わない）。

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"unicode/utf8"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// doseDeviationReasonMaxRunes は transport の技術上限（臨床 taxonomy ではない）。
const doseDeviationReasonMaxRunes = 500

// doseSnapshotMarshal は test から marshal 失敗を注入できる hook（本番は json.Marshal）。
var doseSnapshotMarshal = json.Marshal

const doseAmountUnitMg = "mg"

// doseDeviationRelThreshold は「著しい逸脱」とみなす相対乖離の安全側デフォルト（暫定確定(2026-07-15 PO)・#201 B-5）。
// computed quantity に対して submitted がこの割合を超えて乖離したら逸脱として audit する。
const doseDeviationRelThreshold = 0.20

// doseFormulaVersion はスナップショットに固定する計算式の版。計算モデル変更時にインクリメントする。
const doseFormulaVersion = "v1"

// SavedDoseInput は保存される medicine 明細の dose 再検証入力。
type SavedDoseInput struct {
	Medicine     *model.Medicine
	Param        *model.MedicineDoseParam
	Species      model.MedicineDoseSpecies // pet から正規化済み（Param.Species と一致すべき）
	WeightKg     float64
	SubmittedQty float64
}

// SavedDoseEvaluation は保存時再検証の結果。
type SavedDoseEvaluation struct {
	Computed             DoseCalcResult
	WeightSource         string  // 体重の出典（vital_records:<id> 等）。snapshot 列 dose_weight_source 用
	SavedEffectiveMg     float64 // submitted qty × strength（実際に保存される実投与量）
	UpperCapMg           float64
	HasUpperCap          bool
	ExceedsCapSaved      bool // C1: 保存値の実効 mg が両上限の小さい方を超過
	BelowMinSaved        bool // 保存値の実効 mg が安全域下限を下回る
	DeviatesFromComputed bool // submitted vs computed の相対乖離が閾値超
	Snapshot             DoseSnapshot
}

// IsDeviation は audit すべき逸脱（過量/過少/著しい上書き）が存在するかを返す。
func (e *SavedDoseEvaluation) IsDeviation() bool {
	return e.ExceedsCapSaved || e.BelowMinSaved || e.DeviatesFromComputed
}

// RequiresDeviationReason は上限内の下限割れまたは著しい乖離で clinician 理由が必須かを返す。
// 上限超過は hard reject 済み前提で false（理由で解除しない）。
func (e *SavedDoseEvaluation) RequiresDeviationReason() bool {
	if e == nil || e.ExceedsCapSaved {
		return false
	}
	return e.BelowMinSaved || e.DeviatesFromComputed
}

// DoseSnapshot は treatment.dose_param_snapshot(jsonb) に値で固定する計算根拠。
// マスタの後変更・論理削除があっても当時の計算根拠を保全する（FK 非依存）。
type DoseSnapshot struct {
	Species              model.MedicineDoseSpecies `json:"species"`
	WeightKg             float64                   `json:"weight_kg"`
	DoseBasis            model.MedicineDoseBasis   `json:"dose_basis"`
	DosePerKg            float64                   `json:"dose_per_kg"`
	Strength             float64                   `json:"strength"`
	MinMgPerKg           *float64                  `json:"min_mg_per_kg,omitempty"`
	MaxMgPerKg           *float64                  `json:"max_mg_per_kg,omitempty"`
	AbsoluteMaxDose      *float64                  `json:"absolute_max_dose,omitempty"`
	ComputedQty          float64                   `json:"computed_quantity"`
	SubmittedQty         float64                   `json:"submitted_quantity"`
	EffectiveMg          float64                   `json:"effective_mg"`
	ExceedsMax           bool                      `json:"exceeds_max"`
	BelowMin             bool                      `json:"below_min"`
	DeviatesFromComputed bool                      `json:"deviates_from_computed"`
	// DoseDeviationReason は reason-required 保存時のみ値固定。safe re-evaluation では空にする。
	DoseDeviationReason string `json:"dose_deviation_reason,omitempty"`
	FormulaVersion      string `json:"formula_version"`
}

// doseCalcInputFrom は medicine + param + weight + species から B-1 計算入力を構築する。
func doseCalcInputFrom(in SavedDoseInput) DoseCalcInput {
	return DoseCalcInput{
		CalculationType: in.Medicine.CalculationType,
		MedicineUnit:    in.Medicine.MedicineUnit,
		Strength:        in.Medicine.Strength,
		DoseBasis:       in.Param.DoseBasis,
		FrequencyPerDay: in.Medicine.FrequencyPerDay,
		DosePerKg:       in.Param.DosePerKg,
		MinMgPerKg:      in.Param.MinMgPerKg,
		MaxMgPerKg:      in.Param.MaxMgPerKg,
		AbsoluteMaxDose: in.Param.AbsoluteMaxDose,
		RoundingStep:    in.Param.RoundingStep,
		RoundingMode:    in.Param.RoundingMode,
		WeightKg:        in.WeightKg,
	}
}

// EvaluateSavedDose は保存される quantity の安全性を BE 側で再判定する（純粋関数・B-1 CalculateDose 再利用）。
// 呼び出し側で medicine.CalculationType=per_weight・species 解決済み・weight>0 を保証していること。
func EvaluateSavedDose(in SavedDoseInput) SavedDoseEvaluation {
	calcIn := doseCalcInputFrom(in)
	computed := CalculateDose(calcIn)

	strength := 0.0
	if in.Medicine.Strength != nil {
		strength = *in.Medicine.Strength
	}
	basisFactor := doseBasisFactor(calcIn)
	savedMg := in.SubmittedQty * strength

	eval := SavedDoseEvaluation{
		Computed:         computed,
		SavedEffectiveMg: savedMg,
	}

	// C1: 保存値（丸め後・実際に投与される量）の実効 mg で両上限・下限を判定する。
	upperCap, hasCap := perAdministrationUpperCap(calcIn, basisFactor)
	eval.UpperCapMg = upperCap
	eval.HasUpperCap = hasCap
	if hasCap && savedMg > upperCap+safetyEpsilon {
		eval.ExceedsCapSaved = true
	}
	if in.Param.MinMgPerKg != nil {
		lowerMg := *in.Param.MinMgPerKg * in.WeightKg * basisFactor
		if savedMg < lowerMg-safetyEpsilon {
			eval.BelowMinSaved = true
		}
	}

	// computed（推奨プリフィル）からの相対乖離＝獣医の上書き検出。
	if computed.Eligible && computed.Quantity > 0 {
		rel := (in.SubmittedQty - computed.Quantity) / computed.Quantity
		if rel < 0 {
			rel = -rel
		}
		if rel > doseDeviationRelThreshold {
			eval.DeviatesFromComputed = true
		}
	}

	eval.Snapshot = DoseSnapshot{
		Species:              in.Species,
		WeightKg:             in.WeightKg,
		DoseBasis:            in.Param.DoseBasis,
		DosePerKg:            in.Param.DosePerKg,
		Strength:             strength,
		MinMgPerKg:           in.Param.MinMgPerKg,
		MaxMgPerKg:           in.Param.MaxMgPerKg,
		AbsoluteMaxDose:      in.Param.AbsoluteMaxDose,
		ComputedQty:          computed.Quantity,
		SubmittedQty:         in.SubmittedQty,
		EffectiveMg:          savedMg,
		ExceedsMax:           eval.ExceedsCapSaved,
		BelowMin:             eval.BelowMinSaved,
		DeviatesFromComputed: eval.DeviatesFromComputed,
		FormulaVersion:       doseFormulaVersion,
	}
	return eval
}

// normalizeDoseDeviationReason は transport 上限で free-text 理由を正規化する。
// reason-required でない呼び出しでは使わない（safe 経路で stale 理由を残さないため）。
func normalizeDoseDeviationReason(raw string) (string, error) {
	reason := strings.TrimSpace(raw)
	if reason == "" {
		return "", apperrors.WrapInvalidInput("用量逸脱の理由を入力してください")
	}
	if utf8.RuneCountInString(reason) > doseDeviationReasonMaxRunes {
		return "", apperrors.WrapInvalidInput("用量逸脱の理由は500文字以内で入力してください")
	}
	return reason, nil
}

// applyDeviationReasonToEval は eval に normalized reason を固定する。
// reason-required でない場合は snapshot から理由を必ず除去する（stale 防止）。
func applyDeviationReasonToEval(eval *SavedDoseEvaluation, rawReason string) error {
	if eval == nil {
		return nil
	}
	if !eval.RequiresDeviationReason() {
		eval.Snapshot.DoseDeviationReason = ""
		return nil
	}
	reason, err := normalizeDoseDeviationReason(rawReason)
	if err != nil {
		return err
	}
	eval.Snapshot.DoseDeviationReason = reason
	return nil
}

// marshalDoseSnapshot は計算根拠スナップショットを JSON 化する（失敗はベストエフォートで nil）。
// safe dose など reason-required でない経路専用。reason-required は marshalDoseSnapshotStrict を使う。
func marshalDoseSnapshot(ctx context.Context, eval *SavedDoseEvaluation) []byte {
	b, err := doseSnapshotMarshal(eval.Snapshot)
	if err != nil {
		slog.ErrorContext(ctx, "failed to marshal dose snapshot", "error", err)
		return nil
	}
	return b
}

// marshalDoseSnapshotStrict は reason-required 経路用。serialization 失敗を error で返す（黙って成功させない）。
func marshalDoseSnapshotStrict(eval *SavedDoseEvaluation) ([]byte, error) {
	b, err := doseSnapshotMarshal(eval.Snapshot)
	if err != nil {
		// 理由本文は log に出さない
		return nil, apperrors.Wrap(err, "failed to marshal dose param snapshot")
	}
	return b, nil
}

// doseSnapshotColumns は treatment の dose_* 列に書き込む値を GORM Updates 用 map で返す（Update 経路）。
// reason-required では strict marshal を使い、失敗時は error を返す。
func doseSnapshotColumns(ctx context.Context, eval *SavedDoseEvaluation) (map[string]any, error) {
	cols := map[string]any{
		"dose_weight_kg":   eval.Snapshot.WeightKg,
		"dose_amount_mg":   eval.SavedEffectiveMg,
		"dose_amount_unit": doseAmountUnitMg,
	}
	if eval.WeightSource != "" {
		cols["dose_weight_source"] = eval.WeightSource
	}
	if eval.RequiresDeviationReason() {
		snapJSON, err := marshalDoseSnapshotStrict(eval)
		if err != nil {
			return nil, err
		}
		cols["dose_param_snapshot"] = json.RawMessage(snapJSON)
		return cols, nil
	}
	if snapJSON := marshalDoseSnapshot(ctx, eval); snapJSON != nil {
		cols["dose_param_snapshot"] = json.RawMessage(snapJSON)
	}
	return cols, nil
}

// clearedDoseColumns は dose スナップショット列を NULL クリアする map を返す。
// per_weight 対象でなくなった（薬剤変更・item_type 変更）際に stale スナップショットを除去する（L-3）。
func clearedDoseColumns() map[string]any {
	return map[string]any{
		"dose_weight_kg":      nil,
		"dose_weight_source":  nil,
		"dose_amount_mg":      nil,
		"dose_amount_unit":    nil,
		"dose_param_snapshot": nil,
	}
}

// treatmentHasDoseSnapshot は treatment 行に dose スナップショットが付いているかを返す。
func treatmentHasDoseSnapshot(t *model.Treatment) bool {
	return t != nil && (t.DoseAmountMg != nil || t.DoseWeightKg != nil || len(t.DoseParamSnapshot) > 0)
}

// applyDoseSnapshotToTreatment は計算根拠スナップショットを treatment struct に値で固定する（Create 経路）。
// reason-required では strict marshal 失敗を error で返す。
func applyDoseSnapshotToTreatment(ctx context.Context, t *model.Treatment, eval *SavedDoseEvaluation) error {
	weight := eval.Snapshot.WeightKg
	mg := eval.SavedEffectiveMg
	unit := doseAmountUnitMg
	t.DoseWeightKg = &weight
	t.DoseAmountMg = &mg
	t.DoseAmountUnit = &unit
	if eval.WeightSource != "" {
		src := eval.WeightSource
		t.DoseWeightSource = &src
	}
	if eval.RequiresDeviationReason() {
		snapJSON, err := marshalDoseSnapshotStrict(eval)
		if err != nil {
			return err
		}
		t.DoseParamSnapshot = snapJSON
		return nil
	}
	if snapJSON := marshalDoseSnapshot(ctx, eval); snapJSON != nil {
		t.DoseParamSnapshot = snapJSON
	}
	return nil
}
