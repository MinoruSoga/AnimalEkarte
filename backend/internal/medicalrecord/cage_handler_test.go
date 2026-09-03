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

// ---- mock CageService ----

type mockCageService struct {
	listFn    func(ctx context.Context, clinicID uint64, cageType *string) ([]model.Cage, error)
	getByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Cage, error)
	createFn  func(ctx context.Context, clinicID uint64, input *CreateCageInput) (*model.Cage, error)
	updateFn  func(ctx context.Context, clinicID, id uint64, input *UpdateCageInput) (*model.Cage, error)
	deleteFn  func(ctx context.Context, clinicID, id uint64) error
	reorderFn func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockCageService) List(ctx context.Context, clinicID uint64, cageType *string) ([]model.Cage, error) {
	return m.listFn(ctx, clinicID, cageType)
}

func (m *mockCageService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Cage, error) {
	return m.getByIDFn(ctx, clinicID, id)
}

func (m *mockCageService) Create(ctx context.Context, clinicID uint64, input *CreateCageInput) (*model.Cage, error) {
	return m.createFn(ctx, clinicID, input)
}

func (m *mockCageService) Update(ctx context.Context, clinicID, id uint64, input *UpdateCageInput) (*model.Cage, error) {
	return m.updateFn(ctx, clinicID, id, input)
}

func (m *mockCageService) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockCageService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return m.reorderFn(ctx, clinicID, ids)
}

// ---- test helper ----

func newHandlerWithCageSvc(svc CageService) *CageHandler {
	return NewCageHandler(svc)
}

// ---- ListCages ----

func TestListCages(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockCageService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of cages",
			query:    "",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCageService{
				listFn: func(_ context.Context, clinicID uint64, cageType *string) ([]model.Cage, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Nil(t, cageType)
					return []model.Cage{{ID: 1, Name: "犬用ケージA"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"犬用ケージA"`,
		},
		{
			name:     "passes cage_type filter to service",
			query:    "cage_type=dog",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCageService{
				listFn: func(_ context.Context, _ uint64, cageType *string) ([]model.Cage, error) {
					require.NotNil(t, cageType)
					assert.Equal(t, "dog", *cageType)
					return []model.Cage{}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			query:      "",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCageService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			query:    "",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCageService{
				listFn: func(_ context.Context, _ uint64, _ *string) ([]model.Cage, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCageSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)

			h.ListCages(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetCage ----

func TestGetCage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockCageService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns cage for valid id",
			paramID:  "5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCageService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Cage, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), id)
					return &model.Cage{ID: 5, Name: "猫用ケージB", CageType: model.CageType("cat"), CageSize: model.CageSize("small")}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"猫用ケージB"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCageService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCageService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when cage not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCageService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Cage, error) {
					return nil, apperrors.WrapNotFound("cage", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCageSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.GetCage(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateCage ----

func TestCreateCage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{
			"name":      "ICUケージ1",
			"cage_type": "icu",
			"cage_size": "large",
			"is_active": true,
		}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockCageService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "creates cage successfully",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCageService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateCageInput) (*model.Cage, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "ICUケージ1", input.Name)
					return &model.Cage{
						ID:       10,
						ClinicID: clinicID,
						Name:     input.Name,
						CageType: model.CageType(input.CageType),
						CageSize: model.CageSize(input.CageSize),
					}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"ICUケージ1"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCageService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when name is missing",
			body:       map[string]any{"cage_type": "dog", "cage_size": "small"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCageService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid cage_type",
			body:       map[string]any{"name": "テスト", "cage_type": "invalid", "cage_size": "small"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCageService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCageService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCageService{
				createFn: func(_ context.Context, _ uint64, _ *CreateCageInput) (*model.Cage, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCageSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.CreateCage(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- UpdateCage ----

func TestUpdateCage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockCageService
		wantStatus int
	}{
		{
			name:     "updates cage successfully",
			paramID:  "1",
			body:     map[string]any{"name": "更新ケージ"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCageService{
				updateFn: func(_ context.Context, _, _ uint64, input *UpdateCageInput) (*model.Cage, error) {
					require.NotNil(t, input.Name)
					assert.Equal(t, "更新ケージ", *input.Name)
					return &model.Cage{ID: 1, Name: *input.Name}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCageService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCageService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when cage not found",
			paramID:  "999",
			body:     map[string]any{"name": "テスト"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCageService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateCageInput) (*model.Cage, error) {
					return nil, apperrors.WrapNotFound("cage", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCageSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdateCage(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- DeleteCage ----

func newDeleteCageRouter(svc CageService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithCageSvc(svc)
	r.DELETE("/cages/:id", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteCage)
	return r
}

func TestDeleteCage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockCageService
		wantStatus int
	}{
		{
			name:    "deletes cage successfully",
			paramID: "1",
			svc: &mockCageService{
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
			svc:        &mockCageService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when cage not found",
			paramID: "999",
			svc: &mockCageService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("cage", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "returns 409 when cage is in use",
			paramID: "2",
			svc: &mockCageService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapConflict("このケージは入院データで使用中のため削除できません")
				},
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteCageRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/cages/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithCageSvc(&mockCageService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteCage(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- ReorderCages ----

func newReorderCagesRouter(svc CageService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithCageSvc(svc)
	r.POST("/cages/reorder", func(c *gin.Context) {
		setClinicID(c)
	}, h.ReorderCages)
	return r
}

func TestReorderCages(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("reorders cages successfully", func(t *testing.T) {
		svc := &mockCageService{
			reorderFn: func(_ context.Context, clinicID uint64, ids []uint64) error {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, []uint64{3, 1, 2}, ids)
				return nil
			},
		}
		router := newReorderCagesRouter(svc)
		bodyBytes, err := json.Marshal(map[string]any{"ids": []int{3, 1, 2}})
		require.NoError(t, err)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/cages/reorder", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithCageSvc(&mockCageService{})
		bodyBytes, err := json.Marshal(map[string]any{"ids": []int{1, 2}})
		require.NoError(t, err)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		h.ReorderCages(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		router := newReorderCagesRouter(&mockCageService{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/cages/reorder", bytes.NewBufferString("not-json"))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// SEC-CODEX-UHQPM2 selected-clinic grant
func TestCageSelectedClinicGrant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		invoke func(*CageHandler, *gin.Context)
		svc    *mockCageService
	}{
		{
			name: "ListCages returns 403 when selected clinic lacks master-hospitalization view grant",
			invoke: func(h *CageHandler, c *gin.Context) {
				h.ListCages(c)
			},
			svc: &mockCageService{
				listFn: func(_ context.Context, _ uint64, _ *string) ([]model.Cage, error) {
					t.Fatal("service must not be reached")
					return nil, nil
				},
			},
		},
		{
			name: "GetCage returns 403 when selected clinic lacks master-hospitalization view grant",
			invoke: func(h *CageHandler, c *gin.Context) {
				h.GetCage(c)
			},
			svc: &mockCageService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Cage, error) {
					t.Fatal("service must not be reached")
					return nil, nil
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCageSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			setClinicID(c)
			c.Set("clinic_id", "2")
			c.Set("is_system_admin", false)
			setResourcePermissionOnlyClinic(c, 1, string(model.ResourceMasterHospitalization), "view")
			tt.invoke(h, c)
			assert.Equal(t, http.StatusForbidden, w.Code)
		})
	}
}
