package staff_test

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
	staffdomain "github.com/animal-ekarte/backend/internal/staff"
)

// ---- mock ShiftTemplateService ----

type mockShiftTemplateService struct {
	listFn    func(ctx context.Context, clinicID uint64) ([]model.ShiftTemplate, error)
	getByIDFn func(ctx context.Context, clinicID, id uint64) (*model.ShiftTemplate, error)
	createFn  func(ctx context.Context, clinicID uint64, input *staffdomain.CreateShiftTemplateInput) (*model.ShiftTemplate, error)
	updateFn  func(ctx context.Context, clinicID, id uint64, input *staffdomain.UpdateShiftTemplateInput) (*model.ShiftTemplate, error)
	deleteFn  func(ctx context.Context, clinicID, id uint64) error
	reorderFn func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockShiftTemplateService) List(ctx context.Context, clinicID uint64) ([]model.ShiftTemplate, error) {
	if m.listFn != nil {
		return m.listFn(ctx, clinicID)
	}
	return nil, nil
}

func (m *mockShiftTemplateService) GetByID(ctx context.Context, clinicID, id uint64) (*model.ShiftTemplate, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockShiftTemplateService) Create(ctx context.Context, clinicID uint64, input *staffdomain.CreateShiftTemplateInput) (*model.ShiftTemplate, error) {
	if m.createFn != nil {
		return m.createFn(ctx, clinicID, input)
	}
	return nil, nil
}

func (m *mockShiftTemplateService) Update(ctx context.Context, clinicID, id uint64, input *staffdomain.UpdateShiftTemplateInput) (*model.ShiftTemplate, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, input)
	}
	return nil, nil
}

func (m *mockShiftTemplateService) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockShiftTemplateService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

// ---- test helper ----

func newHandlerWithShiftTemplateSvc(svc staffdomain.ShiftTemplateService) *staffdomain.Handler {
	return staffdomain.NewHandler(nil, nil, nil, nil, svc, nil)
}

// ---- ListShiftTemplates ----

func TestListShiftTemplates_ReturnsOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(c *gin.Context)
		svc        *mockShiftTemplateService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns 200 with shift template list",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockShiftTemplateService{
				listFn: func(_ context.Context, clinicID uint64) ([]model.ShiftTemplate, error) {
					assert.Equal(t, uint64(1), clinicID)
					return []model.ShiftTemplate{
						{ID: 1, ClinicID: 1, Name: "通常勤務", ShiftType: model.ShiftTypeFull, IsActive: true},
					}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"通常勤務"`,
		},
		{
			name:     "returns 200 with empty list when no templates",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockShiftTemplateService{
				listFn: func(_ context.Context, _ uint64) ([]model.ShiftTemplate, error) {
					return []model.ShiftTemplate{}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `[]`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockShiftTemplateService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockShiftTemplateService{
				listFn: func(_ context.Context, _ uint64) ([]model.ShiftTemplate, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithShiftTemplateSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			tt.setupCtx(c)

			h.ListShiftTemplates(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetShiftTemplate ----

func TestGetShiftTemplate_ReturnsOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockShiftTemplateService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns 200 for valid shift template id",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockShiftTemplateService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ShiftTemplate, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					return &model.ShiftTemplate{ID: 1, ClinicID: 1, Name: "午前勤務", ShiftType: model.ShiftTypeMorning, IsActive: true}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"午前勤務"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockShiftTemplateService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockShiftTemplateService{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithShiftTemplateSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.GetShiftTemplate(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetShiftTemplate NotFound ----

func TestGetShiftTemplate_NotFound_Returns404(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &mockShiftTemplateService{
		getByIDFn: func(_ context.Context, _, _ uint64) (*model.ShiftTemplate, error) {
			return nil, apperrors.WrapNotFound("shift_template", "999")
		},
	}
	h := newHandlerWithShiftTemplateSvc(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	setClinicID(c)

	h.GetShiftTemplate(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---- CreateShiftTemplate ----

func TestCreateShiftTemplate_Valid_Returns201(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		body         any
		setupCtx     func(c *gin.Context)
		svc          *mockShiftTemplateService
		wantStatus   int
		wantBody     string
		wantLocation string
	}{
		{
			name: "creates full shift template returns 201",
			body: map[string]any{
				"name":       "通常勤務テンプレート",
				"shift_type": "full",
				"start_time": "09:00",
				"end_time":   "18:00",
			},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockShiftTemplateService{
				createFn: func(_ context.Context, clinicID uint64, input *staffdomain.CreateShiftTemplateInput) (*model.ShiftTemplate, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "通常勤務テンプレート", input.Name)
					assert.Equal(t, "full", input.ShiftType)
					st := "09:00"
					et := "18:00"
					return &model.ShiftTemplate{
						ID:        10,
						ClinicID:  clinicID,
						Name:      input.Name,
						ShiftType: model.ShiftTypeFull,
						StartTime: &st,
						EndTime:   &et,
						IsActive:  true,
					}, nil
				},
			},
			wantStatus:   http.StatusCreated,
			wantBody:     `"name":"通常勤務テンプレート"`,
			wantLocation: "/api/v1/shift-templates/10",
		},
		{
			name: "creates off shift template (no start/end time) returns 201",
			body: map[string]any{
				"name":       "休み",
				"shift_type": "off",
			},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockShiftTemplateService{
				createFn: func(_ context.Context, clinicID uint64, input *staffdomain.CreateShiftTemplateInput) (*model.ShiftTemplate, error) {
					assert.Equal(t, "off", input.ShiftType)
					return &model.ShiftTemplate{ID: 11, ClinicID: clinicID, Name: input.Name, ShiftType: model.ShiftTypeOff, IsActive: true}, nil
				},
			},
			wantStatus:   http.StatusCreated,
			wantBody:     `"shift_type":"off"`,
			wantLocation: "/api/v1/shift-templates/11",
		},
		{
			name:       "returns 400 when name is missing",
			body:       map[string]any{"shift_type": "full"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockShiftTemplateService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when shift_type is missing",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockShiftTemplateService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid shift_type value",
			body:       map[string]any{"name": "テスト", "shift_type": "invalid_type"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockShiftTemplateService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       map[string]any{"name": "テスト", "shift_type": "full"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockShiftTemplateService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockShiftTemplateService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			body:     map[string]any{"name": "テスト", "shift_type": "full"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockShiftTemplateService{
				createFn: func(_ context.Context, _ uint64, _ *staffdomain.CreateShiftTemplateInput) (*model.ShiftTemplate, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithShiftTemplateSvc(tt.svc)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.CreateShiftTemplate(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
			if tt.wantLocation != "" {
				assert.Equal(t, tt.wantLocation, w.Header().Get("Location"))
			}
		})
	}
}

// ---- DeleteShiftTemplate ----

func newDeleteShiftTemplateRouter(svc staffdomain.ShiftTemplateService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithShiftTemplateSvc(svc)
	r.DELETE("/shift-templates/:id", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteShiftTemplate)
	return r
}

func TestDeleteShiftTemplate_Returns204(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockShiftTemplateService
		wantStatus int
	}{
		{
			name:    "deletes shift template successfully returns 204",
			paramID: "1",
			svc: &mockShiftTemplateService{
				deleteFn: func(_ context.Context, clinicID, id uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:    "returns 404 when shift template not found",
			paramID: "999",
			svc: &mockShiftTemplateService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("shift_template", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			svc:        &mockShiftTemplateService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 409 when shift template is in use",
			paramID: "2",
			svc: &mockShiftTemplateService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapConflict("このテンプレートはシフトエントリで使用中のため削除できません")
				},
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteShiftTemplateRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/shift-templates/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithShiftTemplateSvc(&mockShiftTemplateService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteShiftTemplate(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- UpdateShiftTemplate ----

func TestUpdateShiftTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockShiftTemplateService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "updates shift template name successfully returns 200",
			paramID:  "1",
			body:     map[string]any{"name": "更新勤務テンプレート"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockShiftTemplateService{
				updateFn: func(_ context.Context, clinicID, id uint64, input *staffdomain.UpdateShiftTemplateInput) (*model.ShiftTemplate, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					require.NotNil(t, input.Name)
					assert.Equal(t, "更新勤務テンプレート", *input.Name)
					return &model.ShiftTemplate{ID: 1, ClinicID: clinicID, Name: *input.Name, ShiftType: model.ShiftTypeFull, IsActive: true}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"更新勤務テンプレート"`,
		},
		{
			name:     "updates shift type to morning returns 200",
			paramID:  "2",
			body:     map[string]any{"shift_type": "morning"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockShiftTemplateService{
				updateFn: func(_ context.Context, _, _ uint64, input *staffdomain.UpdateShiftTemplateInput) (*model.ShiftTemplate, error) {
					require.NotNil(t, input.ShiftType)
					assert.Equal(t, model.ShiftTypeMorning, *input.ShiftType)
					return &model.ShiftTemplate{ID: 2, ClinicID: 1, Name: "午前", ShiftType: model.ShiftTypeMorning, IsActive: true}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"shift_type":"morning"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockShiftTemplateService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 404 when shift template not found",
			paramID:  "999",
			body:     map[string]any{"name": "テスト"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockShiftTemplateService{
				updateFn: func(_ context.Context, _, _ uint64, _ *staffdomain.UpdateShiftTemplateInput) (*model.ShiftTemplate, error) {
					return nil, apperrors.WrapNotFound("shift_template", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockShiftTemplateService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid shift_type value",
			paramID:    "1",
			body:       map[string]any{"shift_type": "invalid"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockShiftTemplateService{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithShiftTemplateSvc(tt.svc)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdateShiftTemplate(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- ReorderShiftTemplates ----

func newReorderShiftTemplatesRouter(svc staffdomain.ShiftTemplateService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithShiftTemplateSvc(svc)
	r.PATCH("/shift-templates/reorder", func(c *gin.Context) {
		setClinicID(c)
	}, h.ReorderShiftTemplates)
	return r
}

func TestReorderShiftTemplates(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("reorders shift templates successfully", func(t *testing.T) {
		svc := &mockShiftTemplateService{
			reorderFn: func(_ context.Context, clinicID uint64, ids []uint64) error {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, []uint64{3, 1, 2}, ids)
				return nil
			},
		}
		router := newReorderShiftTemplatesRouter(svc)
		bodyBytes, err := json.Marshal(map[string]any{"ids": []int{3, 1, 2}})
		require.NoError(t, err)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/shift-templates/reorder", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithShiftTemplateSvc(&mockShiftTemplateService{})
		bodyBytes, err := json.Marshal(map[string]any{"ids": []int{1, 2}})
		require.NoError(t, err)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		h.ReorderShiftTemplates(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		router := newReorderShiftTemplatesRouter(&mockShiftTemplateService{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/shift-templates/reorder", bytes.NewBufferString("not-json"))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
