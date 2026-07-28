package medicalrecord

// dose_validators_test.go — #201 マスタ書込時バリデーションの table-driven test。
// default-deny・strength 単位整合（誤 per_weight 拒否）・range 二重化を固定する。

import (
	"testing"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestValidateMedicineDoseConfig(t *testing.T) {
	tests := []struct {
		name      string
		calcType  model.MedicineCalculationType
		unit      *model.MedicineUnit
		strength  *float64
		freq      *int
		duration  *int
		wantError bool
	}{
		{"none は計算パラメータ不問（後方互換）", model.MedicineCalculationTypeNone, nil, nil, nil, nil, false},
		{"per_weight + 錠 + strength は通る", model.MedicineCalculationTypePerWeight, muptr(model.MedicineUnitPerTablet), fptr(10), iptr(2), iptr(7), false},
		{"per_weight + mL + strength は通る", model.MedicineCalculationTypePerWeight, muptr(model.MedicineUnitPerML), fptr(50), nil, nil, false},
		{"per_weight + per_dose は拒否（誤 per_weight）", model.MedicineCalculationTypePerWeight, muptr(model.MedicineUnitPerDose), fptr(10), nil, nil, true},
		{"per_weight + unit nil は拒否", model.MedicineCalculationTypePerWeight, nil, fptr(10), nil, nil, true},
		{"per_weight + strength nil は拒否", model.MedicineCalculationTypePerWeight, muptr(model.MedicineUnitPerTablet), nil, nil, nil, true},
		{"per_weight + strength<=0 は拒否", model.MedicineCalculationTypePerWeight, muptr(model.MedicineUnitPerTablet), fptr(0), nil, nil, true},
		{"strength 過大は拒否（タイプミス）", model.MedicineCalculationTypePerWeight, muptr(model.MedicineUnitPerTablet), fptr(maxReasonableStrength + 1), nil, nil, true},
		{"frequency<=0 は拒否", model.MedicineCalculationTypePerWeight, muptr(model.MedicineUnitPerTablet), fptr(10), iptr(0), nil, true},
		{"duration<=0 は拒否", model.MedicineCalculationTypePerWeight, muptr(model.MedicineUnitPerTablet), fptr(10), nil, iptr(0), true},
		{"不正な calculation_type は拒否", model.MedicineCalculationType("bsa"), muptr(model.MedicineUnitPerTablet), fptr(10), nil, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMedicineDoseConfig(tt.calcType, tt.unit, tt.strength, tt.freq, tt.duration)
			if tt.wantError && err == nil {
				t.Fatal("want error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Fatalf("want nil, got %v", err)
			}
			if tt.wantError && err != nil && !apperrors.IsInvalidInput(err) {
				t.Errorf("error should be InvalidInput: %v", err)
			}
		})
	}
}

func TestValidateMedicineDoseParamInput(t *testing.T) {
	valid := func() *MedicineDoseParamInput {
		return &MedicineDoseParamInput{
			Species:    model.MedicineDoseSpeciesDog,
			DoseBasis:  model.MedicineDoseBasisPerAdministration,
			DosePerKg:  5,
			MaxMgPerKg: fptr(10), // per_weight には上限必須（患者安全）
		}
	}
	tests := []struct {
		name      string
		mutate    func(*MedicineDoseParamInput)
		wantError bool
	}{
		{"最小構成（上限あり）は通る", func(*MedicineDoseParamInput) {}, false},
		{"absolute_max_dose のみでも上限要件を満たす", func(in *MedicineDoseParamInput) {
			in.MaxMgPerKg = nil
			in.AbsoluteMaxDose = fptr(300)
		}, false},
		{"上限が一切ないと拒否（過量防止・HIGH-1）", func(in *MedicineDoseParamInput) {
			in.MaxMgPerKg = nil
			in.AbsoluteMaxDose = nil
		}, true},
		{"min/max/abs + 丸めペア全部は通る", func(in *MedicineDoseParamInput) {
			in.MinMgPerKg = fptr(2)
			in.MaxMgPerKg = fptr(10)
			in.AbsoluteMaxDose = fptr(300)
			in.RoundingStep = fptr(0.5)
			in.RoundingMode = rmptr(model.MedicineRoundingModeNearest)
		}, false},
		{"不正 species は拒否（fail-closed）", func(in *MedicineDoseParamInput) { in.Species = model.MedicineDoseSpecies("rabbit") }, true},
		{"空 species は拒否", func(in *MedicineDoseParamInput) { in.Species = "" }, true},
		{"不正 dose_basis は拒否", func(in *MedicineDoseParamInput) { in.DoseBasis = model.MedicineDoseBasis("weekly") }, true},
		{"dose_per_kg<=0 は拒否", func(in *MedicineDoseParamInput) { in.DosePerKg = 0 }, true},
		{"dose_per_kg 過大は拒否", func(in *MedicineDoseParamInput) { in.DosePerKg = maxReasonableDosePerKg + 1 }, true},
		{"min>max は拒否", func(in *MedicineDoseParamInput) { in.MinMgPerKg = fptr(10); in.MaxMgPerKg = fptr(5) }, true},
		{"min<=0 は拒否", func(in *MedicineDoseParamInput) { in.MinMgPerKg = fptr(0) }, true},
		{"dose_per_kg > max は拒否（typo の silent 吸収防止）", func(in *MedicineDoseParamInput) { in.DosePerKg = 50 }, true},
		{"dose_per_kg < min は拒否", func(in *MedicineDoseParamInput) { in.MinMgPerKg = fptr(8) }, true},
		{"dose_per_kg == max は通る（境界・inclusive）", func(in *MedicineDoseParamInput) { in.DosePerKg = 10 }, false},
		{"dose_per_kg == min は通る（境界・inclusive）", func(in *MedicineDoseParamInput) {
			in.MinMgPerKg = fptr(5) // baseline dose_per_kg=5
		}, false},
		{"丸め step だけ指定は拒否（ペア違反）", func(in *MedicineDoseParamInput) { in.RoundingStep = fptr(1) }, true},
		{"丸め mode だけ指定は拒否（ペア違反）", func(in *MedicineDoseParamInput) { in.RoundingMode = rmptr(model.MedicineRoundingModeUp) }, true},
		{"rounding_step<=0 は拒否", func(in *MedicineDoseParamInput) {
			in.RoundingStep = fptr(0)
			in.RoundingMode = rmptr(model.MedicineRoundingModeUp)
		}, true},
		{"不正 rounding_mode は拒否", func(in *MedicineDoseParamInput) {
			in.RoundingStep = fptr(1)
			in.RoundingMode = rmptr(model.MedicineRoundingMode("bankers"))
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := valid()
			tt.mutate(in)
			err := ValidateMedicineDoseParamInput(in)
			if tt.wantError && err == nil {
				t.Fatal("want error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Fatalf("want nil, got %v", err)
			}
			if tt.wantError && err != nil && !apperrors.IsInvalidInput(err) {
				t.Errorf("error should be InvalidInput: %v", err)
			}
		})
	}
}
