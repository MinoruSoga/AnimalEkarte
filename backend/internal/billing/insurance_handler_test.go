package billing

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

// ---- mock InsuranceService ----

type mockInsuranceService struct {
	listFn    func(ctx context.Context, clinicID uint64) ([]model.Insurance, error)
	getByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Insurance, error)
	createFn  func(ctx context.Context, clinicID uint64, input *CreateInsuranceInput) (*model.Insurance, error)
	updateFn  func(ctx context.Context, clinicID, id uint64, input *UpdateInsuranceInput) (*model.Insurance, error)
	deleteFn  func(ctx context.Context, clinicID, id uint64) error
	reorderFn func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockInsuranceService) List(ctx context.Context, clinicID uint64) ([]model.Insurance, error) {
	return m.listFn(ctx, clinicID)
}
func (m *mockInsuranceService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Insurance, error) {
	return m.getByIDFn(ctx, clinicID, id)
}
func (m *mockInsuranceService) Create(ctx context.Context, clinicID uint64, input *CreateInsuranceInput) (*model.Insurance, error) {
	return m.createFn(ctx, clinicID, input)
}
func (m *mockInsuranceService) Update(ctx context.Context, clinicID, id uint64, input *UpdateInsuranceInput) (*model.Insurance, error) {
	return m.updateFn(ctx, clinicID, id, input)
}
func (m *mockInsuranceService) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}
func (m *mockInsuranceService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return m.reorderFn(ctx, clinicID, ids)
}

// ---- helper ----

func newHandlerWithInsuranceSvc(svc InsuranceService) *InsuranceHandler {
	return NewInsuranceHandler(svc)
}

// ---- ListInsurances ----

func TestListInsurances(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(c *gin.Context)
		svc        *mockInsuranceService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of insurances",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInsuranceService{
				listFn: func(_ context.Context, clinicID uint64) ([]model.Insurance, error) {
					assert.Equal(t, uint64(1), clinicID)
					return []model.Insurance{{ID: 1, Name: "アニコム"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"アニコム"`,
		},
		{
			name:     "returns empty list",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInsuranceService{
				listFn: func(_ context.Context, _ uint64) ([]model.Insurance, error) {
					return []model.Insurance{}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockInsuranceService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "returns 403 when selected clinic lacks master-insurance view grant",
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("clinic_id", "2")
				setResourcePermissionOnlyClinic(c, 1, string(model.ResourceMasterInsurance), "view")
			},
			svc: &mockInsuranceService{
				listFn: func(_ context.Context, _ uint64) ([]model.Insurance, error) {
					t.Fatal("insurance service must not be reached")
					return nil, nil
				},
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInsuranceService{
				listFn: func(_ context.Context, _ uint64) ([]model.Insurance, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithInsuranceSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			tt.setupCtx(c)

			h.ListInsurances(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetInsurance ----

func TestGetInsurance(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockInsuranceService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns insurance for valid id",
			paramID:  "2",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInsuranceService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Insurance, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(2), id)
					return &model.Insurance{ID: 2, Name: "イーペット保険"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"イーペット保険"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockInsuranceService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockInsuranceService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 403 when selected clinic lacks master-insurance view grant",
			paramID: "2",
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("clinic_id", "2")
				setResourcePermissionOnlyClinic(c, 1, string(model.ResourceMasterInsurance), "view")
			},
			svc: &mockInsuranceService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Insurance, error) {
					t.Fatal("insurance service must not be reached")
					return nil, nil
				},
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInsuranceService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Insurance, error) {
					return nil, apperrors.WrapNotFound("insurance", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithInsuranceSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.GetInsurance(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateInsurance ----

func TestCreateInsurance(t *testing.T) {
	gin.SetMode(gin.TestMode)

	coverageRate := 70
	validBody := func() map[string]any {
		return map[string]any{
			"name":          "ペット保険A",
			"is_active":     true,
			"coverage_rate": coverageRate,
		}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockInsuranceService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "creates insurance successfully",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInsuranceService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateInsuranceInput) (*model.Insurance, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "ペット保険A", input.Name)
					return &model.Insurance{ID: 1, Name: input.Name}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"ペット保険A"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockInsuranceService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when name is missing",
			body:       map[string]any{"is_active": true},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockInsuranceService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 409 on conflict",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInsuranceService{
				createFn: func(_ context.Context, _ uint64, _ *CreateInsuranceInput) (*model.Insurance, error) {
					return nil, apperrors.WrapAlreadyExists("insurance", "ペット保険A")
				},
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInsuranceService{
				createFn: func(_ context.Context, _ uint64, _ *CreateInsuranceInput) (*model.Insurance, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithInsuranceSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.CreateInsurance(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- UpdateInsurance ----

func TestUpdateInsurance(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockInsuranceService
		wantStatus int
	}{
		{
			name:     "updates insurance successfully",
			paramID:  "1",
			body:     map[string]any{"name": "更新保険"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInsuranceService{
				updateFn: func(_ context.Context, _, _ uint64, input *UpdateInsuranceInput) (*model.Insurance, error) {
					require.NotNil(t, input.Name)
					assert.Equal(t, "更新保険", *input.Name)
					return &model.Insurance{ID: 1, Name: *input.Name}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockInsuranceService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockInsuranceService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			body:     map[string]any{"name": "テスト"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInsuranceService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateInsuranceInput) (*model.Insurance, error) {
					return nil, apperrors.WrapNotFound("insurance", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithInsuranceSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdateInsurance(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- DeleteInsurance ----

func newDeleteInsuranceRouter(svc InsuranceService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithInsuranceSvc(svc)
	r.DELETE("/insurances/:id", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteInsurance)
	return r
}

func TestDeleteInsurance(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockInsuranceService
		wantStatus int
	}{
		{
			name:    "deletes insurance successfully",
			paramID: "1",
			svc: &mockInsuranceService{
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
			paramID:    "abc",
			svc:        &mockInsuranceService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when not found",
			paramID: "999",
			svc: &mockInsuranceService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("insurance", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "returns 409 when in use",
			paramID: "2",
			svc: &mockInsuranceService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapConflict("この保険はペット情報で使用中のため削除できません")
				},
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteInsuranceRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/insurances/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithInsuranceSvc(&mockInsuranceService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteInsurance(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- ReorderInsurances ----

func newReorderInsurancesRouter(svc InsuranceService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithInsuranceSvc(svc)
	r.PUT("/insurances/reorder", func(c *gin.Context) {
		setClinicID(c)
	}, h.ReorderInsurances)
	return r
}

func TestReorderInsurances(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("reorders insurances successfully", func(t *testing.T) {
		svc := &mockInsuranceService{
			reorderFn: func(_ context.Context, clinicID uint64, ids []uint64) error {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, []uint64{2, 1, 3}, ids)
				return nil
			},
		}
		router := newReorderInsurancesRouter(svc)
		body, _ := json.Marshal(map[string]any{"ids": []int{2, 1, 3}})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/insurances/reorder", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithInsuranceSvc(&mockInsuranceService{})
		body, _ := json.Marshal(map[string]any{"ids": []int{1, 2}})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		h.ReorderInsurances(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
