package owner

import (
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func emr491Pet() model.Pet {
	pet := model.Pet{ID: 7, ClinicID: 1, OwnerID: 5, PetNumber: "P-0007"}
	pet.Basic = &model.CommonPetBasicRevision{AnimalSpeciesID: 3, Name: "ポチ", NeuteredStatus: model.PetNeuteredStatusUnknown}
	pet.BasicSubject = &model.CommonSubjectRef{Version: 4, CurrentRevisionID: 40}
	pet.AnimalSpecies = &model.AnimalSpecies{ID: 3, Name: "犬"}
	return pet
}

func emr491Owner() *model.Owner {
	owner := &model.Owner{ID: 5, ClinicID: 1, Remarks: "memo"}
	owner.Basic = &model.CommonOwnerBasicRevision{Name: "山田太郎"}
	owner.BasicSubject = &model.CommonSubjectRef{Version: 3, CurrentRevisionID: 30}
	owner.Pets = []model.Pet{emr491Pet()}
	return owner
}

func emr491Wire(t *testing.T, value any) map[string]any {
	t.Helper()
	raw, err := json.Marshal(value)
	require.NoError(t, err)
	var wire map[string]any
	require.NoError(t, json.Unmarshal(raw, &wire))
	return wire
}

func TestToOwnerResponse_UnresolvedBasicIsAnError(t *testing.T) {
	t.Run("owner was never hydrated", func(t *testing.T) {
		_, err := toOwnerResponse(&model.Owner{ID: 1, ClinicID: 1})
		require.Error(t, err)
		assert.True(t, apperrors.HasCode(err, apperrors.CodeCommonBasicUnresolved))
	})

	t.Run("nil owner", func(t *testing.T) {
		_, err := toOwnerResponse(nil)
		require.Error(t, err)
		assert.True(t, apperrors.HasCode(err, apperrors.CodeCommonBasicUnresolved))
	})

	t.Run("owner has a revision but no subject", func(t *testing.T) {
		owner := emr491Owner()
		owner.BasicSubject = nil
		_, err := toOwnerResponse(owner)
		require.Error(t, err)
		assert.True(t, apperrors.HasCode(err, apperrors.CodeCommonBasicUnresolved))
	})

	t.Run("a nested pet that was never hydrated fails the whole owner", func(t *testing.T) {
		owner := emr491Owner()
		owner.Pets[0].Basic = nil
		_, err := toOwnerResponse(owner)
		require.Error(t, err)
		assert.True(t, apperrors.HasCode(err, apperrors.CodeCommonBasicUnresolved))
	})

	t.Run("a nested pet without its subject fails the whole owner", func(t *testing.T) {
		owner := emr491Owner()
		owner.Pets[0].BasicSubject = nil
		_, err := toOwnerResponse(owner)
		require.Error(t, err)
		assert.True(t, apperrors.HasCode(err, apperrors.CodeCommonBasicUnresolved))
	})
}

func TestToPetInOwnerResponse_UnresolvedBasicIsAnError(t *testing.T) {
	t.Run("pet was never hydrated", func(t *testing.T) {
		_, err := toPetInOwnerResponse(&model.Pet{ID: 1, ClinicID: 1})
		require.Error(t, err)
		assert.True(t, apperrors.HasCode(err, apperrors.CodeCommonBasicUnresolved))
	})

	t.Run("nil pet", func(t *testing.T) {
		_, err := toPetInOwnerResponse(nil)
		require.Error(t, err)
		assert.True(t, apperrors.HasCode(err, apperrors.CodeCommonBasicUnresolved))
	})
}

func TestToOwnerResponse_NullableBasicColumnsAreJSONNullAndVersionsAreExposed(t *testing.T) {
	resp, err := toOwnerResponse(emr491Owner())
	require.NoError(t, err)
	wire := emr491Wire(t, resp)

	assert.Equal(t, "山田太郎", wire["owner_name"])
	assert.EqualValues(t, 3, wire["owner_version"])
	assert.EqualValues(t, 30, wire["owner_basic_revision_id"])
	for _, key := range []string{
		"owner_name_kana", "birth_date", "company", "postal_code", "address1", "address2",
		"home_postal_code", "home_address1", "home_address2", "phone", "company_phone", "email",
	} {
		value, present := wire[key]
		assert.True(t, present, "%s must be on the wire", key)
		assert.Nil(t, value, "%s must be JSON null", key)
	}

	pets, ok := wire["pets"].([]any)
	require.True(t, ok)
	require.Len(t, pets, 1)
	pet, ok := pets[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "ポチ", pet["name"])
	assert.EqualValues(t, 3, pet["animal_species_id"])
	assert.EqualValues(t, 4, pet["pet_version"])
	assert.EqualValues(t, 40, pet["pet_basic_revision_id"])
	assert.Equal(t, "unknown", pet["neutered_status"])
	for _, key := range []string{
		"pet_name_kana", "gender", "status", "birth_date", "breed", "color", "blood_type", "microchip_number",
		"danger_level", "weight", "neutered_date", "acquisition_type", "food", "environment", "last_visit",
		"remarks", "deceased_at", "name_origin", "meeting_story",
	} {
		value, present := pet[key]
		assert.True(t, present, "pets[].%s must be on the wire", key)
		assert.Nil(t, value, "pets[].%s must be JSON null", key)
	}

	species, ok := pet["animal_species"].(map[string]any)
	require.True(t, ok)
	value, present := species["sort_order"]
	assert.True(t, present, "sort_order must be on the wire")
	assert.Nil(t, value, "sort_order must be JSON null")
}

func TestToPetInOwnerResponse_WeightKeepsTheRegisteredDecimal(t *testing.T) {
	pet := emr491Pet()
	pet.Basic.RegistrationWeight = &model.ExactNumeric{Numeric: pgtype.Numeric{Int: big.NewInt(525), Exp: -1, Valid: true}}

	resp, err := toPetInOwnerResponse(&pet)
	require.NoError(t, err)

	raw, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"weight":52.5`)
}

func TestGetOwner_UnresolvedBasicFailsClosedWith422(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithOwnerSvc(&mockOwnerService{
		getByIDForClinicsFn: func(context.Context, []uint64, uint64) (*model.Owner, error) {
			return &model.Owner{ID: 42, ClinicID: 1}, nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	c.Params = gin.Params{{Key: "id", Value: "42"}}
	setClinicID(c)

	h.GetOwner(c)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code, w.Body.String())
	assert.Contains(t, w.Body.String(), `"code":"common_basic_unresolved"`)
	assert.NotContains(t, w.Body.String(), `"owner_name"`)
}

func TestListOwners_UnresolvedBasicFailsTheWholePage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithOwnerSvc(&mockOwnerService{
		listFn: func(context.Context, []uint64, int, int, string) ([]model.Owner, int64, error) {
			return []model.Owner{*emr491Owner(), {ID: 2, ClinicID: 1}}, 2, nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/?page=1&limit=10", http.NoBody)
	setClinicID(c)

	h.ListOwners(c)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code, w.Body.String())
	assert.Contains(t, w.Body.String(), `"code":"common_basic_unresolved"`)
	assert.NotContains(t, w.Body.String(), `"owner_name"`)
}
