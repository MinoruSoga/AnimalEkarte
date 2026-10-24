package pet

import (
	"bytes"
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-491 U3: 基本情報の固定 422（F1/F4/F5/F6）と、hydrate 済み pet の応答変換のテスト。
// DB には触れない（nil の db で書き込みの手前で止まることも検証する）。

// emr491PetService は Update だけを差し替えられる Service。他の操作が呼ばれたらテストを失敗させる。
type emr491PetService struct {
	t        *testing.T
	updateFn func(ctx context.Context, clinicID, id uint64, input *UpdatePetInput) (*model.Pet, error)
}

func (s *emr491PetService) unexpected(name string) error {
	s.t.Helper()
	s.t.Errorf("unexpected service call: %s", name)
	return apperrors.WrapInternalServerError("unexpected call")
}

func (s *emr491PetService) List(context.Context, []uint64, PetListFilters, int, int) ([]model.Pet, int64, error) {
	return nil, 0, s.unexpected("List")
}

func (s *emr491PetService) ListOwnerReportPets(context.Context, []uint64, uint64) ([]model.Pet, error) {
	return nil, s.unexpected("ListOwnerReportPets")
}

func (s *emr491PetService) GetByID(context.Context, uint64, uint64) (*model.Pet, error) {
	return nil, s.unexpected("GetByID")
}

func (s *emr491PetService) GetByIDForClinics(context.Context, []uint64, uint64) (*model.Pet, error) {
	return nil, s.unexpected("GetByIDForClinics")
}

func (s *emr491PetService) Create(context.Context, uint64, *CreatePetInput) (*model.Pet, error) {
	return nil, s.unexpected("Create")
}

func (s *emr491PetService) Update(ctx context.Context, clinicID, id uint64, input *UpdatePetInput) (*model.Pet, error) {
	if s.updateFn == nil {
		return nil, s.unexpected("Update")
	}
	return s.updateFn(ctx, clinicID, id, input)
}

func (s *emr491PetService) Delete(context.Context, uint64, uint64) error {
	return s.unexpected("Delete")
}

func (s *emr491PetService) GetFirstVisitDate(context.Context, uint64, uint64) (*time.Time, error) {
	return nil, s.unexpected("GetFirstVisitDate")
}

// emr491Pet は hydrate 済み（基本情報あり・nullable な列は全て NULL）の pet。
func emr491Pet() *model.Pet {
	return &model.Pet{
		ID:            7,
		ClinicID:      1,
		OwnerID:       2,
		PetNumber:     "2-1",
		Version:       3,
		Basic:         &model.CommonPetBasicRevision{Name: "ポチ", AnimalSpeciesID: 5},
		BasicSubject:  &model.CommonSubjectRef{Version: 4, CurrentRevisionID: 11},
		AnimalSpecies: &model.AnimalSpecies{ID: 5, Name: "犬"},
	}
}

func emr491Keys(t *testing.T, body string) map[string]json.RawMessage {
	t.Helper()
	var keys map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(body), &keys))
	return keys
}

func emr491PatchPet(t *testing.T, svc Service, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	h := NewHandler(svc, nil, nil, allowAllPermission)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader([]byte(body)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	setClinicID(c)
	setStaffID(c, 1)
	h.UpdatePet(c)
	return w
}

func emr491Marshal(t *testing.T, value any) map[string]any {
	t.Helper()
	raw, err := json.Marshal(value)
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(raw, &decoded))
	return decoded
}

func TestRejectPresentPetBasicKeys(t *testing.T) {
	t.Run("rejects every basic key whether it carries a value or an explicit null", func(t *testing.T) {
		for _, name := range petBasicRequestKeys {
			for _, value := range []string{`"x"`, `null`} {
				err := rejectPresentPetBasicKeys(emr491Keys(t, `{"`+name+`":`+value+`}`))
				require.Error(t, err, name+"="+value)
				assert.True(t, apperrors.HasCode(err, apperrors.CodePetBasicRevisionPending), name)
			}
		}
	})

	t.Run("matches key names case-insensitively like the JSON decoder", func(t *testing.T) {
		err := rejectPresentPetBasicKeys(emr491Keys(t, `{"Name_Kana":"x"}`))
		require.Error(t, err)
		assert.True(t, apperrors.HasCode(err, apperrors.CodePetBasicRevisionPending))
	})

	t.Run("lets local-only bodies through, including explicit null on a local key", func(t *testing.T) {
		body := `{"owner_id":2,"insurance_id":null,"pet_number":"2-1"}`
		assert.NoError(t, rejectPresentPetBasicKeys(emr491Keys(t, body)))
	})

	t.Run("lets an empty object through", func(t *testing.T) {
		assert.NoError(t, rejectPresentPetBasicKeys(emr491Keys(t, `{}`)))
	})
}

// updatePetRequest の JSON キーのうち、ローカル列（主飼主・保険・ペット番号）以外は全て基本情報の key 一覧に
// 入っていること。リクエスト項目を足したときに F4 の判定漏れが起きないようにするための照合。
func TestPetBasicRequestKeysCoverEveryNonLocalUpdateField(t *testing.T) {
	local := map[string]bool{"owner_id": true, "insurance_id": true, "pet_number": true}
	requestType := reflect.TypeOf(updatePetRequest{})
	for i := 0; i < requestType.NumField(); i++ {
		name := strings.Split(requestType.Field(i).Tag.Get("json"), ",")[0]
		if name == "" || local[name] {
			continue
		}
		assert.True(t, containsFold(petBasicRequestKeys, name), "update field %q is not covered by petBasicRequestKeys", name)
	}
}

func TestUpdatePet_BasicKeysReturnFixed422BeforeAnyServiceCall(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "basic key with a value", body: `{"name":"ポチ"}`},
		{name: "basic key with explicit null", body: `{"weight":null}`},
		{name: "explicit null on a tri-state key", body: `{"danger_reason":null}`},
		{name: "basic key mixed with local fields", body: `{"owner_id":2,"food":"x"}`},
		{name: "key case differs from the wire name", body: `{"Animal_Species_Id":1}`},
		{name: "retired last_visit", body: `{"last_visit":"2026-01-01"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &emr491PetService{t: t}

			w := emr491PatchPet(t, svc, tt.body)

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code, w.Body.String())
			assert.Contains(t, w.Body.String(), `"code":"pet_basic_revision_pending"`)
		})
	}
}

func TestUpdatePet_LocalOnlyBodyKeepsTheUpdatePath(t *testing.T) {
	var got *UpdatePetInput
	svc := &emr491PetService{
		t: t,
		updateFn: func(_ context.Context, _, _ uint64, input *UpdatePetInput) (*model.Pet, error) {
			got = input
			return emr491Pet(), nil
		},
	}

	w := emr491PatchPet(t, svc, `{"owner_id":2,"insurance_id":null}`)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.NotNil(t, got)
	require.NotNil(t, got.OwnerID)
	assert.False(t, updateInputTouchesBasic(got))
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.EqualValues(t, 4, body["pet_version"])
	assert.EqualValues(t, 11, body["pet_basic_revision_id"])
}

func TestUpdateInputTouchesBasic(t *testing.T) {
	text := "x"
	number := 1.5
	moment := time.Now()
	id := uint64(1)
	textPtr := &text
	idPtr := &id
	full := &UpdatePetInput{
		OwnerID: &id, AnimalSpeciesID: &id, PetNumber: &text, Name: &text, PetNameKana: &text, Gender: &text,
		BirthDate: &moment, Breed: &text, Color: &text, BloodType: &text, MicrochipNumber: &text,
		Weight: &number, NeuteredDate: &moment, NeuteredStatus: &text, AcquisitionType: &text,
		DangerLevel: &text, DangerReason: &textPtr, Food: &text, Environment: &text, Phone: &text,
		LastVisit: &moment, InsuranceID: &idPtr, Remarks: &text, NameOrigin: &textPtr, MeetingStory: &textPtr,
	}
	local := map[string]bool{colPetOwnerID: true, colPetInsuranceID: true, "pet_number": true}

	assert.False(t, updateInputTouchesBasic(nil))
	assert.False(t, updateInputTouchesBasic(&UpdatePetInput{OwnerID: &id, PetNumber: &text, InsuranceID: &idPtr}))
	assert.True(t, updateInputTouchesBasic(&UpdatePetInput{Name: &text}))
	var none *string
	assert.True(t, updateInputTouchesBasic(&UpdatePetInput{DangerReason: &none}), "explicit null is a specified basic key")

	fields := buildPetUpdate(full)
	for key, value := range fields {
		if local[key] {
			continue
		}
		err := rejectRetiredPetFields(map[string]any{key: value})
		require.Error(t, err, key)
		assert.True(t, apperrors.HasCode(err, apperrors.CodePetBasicRevisionPending), key)
	}
}

func TestRejectRetiredPetFields(t *testing.T) {
	assert.NoError(t, rejectRetiredPetFields(map[string]any{colPetOwnerID: uint64(1), colPetInsuranceID: nil, "pet_number": "1-1"}))
	assert.NoError(t, rejectRetiredPetFields(nil))

	t.Run("lifecycle keys win over basic keys regardless of map order", func(t *testing.T) {
		for i := 0; i < 50; i++ {
			err := rejectRetiredPetFields(map[string]any{"name": "x", "status": "deceased", "weight": 1})
			require.Error(t, err)
			assert.True(t, apperrors.HasCode(err, apperrors.CodePetLifecycleWritePending))
		}
	})
}

func TestPetService_Update_BasicFieldsAreRejectedBeforeAnyRead(t *testing.T) {
	name := "x"
	var none *string
	// repo が nil のため、読み取りに進んだ時点で panic する。
	svc := NewService(nil, nil, nil, nil, nil)

	for _, input := range []*UpdatePetInput{{Name: &name}, {DangerReason: &none}} {
		_, err := svc.Update(context.Background(), 1, 1, input)

		require.Error(t, err)
		assert.True(t, apperrors.HasCode(err, apperrors.CodePetBasicRevisionPending))
	}
}

func TestPetService_Create_ReturnsRegistrationPendingBeforeAnyRead(t *testing.T) {
	svc := NewService(nil, nil, nil, nil, nil)

	_, err := svc.Create(context.Background(), 1, &CreatePetInput{})

	require.Error(t, err)
	assert.True(t, apperrors.HasCode(err, apperrors.CodePetRegistrationPending))
}

func TestPetRepository_FixedRejectionsHappenBeforeAnyDatabaseAccess(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(nil)
	name := "x"

	t.Run("RecordDeath is F5", func(t *testing.T) {
		err := repo.RecordDeath(ctx, 1, 1, time.Now(), "reason")
		require.Error(t, err)
		assert.True(t, apperrors.HasCode(err, apperrors.CodePetLifecycleWritePending))
	})

	t.Run("ClearDeath is F5", func(t *testing.T) {
		err := repo.ClearDeath(ctx, 1, 1)
		require.Error(t, err)
		assert.True(t, apperrors.HasCode(err, apperrors.CodePetLifecycleWritePending))
	})

	t.Run("legacy lifecycle column map is F5", func(t *testing.T) {
		err := (&repository{}).update(ctx, 1, 1, map[string]any{
			"status": model.PetStatusDeceased, "deceased_at": time.Now(), "deceased_reason": "x",
		})
		require.Error(t, err)
		assert.True(t, apperrors.HasCode(err, apperrors.CodePetLifecycleWritePending))
	})

	t.Run("Update with a basic field is F4", func(t *testing.T) {
		err := repo.Update(ctx, 1, 1, UpdatePetInput{Name: &name})
		require.Error(t, err)
		assert.True(t, apperrors.HasCode(err, apperrors.CodePetBasicRevisionPending))
	})

	t.Run("UpdateAndFind with a basic field or a danger change is F4", func(t *testing.T) {
		_, err := repo.UpdateAndFind(ctx, 1, 1, PetUpdate{fields: map[string]any{"name": name}})
		require.Error(t, err)
		assert.True(t, apperrors.HasCode(err, apperrors.CodePetBasicRevisionPending))

		level := model.DangerLevelHigh
		_, err = repo.UpdateAndFind(ctx, 1, 1, PetUpdate{fields: map[string]any{}, dangerLevel: &level})
		require.Error(t, err)
		assert.True(t, apperrors.HasCode(err, apperrors.CodePetBasicRevisionPending))
	})

	t.Run("Create is F6", func(t *testing.T) {
		err := repo.Create(ctx, &model.Pet{ClinicID: 1, OwnerID: 1})
		require.Error(t, err)
		assert.True(t, apperrors.HasCode(err, apperrors.CodePetRegistrationPending))
	})

	t.Run("writer paths are F6", func(t *testing.T) {
		_, err := NewWriter(nil).Create(ctx, CreateIntent{})
		require.Error(t, err)
		assert.True(t, apperrors.HasCode(err, apperrors.CodePetRegistrationPending))

		_, err = NewOwnerRegistrationWriter().CreateForOwnerRegistration(ctx, OwnerRegistrationIntent{})
		require.Error(t, err)
		assert.True(t, apperrors.HasCode(err, apperrors.CodePetRegistrationPending))
	})
}

func TestPetConverters_UnhydratedPetIsUnresolvedNotBlank(t *testing.T) {
	unresolved := func(t *testing.T, err error) {
		t.Helper()
		require.Error(t, err)
		assert.True(t, apperrors.HasCode(err, apperrors.CodeCommonBasicUnresolved))
	}
	noSubject := emr491Pet()
	noSubject.BasicSubject = nil

	t.Run("detail", func(t *testing.T) {
		_, err := toResponse(&model.Pet{ID: 1, ClinicID: 1})
		unresolved(t, err)
		_, err = toResponse(noSubject)
		unresolved(t, err)
		_, err = toResponse(nil)
		unresolved(t, err)
	})

	t.Run("list row and list", func(t *testing.T) {
		_, err := toPetListResponse(&model.Pet{ID: 1, ClinicID: 1})
		unresolved(t, err)
		_, err = toPetListResponse(noSubject)
		unresolved(t, err)
		_, err = toPetListResponses([]model.Pet{*emr491Pet(), {ID: 2, ClinicID: 1}})
		unresolved(t, err)
	})

	t.Run("owner report", func(t *testing.T) {
		_, err := toOwnerReportResponse(&model.Pet{ID: 1, ClinicID: 1})
		unresolved(t, err)
		_, err = toOwnerReportResponses([]model.Pet{*emr491Pet(), {ID: 2, ClinicID: 1}})
		unresolved(t, err)
	})

	t.Run("nested owner", func(t *testing.T) {
		owner, err := toPetOwnerNested(nil)
		require.NoError(t, err)
		assert.Nil(t, owner, "a pet without a loaded owner is not an unresolved owner")

		_, err = toPetOwnerNested(&model.Owner{ID: 1, ClinicID: 1})
		unresolved(t, err)

		pet := emr491Pet()
		pet.Owner = &model.Owner{ID: 2, ClinicID: 1}
		_, err = toResponse(pet)
		unresolved(t, err)
	})

	t.Run("secondary owners and shared pets", func(t *testing.T) {
		_, err := mapPetOwnersResponse(
			[]petOwnerResponseLink{{OwnerID: 1}},
			[]*model.Owner{{ID: 1, ClinicID: 1}},
		)
		unresolved(t, err)

		_, err = newSharedPet(&model.Pet{ID: 1, ClinicID: 1}, "family")
		unresolved(t, err)
		withoutSpecies := emr491Pet()
		withoutSpecies.AnimalSpecies = nil
		_, err = newSharedPet(withoutSpecies, "family")
		unresolved(t, err)
	})
}

func TestPetConverters_NullableColumnsStayNullAndVersionsAreExposed(t *testing.T) {
	pet := emr491Pet()
	pet.Basic.RegistrationWeight = &model.ExactNumeric{Numeric: pgtype.Numeric{Int: big.NewInt(125), Exp: -1, Valid: true}}

	detail, err := toResponse(pet)
	require.NoError(t, err)
	list, err := toPetListResponse(pet)
	require.NoError(t, err)
	report, err := toOwnerReportResponse(pet)
	require.NoError(t, err)

	commonKeys := []string{"breed", "color", "gender", "status", "birth_date", "food", "environment", "remarks", "last_visit", "deceased_at"}
	for name, body := range map[string]map[string]any{
		"detail": emr491Marshal(t, detail),
		"list":   emr491Marshal(t, list),
		"report": emr491Marshal(t, report),
	} {
		for _, key := range commonKeys {
			value, present := body[key]
			assert.True(t, present, name+"."+key+" must be present")
			assert.Nil(t, value, name+"."+key+" must be null")
		}
		assert.Equal(t, 12.5, body["weight"], name+".weight")
		assert.Equal(t, "ポチ", body["name"], name)
	}

	for name, body := range map[string]map[string]any{
		"detail": emr491Marshal(t, detail),
		"list":   emr491Marshal(t, list),
	} {
		for _, key := range []string{"phone", "name_origin", "meeting_story"} {
			value, present := body[key]
			assert.True(t, present, name+"."+key+" must be present")
			assert.Nil(t, value, name+"."+key+" must be null")
		}
		assert.EqualValues(t, 4, body["pet_version"], name)
		assert.EqualValues(t, 11, body["pet_basic_revision_id"], name)
		assert.EqualValues(t, 5, body["animal_species_id"], name)
	}
}

func TestRespondPetError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	respond := func(err error) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		respondPetError(c, err)
		return w
	}

	for code, err := range map[string]error{
		apperrors.CodeCommonBasicUnresolved:    apperrors.Wrap(apperrors.WrapCommonBasicUnresolved("pet 1/1"), "failed to get pet"),
		apperrors.CodePetBasicRevisionPending:  errPetBasicRevisionPending(),
		apperrors.CodePetLifecycleWritePending: errPetLifecycleWritePending(),
		apperrors.CodePetRegistrationPending:   errPetRegistrationPending(),
	} {
		w := respond(err)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code, code)
		assert.Contains(t, w.Body.String(), `"code":"`+code+`"`)
	}

	w := respond(apperrors.WrapNotFound("pet", "1"))
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NotContains(t, w.Body.String(), `"code"`)
}
