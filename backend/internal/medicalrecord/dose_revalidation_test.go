package medicalrecord

// dose_revalidation_test.go — #201 B-2: 保存時 BE 再検証（EvaluateSavedDose）の純粋テスト。
// C1 は「保存される実効用量(mg)」で判定する点・両上限・逸脱検出・スナップショット値固定を固定する。

import (
	"context"
	"math"
	"testing"

	"github.com/animal-ekarte/backend/internal/model"
)

func perWeightTabletMedicine(strength float64) *model.Medicine {
	s := strength
	unit := model.MedicineUnitPerTablet
	return &model.Medicine{
		CalculationType: model.MedicineCalculationTypePerWeight,
		MedicineUnit:    &unit,
		Strength:        &s,
	}
}

func TestEvaluateSavedDose(t *testing.T) {
	tests := []struct {
		name string

		strength     float64
		dosePerKg    float64
		maxMgPerKg   *float64
		minMgPerKg   *float64
		absoluteMax  *float64
		weightKg     float64
		submittedQty float64

		wantSavedMg     float64
		wantExceedsCap  bool
		wantBelowMin    bool
		wantDeviates    bool
		wantIsDeviation bool
	}{
		{
			name:     "保存値が推奨どおり・上限内 → 逸脱なし",
			strength: 10, dosePerKg: 5, maxMgPerKg: fptr(5), weightKg: 4, submittedQty: 2,
			wantSavedMg: 20, // computed も 2錠=20mg
		},
		{
			name:     "C1: 保存値(上書き)が上限を超過 → ExceedsCapSaved（silent 過量保存を検出）",
			strength: 10, dosePerKg: 5, maxMgPerKg: fptr(5), weightKg: 4, submittedQty: 3,
			// cap=5*4=20mg, 保存3錠=30mg>20 → 過量
			wantSavedMg: 30, wantExceedsCap: true, wantDeviates: true, wantIsDeviation: true,
		},
		{
			name:     "回帰: per_weight を範囲外 quantity で保存 → 過量フラグ（記録なし silent 保存の閉鎖）",
			strength: 10, dosePerKg: 5, maxMgPerKg: fptr(5), absoluteMax: fptr(40), weightKg: 4, submittedQty: 10,
			// cap=min(20,40)=20mg, 保存10錠=100mg → 大幅過量
			wantSavedMg: 100, wantExceedsCap: true, wantDeviates: true, wantIsDeviation: true,
		},
		{
			name:     "両上限: absolute_max_dose が binding で保存値判定",
			strength: 100, dosePerKg: 10, maxMgPerKg: fptr(10), absoluteMax: fptr(300), weightKg: 40, submittedQty: 4,
			// cap=min(10*40=400, 300)=300mg, 保存4錠=400mg>300 → 過量
			wantSavedMg: 400, wantExceedsCap: true, wantIsDeviation: true,
			// computed=3錠(300mg) なので 4 は rel=0.33>0.2 → deviates も true
			wantDeviates: true,
		},
		{
			name:     "BelowMin: 保存値が安全域下限を下回る",
			strength: 10, dosePerKg: 5, maxMgPerKg: fptr(8), minMgPerKg: fptr(3), weightKg: 4, submittedQty: 1,
			// lowerMg=3*4=12mg, 保存1錠=10mg<12 → 過少
			wantSavedMg: 10, wantBelowMin: true, wantDeviates: true, wantIsDeviation: true,
		},
		{
			name:     "著しい上書き: 上限内だが computed から 20% 超乖離 → DeviatesFromComputed",
			strength: 10, dosePerKg: 5, maxMgPerKg: fptr(10), weightKg: 4, submittedQty: 3,
			// computed: rate=5, raw=20, cap=10*4=40 → 2錠. 保存3錠=30mg<=40(上限内) だが rel=0.5>0.2
			wantSavedMg: 30, wantDeviates: true, wantIsDeviation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			med := perWeightTabletMedicine(tt.strength)
			param := &model.MedicineDoseParam{
				Species:         model.MedicineDoseSpeciesDog,
				DoseBasis:       model.MedicineDoseBasisPerAdministration,
				DosePerKg:       tt.dosePerKg,
				MinMgPerKg:      tt.minMgPerKg,
				MaxMgPerKg:      tt.maxMgPerKg,
				AbsoluteMaxDose: tt.absoluteMax,
			}
			eval := EvaluateSavedDose(SavedDoseInput{
				Medicine:     med,
				Param:        param,
				Species:      model.MedicineDoseSpeciesDog,
				WeightKg:     tt.weightKg,
				SubmittedQty: tt.submittedQty,
			})

			if !floatEq(eval.SavedEffectiveMg, tt.wantSavedMg) {
				t.Errorf("SavedEffectiveMg = %v, want %v", eval.SavedEffectiveMg, tt.wantSavedMg)
			}
			if eval.ExceedsCapSaved != tt.wantExceedsCap {
				t.Errorf("ExceedsCapSaved = %v, want %v", eval.ExceedsCapSaved, tt.wantExceedsCap)
			}
			if eval.BelowMinSaved != tt.wantBelowMin {
				t.Errorf("BelowMinSaved = %v, want %v", eval.BelowMinSaved, tt.wantBelowMin)
			}
			if eval.DeviatesFromComputed != tt.wantDeviates {
				t.Errorf("DeviatesFromComputed = %v, want %v", eval.DeviatesFromComputed, tt.wantDeviates)
			}
			if eval.IsDeviation() != tt.wantIsDeviation {
				t.Errorf("IsDeviation = %v, want %v", eval.IsDeviation(), tt.wantIsDeviation)
			}
		})
	}
}

// TestEvaluateSavedDose_Snapshot はスナップショットが当時の計算根拠を値で保持することを固定する。
func TestEvaluateSavedDose_Snapshot(t *testing.T) {
	med := perWeightTabletMedicine(10)
	param := &model.MedicineDoseParam{
		Species:    model.MedicineDoseSpeciesCat,
		DoseBasis:  model.MedicineDoseBasisPerAdministration,
		DosePerKg:  5,
		MaxMgPerKg: fptr(5),
	}
	eval := EvaluateSavedDose(SavedDoseInput{
		Medicine: med, Param: param, Species: model.MedicineDoseSpeciesCat, WeightKg: 4, SubmittedQty: 2,
	})
	snap := eval.Snapshot
	if snap.Species != model.MedicineDoseSpeciesCat {
		t.Errorf("snapshot species = %v, want cat", snap.Species)
	}
	if !floatEq(snap.WeightKg, 4) || !floatEq(snap.DosePerKg, 5) || !floatEq(snap.Strength, 10) {
		t.Errorf("snapshot core values wrong: %+v", snap)
	}
	if !floatEq(snap.SubmittedQty, 2) || !floatEq(snap.EffectiveMg, 20) {
		t.Errorf("snapshot submitted/effective wrong: %+v", snap)
	}
	if snap.FormulaVersion != doseFormulaVersion {
		t.Errorf("snapshot formula version = %q, want %q", snap.FormulaVersion, doseFormulaVersion)
	}
}

// TestClearedDoseColumns は per_weight 対象でなくなった治療の dose スナップショット列を
// NULL クリアする map の内容を固定する（L-3: stale スナップショット除去）。
func TestClearedDoseColumns(t *testing.T) {
	got := clearedDoseColumns()
	want := map[string]any{
		"dose_weight_kg":      nil,
		"dose_weight_source":  nil,
		"dose_amount_mg":      nil,
		"dose_amount_unit":    nil,
		"dose_param_snapshot": nil,
	}
	if len(got) != len(want) {
		t.Fatalf("clearedDoseColumns() len = %d, want %d", len(got), len(want))
	}
	for k, wantV := range want {
		gotV, ok := got[k]
		if !ok {
			t.Errorf("clearedDoseColumns() missing key %q", k)
			continue
		}
		if gotV != wantV {
			t.Errorf("clearedDoseColumns()[%q] = %v, want %v", k, gotV, wantV)
		}
	}
}

// TestMarshalDoseSnapshot_MarshalError は JSON マーシャル失敗時（例: 非有限浮動小数点値）に
// ベストエフォートで nil を返すことを検証する（監査記録の欠落は許容し治療保存自体は継続させる設計）。
func TestMarshalDoseSnapshot_MarshalError(t *testing.T) {
	med := perWeightTabletMedicine(10)
	param := &model.MedicineDoseParam{
		Species:    model.MedicineDoseSpeciesDog,
		DoseBasis:  model.MedicineDoseBasisPerAdministration,
		DosePerKg:  5,
		MaxMgPerKg: fptr(5),
	}
	// WeightKg=+Inf は json.Marshal で "unsupported value" エラーになる非有限浮動小数点値。
	eval := EvaluateSavedDose(SavedDoseInput{
		Medicine: med, Param: param, Species: model.MedicineDoseSpeciesDog,
		WeightKg: math.Inf(1), SubmittedQty: 2,
	})

	got := marshalDoseSnapshot(context.Background(), &eval)

	if got != nil {
		t.Errorf("marshalDoseSnapshot() with non-finite float = %v, want nil (best-effort)", got)
	}
}

// TestMarshalDoseSnapshot_Success は正常系で有効な JSON バイト列を返すことを検証する。
func TestMarshalDoseSnapshot_Success(t *testing.T) {
	med := perWeightTabletMedicine(10)
	param := &model.MedicineDoseParam{
		Species:    model.MedicineDoseSpeciesDog,
		DoseBasis:  model.MedicineDoseBasisPerAdministration,
		DosePerKg:  5,
		MaxMgPerKg: fptr(5),
	}
	eval := EvaluateSavedDose(SavedDoseInput{
		Medicine: med, Param: param, Species: model.MedicineDoseSpeciesDog, WeightKg: 4, SubmittedQty: 2,
	})

	got := marshalDoseSnapshot(context.Background(), &eval)

	if got == nil {
		t.Fatal("marshalDoseSnapshot() = nil, want non-nil JSON bytes for a valid snapshot")
	}
}
