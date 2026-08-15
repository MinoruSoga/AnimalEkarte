package trimming

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

// ---- mock TrimmingOptionService ----

type mockTrimmingOptionService struct {
	listFn    func(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error)
	getByIDFn func(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error)
	createFn  func(ctx context.Context, clinicID uint64, input *CreateTrimmingOptionInput) (*model.TrimmingOption, error)
	updateFn  func(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingOptionInput) (*model.TrimmingOption, error)
	deleteFn  func(ctx context.Context, clinicID, id uint64) error
	reorderFn func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockTrimmingOptionService) List(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error) {
	if m.listFn != nil {
		return m.listFn(ctx, clinicID)
	}
	return nil, nil
}

func (m *mockTrimmingOptionService) GetByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, clinicID, id)
	}
	return &model.TrimmingOption{ID: id}, nil
}

func (m *mockTrimmingOptionService) Create(ctx context.Context, clinicID uint64, input *CreateTrimmingOptionInput) (*model.TrimmingOption, error) {
	if m.createFn != nil {
		return m.createFn(ctx, clinicID, input)
	}
	return &model.TrimmingOption{ID: 1, Name: input.Name}, nil
}

func (m *mockTrimmingOptionService) Update(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingOptionInput) (*model.TrimmingOption, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, input)
	}
	return &model.TrimmingOption{ID: id}, nil
}

func (m *mockTrimmingOptionService) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockTrimmingOptionService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

// ---- helper ----

func newHandlerWithTrimmingOptionSvc(optionSvc TrimmingOptionService) *Handler {
	return &Handler{svc: &handlerServices{
		TrimmingCourse: &mockTrimmingCourseService{},
		TrimmingOption: optionSvc,
	}}
}

// ========== TrimmingOption tests ==========

// ---- ListTrimmingOptions ----

func TestListTrimmingOptions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(c *gin.Context)
		optionSvc  *mockTrimmingOptionService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of options",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			optionSvc: &mockTrimmingOptionService{
				listFn: func(_ context.Context, clinicID uint64) ([]model.TrimmingOption, error) {
					assert.Equal(t, uint64(1), clinicID)
					return []model.TrimmingOption{{ID: 1, Name: "爪切り"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"爪切り"`,
		},
		{
			name:       "returns 401 when clinic_id missing",
			setupCtx:   func(_ *gin.Context) {},
			optionSvc:  &mockTrimmingOptionService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			optionSvc: &mockTrimmingOptionService{
				listFn: func(_ context.Context, _ uint64) ([]model.TrimmingOption, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithTrimmingOptionSvc(tt.optionSvc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			tt.setupCtx(c)
			h.ListTrimmingOptions(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetTrimmingOption ----

func TestGetTrimmingOption(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		optionSvc  *mockTrimmingOptionService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns option for valid id",
			paramID:  "4",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			optionSvc: &mockTrimmingOptionService{
				getByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingOption, error) {
					return &model.TrimmingOption{ID: id, Name: "歯磨き"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"歯磨き"`,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			optionSvc:  &mockTrimmingOptionService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			optionSvc:  &mockTrimmingOptionService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			optionSvc: &mockTrimmingOptionService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.TrimmingOption, error) {
					return nil, apperrors.WrapNotFound("trimming_option", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithTrimmingOptionSvc(tt.optionSvc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.GetTrimmingOption(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateTrimmingOption ----

func TestCreateTrimmingOption(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{"name": "耳掃除"}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		optionSvc  *mockTrimmingOptionService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "creates option successfully with Location header",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			optionSvc: &mockTrimmingOptionService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateTrimmingOptionInput) (*model.TrimmingOption, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "耳掃除", input.Name)
					return &model.TrimmingOption{ID: 7, Name: input.Name}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"耳掃除"`,
		},
		{
			name:       "returns 401 when clinic_id missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			optionSvc:  &mockTrimmingOptionService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when name is missing",
			body:       map[string]any{"price": 500},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			optionSvc:  &mockTrimmingOptionService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			optionSvc: &mockTrimmingOptionService{
				createFn: func(_ context.Context, _ uint64, _ *CreateTrimmingOptionInput) (*model.TrimmingOption, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithTrimmingOptionSvc(tt.optionSvc)
			b, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)
			h.CreateTrimmingOption(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
			if tt.wantStatus == http.StatusCreated {
				assert.Contains(t, w.Header().Get("Location"), "/trimming-options/")
			}
		})
	}
}

// ---- UpdateTrimmingOption ----

func TestUpdateTrimmingOption(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		optionSvc  *mockTrimmingOptionService
		wantStatus int
	}{
		{
			name:     "updates option successfully",
			paramID:  "1",
			body:     map[string]any{"name": "更新オプション"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			optionSvc: &mockTrimmingOptionService{
				updateFn: func(_ context.Context, _, id uint64, input *UpdateTrimmingOptionInput) (*model.TrimmingOption, error) {
					require.NotNil(t, input.Name)
					assert.Equal(t, "更新オプション", *input.Name)
					return &model.TrimmingOption{ID: id}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "1",
			body:       map[string]any{"name": "test"},
			setupCtx:   func(_ *gin.Context) {},
			optionSvc:  &mockTrimmingOptionService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"name": "test"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			optionSvc:  &mockTrimmingOptionService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			body:     map[string]any{"name": "test"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			optionSvc: &mockTrimmingOptionService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateTrimmingOptionInput) (*model.TrimmingOption, error) {
					return nil, apperrors.WrapNotFound("trimming_option", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithTrimmingOptionSvc(tt.optionSvc)
			b, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(b))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.UpdateTrimmingOption(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- DeleteTrimmingOption ----

func newDeleteTrimmingOptionRouter(optionSvc TrimmingOptionService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithTrimmingOptionSvc(optionSvc)
	r.DELETE("/trimming-options/:id", func(c *gin.Context) { setClinicID(c) }, h.DeleteTrimmingOption)
	return r
}

func TestDeleteTrimmingOption(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		optionSvc  *mockTrimmingOptionService
		wantStatus int
	}{
		{
			name:    "deletes option successfully",
			paramID: "1",
			optionSvc: &mockTrimmingOptionService{
				deleteFn: func(_ context.Context, _, _ uint64) error { return nil },
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:    "returns 404 when not found",
			paramID: "999",
			optionSvc: &mockTrimmingOptionService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("trimming_option", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "returns 409 when in use",
			paramID: "2",
			optionSvc: &mockTrimmingOptionService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapConflict("このオプションは使用中のため削除できません")
				},
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteTrimmingOptionRouter(tt.optionSvc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/trimming-options/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id missing", func(t *testing.T) {
		h := newHandlerWithTrimmingOptionSvc(&mockTrimmingOptionService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteTrimmingOption(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- ReorderTrimmingOptions ----

func newReorderTrimmingOptionsRouter(optionSvc TrimmingOptionService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithTrimmingOptionSvc(optionSvc)
	r.POST("/trimming-options/reorder", func(c *gin.Context) { setClinicID(c) }, h.ReorderTrimmingOptions)
	return r
}

func TestReorderTrimmingOptions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("reorders successfully", func(t *testing.T) {
		svc := &mockTrimmingOptionService{
			reorderFn: func(_ context.Context, clinicID uint64, ids []uint64) error {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, []uint64{2, 3, 1}, ids)
				return nil
			},
		}
		router := newReorderTrimmingOptionsRouter(svc)
		body, _ := json.Marshal(map[string]any{"ids": []int{2, 3, 1}})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/trimming-options/reorder", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("returns 401 when clinic_id missing", func(t *testing.T) {
		h := newHandlerWithTrimmingOptionSvc(&mockTrimmingOptionService{})
		body, _ := json.Marshal(map[string]any{"ids": []int{1, 2}})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		h.ReorderTrimmingOptions(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
