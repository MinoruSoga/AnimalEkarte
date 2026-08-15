package owner

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestDangerReason_NestedOwnerCreatePropagation(t *testing.T) {
	reason := "  診察台で噛む  "
	request := createPetForOwnerRequest{
		Name:            "ポチ",
		AnimalSpeciesID: 3,
		DangerLevel:     string(model.DangerLevelHigh),
		DangerReason:    &reason,
	}

	input := request.toServiceInput()
	pets := buildOwnerPetModels([]CreatePetForOwnerInput{input})
	drafts := ownerRegistrationPetDrafts(pets)

	require.Len(t, drafts, 1)
	require.NotNil(t, drafts[0].DangerReason)
	assert.Equal(t, reason, *drafts[0].DangerReason)
	assert.Equal(t, model.DangerLevelHigh, drafts[0].DangerLevel)
}

func TestDangerReason_OwnerNestedPetResponseOmitsField(t *testing.T) {
	reason := "staff only"
	response := toOwnerResponse(&model.Owner{
		ID: 1,
		Pets: []model.Pet{{
			ID:           2,
			DangerLevel:  model.DangerLevelHigh,
			DangerReason: &reason,
		}},
	})

	body, err := json.Marshal(response)
	require.NoError(t, err)
	var payload struct {
		Pets []map[string]any `json:"pets"`
	}
	require.NoError(t, json.Unmarshal(body, &payload))
	require.Len(t, payload.Pets, 1)
	_, exposed := payload.Pets[0]["danger_reason"]
	assert.False(t, exposed, "owner nested pets must physically omit danger_reason: %s", body)
}
