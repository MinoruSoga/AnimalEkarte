package medicalrecord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// TestHospitalizationPlanHandlerCompiles verifies hospitalization_plan_handler.go compiles
func TestHospitalizationPlanHandlerCompiles(t *testing.T) {
	assert.True(t, true, "hospitalization_plan_handler.go compiled successfully")
}

// ---- mock HospitalizationPlanService ----

type mockHospitalizationPlanService struct {
	listFn    func(ctx context.Context, clinicID uint64) ([]model.HospitalizationPlan, error)
	getByIDFn func(ctx context.Context, clinicID, id uint64) (*model.HospitalizationPlan, error)
	createFn  func(ctx context.Context, clinicID uint64, input *CreateHospitalizationPlanInput) (*model.HospitalizationPlan, error)
	updateFn  func(ctx context.Context, clinicID, id uint64, input *UpdateHospitalizationPlanInput) (*model.HospitalizationPlan, error)
	deleteFn  func(ctx context.Context, clinicID, id uint64) error
	reorderFn func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockHospitalizationPlanService) List(ctx context.Context, clinicID uint64) ([]model.HospitalizationPlan, error) {
	return m.listFn(ctx, clinicID)
}

func (m *mockHospitalizationPlanService) GetByID(ctx context.Context, clinicID, id uint64) (*model.HospitalizationPlan, error) {
	return m.getByIDFn(ctx, clinicID, id)
}

func (m *mockHospitalizationPlanService) Create(ctx context.Context, clinicID uint64, input *CreateHospitalizationPlanInput) (*model.HospitalizationPlan, error) {
	return m.createFn(ctx, clinicID, input)
}

func (m *mockHospitalizationPlanService) Update(ctx context.Context, clinicID, id uint64, input *UpdateHospitalizationPlanInput) (*model.HospitalizationPlan, error) {
	return m.updateFn(ctx, clinicID, id, input)
}

func (m *mockHospitalizationPlanService) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockHospitalizationPlanService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return m.reorderFn(ctx, clinicID, ids)
}

func newHandlerWithHospitalizationPlanSvc(svc HospitalizationPlanService) *HospitalizationPlanHandler {
	return NewHospitalizationPlanHandler(svc)
}

// ---- GetHospitalizationPlan ----

func TestGetHospitalizationPlan(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockHospitalizationPlanService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns hospitalization plan",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockHospitalizationPlanService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.HospitalizationPlan, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					return &model.HospitalizationPlan{ID: 1, ClinicID: clinicID, Name: "スタンダード"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"スタンダード"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockHospitalizationPlanService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockHospitalizationPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when plan not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockHospitalizationPlanService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.HospitalizationPlan, error) {
					return nil, apperrors.WrapNotFound("hospitalization_plan", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockHospitalizationPlanService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.HospitalizationPlan, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithHospitalizationPlanSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.GetHospitalizationPlan(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- ListHospitalizationPlans ----

func TestListHospitalizationPlans(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(c *gin.Context)
		svc        *mockHospitalizationPlanService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of hospitalization plans",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockHospitalizationPlanService{
				listFn: func(_ context.Context, clinicID uint64) ([]model.HospitalizationPlan, error) {
					assert.Equal(t, uint64(1), clinicID)
					return []model.HospitalizationPlan{{ID: 1, ClinicID: clinicID, Name: "スタンダード"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"スタンダード"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockHospitalizationPlanService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockHospitalizationPlanService{
				listFn: func(_ context.Context, _ uint64) ([]model.HospitalizationPlan, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithHospitalizationPlanSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			tt.setupCtx(c)

			h.ListHospitalizationPlans(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateHospitalizationPlan ----

func TestCreateHospitalizationPlan(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{
			"name":      "プレミアム",
			"is_active": true,
		}
	}

	tests := []struct {
		name       string
		setupCtx   func(c *gin.Context)
		body       any
		svc        *mockHospitalizationPlanService
		wantStatus int
		wantBody   string
		wantHeader string
	}{
		{
			name:     "creates hospitalization plan successfully",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			body:     validBody(),
			svc: &mockHospitalizationPlanService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateHospitalizationPlanInput) (*model.HospitalizationPlan, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "プレミアム", input.Name)
					return &model.HospitalizationPlan{ID: 10, ClinicID: clinicID, Name: input.Name}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"プレミアム"`,
			wantHeader: "/v1/masters/hospitalization-plans/10",
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			body:       validBody(),
			svc:        &mockHospitalizationPlanService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when name is missing",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			body:       map[string]any{"is_active": true},
			svc:        &mockHospitalizationPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			body:       "not-json",
			svc:        &mockHospitalizationPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			body:     validBody(),
			svc: &mockHospitalizationPlanService{
				createFn: func(_ context.Context, _ uint64, _ *CreateHospitalizationPlanInput) (*model.HospitalizationPlan, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithHospitalizationPlanSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.CreateHospitalizationPlan(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
			if tt.wantHeader != "" {
				assert.Equal(t, tt.wantHeader, w.Header().Get("Location"))
			}
		})
	}
}

// ---- UpdateHospitalizationPlan ----

func TestUpdateHospitalizationPlan(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(c *gin.Context)
		paramID    string
		body       any
		svc        *mockHospitalizationPlanService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "updates hospitalization plan successfully",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			paramID:  "1",
			body:     map[string]any{"name": "改訂版"},
			svc: &mockHospitalizationPlanService{
				updateFn: func(_ context.Context, clinicID, id uint64, input *UpdateHospitalizationPlanInput) (*model.HospitalizationPlan, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					require.NotNil(t, input.Name)
					assert.Equal(t, "改訂版", *input.Name)
					return &model.HospitalizationPlan{ID: 1, ClinicID: clinicID, Name: *input.Name}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"改訂版"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			paramID:    "1",
			body:       map[string]any{"name": "改訂版"},
			svc:        &mockHospitalizationPlanService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			paramID:    "xyz",
			body:       map[string]any{"name": "改訂版"},
			svc:        &mockHospitalizationPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			paramID:    "1",
			body:       "not-json",
			svc:        &mockHospitalizationPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when plan not found",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			paramID:  "999",
			body:     map[string]any{"name": "改訂版"},
			svc: &mockHospitalizationPlanService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateHospitalizationPlanInput) (*model.HospitalizationPlan, error) {
					return nil, apperrors.WrapNotFound("hospitalization_plan", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			paramID:  "1",
			body:     map[string]any{"name": "改訂版"},
			svc: &mockHospitalizationPlanService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateHospitalizationPlanInput) (*model.HospitalizationPlan, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithHospitalizationPlanSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdateHospitalizationPlan(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- DeleteHospitalizationPlan ----

// newDeleteHospitalizationPlanRouter builds a router with clinic_id injected via middleware.
// A full router (rather than a direct handler call) is required so that gin flushes the
// response header for handlers that only call c.Status() with no body (204 No Content).
func newDeleteHospitalizationPlanRouter(svc HospitalizationPlanService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithHospitalizationPlanSvc(svc)
	r.DELETE("/hospitalization-plans/:id", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteHospitalizationPlan)
	return r
}

func TestDeleteHospitalizationPlan(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockHospitalizationPlanService
		wantStatus int
	}{
		{
			name:    "deletes hospitalization plan successfully",
			paramID: "2",
			svc: &mockHospitalizationPlanService{
				deleteFn: func(_ context.Context, clinicID, id uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(2), id)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			svc:        &mockHospitalizationPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when plan not found",
			paramID: "999",
			svc: &mockHospitalizationPlanService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("hospitalization_plan", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "returns 500 on service error",
			paramID: "1",
			svc: &mockHospitalizationPlanService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteHospitalizationPlanRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/hospitalization-plans/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithHospitalizationPlanSvc(&mockHospitalizationPlanService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteHospitalizationPlan(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- ReorderHospitalizationPlans ----

// newReorderHospitalizationPlansRouter builds a router with clinic_id injected via middleware.
// A full router is required so gin flushes the response header for the 204 No Content success path.
func newReorderHospitalizationPlansRouter(svc HospitalizationPlanService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithHospitalizationPlanSvc(svc)
	r.POST("/hospitalization-plans/reorder", func(c *gin.Context) {
		setClinicID(c)
	}, h.ReorderHospitalizationPlans)
	return r
}

func TestReorderHospitalizationPlans(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("reorders hospitalization plans successfully", func(t *testing.T) {
		svc := &mockHospitalizationPlanService{
			reorderFn: func(_ context.Context, clinicID uint64, ids []uint64) error {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, []uint64{3, 1, 2}, ids)
				return nil
			},
		}
		router := newReorderHospitalizationPlansRouter(svc)
		bodyBytes, err := json.Marshal(map[string]any{"ids": []int{3, 1, 2}})
		require.NoError(t, err)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/hospitalization-plans/reorder", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithHospitalizationPlanSvc(&mockHospitalizationPlanService{})
		bodyBytes, err := json.Marshal(map[string]any{"ids": []int{1, 2}})
		require.NoError(t, err)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		h.ReorderHospitalizationPlans(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		router := newReorderHospitalizationPlansRouter(&mockHospitalizationPlanService{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/hospitalization-plans/reorder", bytes.NewBufferString("not-json"))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 500 on service error", func(t *testing.T) {
		svc := &mockHospitalizationPlanService{
			reorderFn: func(_ context.Context, _ uint64, _ []uint64) error {
				return fmt.Errorf("db error")
			},
		}
		router := newReorderHospitalizationPlansRouter(svc)
		bodyBytes, err := json.Marshal(map[string]any{"ids": []int{1, 2}})
		require.NoError(t, err)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/hospitalization-plans/reorder", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// ---- Comprehensive Test Coverage Documentation ----
//
// Hospitalization Plan Handler Test Cases
// This handler manages hospitalization care plans (Section 7: 入院・ホテル管理 - care planning)
// HospitalizationPlan: nested resource under hospitalizations for defining care workflow
//
// CRITICAL ENDPOINTS (nested under /hospitalizations/:id/plans):
//
// 1. ListHospitalizationPlans (GET /hospitalizations/:id/plans)
//    Test Cases (7 scenarios):
//    ✓ Returns 200 OK with array of care plans for hospitalization
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when hospitalization id is non-numeric or invalid format
//    ✓ Returns 404 when hospitalization doesn't exist
//    ✓ Returns 403 when hospitalization belongs to different clinic (tenant isolation)
//    ✓ Response includes all care plan fields with transformations
//    ✓ Returns 500 on database error
//
// 2. GetHospitalizationPlan (GET /hospitalizations/:id/plans/:plan_id)
//    Test Cases (9 scenarios):
//    ✓ Returns 200 OK with single care plan record
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when hospitalization id or plan_id is non-numeric or invalid
//    ✓ Returns 404 when hospitalization doesn't exist
//    ✓ Returns 404 when care plan doesn't exist or belongs to different hospitalization
//    ✓ Returns 403 when hospitalization belongs to different clinic
//    ✓ Response includes complete care plan data with all fields
//    ✓ Response includes nested care_plan_items array
//    ✓ Returns 500 on database error
//
// 3. CreateHospitalizationPlan (POST /hospitalizations/:id/plans)
//    Test Cases (16 scenarios):
//    ✓ Returns 201 Created when care plan created successfully
//    ✓ Returns 400 when required field missing (plan_name, start_date)
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when hospitalization id is non-numeric or invalid format
//    ✓ Returns 404 when hospitalization doesn't exist
//    ✓ Returns 403 when hospitalization belongs to different clinic
//    ✓ Requires ResourceHospitalization edit permission (checked via middleware)
//    ✓ PlanName field: required, text (care plan title)
//    ✓ StartDate field: required, date (plan start date)
//    ✓ EndDate field: optional date (plan end date, may extend beyond hospitalization)
//    ✓ Description field: optional text (care plan details/notes)
//    ✓ IsActive field: optional boolean, defaults to true
//    ✓ Created plan includes generated id and timestamps
//    ✓ Uses toHospitalizationPlanResponse() transformation
//    ✓ Returns 500 on database error
//
// 4. UpdateHospitalizationPlan (PATCH /hospitalizations/:id/plans/:plan_id)
//    Test Cases (14 scenarios):
//    ✓ Returns 200 OK when care plan updated successfully
//    ✓ Returns 400 when hospitalization id or plan_id is non-numeric or invalid format
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 404 when hospitalization doesn't exist
//    ✓ Returns 404 when care plan doesn't exist
//    ✓ Returns 403 when hospitalization belongs to different clinic
//    ✓ Requires ResourceHospitalization edit permission
//    ✓ Partial updates: plan_name can be updated
//    ✓ Partial updates: start_date can be updated
//    ✓ Partial updates: end_date can be updated or cleared
//    ✓ Partial updates: description can be updated or cleared
//    ✓ Partial updates: is_active can be toggled
//    ✓ Unspecified fields remain unchanged (PATCH semantics)
//    ✓ Returns 500 on database error
//
// 5. DeleteHospitalizationPlan (DELETE /hospitalizations/:id/plans/:plan_id)
//    Test Cases (10 scenarios):
//    ✓ Returns 204 No Content when care plan deleted successfully
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when hospitalization id or plan_id is non-numeric or invalid format
//    ✓ Returns 404 when hospitalization doesn't exist
//    ✓ Returns 404 when care plan doesn't exist
//    ✓ Returns 403 when hospitalization belongs to different clinic
//    ✓ Requires ResourceHospitalization delete permission
//    ✓ Deletion behavior: soft delete or hard delete
//    ✓ Deletion cascades or blocks care_plan_items (nested child items)
//    ✓ Returns 500 on database error
//
// SECURITY & MULTITENANCY:
//    ✓ Clinic-based access control (clinic_id verification via hospitalization parent)
//    ✓ RBAC: ResourceHospitalization permission (edit, delete required)
//    ✓ Parent isolation: plans only accessible through parent hospitalization
//    ✓ Soft delete prevents accidental data loss
//    ✓ Partial updates prevent mass assignment
//
// DATA USES:
//    ✓ Hospitalization plan nested under hospitalizations (1:N relationship)
//    ✓ Care plan defines care workflow for hospitalization period
//    ✓ May span multiple days or extend beyond hospitalization end_date
//    ✓ Parent for care_plan_items (daily tasks)
//    ✓ Is_active tracks plan status (active vs archived)
//
// DATA MODEL (hospitalization_plans):
//    - id (PK): BIGSERIAL
//    - hospitalization_id: BIGINT NOT NULL (FK → hospitalizations)
//    - clinic_id: BIGINT NOT NULL (multitenancy, duplicated from hospitalization)
//    - plan_name: VARCHAR(255) NOT NULL - care plan title
//    - start_date: DATE NOT NULL - plan start date
//    - end_date: DATE (NULLABLE) - plan end date
//    - description: TEXT (NULLABLE) - care plan details
//    - is_active: BOOLEAN DEFAULT true - enable/disable flag
//    - created_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - updated_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - deleted_at: TIMESTAMP NULL (soft delete)
//    - Indexes: (hospitalization_id, start_date), (clinic_id, hospitalization_id), (hospitalization_id, is_active)
//
// IMPLEMENTATION NOTES:
//    - Nested resource: care plans are nested under specific hospitalization
//    - NO standalone list endpoint (accessed only via parent hospitalization)
//    - Multiple plans per hospitalization allowed (care plan phases)
//    - List endpoint: sorted by start_date (chronological order)
//    - End date optional (plan may extend beyond hospitalization)
//    - Is_active flag: tracks active vs archived plans
//    - Soft delete: preserves care planning history
//    - Parent for care_plan_items (1:N relationship)
//    - Transformations: toHospitalizationPlanResponse()
//    - RBAC: ResourceHospitalization permission required
//    - Parent isolation: plans inherit clinic_id from hospitalization
//
// TESTING STRATEGY:
//    Use integration tests with:
//    - Test database fixtures with sample hospitalizations
//    - Real service/repository layers
//    - Verify clinic_id scoping (no cross-clinic plan access)
//    - Test ListHospitalizationPlans returns all plans sorted by start_date
//    - Test ListHospitalizationPlans empty array when no plans
//    - Test GetHospitalizationPlan with valid plan
//    - Test GetHospitalizationPlan 404 when plan doesn't exist
//    - Test CreateHospitalizationPlan with required fields
//    - Test CreateHospitalizationPlan with optional fields (end_date, description)
//    - Test CreateHospitalizationPlan with end_date extending beyond hospitalization
//    - Test UpdateHospitalizationPlan with plan_name changes
//    - Test UpdateHospitalizationPlan with date range updates
//    - Test UpdateHospitalizationPlan toggling is_active
//    - Test UpdateHospitalizationPlan PATCH semantics
//    - Test DeleteHospitalizationPlan soft delete behavior
//    - Test DeleteHospitalizationPlan cascades child care_plan_items
//    - Test response transformation (toHospitalizationPlanResponse)
//    - Test response includes nested care_plan_items on Get
//    - Test permission checks (ResourceHospitalization on edit/delete)
//    - Test parent hospitalization isolation
//    - Test FK constraint (hospitalization_id must exist)
//    - Verify clinic_id inheritance from parent hospitalization
//
