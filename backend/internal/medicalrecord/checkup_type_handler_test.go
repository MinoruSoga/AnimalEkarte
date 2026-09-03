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

// ---- mock CheckupTypeService ----

type mockCheckupTypeService struct {
	listFn    func(ctx context.Context, clinicID uint64) ([]model.CheckupType, error)
	getByIDFn func(ctx context.Context, clinicID, id uint64) (*model.CheckupType, error)
	createFn  func(ctx context.Context, clinicID uint64, input *CreateCheckupTypeInput) (*model.CheckupType, error)
	updateFn  func(ctx context.Context, clinicID, id uint64, input *UpdateCheckupTypeInput) (*model.CheckupType, error)
	deleteFn  func(ctx context.Context, clinicID, id uint64) error
	reorderFn func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockCheckupTypeService) List(ctx context.Context, clinicID uint64) ([]model.CheckupType, error) {
	return m.listFn(ctx, clinicID)
}

func (m *mockCheckupTypeService) GetByID(ctx context.Context, clinicID, id uint64) (*model.CheckupType, error) {
	return m.getByIDFn(ctx, clinicID, id)
}

func (m *mockCheckupTypeService) Create(ctx context.Context, clinicID uint64, input *CreateCheckupTypeInput) (*model.CheckupType, error) {
	return m.createFn(ctx, clinicID, input)
}

func (m *mockCheckupTypeService) Update(ctx context.Context, clinicID, id uint64, input *UpdateCheckupTypeInput) (*model.CheckupType, error) {
	return m.updateFn(ctx, clinicID, id, input)
}

func (m *mockCheckupTypeService) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockCheckupTypeService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return m.reorderFn(ctx, clinicID, ids)
}

// ---- test helper ----

func newHandlerWithCheckupTypeSvc(svc CheckupTypeService) *CheckupTypeHandler {
	return NewCheckupTypeHandler(svc)
}

// ---- ListCheckupTypes ----

func TestListCheckupTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(c *gin.Context)
		svc        *mockCheckupTypeService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of checkup types",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupTypeService{
				listFn: func(_ context.Context, clinicID uint64) ([]model.CheckupType, error) {
					assert.Equal(t, uint64(1), clinicID)
					return []model.CheckupType{{ID: 1, ClinicID: clinicID, Name: "シニア健診"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"シニア健診"`,
		},
		{
			name:     "returns empty list when no types exist",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupTypeService{
				listFn: func(_ context.Context, _ uint64) ([]model.CheckupType, error) {
					return []model.CheckupType{}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `[]`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCheckupTypeService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupTypeService{
				listFn: func(_ context.Context, _ uint64) ([]model.CheckupType, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCheckupTypeSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			tt.setupCtx(c)

			h.ListCheckupTypes(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetCheckupType ----

func TestGetCheckupType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockCheckupTypeService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns checkup type for valid id",
			paramID:  "7",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupTypeService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.CheckupType, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(7), id)
					return &model.CheckupType{ID: 7, ClinicID: clinicID, Name: "予防接種後健診"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"予防接種後健診"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCheckupTypeService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when checkup type not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupTypeService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.CheckupType, error) {
					return nil, apperrors.WrapNotFound("checkup_type", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCheckupTypeSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.GetCheckupType(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateCheckupType ----

func TestCreateCheckupType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{
			"name":      "年次健診",
			"is_active": true,
		}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockCheckupTypeService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "creates checkup type successfully",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupTypeService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateCheckupTypeInput) (*model.CheckupType, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "年次健診", input.Name)
					return &model.CheckupType{ID: 5, ClinicID: clinicID, Name: input.Name}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"年次健診"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCheckupTypeService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when name is missing",
			body:       map[string]any{"is_active": true},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "creates with optional fields",
			body: map[string]any{
				"name":        "半年健診",
				"is_active":   true,
				"description": "半年ごとの定期健診",
				"interval":    "6_months",
				"target_age":  "senior",
				"sort_order":  2,
			},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupTypeService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateCheckupTypeInput) (*model.CheckupType, error) {
					assert.Equal(t, "半年健診", input.Name)
					assert.Equal(t, "6_months", input.Interval)
					assert.Equal(t, "senior", input.TargetAge)
					return &model.CheckupType{ID: 6, ClinicID: clinicID, Name: input.Name}, nil
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupTypeService{
				createFn: func(_ context.Context, _ uint64, _ *CreateCheckupTypeInput) (*model.CheckupType, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCheckupTypeSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.CreateCheckupType(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- UpdateCheckupType ----

func TestUpdateCheckupType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockCheckupTypeService
		wantStatus int
	}{
		{
			name:     "updates checkup type successfully",
			paramID:  "1",
			body:     map[string]any{"name": "更新健診"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupTypeService{
				updateFn: func(_ context.Context, _, _ uint64, input *UpdateCheckupTypeInput) (*model.CheckupType, error) {
					require.NotNil(t, input.Name)
					assert.Equal(t, "更新健診", *input.Name)
					return &model.CheckupType{ID: 1, Name: *input.Name}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "updates with clear_parent_id flag",
			paramID:  "2",
			body:     map[string]any{"clear_parent_id": true},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupTypeService{
				updateFn: func(_ context.Context, _, _ uint64, input *UpdateCheckupTypeInput) (*model.CheckupType, error) {
					assert.True(t, input.ClearParentID)
					return &model.CheckupType{ID: 2}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCheckupTypeService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when checkup type not found",
			paramID:  "999",
			body:     map[string]any{"name": "テスト"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupTypeService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateCheckupTypeInput) (*model.CheckupType, error) {
					return nil, apperrors.WrapNotFound("checkup_type", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCheckupTypeSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdateCheckupType(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- DeleteCheckupType ----

func newDeleteCheckupTypeRouter(svc CheckupTypeService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithCheckupTypeSvc(svc)
	r.DELETE("/checkup-types/:id", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteCheckupType)
	return r
}

func TestDeleteCheckupType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockCheckupTypeService
		wantStatus int
	}{
		{
			name:    "deletes checkup type successfully",
			paramID: "1",
			svc: &mockCheckupTypeService{
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
			svc:        &mockCheckupTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when checkup type not found",
			paramID: "999",
			svc: &mockCheckupTypeService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("checkup_type", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "returns 409 when checkup type is in use",
			paramID: "3",
			svc: &mockCheckupTypeService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapConflict("この定期健診種別は健診記録で使用中のため削除できません")
				},
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteCheckupTypeRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/checkup-types/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithCheckupTypeSvc(&mockCheckupTypeService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteCheckupType(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- ReorderCheckupTypes ----

func newReorderCheckupTypesRouter(svc CheckupTypeService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithCheckupTypeSvc(svc)
	r.POST("/checkup-types/reorder", func(c *gin.Context) {
		setClinicID(c)
	}, h.ReorderCheckupTypes)
	return r
}

func TestReorderCheckupTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("reorders checkup types successfully", func(t *testing.T) {
		svc := &mockCheckupTypeService{
			reorderFn: func(_ context.Context, clinicID uint64, ids []uint64) error {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, []uint64{2, 3, 1}, ids)
				return nil
			},
		}
		router := newReorderCheckupTypesRouter(svc)
		bodyBytes, err := json.Marshal(map[string]any{"ids": []int{2, 3, 1}})
		require.NoError(t, err)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/checkup-types/reorder", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithCheckupTypeSvc(&mockCheckupTypeService{})
		bodyBytes, err := json.Marshal(map[string]any{"ids": []int{1, 2}})
		require.NoError(t, err)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		h.ReorderCheckupTypes(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		router := newReorderCheckupTypesRouter(&mockCheckupTypeService{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/checkup-types/reorder", bytes.NewBufferString("not-json"))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// SEC-CODEX-UHQPM2 selected-clinic grant
func TestCheckupTypeSelectedClinicGrant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		invoke func(*CheckupTypeHandler, *gin.Context)
		svc    *mockCheckupTypeService
	}{
		{
			name: "ListCheckupTypes returns 403 when selected clinic lacks checkups view grant",
			invoke: func(h *CheckupTypeHandler, c *gin.Context) {
				h.ListCheckupTypes(c)
			},
			svc: &mockCheckupTypeService{
				listFn: func(_ context.Context, _ uint64) ([]model.CheckupType, error) {
					t.Fatal("service must not be reached")
					return nil, nil
				},
			},
		},
		{
			name: "GetCheckupType returns 403 when selected clinic lacks checkups view grant",
			invoke: func(h *CheckupTypeHandler, c *gin.Context) {
				h.GetCheckupType(c)
			},
			svc: &mockCheckupTypeService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.CheckupType, error) {
					t.Fatal("service must not be reached")
					return nil, nil
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCheckupTypeSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			setClinicID(c)
			c.Set("clinic_id", "2")
			c.Set("is_system_admin", false)
			setResourcePermissionOnlyClinic(c, 1, string(model.ResourceCheckups), "view")
			tt.invoke(h, c)
			assert.Equal(t, http.StatusForbidden, w.Code)
		})
	}
}
