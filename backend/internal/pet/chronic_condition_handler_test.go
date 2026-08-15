package pet

import (
	"bytes"
	"context"
	"encoding/json"
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

// ---- mock ChronicConditionService ----

type mockChronicConditionService struct {
	listFn   func(ctx context.Context, clinicID, petID uint64) ([]model.PetChronicCondition, error)
	createFn func(ctx context.Context, clinicID, petID uint64, input CreateChronicConditionInput) (*model.PetChronicCondition, error)
	updateFn func(ctx context.Context, clinicID, petID, id uint64, input UpdateChronicConditionInput) (*model.PetChronicCondition, error)
	deleteFn func(ctx context.Context, clinicID, petID, id uint64) error
}

func (m *mockChronicConditionService) List(ctx context.Context, clinicID, petID uint64) ([]model.PetChronicCondition, error) {
	return m.listFn(ctx, clinicID, petID)
}

func (m *mockChronicConditionService) Create(
	ctx context.Context, clinicID, petID uint64, input CreateChronicConditionInput,
) (*model.PetChronicCondition, error) {
	return m.createFn(ctx, clinicID, petID, input)
}

func (m *mockChronicConditionService) Update(
	ctx context.Context, clinicID, petID, id uint64, input UpdateChronicConditionInput,
) (*model.PetChronicCondition, error) {
	return m.updateFn(ctx, clinicID, petID, id, input)
}

func (m *mockChronicConditionService) Delete(ctx context.Context, clinicID, petID, id uint64) error {
	return m.deleteFn(ctx, clinicID, petID, id)
}

func newHandlerWithChronicConditionSvc(svc ChronicConditionService) *Handler {
	return NewHandler(nil, nil, svc, allowAllPermission)
}

// ---- toChronicConditionResponse / toChronicConditionListResponse ----

func TestToChronicConditionResponse(t *testing.T) {
	diagnosedAt := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 5, 27, 9, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 29, 9, 0, 0, 0, time.UTC)
	notes := "requires monitoring"

	resp := toChronicConditionResponse(&model.PetChronicCondition{
		ID:            1,
		ClinicID:      2,
		PetID:         3,
		ConditionCode: "heart_disease",
		ConditionName: "Heart disease",
		DiagnosedAt:   diagnosedAt,
		Notes:         &notes,
		IsActive:      true,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	})

	assert.Equal(t, uint64(1), resp.ID)
	assert.Equal(t, uint64(2), resp.ClinicID)
	assert.Equal(t, uint64(3), resp.PetID)
	assert.Equal(t, "heart_disease", resp.ConditionCode)
	assert.Equal(t, "Heart disease", resp.ConditionName)
	assert.Equal(t, "2026-05-28", resp.DiagnosedAt)
	require.NotNil(t, resp.Notes)
	assert.Equal(t, notes, *resp.Notes)
	assert.True(t, resp.IsActive)
	assert.Equal(t, createdAt.In(time.Local).Format(time.RFC3339), resp.CreatedAt)
	assert.Equal(t, updatedAt.In(time.Local).Format(time.RFC3339), resp.UpdatedAt)
}

func TestToChronicConditionResponse_NilNotes(t *testing.T) {
	resp := toChronicConditionResponse(&model.PetChronicCondition{
		ID:            5,
		ConditionCode: "kidney_disease",
		ConditionName: "Kidney disease",
		DiagnosedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Notes:         nil,
		IsActive:      false,
	})

	assert.Equal(t, uint64(5), resp.ID)
	assert.Nil(t, resp.Notes)
	assert.False(t, resp.IsActive)
}

func TestToChronicConditionListResponse(t *testing.T) {
	records := []model.PetChronicCondition{
		{ID: 1, ConditionCode: "a", ConditionName: "A", DiagnosedAt: time.Now()},
		{ID: 2, ConditionCode: "b", ConditionName: "B", DiagnosedAt: time.Now()},
	}

	resp := toChronicConditionListResponse(records)

	require.Len(t, resp, 2)
	assert.Equal(t, uint64(1), resp[0].ID)
	assert.Equal(t, uint64(2), resp[1].ID)
}

func TestToChronicConditionListResponse_Empty(t *testing.T) {
	resp := toChronicConditionListResponse(nil)
	assert.Empty(t, resp)
	assert.NotNil(t, resp)
}

// ---- ListChronicConditions ----

func TestListChronicConditions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockChronicConditionService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of chronic conditions",
			paramID:  "3",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockChronicConditionService{
				listFn: func(_ context.Context, clinicID, petID uint64) ([]model.PetChronicCondition, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(3), petID)
					return []model.PetChronicCondition{{ID: 1, ConditionCode: "heart_disease", ConditionName: "Heart disease", DiagnosedAt: time.Now()}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"condition_code":"heart_disease"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "3",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockChronicConditionService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric pet id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockChronicConditionService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "3",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockChronicConditionService{
				listFn: func(_ context.Context, _, _ uint64) ([]model.PetChronicCondition, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithChronicConditionSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.ListChronicConditions(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateChronicCondition ----

func TestCreateChronicCondition(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{
			"condition_code": "heart_disease",
			"condition_name": "Heart disease",
			"diagnosed_at":   "2026-05-28",
		}
	}

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockChronicConditionService
		wantStatus int
		wantBody   string
		wantHeader bool
	}{
		{
			name:     "creates chronic condition successfully",
			paramID:  "3",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockChronicConditionService{
				createFn: func(_ context.Context, clinicID, petID uint64, input CreateChronicConditionInput) (*model.PetChronicCondition, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(3), petID)
					assert.Equal(t, "heart_disease", input.ConditionCode)
					return &model.PetChronicCondition{
						ID:            10,
						ClinicID:      clinicID,
						PetID:         petID,
						ConditionCode: input.ConditionCode,
						ConditionName: input.ConditionName,
						DiagnosedAt:   input.DiagnosedAt,
						IsActive:      input.IsActive,
					}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"id":10`,
			wantHeader: true,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "3",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockChronicConditionService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric pet id",
			paramID:    "abc",
			body:       validBody(),
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockChronicConditionService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when required fields are missing",
			paramID:    "3",
			body:       map[string]any{"condition_name": "Heart disease"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockChronicConditionService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for malformed JSON",
			paramID:    "3",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockChronicConditionService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid diagnosed_at format",
			paramID:    "3",
			body:       map[string]any{"condition_code": "a", "condition_name": "A", "diagnosed_at": "2026/05/28"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockChronicConditionService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "3",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockChronicConditionService{
				createFn: func(_ context.Context, _, _ uint64, _ CreateChronicConditionInput) (*model.PetChronicCondition, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithChronicConditionSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.CreateChronicCondition(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
			if tt.wantHeader {
				assert.NotEmpty(t, w.Header().Get("Location"))
			}
		})
	}
}

// ---- UpdateChronicCondition ----

func TestUpdateChronicCondition(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		petParamID string
		ccParamID  string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockChronicConditionService
		wantStatus int
	}{
		{
			name:       "updates chronic condition successfully",
			petParamID: "3",
			ccParamID:  "1",
			body:       map[string]any{"condition_name": "更新後の病名"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc: &mockChronicConditionService{
				updateFn: func(_ context.Context, clinicID, petID, id uint64, input UpdateChronicConditionInput) (*model.PetChronicCondition, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(3), petID)
					assert.Equal(t, uint64(1), id)
					require.NotNil(t, input.ConditionName)
					assert.Equal(t, "更新後の病名", *input.ConditionName)
					return &model.PetChronicCondition{ID: 1, PetID: petID, ConditionName: *input.ConditionName}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			petParamID: "3",
			ccParamID:  "1",
			body:       map[string]any{"condition_name": "テスト"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockChronicConditionService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric pet id",
			petParamID: "abc",
			ccParamID:  "1",
			body:       map[string]any{"condition_name": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockChronicConditionService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for non-numeric cc id",
			petParamID: "3",
			ccParamID:  "xyz",
			body:       map[string]any{"condition_name": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockChronicConditionService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for malformed JSON",
			petParamID: "3",
			ccParamID:  "1",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockChronicConditionService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid diagnosed_at format",
			petParamID: "3",
			ccParamID:  "1",
			body:       map[string]any{"diagnosed_at": "2026/05/28"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockChronicConditionService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 404 when chronic condition not found",
			petParamID: "3",
			ccParamID:  "999",
			body:       map[string]any{"condition_name": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc: &mockChronicConditionService{
				updateFn: func(_ context.Context, _, _, _ uint64, _ UpdateChronicConditionInput) (*model.PetChronicCondition, error) {
					return nil, apperrors.WrapNotFound("chronic_condition", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithChronicConditionSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.petParamID}, {Key: "cc_id", Value: tt.ccParamID}}
			tt.setupCtx(c)

			h.UpdateChronicCondition(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- DeleteChronicCondition ----

func TestDeleteChronicCondition(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		petParamID string
		ccParamID  string
		svc        *mockChronicConditionService
		wantStatus int
	}{
		{
			name:       "deletes chronic condition successfully",
			petParamID: "3",
			ccParamID:  "1",
			svc: &mockChronicConditionService{
				deleteFn: func(_ context.Context, clinicID, petID, id uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(3), petID)
					assert.Equal(t, uint64(1), id)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 400 for non-numeric pet id",
			petParamID: "abc",
			ccParamID:  "1",
			svc:        &mockChronicConditionService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for non-numeric cc id",
			petParamID: "3",
			ccParamID:  "xyz",
			svc:        &mockChronicConditionService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 404 when chronic condition not found",
			petParamID: "3",
			ccParamID:  "999",
			svc: &mockChronicConditionService{
				deleteFn: func(_ context.Context, _, _, _ uint64) error {
					return apperrors.WrapNotFound("chronic_condition", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteChronicConditionRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/pets/"+tt.petParamID+"/chronic-conditions/"+tt.ccParamID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithChronicConditionSvc(&mockChronicConditionService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "3"}, {Key: "cc_id", Value: "1"}}
		h.DeleteChronicCondition(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func newDeleteChronicConditionRouter(svc ChronicConditionService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithChronicConditionSvc(svc)
	r.DELETE("/pets/:id/chronic-conditions/:cc_id", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteChronicCondition)
	return r
}
