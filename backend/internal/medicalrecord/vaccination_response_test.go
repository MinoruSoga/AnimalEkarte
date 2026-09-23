package medicalrecord

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// UAT-R2-EXCLUSIVE-LOCK: 応答 DTO が現在の行 version を返すこと。
// FE はこの値を次の PATCH の expectedVersion として送り返す。
func TestToVaccinationResponse_IncludesVersion(t *testing.T) {
	petID := uint64(7)
	vaccination := &model.Vaccination{
		ID:        21,
		ClinicID:  3,
		PetID:     &petID,
		VaccineID: 2,
		Date:      time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		Version:   5,
	}

	resp := toVaccinationResponse(vaccination)

	assert.Equal(t, 5, resp.Version)

	body, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"version":5`)
}
