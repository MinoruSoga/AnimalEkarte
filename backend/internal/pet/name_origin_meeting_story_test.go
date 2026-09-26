package pet

// EMR-174: pets.name_origin（名前の由来）/ pets.meeting_story（出逢いのストーリー）の
// 作成・更新・レスポンス・飼主登録ネスト経路の配線テスト。
// 両フィールドは nullable text（未記録は NULL で保持し、PATCH の tri-state で
// 変更なし / NULL クリア / 値更新 を区別する）。

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	ownerdomain "github.com/animal-ekarte/backend/internal/owner"
)

const (
	testNameOrigin   = "生まれた神社の名前から"
	testMeetingStory = "里親募集サイトで出会った"
)

func TestCreatePetRequest_NameOriginMeetingStory_BindsToServiceInputAndModel(t *testing.T) {
	var req createPetRequest
	require.NoError(t, json.Unmarshal([]byte(`{
		"owner_id": 7,
		"animal_species_id": 3,
		"name": "ポチ",
		"name_origin": "`+testNameOrigin+`",
		"meeting_story": "`+testMeetingStory+`"
	}`), &req))

	input := req.toServiceInput()
	require.NotNil(t, input.NameOrigin)
	assert.Equal(t, testNameOrigin, *input.NameOrigin)
	require.NotNil(t, input.MeetingStory)
	assert.Equal(t, testMeetingStory, *input.MeetingStory)

	pet := buildPetModel(11, "7-1", input)
	require.NotNil(t, pet.NameOrigin)
	assert.Equal(t, testNameOrigin, *pet.NameOrigin)
	require.NotNil(t, pet.MeetingStory)
	assert.Equal(t, testMeetingStory, *pet.MeetingStory)
}

func TestCreatePetRequest_NameOriginMeetingStory_OmittedStaysNull(t *testing.T) {
	var req createPetRequest
	require.NoError(t, json.Unmarshal(
		[]byte(`{"owner_id":7,"animal_species_id":3,"name":"ポチ"}`), &req))

	input := req.toServiceInput()
	assert.Nil(t, input.NameOrigin)
	assert.Nil(t, input.MeetingStory)

	pet := buildPetModel(11, "7-1", input)
	assert.Nil(t, pet.NameOrigin, "未入力は NULL のまま（空文字を記録済みと誤表示しない）")
	assert.Nil(t, pet.MeetingStory, "未入力は NULL のまま（空文字を記録済みと誤表示しない）")
}

func TestUpdatePetRequest_NameOriginMeetingStory_TriState(t *testing.T) {
	cases := []struct {
		name      string
		body      string
		field     string
		wantSet   bool
		wantClear bool
		wantValue string
	}{
		{name: "name_origin 省略=変更なし", body: `{}`, field: "name_origin", wantSet: false},
		{name: "name_origin null=NULLクリア", body: `{"name_origin":null}`, field: "name_origin", wantSet: true, wantClear: true},
		{name: "name_origin 値=更新", body: `{"name_origin":"` + testNameOrigin + `"}`, field: "name_origin", wantSet: true, wantValue: testNameOrigin},
		{name: "meeting_story 省略=変更なし", body: `{}`, field: "meeting_story", wantSet: false},
		{name: "meeting_story null=NULLクリア", body: `{"meeting_story":null}`, field: "meeting_story", wantSet: true, wantClear: true},
		{name: "meeting_story 値=更新", body: `{"meeting_story":"` + testMeetingStory + `"}`, field: "meeting_story", wantSet: true, wantValue: testMeetingStory},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var req updatePetRequest
			require.NoError(t, json.Unmarshal([]byte(tc.body), &req))
			input := req.toServiceInput()

			var field **string
			switch tc.field {
			case "name_origin":
				field = input.NameOrigin
			case "meeting_story":
				field = input.MeetingStory
			default:
				t.Fatalf("unknown field %q", tc.field)
			}

			if !tc.wantSet {
				assert.Nil(t, field, "省略時は nil（変更なし）")
				return
			}
			require.NotNil(t, field, "JSON で送信されたフィールドは非 nil")
			if tc.wantClear {
				assert.Nil(t, *field, "JSON null は NULL クリア（&nil）")
				return
			}
			require.NotNil(t, *field)
			assert.Equal(t, tc.wantValue, **field)
		})
	}
}

func TestBuildPetUpdate_NameOriginMeetingStory(t *testing.T) {
	t.Run("省略フィールドは map に含めない", func(t *testing.T) {
		fields := buildPetUpdate(&UpdatePetInput{})
		_, hasOrigin := fields[colPetNameOrigin]
		_, hasStory := fields[colPetMeetingStory]
		assert.False(t, hasOrigin)
		assert.False(t, hasStory)
	})

	t.Run("値セットは文字列として map に入る", func(t *testing.T) {
		input := &UpdatePetInput{
			NameOrigin:   stringPointerPointer(ptrString(testNameOrigin)),
			MeetingStory: stringPointerPointer(ptrString(testMeetingStory)),
		}
		fields := buildPetUpdate(input)
		assert.Equal(t, testNameOrigin, *fields[colPetNameOrigin].(*string))
		assert.Equal(t, testMeetingStory, *fields[colPetMeetingStory].(*string))
	})

	t.Run("&nil は NULL クリアとして map に入る", func(t *testing.T) {
		input := &UpdatePetInput{
			NameOrigin:   stringPointerPointer(nil),
			MeetingStory: stringPointerPointer(nil),
		}
		fields := buildPetUpdate(input)
		assert.Contains(t, fields, colPetNameOrigin)
		assert.Contains(t, fields, colPetMeetingStory)
		assert.Equal(t, (*string)(nil), fields[colPetNameOrigin])
		assert.Equal(t, (*string)(nil), fields[colPetMeetingStory])
	})
}

func TestPetResponses_NameOriginMeetingStory(t *testing.T) {
	p := &model.Pet{
		ID:           42,
		NameOrigin:   ptrString(testNameOrigin),
		MeetingStory: ptrString(testMeetingStory),
	}

	detail := toResponse(p)
	require.NotNil(t, detail.NameOrigin)
	assert.Equal(t, testNameOrigin, *detail.NameOrigin)
	require.NotNil(t, detail.MeetingStory)
	assert.Equal(t, testMeetingStory, *detail.MeetingStory)

	list := toPetListResponse(p)
	require.NotNil(t, list.NameOrigin)
	assert.Equal(t, testNameOrigin, *list.NameOrigin)
	require.NotNil(t, list.MeetingStory)
	assert.Equal(t, testMeetingStory, *list.MeetingStory)

	body, err := json.Marshal(detail)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"name_origin":"`+testNameOrigin+`"`)
	assert.Contains(t, string(body), `"meeting_story":"`+testMeetingStory+`"`)
}

func TestPetResponses_NameOriginMeetingStory_OmittedWhenNull(t *testing.T) {
	p := &model.Pet{ID: 42}

	detail := toResponse(p)
	assert.Nil(t, detail.NameOrigin)
	assert.Nil(t, detail.MeetingStory)

	body, err := json.Marshal(detail)
	require.NoError(t, err)
	assert.NotContains(t, string(body), "name_origin", "NULL は omitempty で物理欠落させる")
	assert.NotContains(t, string(body), "meeting_story", "NULL は omitempty で物理欠落させる")
}

func TestCreatePetDraft_NameOriginMeetingStoryRoundTrip(t *testing.T) {
	m := model.Pet{
		NameOrigin:   ptrString(testNameOrigin),
		MeetingStory: ptrString(testMeetingStory),
	}
	draft := CreatePetDraftFromModel(m)
	require.NotNil(t, draft.NameOrigin)
	assert.Equal(t, testNameOrigin, *draft.NameOrigin)
	require.NotNil(t, draft.MeetingStory)
	assert.Equal(t, testMeetingStory, *draft.MeetingStory)

	created := draft.model(1, 2, "2-1")
	require.NotNil(t, created.NameOrigin)
	assert.Equal(t, testNameOrigin, *created.NameOrigin)
	require.NotNil(t, created.MeetingStory)
	assert.Equal(t, testMeetingStory, *created.MeetingStory)
}

// EMR-174 L2: rune 上限と空白正規化の境界値（create/update の service 層検証）。
func TestValidatePetInput_NameOriginMeetingStory_RuneLimitsAndWhitespace(t *testing.T) {
	atOrigin := strings.Repeat("あ", nameOriginMaxRunes)      // 500
	overOrigin := strings.Repeat("あ", nameOriginMaxRunes+1)  // 501
	overStory := strings.Repeat("い", meetingStoryMaxRunes+1) // 2001
	whitespace := " 　\n\t "

	t.Run("create: 上限以下は受理し空白のみはNULL正規化", func(t *testing.T) {
		input := &CreatePetInput{
			Name:         "ポチ",
			NameOrigin:   ptrString(atOrigin),
			MeetingStory: ptrString(whitespace),
		}
		require.NoError(t, validateCreatePetInput(input))
		require.NotNil(t, input.NameOrigin)
		assert.Equal(t, atOrigin, *input.NameOrigin)
		assert.Nil(t, input.MeetingStory, "空白のみは NULL（未記録）へ正規化")
	})

	t.Run("create: 上限超過は invalid input", func(t *testing.T) {
		assert.Error(t, validateCreatePetInput(&CreatePetInput{
			Name: "ポチ", NameOrigin: ptrString(overOrigin),
		}))
		assert.Error(t, validateCreatePetInput(&CreatePetInput{
			Name: "ポチ", MeetingStory: ptrString(overStory),
		}))
	})

	t.Run("update: 上限以下は受理し空白のみはNULLクリア", func(t *testing.T) {
		input := &UpdatePetInput{
			NameOrigin:   stringPointerPointer(ptrString(atOrigin)),
			MeetingStory: stringPointerPointer(ptrString(whitespace)),
		}
		require.NoError(t, validateUpdatePetInput(input))
		require.NotNil(t, input.NameOrigin)
		require.NotNil(t, *input.NameOrigin)
		assert.Equal(t, atOrigin, **input.NameOrigin)
		require.NotNil(t, input.MeetingStory)
		assert.Nil(t, *input.MeetingStory, "空白のみは &nil = NULL クリア")
	})

	t.Run("update: 上限超過は invalid input", func(t *testing.T) {
		assert.Error(t, validateUpdatePetInput(&UpdatePetInput{
			NameOrigin: stringPointerPointer(ptrString(overOrigin)),
		}))
		assert.Error(t, validateUpdatePetInput(&UpdatePetInput{
			MeetingStory: stringPointerPointer(ptrString(overStory)),
		}))
	})
}

// EMR-174 L1: createPetsInTransaction の共有書込境界でも叙述フィールドを正規化する。
func TestNormalizeCreatePetDrafts_NameOriginMeetingStory(t *testing.T) {
	whitespace := " 　 "
	overOrigin := strings.Repeat("あ", nameOriginMaxRunes+1)

	normalized, err := normalizeCreatePetDrafts([]CreatePetDraft{{
		NameOrigin:   &whitespace,
		MeetingStory: ptrString(testMeetingStory),
	}})
	require.NoError(t, err)
	assert.Nil(t, normalized[0].NameOrigin, "空白のみは境界で NULL へ正規化")
	require.NotNil(t, normalized[0].MeetingStory)
	assert.Equal(t, testMeetingStory, *normalized[0].MeetingStory)

	_, err = normalizeCreatePetDrafts([]CreatePetDraft{{NameOrigin: &overOrigin}})
	assert.Error(t, err, "上限超過は境界でも拒否")
}

type captureOwnerRegistrationWriter struct {
	captured OwnerRegistrationIntent
}

func (w *captureOwnerRegistrationWriter) CreateForOwnerRegistration(
	_ context.Context,
	intent OwnerRegistrationIntent,
) ([]model.Pet, error) {
	w.captured = intent
	return []model.Pet{}, nil
}

func TestOwnerRegistrationAdapter_NameOriginMeetingStoryPropagates(t *testing.T) {
	writer := &captureOwnerRegistrationWriter{}
	adapter := NewOwnerRegistrationAdapter(writer)

	origin := testNameOrigin
	story := testMeetingStory
	_, err := adapter.CreateForOwnerRegistration(context.Background(), ownerdomain.PetRegistrationIntent{
		ClinicID: 1,
		OwnerID:  2,
		Pets: []ownerdomain.PetRegistrationDraft{{
			AnimalSpeciesID: 3,
			Name:            "ポチ",
			NameOrigin:      &origin,
			MeetingStory:    &story,
		}},
	})
	require.NoError(t, err)
	require.Len(t, writer.captured.Pets, 1)
	require.NotNil(t, writer.captured.Pets[0].NameOrigin)
	assert.Equal(t, testNameOrigin, *writer.captured.Pets[0].NameOrigin)
	require.NotNil(t, writer.captured.Pets[0].MeetingStory)
	assert.Equal(t, testMeetingStory, *writer.captured.Pets[0].MeetingStory)
}
