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

// ---- mock TrimmingCourseTypeService ----

type mockTrimmingCourseTypeService struct {
	listFn    func(ctx context.Context, clinicID uint64) ([]model.TrimmingCourseType, error)
	getByIDFn func(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourseType, error)
	createFn  func(ctx context.Context, clinicID uint64, input *CreateTrimmingCourseTypeInput) (*model.TrimmingCourseType, error)
	updateFn  func(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingCourseTypeInput) (*model.TrimmingCourseType, error)
	deleteFn  func(ctx context.Context, clinicID, id uint64) error
	reorderFn func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockTrimmingCourseTypeService) List(ctx context.Context, clinicID uint64) ([]model.TrimmingCourseType, error) {
	if m.listFn != nil {
		return m.listFn(ctx, clinicID)
	}
	return nil, nil
}

func (m *mockTrimmingCourseTypeService) GetByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourseType, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockTrimmingCourseTypeService) Create(ctx context.Context, clinicID uint64, input *CreateTrimmingCourseTypeInput) (*model.TrimmingCourseType, error) {
	if m.createFn != nil {
		return m.createFn(ctx, clinicID, input)
	}
	return nil, nil
}

func (m *mockTrimmingCourseTypeService) Update(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingCourseTypeInput) (*model.TrimmingCourseType, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, input)
	}
	return nil, nil
}

func (m *mockTrimmingCourseTypeService) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockTrimmingCourseTypeService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

func newHandlerWithTrimmingCourseTypeSvc(svc TrimmingCourseTypeService) *Handler {
	return &Handler{svc: &handlerServices{TrimmingCourseType: svc}}
}

// ---- ListTrimmingCourseTypes ----

func TestListTrimmingCourseTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(c *gin.Context)
		svc        *mockTrimmingCourseTypeService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns 200 with list",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockTrimmingCourseTypeService{
				listFn: func(_ context.Context, clinicID uint64) ([]model.TrimmingCourseType, error) {
					assert.Equal(t, uint64(1), clinicID)
					return []model.TrimmingCourseType{{ID: 1, Name: "フルコース"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"フルコース"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockTrimmingCourseTypeService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockTrimmingCourseTypeService{
				listFn: func(_ context.Context, _ uint64) ([]model.TrimmingCourseType, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithTrimmingCourseTypeSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			tt.setupCtx(c)

			h.ListTrimmingCourseTypes(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetTrimmingCourseType ----

func TestGetTrimmingCourseType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockTrimmingCourseTypeService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns 200 with item",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockTrimmingCourseTypeService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.TrimmingCourseType, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					return &model.TrimmingCourseType{ID: 1, Name: "シャンプーコース"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"シャンプーコース"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockTrimmingCourseTypeService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockTrimmingCourseTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockTrimmingCourseTypeService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.TrimmingCourseType, error) {
					return nil, apperrors.WrapNotFound("trimming_course_type", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithTrimmingCourseTypeSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.GetTrimmingCourseType(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateTrimmingCourseType ----

func TestCreateTrimmingCourseType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockTrimmingCourseTypeService
		wantStatus int
		wantBody   string
		wantHeader bool
	}{
		{
			name:     "creates and returns 201 with Location header",
			body:     map[string]any{"name": "爪切りコース", "sort_order": 2},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockTrimmingCourseTypeService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateTrimmingCourseTypeInput) (*model.TrimmingCourseType, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "爪切りコース", input.Name)
					assert.Equal(t, 2, input.SortOrder)
					return &model.TrimmingCourseType{ID: 5, Name: input.Name, SortOrder: input.SortOrder}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"爪切りコース"`,
			wantHeader: true,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockTrimmingCourseTypeService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when name is missing",
			body:       map[string]any{"sort_order": 1},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockTrimmingCourseTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			body:     map[string]any{"name": "エラーコース"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockTrimmingCourseTypeService{
				createFn: func(_ context.Context, _ uint64, _ *CreateTrimmingCourseTypeInput) (*model.TrimmingCourseType, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithTrimmingCourseTypeSvc(tt.svc)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.CreateTrimmingCourseType(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
			if tt.wantHeader {
				assert.Equal(t, "/v1/masters/trimming-course-types/5", w.Header().Get("Location"))
			}
		})
	}
}

// ---- UpdateTrimmingCourseType ----

func TestUpdateTrimmingCourseType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockTrimmingCourseTypeService
		wantStatus int
	}{
		{
			name:     "updates successfully and returns 200",
			paramID:  "3",
			body:     map[string]any{"name": "更新後コース"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockTrimmingCourseTypeService{
				updateFn: func(_ context.Context, clinicID, id uint64, input *UpdateTrimmingCourseTypeInput) (*model.TrimmingCourseType, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(3), id)
					require.NotNil(t, input.Name)
					assert.Equal(t, "更新後コース", *input.Name)
					return &model.TrimmingCourseType{ID: 3, Name: *input.Name}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "3",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockTrimmingCourseTypeService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockTrimmingCourseTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON body",
			paramID:    "3",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockTrimmingCourseTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			body:     map[string]any{"name": "テスト"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockTrimmingCourseTypeService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateTrimmingCourseTypeInput) (*model.TrimmingCourseType, error) {
					return nil, apperrors.WrapNotFound("trimming_course_type", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithTrimmingCourseTypeSvc(tt.svc)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdateTrimmingCourseType(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- ReorderTrimmingCourseTypes ----

func TestReorderTrimmingCourseTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockTrimmingCourseTypeService
		wantStatus int
	}{
		{
			name:     "reorders successfully and returns 204",
			body:     map[string]any{"ids": []uint64{3, 1, 2}},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockTrimmingCourseTypeService{
				reorderFn: func(_ context.Context, clinicID uint64, ids []uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, []uint64{3, 1, 2}, ids)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       map[string]any{"ids": []uint64{1}},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockTrimmingCourseTypeService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when ids is empty",
			body:       map[string]any{"ids": []uint64{}},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockTrimmingCourseTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			body:     map[string]any{"ids": []uint64{1, 2}},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockTrimmingCourseTypeService{
				reorderFn: func(_ context.Context, _ uint64, _ []uint64) error {
					return fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithTrimmingCourseTypeSvc(tt.svc)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.ReorderTrimmingCourseTypes(c)
			c.Writer.WriteHeaderNow() // flush a bare c.Status() (no body) to the recorder

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- DeleteTrimmingCourseType ----

func TestDeleteTrimmingCourseType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockTrimmingCourseTypeService
		wantStatus int
	}{
		{
			name:     "deletes successfully and returns 204",
			paramID:  "4",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockTrimmingCourseTypeService{
				deleteFn: func(_ context.Context, clinicID, id uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(4), id)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "4",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockTrimmingCourseTypeService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockTrimmingCourseTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 409 when in use",
			paramID:  "4",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockTrimmingCourseTypeService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapConflict("このコースは使用中のため削除できません")
				},
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithTrimmingCourseTypeSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.DeleteTrimmingCourseType(c)
			c.Writer.WriteHeaderNow() // flush a bare c.Status() (no body) to the recorder

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
