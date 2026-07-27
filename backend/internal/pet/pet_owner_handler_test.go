package pet

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type petOwnerHandlerServiceDouble struct {
	getByPetIDFn             func(context.Context, uint64, uint64) ([]model.PetOwner, error)
	getSharedPetsByOwnerIDFn func(context.Context, uint64, uint64) ([]SharedPet, error)
	replaceForPetFn          func(context.Context, uint64, uint64, *ReplacePetOwnersInput) error
}

func (d *petOwnerHandlerServiceDouble) GetByPetID(
	ctx context.Context,
	clinicID, petID uint64,
) ([]model.PetOwner, error) {
	if d.getByPetIDFn == nil {
		return nil, nil
	}
	return d.getByPetIDFn(ctx, clinicID, petID)
}

func (d *petOwnerHandlerServiceDouble) ReplaceForPet(
	ctx context.Context,
	clinicID, petID uint64,
	input *ReplacePetOwnersInput,
) error {
	if d.replaceForPetFn == nil {
		return nil
	}
	return d.replaceForPetFn(ctx, clinicID, petID, input)
}

func (d *petOwnerHandlerServiceDouble) GetSharedPetsByOwnerID(
	ctx context.Context,
	clinicID, ownerID uint64,
) ([]SharedPet, error) {
	if d.getSharedPetsByOwnerIDFn == nil {
		return nil, nil
	}
	return d.getSharedPetsByOwnerIDFn(ctx, clinicID, ownerID)
}

type petOwnerDetailsFinderDouble struct {
	findByIDsFn func(context.Context, uint64, []uint64) ([]*model.Owner, error)
}

func (d *petOwnerDetailsFinderDouble) FindByIDs(
	ctx context.Context,
	clinicID uint64,
	ownerIDs []uint64,
) ([]*model.Owner, error) {
	if d.findByIDsFn == nil {
		return nil, nil
	}
	return d.findByIDsFn(ctx, clinicID, ownerIDs)
}

func newPetOwnerHandlerForTest(
	service PetOwnerService,
	owners PetOwnerDetailsFinder,
) *Handler {
	return NewHandlerWithPetOwners(
		nil,
		nil,
		nil,
		service,
		owners,
		allowAllPermission,
	)
}

func TestPetOwnerHandler_ListReturnsExplicitResponseAndScopesOwnerLookup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &petOwnerHandlerServiceDouble{
		getByPetIDFn: func(_ context.Context, clinicID, petID uint64) ([]model.PetOwner, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(7), petID)
			return []model.PetOwner{{OwnerID: 12, Relationship: "妻"}}, nil
		},
	}
	owners := &petOwnerDetailsFinderDouble{
		findByIDsFn: func(_ context.Context, clinicID uint64, ownerIDs []uint64) ([]*model.Owner, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, []uint64{12}, ownerIDs)
			return []*model.Owner{{
				ID:       12,
				ClinicID: 1,
				Name:     "山田 花子",
				NameKana: "ヤマダ ハナコ",
			}}, nil
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Params = gin.Params{{Key: "id", Value: "7"}}
	setClinicID(c)

	newPetOwnerHandlerForTest(service, owners).ListPetOwners(c)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.ElementsMatch(t, []string{"sub_owners"}, objectKeys(body))
	items, ok := body["sub_owners"].([]any)
	require.True(t, ok)
	require.Len(t, items, 1)
	item, ok := items[0].(map[string]any)
	require.True(t, ok)
	assert.ElementsMatch(
		t,
		[]string{"owner_id", "name", "name_kana", "relationship"},
		objectKeys(item),
	)
	assert.Equal(t, float64(12), item["owner_id"])
	assert.Equal(t, "山田 花子", item["name"])
	assert.Equal(t, "ヤマダ ハナコ", item["name_kana"])
	assert.Equal(t, "妻", item["relationship"])
}

func TestPetOwnerHandler_ListSharedPetsReturnsExplicitResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	birthDate := time.Date(2020, time.January, 2, 0, 0, 0, 0, time.Local)
	weight := 4.2
	service := &petOwnerHandlerServiceDouble{
		getSharedPetsByOwnerIDFn: func(
			_ context.Context,
			clinicID, ownerID uint64,
		) ([]SharedPet, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(12), ownerID)
			return []SharedPet{{
				ID:                7,
				PetNumber:         "P-0007",
				Name:              "ポチ",
				Status:            model.PetStatusDeceased,
				AnimalSpeciesName: "犬",
				Gender:            model.PetGenderMale,
				BirthDate:         &birthDate,
				Color:             "茶",
				Weight:            &weight,
				Environment:       "室内",
				Remarks:           "共同飼育",
				Relationship:      "妻",
			}}, nil
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Params = gin.Params{{Key: "id", Value: "12"}}
	setClinicID(c)

	newPetOwnerHandlerForTest(service, &petOwnerDetailsFinderDouble{}).ListOwnerSharedPets(c)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.ElementsMatch(t, []string{"shared_pets"}, objectKeys(body))
	items, ok := body["shared_pets"].([]any)
	require.True(t, ok)
	require.Len(t, items, 1)
	item, ok := items[0].(map[string]any)
	require.True(t, ok)
	assert.ElementsMatch(t, []string{
		"id",
		"pet_number",
		"name",
		"status",
		"animal_species",
		"gender",
		"birth_date",
		"color",
		"weight",
		"environment",
		"remarks",
		"relationship",
	}, objectKeys(item))
	assert.Equal(t, float64(7), item["id"])
	assert.Equal(t, "deceased", item["status"])
	assert.Equal(t, "2020-01-02", item["birth_date"])
	assert.Equal(t, "妻", item["relationship"])
	species, ok := item["animal_species"].(map[string]any)
	require.True(t, ok)
	assert.ElementsMatch(t, []string{"name"}, objectKeys(species))
	assert.Equal(t, "犬", species["name"])
}

func TestPetOwnerHandler_ListSharedPetsReturnsEmptyArray(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &petOwnerHandlerServiceDouble{
		getSharedPetsByOwnerIDFn: func(
			context.Context,
			uint64,
			uint64,
		) ([]SharedPet, error) {
			return []SharedPet{}, nil
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Params = gin.Params{{Key: "id", Value: "12"}}
	setClinicID(c)

	newPetOwnerHandlerForTest(service, &petOwnerDetailsFinderDouble{}).ListOwnerSharedPets(c)

	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"shared_pets":[]}`, w.Body.String())
}

func TestPetOwnerHandler_ListSharedPetsRejectsInvalidOwnerIDBeforeService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serviceCalled := false
	service := &petOwnerHandlerServiceDouble{
		getSharedPetsByOwnerIDFn: func(context.Context, uint64, uint64) ([]SharedPet, error) {
			serviceCalled = true
			return nil, nil
		},
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	setClinicID(c)

	newPetOwnerHandlerForTest(service, &petOwnerDetailsFinderDouble{}).ListOwnerSharedPets(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, serviceCalled)
}

func TestPetOwnerHandler_ListSharedPetsRequiresClinicContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serviceCalled := false
	service := &petOwnerHandlerServiceDouble{
		getSharedPetsByOwnerIDFn: func(context.Context, uint64, uint64) ([]SharedPet, error) {
			serviceCalled = true
			return nil, nil
		},
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Params = gin.Params{{Key: "id", Value: "12"}}

	newPetOwnerHandlerForTest(service, &petOwnerDetailsFinderDouble{}).ListOwnerSharedPets(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.False(t, serviceCalled)
}

func TestPetOwnerHandler_ListSharedPetsMapsServiceErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		serviceErr error
		wantStatus int
	}{
		{
			name:       "owner not found",
			serviceErr: apperrors.WrapNotFound("owner", "12"),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "repository failure",
			serviceErr: apperrors.WrapInternalServerError("database detail must not leak"),
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &petOwnerHandlerServiceDouble{
				getSharedPetsByOwnerIDFn: func(context.Context, uint64, uint64) ([]SharedPet, error) {
					return nil, tt.serviceErr
				},
			}
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			c.Params = gin.Params{{Key: "id", Value: "12"}}
			setClinicID(c)

			newPetOwnerHandlerForTest(service, &petOwnerDetailsFinderDouble{}).ListOwnerSharedPets(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.NotContains(t, w.Body.String(), "database detail must not leak")
		})
	}
}

func TestPetOwnerHandler_ListMapsCrossClinicPetToNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ownerLookupCalled := false
	service := &petOwnerHandlerServiceDouble{
		getByPetIDFn: func(context.Context, uint64, uint64) ([]model.PetOwner, error) {
			return nil, apperrors.WrapNotFound("pet", "7")
		},
	}
	owners := &petOwnerDetailsFinderDouble{
		findByIDsFn: func(context.Context, uint64, []uint64) ([]*model.Owner, error) {
			ownerLookupCalled = true
			return nil, nil
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Params = gin.Params{{Key: "id", Value: "7"}}
	setClinicID(c)

	newPetOwnerHandlerForTest(service, owners).ListPetOwners(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.False(t, ownerLookupCalled)
}

func TestPetOwnerHandler_ListFailsClosedWhenOwnerDetailsAreOutsideClinic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &petOwnerHandlerServiceDouble{
		getByPetIDFn: func(context.Context, uint64, uint64) ([]model.PetOwner, error) {
			return []model.PetOwner{{OwnerID: 12, Relationship: "妻"}}, nil
		},
	}
	owners := &petOwnerDetailsFinderDouble{
		findByIDsFn: func(context.Context, uint64, []uint64) ([]*model.Owner, error) {
			return []*model.Owner{}, nil
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Params = gin.Params{{Key: "id", Value: "7"}}
	setClinicID(c)

	newPetOwnerHandlerForTest(service, owners).ListPetOwners(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestPetOwnerHandler_ReplaceReturnsNoContentAndPassesAuditActor(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &petOwnerHandlerServiceDouble{
		replaceForPetFn: func(
			_ context.Context,
			clinicID, petID uint64,
			input *ReplacePetOwnersInput,
		) error {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(7), petID)
			require.NotNil(t, input)
			require.NotNil(t, input.Version)
			assert.Equal(t, 3, *input.Version)
			assert.Equal(t, []PetOwnerLinkInput{{OwnerID: 12, Relationship: "妻"}}, input.Links)
			require.NotNil(t, input.ActorID)
			assert.Equal(t, uint64(42), *input.ActorID)
			assert.Equal(t, model.AuditActorTypeStaff, input.ActorType)
			assert.Equal(t, "203.0.113.10", input.IPAddress)
			assert.Equal(t, "unit5-test-agent", input.UserAgent)
			return nil
		},
	}

	body := []byte(`{"version":3,"sub_owners":[{"owner_id":12,"relationship":" 妻 "}]}`)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("User-Agent", "unit5-test-agent")
	c.Request.RemoteAddr = "203.0.113.10:12345"
	c.Params = gin.Params{{Key: "id", Value: "7"}}
	setClinicID(c)
	setStaffID(c, 42)

	newPetOwnerHandlerForTest(service, &petOwnerDetailsFinderDouble{}).ReplacePetOwners(c)

	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	assert.Empty(t, w.Body.Bytes())
}

func TestPetOwnerHandler_ReplaceRejectsInvalidVersionWithoutCallingService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		body string
	}{
		{name: "missing", body: `{"sub_owners":[]}`},
		{name: "null", body: `{"version":null,"sub_owners":[]}`},
		{name: "wrong type", body: `{"version":"3","sub_owners":[]}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceCalled := false
			ownerLookupCalled := false
			service := &petOwnerHandlerServiceDouble{
				replaceForPetFn: func(context.Context, uint64, uint64, *ReplacePetOwnersInput) error {
					serviceCalled = true
					return nil
				},
			}
			owners := &petOwnerDetailsFinderDouble{
				findByIDsFn: func(context.Context, uint64, []uint64) ([]*model.Owner, error) {
					ownerLookupCalled = true
					return nil, nil
				},
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewBufferString(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: "7"}}
			setClinicID(c)
			setStaffID(c, 42)

			newPetOwnerHandlerForTest(service, owners).ReplacePetOwners(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.False(t, serviceCalled)
			assert.False(t, ownerLookupCalled)
		})
	}
}

func TestPetOwnerHandler_ReplaceRequiresAuthenticatedStaffActor(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serviceCalled := false
	service := &petOwnerHandlerServiceDouble{
		replaceForPetFn: func(context.Context, uint64, uint64, *ReplacePetOwnersInput) error {
			serviceCalled = true
			return nil
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(
		http.MethodPut,
		"/",
		bytes.NewBufferString(`{"version":3,"sub_owners":[]}`),
	)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "7"}}
	setClinicID(c)

	newPetOwnerHandlerForTest(service, &petOwnerDetailsFinderDouble{}).ReplacePetOwners(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.False(t, serviceCalled)
}

func TestPetOwnerHandler_ReplaceRequiresExplicitSubOwnersArray(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		body string
	}{
		{name: "missing", body: `{"version":3}`},
		{name: "null", body: `{"version":3,"sub_owners":null}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceCalled := false
			service := &petOwnerHandlerServiceDouble{
				replaceForPetFn: func(context.Context, uint64, uint64, *ReplacePetOwnersInput) error {
					serviceCalled = true
					return nil
				},
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewBufferString(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: "7"}}
			setClinicID(c)
			setStaffID(c, 42)

			newPetOwnerHandlerForTest(service, &petOwnerDetailsFinderDouble{}).ReplacePetOwners(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.False(t, serviceCalled)
		})
	}
}

func TestPetOwnerHandler_ReplaceAcceptsEmptyArray(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &petOwnerHandlerServiceDouble{
		replaceForPetFn: func(
			_ context.Context,
			_, _ uint64,
			input *ReplacePetOwnersInput,
		) error {
			require.NotNil(t, input)
			assert.Empty(t, input.Links)
			require.NotNil(t, input.ActorID)
			return nil
		},
	}

	body := []byte(`{"version":3,"sub_owners":[]}`)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "7"}}
	setClinicID(c)
	setStaffID(c, 42)

	newPetOwnerHandlerForTest(service, &petOwnerDetailsFinderDouble{}).ReplacePetOwners(c)

	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	assert.Empty(t, w.Body.Bytes())
}

func TestPetOwnerHandler_ReplaceMapsServiceErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		serviceErr error
		wantStatus int
	}{
		{
			name:       "not found",
			serviceErr: apperrors.WrapNotFound("pet", "7"),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "primary owner duplicate",
			serviceErr: apperrors.WrapInvalidInput("primary owner duplicate"),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "request duplicate",
			serviceErr: apperrors.WrapInvalidInput("request duplicate"),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "relationship too long",
			serviceErr: apperrors.WrapInvalidInput("relationship too long"),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "version conflict",
			serviceErr: apperrors.WrapConflict("version conflict"),
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &petOwnerHandlerServiceDouble{
				replaceForPetFn: func(context.Context, uint64, uint64, *ReplacePetOwnersInput) error {
					return tt.serviceErr
				},
			}
			body := []byte(`{"version":3,"sub_owners":[{"owner_id":12,"relationship":"妻"}]}`)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: "7"}}
			setClinicID(c)
			setStaffID(c, 42)

			newPetOwnerHandlerForTest(service, &petOwnerDetailsFinderDouble{}).ReplacePetOwners(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func objectKeys(value map[string]any) []string {
	keys := make([]string, 0, len(value))
	for key := range value {
		keys = append(keys, key)
	}
	return keys
}
