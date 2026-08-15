package medicalrecord

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExamReferenceRangeValidation(t *testing.T) {
	negative, positive := "(-)", "(+)"
	min, max := 1.0, 2.0

	tests := []struct {
		name    string
		input   []ReferenceRangeInput
		wantErr bool
	}{
		{name: "accepts numeric bounds", input: []ReferenceRangeInput{{AnimalSpeciesID: 1, RefMin: &min, RefMax: &max}}},
		{name: "accepts ordered qualitative bounds", input: []ReferenceRangeInput{{AnimalSpeciesID: 1, QualitativeMin: &negative, QualitativeMax: &positive}}},
		{name: "rejects duplicate species", input: []ReferenceRangeInput{{AnimalSpeciesID: 1}, {AnimalSpeciesID: 1}}, wantErr: true},
		{name: "rejects numeric and qualitative coexistence", input: []ReferenceRangeInput{{AnimalSpeciesID: 1, RefMin: &min, QualitativeMax: &positive}}, wantErr: true},
		{name: "rejects non finite numeric bound", input: []ReferenceRangeInput{{AnimalSpeciesID: 1, RefMin: examRangeFloat64Ptr(math.Inf(1))}}, wantErr: true},
		{name: "rejects reversed numeric range", input: []ReferenceRangeInput{{AnimalSpeciesID: 1, RefMin: &max, RefMax: &min}}, wantErr: true},
		{name: "rejects unsupported qualitative value", input: []ReferenceRangeInput{{AnimalSpeciesID: 1, QualitativeMin: stringPtr("unknown")}}, wantErr: true},
		{name: "rejects reversed qualitative range", input: []ReferenceRangeInput{{AnimalSpeciesID: 1, QualitativeMin: &positive, QualitativeMax: &negative}}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateReferenceRangeInputs(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func examRangeFloat64Ptr(value float64) *float64 { return &value }
func stringPtr(value string) *string             { return &value }
