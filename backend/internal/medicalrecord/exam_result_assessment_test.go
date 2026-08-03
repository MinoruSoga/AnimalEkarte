package medicalrecord

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestQualitativeValueOrder(t *testing.T) {
	values := []struct {
		value string
		index int
	}{
		{value: "(-)", index: 0},
		{value: "(±)", index: 1},
		{value: "(+)", index: 2},
		{value: "(++)", index: 3},
		{value: "(+++)", index: 4},
	}

	for _, tt := range values {
		got, ok := qualitativeValueIndex(tt.value)
		assert.True(t, ok, "qualitativeValueIndex(%q) must recognize the canonical value", tt.value)
		assert.Equal(t, tt.index, got, "qualitativeValueIndex(%q)", tt.value)
	}

	for _, left := range values {
		for _, right := range values {
			leftIndex, leftOK := qualitativeValueIndex(left.value)
			rightIndex, rightOK := qualitativeValueIndex(right.value)
			if !leftOK || !rightOK {
				t.Fatalf("canonical pair (%q, %q) must be recognized", left.value, right.value)
			}
			assert.Equal(
				t,
				sign(left.index-right.index),
				sign(leftIndex-rightIndex),
				"comparison order for %q and %q",
				left.value,
				right.value,
			)
		}
	}

	normalizationCases := []struct {
		input string
		index int
	}{
		{input: " (-) ", index: 0},
		{input: "（±）", index: 1},
		{input: "( + )", index: 2},
		{input: "（ + + ）", index: 3},
		{input: "\t（+++）\n", index: 4},
	}
	for _, tt := range normalizationCases {
		got, ok := qualitativeValueIndex(tt.input)
		assert.True(t, ok, "qualitativeValueIndex(%q) must normalize the supported notation", tt.input)
		assert.Equal(t, tt.index, got, "qualitativeValueIndex(%q)", tt.input)
	}

	for _, input := range []string{"", "陰性", "(++++)", "＋", "(-"} {
		_, ok := qualitativeValueIndex(input)
		assert.False(t, ok, "qualitativeValueIndex(%q) must reject non-canonical notation", input)
	}
}

func TestComputeExamResultAssessment(t *testing.T) {
	numericMin, numericMax := 1.0, 10.0
	numericNaN := math.NaN()
	numericPositiveInfinity := math.Inf(1)
	numericNegativeInfinity := math.Inf(-1)
	invertedNumericMin, invertedNumericMax := 10.0, 1.0
	qualitativeNegative := "(-)"
	qualitativePlusMinus := "(±)"
	qualitativePositive := "(+)"
	qualitativeDoublePositive := "(++)"
	qualitativeTriplePositive := "(+++)"
	unknown := "陰性"

	tests := []struct {
		name           string
		inspection     string
		refMin         *float64
		refMax         *float64
		qualitativeMin *string
		qualitativeMax *string
		wantStatus     model.ExaminationResultStatus
		wantAbnormal   bool
		wantAssessed   bool
	}{
		{
			name:         "numeric in range is assessed normal",
			inspection:   "5",
			refMin:       &numericMin,
			refMax:       &numericMax,
			wantStatus:   model.ExaminationResultStatusNormal,
			wantAssessed: true,
		},
		{
			name:         "numeric lower-bound equality is assessed normal",
			inspection:   "1",
			refMin:       &numericMin,
			refMax:       &numericMax,
			wantStatus:   model.ExaminationResultStatusNormal,
			wantAssessed: true,
		},
		{
			name:         "numeric upper-bound equality is assessed normal",
			inspection:   "10",
			refMin:       &numericMin,
			refMax:       &numericMax,
			wantStatus:   model.ExaminationResultStatusNormal,
			wantAssessed: true,
		},
		{
			name:         "numeric below range is assessed low",
			inspection:   "0.5",
			refMin:       &numericMin,
			refMax:       &numericMax,
			wantStatus:   model.ExaminationResultStatusLow,
			wantAbnormal: true,
			wantAssessed: true,
		},
		{
			name:         "numeric above range is assessed high",
			inspection:   "10.1",
			refMin:       &numericMin,
			refMax:       &numericMax,
			wantStatus:   model.ExaminationResultStatusHigh,
			wantAbnormal: true,
			wantAssessed: true,
		},
		{
			name:       "numeric bounds with nonnumeric input are unassessed",
			inspection: unknown,
			refMin:     &numericMin,
			refMax:     &numericMax,
			wantStatus: model.ExaminationResultStatusNormal,
		},
		{
			name:       "numeric bounds with empty input are unassessed",
			inspection: "",
			refMin:     &numericMin,
			refMax:     &numericMax,
			wantStatus: model.ExaminationResultStatusNormal,
		},
		{
			name:       "numeric bounds with NaN are unassessed",
			inspection: "NaN",
			refMin:     &numericMin,
			refMax:     &numericMax,
			wantStatus: model.ExaminationResultStatusNormal,
		},
		{
			name:       "numeric bounds with infinity are unassessed",
			inspection: "Inf",
			refMin:     &numericMin,
			refMax:     &numericMax,
			wantStatus: model.ExaminationResultStatusNormal,
		},
		{
			name:       "NaN numeric minimum is unassessed",
			inspection: "5",
			refMin:     &numericNaN,
			refMax:     &numericMax,
			wantStatus: model.ExaminationResultStatusNormal,
		},
		{
			name:       "positive infinite numeric maximum is unassessed",
			inspection: "5",
			refMin:     &numericMin,
			refMax:     &numericPositiveInfinity,
			wantStatus: model.ExaminationResultStatusNormal,
		},
		{
			name:       "negative infinite numeric minimum is unassessed",
			inspection: "5",
			refMin:     &numericNegativeInfinity,
			refMax:     &numericMax,
			wantStatus: model.ExaminationResultStatusNormal,
		},
		{
			name:       "inverted numeric boundaries are unassessed",
			inspection: "5",
			refMin:     &invertedNumericMin,
			refMax:     &invertedNumericMax,
			wantStatus: model.ExaminationResultStatusNormal,
		},
		{
			name:           "qualitative lower boundary is inclusive",
			inspection:     qualitativeNegative,
			qualitativeMin: &qualitativeNegative,
			qualitativeMax: &qualitativePositive,
			wantStatus:     model.ExaminationResultStatusNormal,
			wantAssessed:   true,
		},
		{
			name:           "qualitative upper boundary is inclusive",
			inspection:     qualitativePositive,
			qualitativeMin: &qualitativeNegative,
			qualitativeMax: &qualitativePositive,
			wantStatus:     model.ExaminationResultStatusNormal,
			wantAssessed:   true,
		},
		{
			name:           "qualitative above range is assessed high",
			inspection:     qualitativeDoublePositive,
			qualitativeMin: &qualitativeNegative,
			qualitativeMax: &qualitativePositive,
			wantStatus:     model.ExaminationResultStatusHigh,
			wantAbnormal:   true,
			wantAssessed:   true,
		},
		{
			name:           "qualitative below range is assessed low",
			inspection:     qualitativeNegative,
			qualitativeMin: &qualitativePlusMinus,
			qualitativeMax: &qualitativeTriplePositive,
			wantStatus:     model.ExaminationResultStatusLow,
			wantAbnormal:   true,
			wantAssessed:   true,
		},
		{
			name:           "qualitative normalized input is assessed",
			inspection:     " （ + + ） ",
			qualitativeMin: &qualitativeNegative,
			qualitativeMax: &qualitativePositive,
			wantStatus:     model.ExaminationResultStatusHigh,
			wantAbnormal:   true,
			wantAssessed:   true,
		},
		{
			name:           "qualitative minimum only permits values at or above the minimum",
			inspection:     qualitativePositive,
			qualitativeMin: &qualitativePlusMinus,
			wantStatus:     model.ExaminationResultStatusNormal,
			wantAssessed:   true,
		},
		{
			name:           "qualitative minimum only marks lower values low",
			inspection:     qualitativeNegative,
			qualitativeMin: &qualitativePlusMinus,
			wantStatus:     model.ExaminationResultStatusLow,
			wantAbnormal:   true,
			wantAssessed:   true,
		},
		{
			name:           "qualitative maximum only permits values at or below the maximum",
			inspection:     qualitativePositive,
			qualitativeMax: &qualitativeDoublePositive,
			wantStatus:     model.ExaminationResultStatusNormal,
			wantAssessed:   true,
		},
		{
			name:           "qualitative maximum only marks higher values high",
			inspection:     qualitativeTriplePositive,
			qualitativeMax: &qualitativeDoublePositive,
			wantStatus:     model.ExaminationResultStatusHigh,
			wantAbnormal:   true,
			wantAssessed:   true,
		},
		{
			name:           "qualitative bounds with unknown input are unassessed",
			inspection:     unknown,
			qualitativeMin: &qualitativeNegative,
			qualitativeMax: &qualitativePositive,
			wantStatus:     model.ExaminationResultStatusNormal,
		},
		{
			name:           "invalid qualitative boundary is unassessed",
			inspection:     qualitativePositive,
			qualitativeMin: &unknown,
			qualitativeMax: &qualitativeTriplePositive,
			wantStatus:     model.ExaminationResultStatusNormal,
		},
		{
			name:           "inverted qualitative boundaries are unassessed",
			inspection:     qualitativePositive,
			qualitativeMin: &qualitativeTriplePositive,
			qualitativeMax: &qualitativeNegative,
			wantStatus:     model.ExaminationResultStatusNormal,
		},
		{
			name:           "coexisting numeric and qualitative bounds fail closed for numeric input",
			inspection:     "5",
			refMin:         &numericMin,
			refMax:         &numericMax,
			qualitativeMin: &qualitativeNegative,
			qualitativeMax: &qualitativePositive,
			wantStatus:     model.ExaminationResultStatusNormal,
		},
		{
			name:           "coexisting numeric and qualitative bounds fail closed for qualitative input",
			inspection:     qualitativePositive,
			refMin:         &numericMin,
			refMax:         &numericMax,
			qualitativeMin: &qualitativeNegative,
			qualitativeMax: &qualitativePositive,
			wantStatus:     model.ExaminationResultStatusNormal,
		},
		{
			name:       "no bounds are unassessed",
			inspection: "5",
			wantStatus: model.ExaminationResultStatusNormal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := assessExamResult(
				tt.inspection,
				tt.refMin,
				tt.refMax,
				tt.qualitativeMin,
				tt.qualitativeMax,
			)
			assert.Equal(t, tt.wantStatus, got.status)
			assert.Equal(t, tt.wantAbnormal, got.isAbnormal)
			assert.Equal(t, tt.wantAssessed, got.isAssessed)
		})
	}
}

func sign(value int) int {
	switch {
	case value < 0:
		return -1
	case value > 0:
		return 1
	default:
		return 0
	}
}
