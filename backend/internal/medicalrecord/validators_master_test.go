package medicalrecord

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

// TestValidateVaccineSpecies moved here with validateVaccineSpecies (BE9-2D), which relocated
// from internal/service/validators_master.go since vaccineService is its only consumer.
func TestValidateVaccineSpecies(t *testing.T) {
	assert.NoError(t, validateVaccineSpecies(""))
	assert.NoError(t, validateVaccineSpecies(string(model.VaccineSpeciesDog)))
	assert.Error(t, validateVaccineSpecies("invalid_species"))
}

// TestValidateNonNegativePrice covers the documented duplicate of internal/service's helper
// (validators_accounting.go) that this package's vaccineService uses.
func TestValidateNonNegativePrice(t *testing.T) {
	assert.NoError(t, validateNonNegativePrice(nil))
	var zero int64
	assert.NoError(t, validateNonNegativePrice(&zero))
	var neg int64 = -100
	assert.Error(t, validateNonNegativePrice(&neg))
}
