package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

func TestValidatePetForOwnerInput(t *testing.T) {
	tests := []struct {
		name    string
		input   CreatePetForOwnerInput
		wantErr bool
	}{
		{
			name:    "all fields empty (zero value): valid",
			input:   CreatePetForOwnerInput{Name: "ポチ", AnimalSpeciesID: 1},
			wantErr: false,
		},
		{
			name: "all fields valid",
			input: CreatePetForOwnerInput{
				Name:            "ポチ",
				AnimalSpeciesID: 1,
				Gender:          "male",
				Status:          "alive",
				AcquisitionType: "purchased",
				DangerLevel:     "low",
			},
			wantErr: false,
		},
		{
			name: "invalid gender",
			input: CreatePetForOwnerInput{
				Name:            "ポチ",
				AnimalSpeciesID: 1,
				Gender:          "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			input: CreatePetForOwnerInput{
				Name:            "ポチ",
				AnimalSpeciesID: 1,
				Status:          "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid acquisition type",
			input: CreatePetForOwnerInput{
				Name:            "ポチ",
				AnimalSpeciesID: 1,
				AcquisitionType: "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid danger level",
			input: CreatePetForOwnerInput{
				Name:            "ポチ",
				AnimalSpeciesID: 1,
				DangerLevel:     "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePetForOwnerInput(&tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, apperrors.IsInvalidInput(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
