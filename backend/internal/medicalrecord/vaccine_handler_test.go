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

// ---- mock VaccineService ----

type mockVaccineService struct {
	listFn    func(ctx context.Context, clinicID uint64, species *string) ([]model.Vaccine, error)
	getByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Vaccine, error)
	createFn  func(ctx context.Context, clinicID uint64, input *CreateVaccineInput) (*model.Vaccine, error)
	updateFn  func(ctx context.Context, clinicID, id uint64, input *UpdateVaccineInput) (*model.Vaccine, error)
	deleteFn  func(ctx context.Context, clinicID, id uint64) error
	reorderFn func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockVaccineService) List(ctx context.Context, clinicID uint64, species *string) ([]model.Vaccine, error) {
	return m.listFn(ctx, clinicID, species)
}
func (m *mockVaccineService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Vaccine, error) {
	return m.getByIDFn(ctx, clinicID, id)
}
func (m *mockVaccineService) Create(ctx context.Context, clinicID uint64, input *CreateVaccineInput) (*model.Vaccine, error) {
	return m.createFn(ctx, clinicID, input)
}
func (m *mockVaccineService) Update(ctx context.Context, clinicID, id uint64, input *UpdateVaccineInput) (*model.Vaccine, error) {
	return m.updateFn(ctx, clinicID, id, input)
}
func (m *mockVaccineService) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}
func (m *mockVaccineService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return m.reorderFn(ctx, clinicID, ids)
}

// ---- helper ----

func newHandlerWithVaccineSvc(svc VaccineService) *VaccineHandler {
	return NewVaccineHandler(svc)
}

// ---- ListVaccines ----

func TestListVaccines(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockVaccineService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of vaccines",
			query:    "",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockVaccineService{
				listFn: func(_ context.Context, clinicID uint64, species *string) ([]model.Vaccine, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Nil(t, species)
					return []model.Vaccine{{ID: 1, Name: "狂犬病ワクチン"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"狂犬病ワクチン"`,
		},
		{
			name:     "passes species filter to service",
			query:    "species=dog",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockVaccineService{
				listFn: func(_ context.Context, _ uint64, species *string) ([]model.Vaccine, error) {
					require.NotNil(t, species)
					assert.Equal(t, "dog", *species)
					return []model.Vaccine{{ID: 2, Name: "混合ワクチン"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"混合ワクチン"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			query:      "",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockVaccineService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			query:    "",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockVaccineService{
				listFn: func(_ context.Context, _ uint64, _ *string) ([]model.Vaccine, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithVaccineSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)

			h.ListVaccines(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetVaccine ----

func TestGetVaccine(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockVaccineService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns vaccine for valid id",
			paramID:  "6",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockVaccineService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Vaccine, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(6), id)
					return &model.Vaccine{ID: 6, Name: "猫3種混合ワクチン"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"猫3種混合ワクチン"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockVaccineService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockVaccineService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockVaccineService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccine, error) {
					return nil, apperrors.WrapNotFound("vaccine", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithVaccineSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.GetVaccine(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateVaccine ----

func TestCreateVaccine(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{
			"name":      "5種混合ワクチン",
			"is_active": true,
			"species":   "dog",
		}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockVaccineService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "creates vaccine successfully",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockVaccineService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateVaccineInput) (*model.Vaccine, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "5種混合ワクチン", input.Name)
					return &model.Vaccine{ID: 1, Name: input.Name}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"5種混合ワクチン"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockVaccineService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when name is missing",
			body:       map[string]any{"is_active": true},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockVaccineService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when species is invalid",
			body:       map[string]any{"name": "テスト", "species": "fish"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockVaccineService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 409 on conflict",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockVaccineService{
				createFn: func(_ context.Context, _ uint64, _ *CreateVaccineInput) (*model.Vaccine, error) {
					return nil, apperrors.WrapAlreadyExists("vaccine", "5種混合ワクチン")
				},
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockVaccineService{
				createFn: func(_ context.Context, _ uint64, _ *CreateVaccineInput) (*model.Vaccine, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithVaccineSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.CreateVaccine(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- UpdateVaccine ----

func TestUpdateVaccine(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockVaccineService
		wantStatus int
	}{
		{
			name:     "updates vaccine successfully",
			paramID:  "1",
			body:     map[string]any{"name": "更新ワクチン"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockVaccineService{
				updateFn: func(_ context.Context, _, _ uint64, input *UpdateVaccineInput) (*model.Vaccine, error) {
					require.NotNil(t, input.Name)
					assert.Equal(t, "更新ワクチン", *input.Name)
					return &model.Vaccine{ID: 1, Name: *input.Name}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockVaccineService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockVaccineService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			body:     map[string]any{"name": "テスト"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockVaccineService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateVaccineInput) (*model.Vaccine, error) {
					return nil, apperrors.WrapNotFound("vaccine", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithVaccineSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdateVaccine(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- DeleteVaccine ----

func newDeleteVaccineRouter(svc VaccineService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithVaccineSvc(svc)
	r.DELETE("/vaccines/:id", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteVaccine)
	return r
}

func TestDeleteVaccine(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockVaccineService
		wantStatus int
	}{
		{
			name:    "deletes vaccine successfully",
			paramID: "1",
			svc: &mockVaccineService{
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
			svc:        &mockVaccineService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when not found",
			paramID: "999",
			svc: &mockVaccineService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("vaccine", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "returns 409 when in use",
			paramID: "2",
			svc: &mockVaccineService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapConflict("このワクチンはワクチン接種記録で使用中のため削除できません")
				},
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteVaccineRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/vaccines/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithVaccineSvc(&mockVaccineService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteVaccine(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- ReorderVaccines ----

func newReorderVaccinesRouter(svc VaccineService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithVaccineSvc(svc)
	r.PUT("/vaccines/reorder", func(c *gin.Context) {
		setClinicID(c)
	}, h.ReorderVaccines)
	return r
}

func TestReorderVaccines(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("reorders vaccines successfully", func(t *testing.T) {
		svc := &mockVaccineService{
			reorderFn: func(_ context.Context, clinicID uint64, ids []uint64) error {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, []uint64{3, 1, 2}, ids)
				return nil
			},
		}
		router := newReorderVaccinesRouter(svc)
		body, _ := json.Marshal(map[string]any{"ids": []int{3, 1, 2}})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/vaccines/reorder", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithVaccineSvc(&mockVaccineService{})
		body, _ := json.Marshal(map[string]any{"ids": []int{1, 2}})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		h.ReorderVaccines(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// SEC-CODEX-UHQPM2 selected-clinic grant
func TestVaccineSelectedClinicGrant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		invoke func(*VaccineHandler, *gin.Context)
		svc    *mockVaccineService
	}{
		{
			name: "ListVaccines returns 403 when selected clinic lacks master-medical view grant",
			invoke: func(h *VaccineHandler, c *gin.Context) {
				h.ListVaccines(c)
			},
			svc: &mockVaccineService{
				listFn: func(_ context.Context, _ uint64, _ *string) ([]model.Vaccine, error) {
					t.Fatal("service must not be reached")
					return nil, nil
				},
			},
		},
		{
			name: "GetVaccine returns 403 when selected clinic lacks master-medical view grant",
			invoke: func(h *VaccineHandler, c *gin.Context) {
				h.GetVaccine(c)
			},
			svc: &mockVaccineService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccine, error) {
					t.Fatal("service must not be reached")
					return nil, nil
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithVaccineSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			setClinicID(c)
			c.Set("clinic_id", "2")
			c.Set("is_system_admin", false)
			setResourcePermissionOnlyClinic(c, 1, string(model.ResourceMasterMedical), "view")
			tt.invoke(h, c)
			assert.Equal(t, http.StatusForbidden, w.Code)
		})
	}
}
