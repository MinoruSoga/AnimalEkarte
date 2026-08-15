package medicalrecord

// dose_calc_test.go — #201 投与量自動計算の純粋関数 golden table-driven test。
// 計算モデル（#201）を仕様の単一の真実として固定する。FE(TS) 移植はこの期待値に一致させること。

import (
	"testing"

	"github.com/animal-ekarte/backend/internal/model"
)

// --- pointer helpers ---（fptr は service_deps_mock_test.go の既存 mirror を使用）

func iptr(v int) *int                                                { return &v }
func muptr(v model.MedicineUnit) *model.MedicineUnit                 { return &v }
func rmptr(v model.MedicineRoundingMode) *model.MedicineRoundingMode { return &v }

const floatTol = 1e-6

func floatEq(a, b float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= floatTol
}

func TestNormalizeDoseSpecies(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wantSp model.MedicineDoseSpecies
		wantOK bool
	}{
		{"犬 漢字", "犬", model.MedicineDoseSpeciesDog, true},
		{"いぬ ひらがな", "いぬ", model.MedicineDoseSpeciesDog, true},
		{"dog 英語 大文字混在", "Dog", model.MedicineDoseSpeciesDog, true},
		{"前後空白 trim", "  cat  ", model.MedicineDoseSpeciesCat, true},
		{"猫 漢字", "猫", model.MedicineDoseSpeciesCat, true},
		{"feline 英語", "feline", model.MedicineDoseSpeciesCat, true},
		{"マップ不能種は fail-closed（うさぎ）", "うさぎ", "", false},
		{"マップ不能種は fail-closed（rabbit）", "rabbit", "", false},
		{"空文字は fail-closed", "", "", false},
		{"未知の犬関連語は安全側で fail-closed（狂犬病等の取り違え防止）", "狂犬病", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSp, gotOK := NormalizeDoseSpecies(tt.input)
			if gotOK != tt.wantOK || gotSp != tt.wantSp {
				t.Fatalf("NormalizeDoseSpecies(%q) = (%q, %v), want (%q, %v)", tt.input, gotSp, gotOK, tt.wantSp, tt.wantOK)
			}
		})
	}
}

// TestNormalizeDoseSpecies_NoDogCatCrossover は犬↔猫の取り違えが起きないことを明示的に保証する。
func TestNormalizeDoseSpecies_NoDogCatCrossover(t *testing.T) {
	dogInputs := []string{"犬", "いぬ", "イヌ", "ドッグ", "dog", "canine"}
	for _, in := range dogInputs {
		if sp, _ := NormalizeDoseSpecies(in); sp == model.MedicineDoseSpeciesCat {
			t.Fatalf("犬入力 %q が cat に写像された（致死的取り違え）", in)
		}
	}
	catInputs := []string{"猫", "ねこ", "ネコ", "キャット", "cat", "feline"}
	for _, in := range catInputs {
		if sp, _ := NormalizeDoseSpecies(in); sp == model.MedicineDoseSpeciesDog {
			t.Fatalf("猫入力 %q が dog に写像された", in)
		}
	}
}

func TestNormalizeWeightToKg(t *testing.T) {
	tests := []struct {
		name   string
		weight float64
		unit   model.BodyWeightUnit
		wantKg float64
		wantOK bool
	}{
		{"Kg はそのまま", 4.2, model.BodyWeightUnitKg, 4.2, true},
		{"g は kg へ", 850, model.BodyWeightUnitG, 0.85, true},
		{"単位未指定（既定 Kg）", 3.0, model.BodyWeightUnit(""), 3.0, true},
		{"0 は fail-closed", 0, model.BodyWeightUnitKg, 0, false},
		{"負値は fail-closed", -1, model.BodyWeightUnitKg, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKg, gotOK := NormalizeWeightToKg(tt.weight, tt.unit)
			if gotOK != tt.wantOK || !floatEq(gotKg, tt.wantKg) {
				t.Fatalf("NormalizeWeightToKg(%v,%q) = (%v,%v), want (%v,%v)", tt.weight, tt.unit, gotKg, gotOK, tt.wantKg, tt.wantOK)
			}
		})
	}
}

func TestCalculateDose_Eligibility_FailClosed(t *testing.T) {
	base := func() DoseCalcInput {
		return DoseCalcInput{
			CalculationType: model.MedicineCalculationTypePerWeight,
			MedicineUnit:    muptr(model.MedicineUnitPerTablet),
			Strength:        fptr(10),
			DoseBasis:       model.MedicineDoseBasisPerAdministration,
			DosePerKg:       5,
			WeightKg:        4,
		}
	}
	tests := []struct {
		name   string
		mutate func(*DoseCalcInput)
	}{
		{"calculation_type=none", func(in *DoseCalcInput) { in.CalculationType = model.MedicineCalculationTypeNone }},
		{"medicine_unit nil", func(in *DoseCalcInput) { in.MedicineUnit = nil }},
		{"medicine_unit per_dose は不適格", func(in *DoseCalcInput) { in.MedicineUnit = muptr(model.MedicineUnitPerDose) }},
		{"strength nil", func(in *DoseCalcInput) { in.Strength = nil }},
		{"strength <= 0", func(in *DoseCalcInput) { in.Strength = fptr(0) }},
		{"dose_per_kg <= 0", func(in *DoseCalcInput) { in.DosePerKg = 0 }},
		{"weight <= 0（未記録）", func(in *DoseCalcInput) { in.WeightKg = 0 }},
		{"per_day で frequency 欠落", func(in *DoseCalcInput) {
			in.DoseBasis = model.MedicineDoseBasisPerDay
			in.FrequencyPerDay = nil
		}},
		{"per_day で frequency <= 0", func(in *DoseCalcInput) {
			in.DoseBasis = model.MedicineDoseBasisPerDay
			in.FrequencyPerDay = iptr(0)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := base()
			tt.mutate(&in)
			got := CalculateDose(in)
			if got.Eligible {
				t.Fatalf("Eligible=true, want false (fail-closed): reason=%q", got.IneligibleReason)
			}
			if got.IneligibleReason == "" {
				t.Fatal("IneligibleReason should explain why calculation was skipped")
			}
		})
	}
}

func TestCalculateDose_Calculation(t *testing.T) {
	tests := []struct {
		name string
		in   DoseCalcInput

		wantQty        float64
		wantEffMg      float64
		wantBelowMin   bool
		wantExceedsMax bool
	}{
		{
			name: "基本: 錠剤 per_administration 丸めなし",
			in: DoseCalcInput{
				CalculationType: model.MedicineCalculationTypePerWeight,
				MedicineUnit:    muptr(model.MedicineUnitPerTablet),
				Strength:        fptr(10), DoseBasis: model.MedicineDoseBasisPerAdministration,
				DosePerKg: 5, WeightKg: 4,
			},
			wantQty: 2, wantEffMg: 20,
		},
		{
			name: "液剤: per_ml で小数数量（silent 丸めなし）",
			in: DoseCalcInput{
				CalculationType: model.MedicineCalculationTypePerWeight,
				MedicineUnit:    muptr(model.MedicineUnitPerML),
				Strength:        fptr(50), DoseBasis: model.MedicineDoseBasisPerAdministration,
				DosePerKg: 10, WeightKg: 2.5,
			},
			wantQty: 0.5, wantEffMg: 25,
		},
		{
			name: "丸め上げ（上限未設定なので逸脱なし）",
			in: DoseCalcInput{
				CalculationType: model.MedicineCalculationTypePerWeight,
				MedicineUnit:    muptr(model.MedicineUnitPerTablet),
				Strength:        fptr(20), DoseBasis: model.MedicineDoseBasisPerAdministration,
				DosePerKg: 5, WeightKg: 3,
				RoundingStep: fptr(1), RoundingMode: rmptr(model.MedicineRoundingModeUp),
			},
			wantQty: 1, wantEffMg: 20,
		},
		{
			name: "C1 回帰: 切上で実効用量が上限を越える → ExceedsMax フラグ（silent 越え禁止）",
			in: DoseCalcInput{
				CalculationType: model.MedicineCalculationTypePerWeight,
				MedicineUnit:    muptr(model.MedicineUnitPerTablet),
				Strength:        fptr(10), DoseBasis: model.MedicineDoseBasisPerAdministration,
				DosePerKg: 5, WeightKg: 1.7, MaxMgPerKg: fptr(5),
				RoundingStep: fptr(1), RoundingMode: rmptr(model.MedicineRoundingModeUp),
			},
			// rawMg=8.5, cap=8.5, ceil(0.85)=1錠, effMg=10 > 8.5 → ExceedsMax
			wantQty: 1, wantEffMg: 10, wantExceedsMax: true,
		},
		{
			name: "両上限: 大型犬で absolute_max_dose が binding（小さい方で cap）",
			in: DoseCalcInput{
				CalculationType: model.MedicineCalculationTypePerWeight,
				MedicineUnit:    muptr(model.MedicineUnitPerTablet),
				Strength:        fptr(100), DoseBasis: model.MedicineDoseBasisPerAdministration,
				DosePerKg: 10, WeightKg: 40, MaxMgPerKg: fptr(10), AbsoluteMaxDose: fptr(300),
			},
			// rawMg=400, weightCap=400, absCap=300 → cap=300 → 3錠, effMg=300（境界一致でフラグなし）
			wantQty: 3, wantEffMg: 300,
		},
		{
			name: "min クランプ: dose_per_kg が下限未満 → 下限へ引上げ",
			in: DoseCalcInput{
				CalculationType: model.MedicineCalculationTypePerWeight,
				MedicineUnit:    muptr(model.MedicineUnitPerTablet),
				Strength:        fptr(10), DoseBasis: model.MedicineDoseBasisPerAdministration,
				DosePerKg: 2, MinMgPerKg: fptr(5), WeightKg: 4,
			},
			wantQty: 2, wantEffMg: 20,
		},
		{
			name: "max クランプ: dose_per_kg が上限超過 → 上限へ引下げ",
			in: DoseCalcInput{
				CalculationType: model.MedicineCalculationTypePerWeight,
				MedicineUnit:    muptr(model.MedicineUnitPerTablet),
				Strength:        fptr(10), DoseBasis: model.MedicineDoseBasisPerAdministration,
				DosePerKg: 20, MaxMgPerKg: fptr(5), WeightKg: 2,
			},
			wantQty: 1, wantEffMg: 10,
		},
		{
			name: "BelowMin: 丸め下げで実効用量が下限割れ → フラグ",
			in: DoseCalcInput{
				CalculationType: model.MedicineCalculationTypePerWeight,
				MedicineUnit:    muptr(model.MedicineUnitPerTablet),
				Strength:        fptr(10), DoseBasis: model.MedicineDoseBasisPerAdministration,
				DosePerKg: 5, MinMgPerKg: fptr(4), WeightKg: 1.5,
				RoundingStep: fptr(1), RoundingMode: rmptr(model.MedicineRoundingModeDown),
			},
			// rawMg=7.5, floor(0.75)=0錠, effMg=0 < lower(6) → BelowMin
			wantQty: 0, wantEffMg: 0, wantBelowMin: true,
		},
		{
			name: "per_day 基準: frequency で 1回量へ按分",
			in: DoseCalcInput{
				CalculationType: model.MedicineCalculationTypePerWeight,
				MedicineUnit:    muptr(model.MedicineUnitPerTablet),
				Strength:        fptr(25), DoseBasis: model.MedicineDoseBasisPerDay,
				FrequencyPerDay: iptr(2), DosePerKg: 10, WeightKg: 5,
			},
			// basisFactor=0.5, rawMg=10*5*0.5=25, 25/25=1錠
			wantQty: 1, wantEffMg: 25,
		},
		{
			name: "per_day 基準: absolute_max_dose を1日総量とみなし按分して cap",
			in: DoseCalcInput{
				CalculationType: model.MedicineCalculationTypePerWeight,
				MedicineUnit:    muptr(model.MedicineUnitPerML),
				Strength:        fptr(50), DoseBasis: model.MedicineDoseBasisPerDay,
				FrequencyPerDay: iptr(2), DosePerKg: 40, WeightKg: 10, AbsoluteMaxDose: fptr(300),
			},
			// basisFactor=0.5, rawMg=40*10*0.5=200, absCap=300*0.5=150 → cap=150 → 3mL, effMg=150
			wantQty: 3, wantEffMg: 150,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateDose(tt.in)
			if !got.Eligible {
				t.Fatalf("Eligible=false unexpectedly: %q", got.IneligibleReason)
			}
			if !floatEq(got.Quantity, tt.wantQty) {
				t.Errorf("Quantity = %v, want %v", got.Quantity, tt.wantQty)
			}
			if !floatEq(got.EffectiveDoseMg, tt.wantEffMg) {
				t.Errorf("EffectiveDoseMg = %v, want %v", got.EffectiveDoseMg, tt.wantEffMg)
			}
			if got.BelowMin != tt.wantBelowMin {
				t.Errorf("BelowMin = %v, want %v", got.BelowMin, tt.wantBelowMin)
			}
			if got.ExceedsMax != tt.wantExceedsMax {
				t.Errorf("ExceedsMax = %v, want %v", got.ExceedsMax, tt.wantExceedsMax)
			}
		})
	}
}

// TestRoundQuantity は roundQuantity の丸め分岐を直接網羅する（CalculateDose 経由では
// Nearest モードや無効な mode・step<=0・mode nil のケースを再現しづらいため）。
func TestRoundQuantity(t *testing.T) {
	up := model.MedicineRoundingModeUp
	down := model.MedicineRoundingModeDown
	nearest := model.MedicineRoundingModeNearest
	invalid := model.MedicineRoundingMode("invalid_mode")

	tests := []struct {
		name  string
		value float64
		step  *float64
		mode  *model.MedicineRoundingMode
		want  float64
	}{
		{"step nil は丸めない", 1.234, nil, &up, 1.234},
		{"mode nil は丸めない", 1.234, fptr(1), nil, 1.234},
		{"step も mode も nil", 1.234, nil, nil, 1.234},
		{"step<=0 は丸めない", 1.234, fptr(0), &up, 1.234},
		{"step 負値は丸めない", 1.234, fptr(-1), &up, 1.234},
		{"Up: 切上", 1.1, fptr(1), &up, 2},
		{"Down: 切下", 1.9, fptr(1), &down, 1},
		{"Nearest: 四捨五入（切上側）", 1.5, fptr(1), &nearest, 2},
		{"Nearest: 四捨五入（切下側）", 1.4, fptr(1), &nearest, 1},
		{"Nearest: step=0.5刻み", 1.3, fptr(0.5), &nearest, 1.5},
		{"未知の mode は丸めない（fail-safe）", 1.234, fptr(1), &invalid, 1.234},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := roundQuantity(tt.value, tt.step, tt.mode)
			if !floatEq(got, tt.want) {
				t.Fatalf("roundQuantity(%v, step=%v, mode=%v) = %v, want %v", tt.value, tt.step, tt.mode, got, tt.want)
			}
		})
	}
}

// TestCalculateDose_C1_NoSilentBreach は #201 の最重要回帰: 上限が設定されている限り、
// 丸め後の実効用量が上限を越えるなら必ず ExceedsMax=true になることを境界で網羅する。
func TestCalculateDose_C1_NoSilentBreach(t *testing.T) {
	// strength=10 mg/錠, max=5 mg/kg, 切上, 体重を変えて ceil が上限を越える領域を走査。
	for _, w := range []float64{1.1, 1.3, 1.7, 1.9, 2.1, 2.3, 2.7, 2.9} {
		in := DoseCalcInput{
			CalculationType: model.MedicineCalculationTypePerWeight,
			MedicineUnit:    muptr(model.MedicineUnitPerTablet),
			Strength:        fptr(10), DoseBasis: model.MedicineDoseBasisPerAdministration,
			DosePerKg: 5, WeightKg: w, MaxMgPerKg: fptr(5),
			RoundingStep: fptr(1), RoundingMode: rmptr(model.MedicineRoundingModeUp),
		}
		got := CalculateDose(in)
		upper := 5 * w // weight × max_mg/kg（basisFactor=1）
		if got.EffectiveDoseMg > upper+1e-6 && !got.ExceedsMax {
			t.Fatalf("weight=%v: 実効用量 %v が上限 %v を silent に越えた（ExceedsMax=false）", w, got.EffectiveDoseMg, upper)
		}
		if got.EffectiveDoseMg <= upper+1e-6 && got.ExceedsMax {
			t.Fatalf("weight=%v: 上限内 %v<=%v なのに ExceedsMax=true（過検出）", w, got.EffectiveDoseMg, upper)
		}
	}
}
