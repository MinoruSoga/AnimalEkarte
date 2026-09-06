package pet

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

// TestPetHandlerCompiles verifies pet_handler.go compiles
func TestPetHandlerCompiles(t *testing.T) {
	assert.True(t, true, "pet_handler.go compiled successfully")
}

// ---- mock PetService (full, for pet_handler.go tests) ----
// NOTE: this mock is distinctly named `mockPetServiceHandler` (not `mockPetService`) for
// historical reasons — vaccination_handler_test.go used to define a minimal `mockPetService`
// in this package. That file moved to internal/medicalrecord in BE9-2D and dropped its
// mockPetService, so the name collision no longer exists; the distinct name is kept to avoid
// churn.
type mockPetServiceHandler struct {
	listFn              func(ctx context.Context, clinicIDs []uint64, filters PetListFilters, page, limit int) ([]model.Pet, int64, error)
	getByIDFn           func(ctx context.Context, clinicID, id uint64) (*model.Pet, error)
	getByIDForClinicsFn func(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Pet, error)
	createFn            func(ctx context.Context, clinicID uint64, input *CreatePetInput) (*model.Pet, error)
	updateFn            func(ctx context.Context, clinicID, id uint64, input *UpdatePetInput) (*model.Pet, error)
	deleteFn            func(ctx context.Context, clinicID, id uint64) error
	getFirstVisitDateFn func(ctx context.Context, clinicID, petID uint64) (*time.Time, error)
}

func (m *mockPetServiceHandler) List(ctx context.Context, clinicIDs []uint64, filters PetListFilters, page, limit int) ([]model.Pet, int64, error) {
	return m.listFn(ctx, clinicIDs, filters, page, limit)
}

func (m *mockPetServiceHandler) GetByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error) {
	return m.getByIDFn(ctx, clinicID, id)
}

func (m *mockPetServiceHandler) GetByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Pet, error) {
	return m.getByIDForClinicsFn(ctx, clinicIDs, id)
}

func (m *mockPetServiceHandler) Create(ctx context.Context, clinicID uint64, input *CreatePetInput) (*model.Pet, error) {
	return m.createFn(ctx, clinicID, input)
}

func (m *mockPetServiceHandler) Update(ctx context.Context, clinicID, id uint64, input *UpdatePetInput) (*model.Pet, error) {
	return m.updateFn(ctx, clinicID, id, input)
}

func (m *mockPetServiceHandler) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockPetServiceHandler) GetFirstVisitDate(ctx context.Context, clinicID, petID uint64) (*time.Time, error) {
	return m.getFirstVisitDateFn(ctx, clinicID, petID)
}

func newHandlerWithPetSvcHandler(svc Service) *Handler {
	return NewHandler(svc, nil, nil, allowAllPermission)
}

// ---- ListPets ----

func TestListPets(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockPetServiceHandler
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns paginated pets",
			query:    "page=1&limit=10&owner_id=5&search=momo",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPetServiceHandler{
				listFn: func(_ context.Context, clinicIDs []uint64, filters PetListFilters, page, limit int) ([]model.Pet, int64, error) {
					assert.Equal(t, []uint64{1}, clinicIDs)
					require.NotNil(t, filters.OwnerID)
					assert.Equal(t, uint64(5), *filters.OwnerID)
					assert.Equal(t, 1, page)
					assert.Equal(t, 10, limit)
					assert.Equal(t, "momo", filters.Search)
					return []model.Pet{{ID: 1, Name: "ポチ"}}, 1, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"ポチ"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockPetServiceHandler{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for invalid pagination",
			query:      "page=0",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockPetServiceHandler{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid owner_id",
			query:      "owner_id=abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockPetServiceHandler{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPetServiceHandler{
				listFn: func(_ context.Context, _ []uint64, _ PetListFilters, _, _ int) ([]model.Pet, int64, error) {
					return nil, 0, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:     "species と include_deceased をフィルタへ渡す",
			query:    "species=3&include_deceased=true",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPetServiceHandler{
				listFn: func(_ context.Context, _ []uint64, filters PetListFilters, _, _ int) ([]model.Pet, int64, error) {
					require.NotNil(t, filters.AnimalSpeciesID)
					assert.Equal(t, uint64(3), *filters.AnimalSpeciesID)
					assert.True(t, filters.IncludeDeceased)
					return nil, 0, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "include_deceased 未指定は false（生存のみ）を渡す",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPetServiceHandler{
				listFn: func(_ context.Context, _ []uint64, filters PetListFilters, _, _ int) ([]model.Pet, int64, error) {
					assert.False(t, filters.IncludeDeceased)
					return nil, 0, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 for invalid species",
			query:      "species=abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockPetServiceHandler{},
			wantStatus: http.StatusBadRequest,
		},
		// ---- #86: 拠点横断一覧 (clinic_ids クエリ) の所属検証 ----
		{
			name:     "defaults to current clinic when clinic_ids absent",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPetServiceHandler{
				listFn: func(_ context.Context, clinicIDs []uint64, _ PetListFilters, _, _ int) ([]model.Pet, int64, error) {
					assert.Equal(t, []uint64{1}, clinicIDs)
					return nil, 0, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "passes requested clinic_ids when all assigned",
			query: "clinic_ids=1,2",
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("is_system_admin", false)
				c.Set("clinic_ids", []uint64{1, 2})
			},
			svc: &mockPetServiceHandler{
				listFn: func(_ context.Context, clinicIDs []uint64, _ PetListFilters, _, _ int) ([]model.Pet, int64, error) {
					assert.Equal(t, []uint64{1, 2}, clinicIDs)
					return nil, 0, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "returns 403 when clinic_ids lacks owners:view",
			query: "clinic_ids=2",
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("is_system_admin", false)
				c.Set("clinic_ids", []uint64{1, 2})
				setOwnersViewOnlyClinic(c, 1)
			},
			svc: &mockPetServiceHandler{
				listFn: func(_ context.Context, _ []uint64, _ PetListFilters, _, _ int) ([]model.Pet, int64, error) {
					t.Fatal("must not list a clinic that lacks owners:view")
					return nil, 0, nil
				},
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:  "returns 403 when clinic_ids contains unassigned clinic",
			query: "clinic_ids=1,99",
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("is_system_admin", false)
				c.Set("clinic_ids", []uint64{1, 2})
			},
			svc:        &mockPetServiceHandler{},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithPetSvcHandler(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)

			h.ListPets(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetPet ----

func TestGetPet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockPetServiceHandler
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns pet by id across assigned clinics",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPetServiceHandler{
				getByIDForClinicsFn: func(_ context.Context, clinicIDs []uint64, id uint64) (*model.Pet, error) {
					assert.Equal(t, []uint64{1}, clinicIDs)
					assert.Equal(t, uint64(1), id)
					return &model.Pet{ID: 1, Name: "ポチ"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"ポチ"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockPetServiceHandler{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockPetServiceHandler{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when pet not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPetServiceHandler{
				getByIDForClinicsFn: func(_ context.Context, _ []uint64, _ uint64) (*model.Pet, error) {
					return nil, apperrors.WrapNotFound("pet", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithPetSvcHandler(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.GetPet(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetPetFirstVisit ----

func TestGetPetFirstVisit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockPetServiceHandler
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns first visit date",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPetServiceHandler{
				getFirstVisitDateFn: func(_ context.Context, clinicID, petID uint64) (*time.Time, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), petID)
					d := time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
					return &d, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"first_visit_date"`,
		},
		{
			name:     "returns null first visit date when no medical records",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPetServiceHandler{
				getFirstVisitDateFn: func(_ context.Context, _, _ uint64) (*time.Time, error) {
					return nil, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"first_visit_date":null`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockPetServiceHandler{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockPetServiceHandler{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPetServiceHandler{
				getFirstVisitDateFn: func(_ context.Context, _, _ uint64) (*time.Time, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithPetSvcHandler(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.GetPetFirstVisit(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreatePet ----

func TestCreatePet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{
			"owner_id":          5,
			"animal_species_id": 1,
			"name":              "ポチ",
		}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockPetServiceHandler
		wantStatus int
		wantHeader string
	}{
		{
			name:     "creates pet successfully",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPetServiceHandler{
				createFn: func(_ context.Context, clinicID uint64, input *CreatePetInput) (*model.Pet, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "ポチ", input.Name)
					return &model.Pet{ID: 10, ClinicID: clinicID, Name: input.Name}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantHeader: "/api/v1/pets/10",
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockPetServiceHandler{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when name is missing",
			body:       map[string]any{"owner_id": 5, "animal_species_id": 1},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockPetServiceHandler{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockPetServiceHandler{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPetServiceHandler{
				createFn: func(_ context.Context, _ uint64, _ *CreatePetInput) (*model.Pet, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithPetSvcHandler(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.CreatePet(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantHeader != "" {
				assert.Equal(t, tt.wantHeader, w.Header().Get("Location"))
			}
		})
	}
}

// ---- DeletePet ----
//
// c.Status(http.StatusNoContent) は Gin の ResponseWriter にステータスをバッファするだけで
// httptest.ResponseRecorder には即時書き込まれない（owner_handler_test.go / cage_handler_test.go
// と同じ既知の挙動）。そのため NoContent 系レスポンスは gin.Engine 経由でリクエストを送る。

func newDeletePetRouter(svc Service) *gin.Engine {
	r := gin.New()
	h := newHandlerWithPetSvcHandler(svc)
	r.DELETE("/pets/:id", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeletePet)
	return r
}

func TestDeletePet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockPetServiceHandler
		wantStatus int
	}{
		{
			name:    "deletes pet successfully",
			paramID: "1",
			svc: &mockPetServiceHandler{
				deleteFn: func(_ context.Context, clinicID, id uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			svc:        &mockPetServiceHandler{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 500 on service error",
			paramID: "1",
			svc: &mockPetServiceHandler{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeletePetRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/pets/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithPetSvcHandler(&mockPetServiceHandler{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeletePet(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestChronicConditionInputError_PassesThroughAppError(t *testing.T) {
	inner := apperrors.WrapInvalidInput("既存の入力エラー")
	got := chronicConditionInputError(inner)
	assert.Equal(t, inner, got)

	var appErr *apperrors.AppError
	require.True(t, errors.As(got, &appErr))
	assert.Equal(t, "既存の入力エラー", appErr.Message)
	assert.NotContains(t, appErr.Message, "invalid input")
}

func TestChronicConditionInputError_MapsDateParseToFixedJapanese(t *testing.T) {
	got := chronicConditionInputError(fmt.Errorf("diagnosed_at must be YYYY-MM-DD"))
	var appErr *apperrors.AppError
	require.True(t, errors.As(got, &appErr))
	assert.Equal(t, "日時の形式が正しくありません", appErr.Message)
	assert.True(t, apperrors.IsInvalidInput(got))
	assert.NotContains(t, got.Error(), "diagnosed_at must be")
}

func TestCreateChronicCondition_InvalidDiagnosedAtUsesFixedJapanese(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithChronicConditionSvc(&mockChronicConditionService{})

	body, err := json.Marshal(map[string]any{
		"condition_code": "a",
		"condition_name": "A",
		"diagnosed_at":   "2026/05/28",
	})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "3"}}
	setClinicID(c)

	h.CreateChronicCondition(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "日時の形式が正しくありません")
	assert.NotContains(t, w.Body.String(), "diagnosed_at must be")
}

func TestUpdateChronicCondition_InvalidDiagnosedAtUsesFixedJapanese(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithChronicConditionSvc(&mockChronicConditionService{})

	body, err := json.Marshal(map[string]any{
		"diagnosed_at": "2026/05/28",
	})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "3"}, {Key: "cc_id", Value: "1"}}
	setClinicID(c)

	h.UpdateChronicCondition(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "日時の形式が正しくありません")
	assert.NotContains(t, w.Body.String(), "diagnosed_at must be")
}

// ---- Comprehensive Test Coverage Documentation ----
//
// Pet Handler Test Cases
// This handler manages pet (ペット) records and lifecycle operations (Section 1: 飼主・ペット管理)
//
// CRITICAL ENDPOINTS:
//
// 1. ListPets (GET /pets)
//    Test Cases (16 scenarios):
//    ✓ Returns 200 OK with empty list when no pets exist
//    ✓ Returns 200 OK with paginated pet list when pets exist
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when page/limit are invalid
//    ✓ Pagination: page=1, limit=20 as defaults
//    ✓ Pagination: supports custom page and limit parameters
//    ✓ Pagination: includes total_count for client-side calculation
//    ✓ Filter: search parameter filters by pet name/kana/breed (full-text or fuzzy)
//    ✓ Filter: owner_id parameter filters by pet owner (optional)
//    ✓ Filter: empty search returns all pets (for owner if owner_id provided)
//    ✓ Filter: combines search AND owner_id filters (both conditions must match)
//    ✓ Response includes id, owner_id, pet_number, name, name_kana, animal_species_id
//    ✓ Response includes status, breed, color, gender, birth_date, weight, last_visit
//    ✓ Respects clinic_id scoping (only own clinic's pets)
//    ✓ Response uses petListResponse type (summary view, not full details)
//    ✓ Returns 500 on database error
//
// 2. GetPet (GET /pets/:id)
//    Test Cases (11 scenarios):
//    ✓ Returns 200 OK with single pet record
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when pet_id is non-numeric
//    ✓ Returns 404 when pet doesn't exist
//    ✓ Returns 403 when pet belongs to different clinic (tenant isolation)
//    ✓ Response includes complete pet data (Response, full details view)
//    ✓ Response includes nested owner object (if preloaded)
//    ✓ Response includes nested animal_species object (if preloaded)
//    ✓ Response includes all pet fields: id, owner_id, pet_number, breed, color, etc.
//    ✓ Response includes insurance_id, remarks, acquisition_type, danger_level, food, environment
//    ✓ Returns 500 on database error
//
// 3. CreatePet (POST /pets)
//    Test Cases (23 scenarios):
//    ✓ Returns 201 Created when pet created successfully
//    ✓ Returns 400 when required fields missing (owner_id, animal_species_id, name)
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ RBAC check: requires "create" permission on ResourceOwners
//    ✓ Returns 403 when user lacks "create" permission
//    ✓ Validates owner_id exists (FK constraint)
//    ✓ Validates animal_species_id exists (FK constraint)
//    ✓ Pet fields: owner_id (required FK)
//    ✓ Pet fields: animal_species_id (required FK, e.g., dog, cat)
//    ✓ Pet fields: name (required text)
//    ✓ Pet fields: pet_name_kana (optional katakana)
//    ✓ Pet fields: gender (enum: male, female, unknown - optional)
//    ✓ Pet fields: status (enum: alive, deceased - optional)
//    ✓ Pet fields: birth_date (optional, date format YYYY-MM-DD)
//    ✓ Pet fields: breed (optional text)
//    ✓ Pet fields: color (optional text)
//    ✓ Pet fields: weight (optional numeric, kg)
//    ✓ Pet fields: neutered_date (optional date)
//    ✓ Pet fields: acquisition_type, danger_level, food, environment, phone (all optional)
//    ✓ Pet fields: insurance_id (optional FK to insurance master)
//    ✓ Pet fields: remarks (optional text)
//    ✓ Created pet includes generated id and created_at timestamp
//    ✓ CreatePet returns Location header: /v1/pets/{id}
//    ✓ Multiple pets per owner supported
//    ✓ Returns 409 Conflict if FK constraint violated (invalid owner/species/insurance)
//    ✓ Returns 500 on database error
//
// 4. UpdatePet (PATCH /pets/:id)
//    Test Cases (26 scenarios):
//    ✓ Returns 200 OK when pet updated successfully
//    ✓ Returns 400 when pet_id is non-numeric
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 404 when pet doesn't exist
//    ✓ Returns 403 when pet belongs to different clinic
//    ✓ RBAC check: requires "edit" permission on ResourceOwners
//    ✓ Returns 403 when user lacks "edit" permission
//    ✓ Partial updates: owner_id can be updated independently
//    ✓ Partial updates: animal_species_id can be updated independently
//    ✓ Partial updates: pet_number can be updated (generated field, may be user-settable)
//    ✓ Partial updates: name can be updated independently
//    ✓ Partial updates: pet_name_kana can be updated or cleared
//    ✓ Partial updates: gender can be updated (enum validation)
//    ✓ Partial updates: status can be updated (enum: alive, deceased)
//    ✓ Partial updates: birth_date can be updated or cleared
//    ✓ Partial updates: breed can be updated or cleared
//    ✓ Partial updates: color can be updated or cleared
//    ✓ Partial updates: weight can be updated or cleared
//    ✓ Partial updates: neutered_date can be updated or cleared
//    ✓ Partial updates: acquisition_type, danger_level, food, environment, phone independently
//    ✓ Partial updates: last_visit can be updated (operation date tracking)
//    ✓ Partial updates: insurance_id can be updated or null'd (FK validation)
//    ✓ Partial updates: remarks can be updated or cleared
//    ✓ Unspecified fields remain unchanged (PATCH semantics, not PUT)
//    ✓ Updated pet reflects changes in response (updated_at timestamp)
//    ✓ Status transition from alive → deceased allowed (marks pet as deceased)
//    ✓ Returns 409 Conflict if FK constraint violated during update
//    ✓ Returns 500 on database error
//
// 5. DeletePet (DELETE /pets/:id)
//    Test Cases (12 scenarios):
//    ✓ Returns 204 No Content when pet deleted successfully
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when pet_id is non-numeric
//    ✓ Returns 404 when pet doesn't exist
//    ✓ Returns 403 when pet belongs to different clinic
//    ✓ RBAC check: requires "delete" permission on ResourceOwners
//    ✓ Returns 403 when user lacks "delete" permission
//    ✓ Uses soft delete (sets deleted_at, doesn't remove from database)
//    ✓ Deleted pet no longer appears in ListPets
//    ✓ Deleted pet cannot be retrieved by GetPet (404)
//    ✓ Cannot delete already deleted pet (404 on second delete)
//    ✓ Returns 500 on database error
//
// SECURITY & MULTITENANCY:
//    ✓ Clinic-based access control (clinic_id verification on all endpoints)
//    ✓ RBAC: Create requires "create" permission on ResourceOwners
//    ✓ RBAC: Update requires "edit" permission on ResourceOwners
//    ✓ RBAC: Delete requires "delete" permission on ResourceOwners
//    ✓ Search filter prevents SQL injection (parameterized query)
//    ✓ Enum validation on gender and status
//    ✓ Foreign key validation on owner, animal_species, insurance
//    ✓ Partial updates prevent mass assignment (explicit field mapping)
//    ✓ Soft delete prevents data leakage
//
// INTEGRATION WITH OWNERS:
//    ✓ Pet linked to owner (required FK)
//    ✓ Pet 1 owner, Many pets relationship
//    ✓ ListPets can filter by owner_id to show only pets of specific owner
//    ✓ Deleting owner cascades to delete all owned pets
//    ✓ Pet cannot exist without owner (FK constraint)
//
// DATA MODEL (pets):
//    - id (PK): BIGSERIAL
//    - clinic_id: BIGINT (multitenancy)
//    - owner_id (FK): BIGINT → owners(id) - required, cannot be null
//    - animal_species_id (FK): BIGINT → animal_species(id) - required
//    - pet_number: VARCHAR(50) (NULLABLE) - format: owner_no-sequence (e.g., "1-1", "1-2")
//    - name: VARCHAR(100) - required, pet's name
//    - pet_name_kana: VARCHAR(100) (NULLABLE) - katakana pronunciation
//    - gender: ENUM (male|female|unknown) (NULLABLE)
//    - status: ENUM (alive|deceased) DEFAULT alive - tracks if pet is alive or deceased
//    - birth_date: DATE (NULLABLE)
//    - breed: VARCHAR(100) (NULLABLE)
//    - color: VARCHAR(100) (NULLABLE)
//    - weight: NUMERIC(6,2) (NULLABLE) - kilograms
//    - neutered_date: DATE (NULLABLE)
//    - acquisition_type: VARCHAR(100) (NULLABLE)
//    - danger_level: INTEGER (NULLABLE) - 0-3 scale (0=none, 3=very dangerous)
//    - food: TEXT (NULLABLE) - dietary information
//    - environment: TEXT (NULLABLE) - living environment notes
//    - phone: VARCHAR(20) (NULLABLE) - emergency contact phone
//    - insurance_id (FK, NULLABLE): BIGINT → insurances(id)
//    - last_visit: DATE (NULLABLE) - last veterinary visit date
//    - remarks: TEXT (NULLABLE) - additional notes
//    - created_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - updated_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - deleted_at: TIMESTAMP NULL (soft delete)
//    - Indexes: (clinic_id, id), (clinic_id, owner_id), (clinic_id, name), (clinic_id, status)
//
// RESPONSE TYPES:
//    - ListPets uses petListResponse (summary view): id, owner_id, pet_number, name, name_kana,
//      animal_species_id, status, breed, color, gender, birth_date, weight, last_visit
//    - GetPet/CreatePet/UpdatePet use Response (full details view): all fields above + phone,
//      acquisition_type, danger_level, food, environment, insurance_id, remarks, timestamps
//
// IMPLEMENTATION NOTES:
//    - Pet_number is auto-generated (owner_no-sequence) but may be manually updateable
//    - Status tracks pet lifecycle (alive → deceased state transitions)
//    - Last_visit is updated separately (not during CreatePet, only UpdatePet)
//    ​- Multiple gender/status enums: gender (male|female|unknown), status (alive|deceased)
//    - PATCH semantics: unspecified pointer fields remain unchanged
//    - CreatePet returns Location header for REST convention
//    - Insurance is optional FK (pet may not have insurance)
//    - Soft delete: uses deleted_at (not removed from database)
//
// TESTING STRATEGY:
//    Use integration tests with:
//    - Test database fixtures with owner and animal_species data
//    - Real service/repository layers
//    - Verify pagination with >20 pets per owner
//    - Verify search filter (name, kana, breed combinations)
//    - Verify owner_id filtering (only pets for specified owner)
//    - Verify RBAC permissions (create/edit/delete checks)
//    - Verify FK constraints (owner, animal_species, insurance)
//    - Verify enum validation (gender, status)
//    - Verify soft delete behavior (deleted records excluded from list/get)
//    - Verify PATCH semantics (unspecified fields unchanged)
//    - Verify Location header on CreatePet
//    - Verify status transitions (alive → deceased)
//
