package owner

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestToPetInOwnerResponse_NormalizesDatesAndKeepsClinicalReasonPrivate(t *testing.T) {
	originalLocation := time.Local
	time.Local = config.JST
	t.Cleanup(func() {
		time.Local = originalLocation
	})

	birthDate := time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC)
	neuteredDate := time.Date(2021, 2, 16, 0, 0, 0, 0, time.UTC)
	lastVisit := time.Date(2022, 3, 17, 0, 0, 0, 0, time.UTC)
	deceasedAt := time.Date(2026, 7, 10, 3, 0, 0, 0, time.UTC)
	deceasedReason := "clinical note that must not be exposed"

	response := toPetInOwnerResponse(&model.Pet{
		ID:             7,
		OwnerID:        42,
		Name:           "ポチ",
		Status:         model.PetStatusDeceased,
		BirthDate:      &birthDate,
		NeuteredDate:   &neuteredDate,
		LastVisit:      &lastVisit,
		DeceasedAt:     &deceasedAt,
		DeceasedReason: &deceasedReason,
	})

	body, err := json.Marshal(response)
	require.NoError(t, err)
	jsonBody := string(body)

	assert.Contains(t, jsonBody, `"birth_date":"2020-01-15T09:00:00+09:00"`)
	assert.Contains(t, jsonBody, `"neutered_date":"2021-02-16T09:00:00+09:00"`)
	assert.Contains(t, jsonBody, `"last_visit":"2022-03-17T09:00:00+09:00"`)
	assert.Contains(t, jsonBody, `"deceased_at":"2026-07-10T12:00:00+09:00"`)
	assert.NotContains(t, jsonBody, "deceased_reason")
	assert.NotContains(t, jsonBody, deceasedReason)
}

func TestToPetInOwnerResponse_OmitsDeceasedAtForLivingPet(t *testing.T) {
	response := toPetInOwnerResponse(&model.Pet{
		ID:     7,
		Name:   "ポチ",
		Status: model.PetStatusAlive,
	})

	assert.Nil(t, response.DeceasedAt)
	body, err := json.Marshal(response)
	require.NoError(t, err)
	assert.NotContains(t, string(body), "deceased_at")
}
