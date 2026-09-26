package owner

// EMR-174: pets.name_origin（名前の由来）/ pets.meeting_story（出逢いのストーリー）の
// 飼主登録ネスト経路（request → service input → model → PetRegistrationDraft）と
// OwnerResponse ネスト（PetInOwnerResponse）の配線テスト。

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

const (
	testPetNameOrigin   = "生まれた神社の名前から"
	testPetMeetingStory = "里親募集サイトで出会った"
)

func TestCreatePetForOwnerRequest_NameOriginMeetingStory_BindsToServiceInput(t *testing.T) {
	var req createPetForOwnerRequest
	require.NoError(t, json.Unmarshal([]byte(`{
		"name": "ポチ",
		"animal_species_id": 3,
		"name_origin": "`+testPetNameOrigin+`",
		"meeting_story": "`+testPetMeetingStory+`"
	}`), &req))

	input := req.toServiceInput()
	require.NotNil(t, input.NameOrigin)
	assert.Equal(t, testPetNameOrigin, *input.NameOrigin)
	require.NotNil(t, input.MeetingStory)
	assert.Equal(t, testPetMeetingStory, *input.MeetingStory)
}

func TestCreatePetForOwnerRequest_NameOriginMeetingStory_OmittedStaysNull(t *testing.T) {
	var req createPetForOwnerRequest
	require.NoError(t, json.Unmarshal(
		[]byte(`{"name":"ポチ","animal_species_id":3}`), &req))

	input := req.toServiceInput()
	assert.Nil(t, input.NameOrigin)
	assert.Nil(t, input.MeetingStory)
}

func TestNestedPetRegistrationChain_NameOriginMeetingStoryPropagates(t *testing.T) {
	origin := testPetNameOrigin
	story := testPetMeetingStory
	inputs := []CreatePetForOwnerInput{{
		Name:            "ポチ",
		AnimalSpeciesID: 3,
		NameOrigin:      &origin,
		MeetingStory:    &story,
	}}

	pets := buildOwnerPetModels(inputs)
	require.Len(t, pets, 1)
	require.NotNil(t, pets[0].NameOrigin)
	assert.Equal(t, testPetNameOrigin, *pets[0].NameOrigin)
	require.NotNil(t, pets[0].MeetingStory)
	assert.Equal(t, testPetMeetingStory, *pets[0].MeetingStory)

	drafts := ownerRegistrationPetDrafts(pets)
	require.Len(t, drafts, 1)
	require.NotNil(t, drafts[0].NameOrigin)
	assert.Equal(t, testPetNameOrigin, *drafts[0].NameOrigin)
	require.NotNil(t, drafts[0].MeetingStory)
	assert.Equal(t, testPetMeetingStory, *drafts[0].MeetingStory)
}

func TestPetInOwnerResponse_NameOriginMeetingStory(t *testing.T) {
	p := &model.Pet{
		ID:           42,
		NameOrigin:   ptrString(testPetNameOrigin),
		MeetingStory: ptrString(testPetMeetingStory),
	}

	resp := toPetInOwnerResponse(p)
	require.NotNil(t, resp.NameOrigin)
	assert.Equal(t, testPetNameOrigin, *resp.NameOrigin)
	require.NotNil(t, resp.MeetingStory)
	assert.Equal(t, testPetMeetingStory, *resp.MeetingStory)

	body, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"name_origin":"`+testPetNameOrigin+`"`)
	assert.Contains(t, string(body), `"meeting_story":"`+testPetMeetingStory+`"`)
}

func TestPetInOwnerResponse_NameOriginMeetingStory_OmittedWhenNull(t *testing.T) {
	resp := toPetInOwnerResponse(&model.Pet{ID: 42})
	assert.Nil(t, resp.NameOrigin)
	assert.Nil(t, resp.MeetingStory)

	body, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.NotContains(t, string(body), "name_origin", "NULL は omitempty で物理欠落させる")
	assert.NotContains(t, string(body), "meeting_story", "NULL は omitempty で物理欠落させる")
}
