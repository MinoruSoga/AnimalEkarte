package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestValidMedicineDoseSpecies guards the #201 fail-closed species allowlist: an
// arbitrary string cast to MedicineDoseSpecies (e.g. "rabbit") must never validate,
// since mg/kg dosing tables only cover dog/cat.
func TestValidMedicineDoseSpecies(t *testing.T) {
	tests := []struct {
		name string
		in   MedicineDoseSpecies
		want bool
	}{
		{"dog is valid", MedicineDoseSpeciesDog, true},
		{"cat is valid", MedicineDoseSpeciesCat, true},
		{"unknown species is invalid", MedicineDoseSpecies("rabbit"), false},
		{"empty species is invalid", MedicineDoseSpecies(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ValidMedicineDoseSpecies(tt.in))
		})
	}
}
