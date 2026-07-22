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

// ---- mock TrimmingCourseService ----

type mockTrimmingCourseService struct {
	listFn    func(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error)
	getByIDFn func(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error)
	createFn  func(ctx context.Context, clinicID uint64, input *CreateTrimmingCourseInput) (*model.TrimmingCourse, error)
	updateFn  func(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingCourseInput) (*model.TrimmingCourse, error)
	deleteFn  func(ctx context.Context, clinicID, id uint64) error
	reorderFn func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockTrimmingCourseService) List(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error) {
	if m.listFn != nil {
		return m.listFn(ctx, clinicID)
	}
	return nil, nil
}

func (m *mockTrimmingCourseService) GetByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, clinicID, id)
	}
	return &model.TrimmingCourse{ID: id}, nil
}

func (m *mockTrimmingCourseService) Create(ctx context.Context, clinicID uint64, input *CreateTrimmingCourseInput) (*model.TrimmingCourse, error) {
	if m.createFn != nil {
		return m.createFn(ctx, clinicID, input)
	}
	return &model.TrimmingCourse{ID: 1, Name: input.Name}, nil
}

func (m *mockTrimmingCourseService) Update(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingCourseInput) (*model.TrimmingCourse, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, input)
	}
	return &model.TrimmingCourse{ID: id}, nil
}

func (m *mockTrimmingCourseService) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockTrimmingCourseService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

// ---- helper ----

func newHandlerWithTrimmingCourseSvc(courseSvc TrimmingCourseService) *Handler {
	return &Handler{svc: &handlerServices{
		TrimmingCourse: courseSvc,
		TrimmingOption: &mockTrimmingOptionService{},
	}}
}

// ========== TrimmingCourse tests ==========

// ---- ListTrimmingCourses ----

func TestListTrimmingCourses(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(c *gin.Context)
		courseSvc  *mockTrimmingCourseService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of courses",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			courseSvc: &mockTrimmingCourseService{
				listFn: func(_ context.Context, clinicID uint64) ([]model.TrimmingCourse, error) {
					assert.Equal(t, uint64(1), clinicID)
					return []model.TrimmingCourse{{ID: 1, Name: "スタンダードカット"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"スタンダードカット"`,
		},
		{
			name:       "returns 401 when clinic_id missing",
			setupCtx:   func(_ *gin.Context) {},
			courseSvc:  &mockTrimmingCourseService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			courseSvc: &mockTrimmingCourseService{
				listFn: func(_ context.Context, _ uint64) ([]model.TrimmingCourse, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithTrimmingCourseSvc(tt.courseSvc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			tt.setupCtx(c)
			h.ListTrimmingCourses(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetTrimmingCourse ----

func TestGetTrimmingCourse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		courseSvc  *mockTrimmingCourseService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns course for valid id",
			paramID:  "3",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			courseSvc: &mockTrimmingCourseService{
				getByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingCourse, error) {
					return &model.TrimmingCourse{ID: id, Name: "フルカット"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"フルカット"`,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			courseSvc:  &mockTrimmingCourseService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			courseSvc:  &mockTrimmingCourseService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			courseSvc: &mockTrimmingCourseService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.TrimmingCourse, error) {
					return nil, apperrors.WrapNotFound("trimming_course", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithTrimmingCourseSvc(tt.courseSvc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.GetTrimmingCourse(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateTrimmingCourse ----

func TestCreateTrimmingCourse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{"name": "シャンプーコース"}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		courseSvc  *mockTrimmingCourseService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "creates course successfully with Location header",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			courseSvc: &mockTrimmingCourseService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateTrimmingCourseInput) (*model.TrimmingCourse, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "シャンプーコース", input.Name)
					return &model.TrimmingCourse{ID: 5, Name: input.Name}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"シャンプーコース"`,
		},
		{
			name:       "returns 401 when clinic_id missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			courseSvc:  &mockTrimmingCourseService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when name is missing",
			body:       map[string]any{"price": 5000},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			courseSvc:  &mockTrimmingCourseService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			courseSvc: &mockTrimmingCourseService{
				createFn: func(_ context.Context, _ uint64, _ *CreateTrimmingCourseInput) (*model.TrimmingCourse, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithTrimmingCourseSvc(tt.courseSvc)
			b, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)
			h.CreateTrimmingCourse(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
			if tt.wantStatus == http.StatusCreated {
				assert.Contains(t, w.Header().Get("Location"), "/trimming-courses/")
			}
		})
	}
}

// ---- UpdateTrimmingCourse ----

func TestUpdateTrimmingCourse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		courseSvc  *mockTrimmingCourseService
		wantStatus int
	}{
		{
			name:     "updates course successfully",
			paramID:  "1",
			body:     map[string]any{"name": "更新コース"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			courseSvc: &mockTrimmingCourseService{
				updateFn: func(_ context.Context, _, id uint64, input *UpdateTrimmingCourseInput) (*model.TrimmingCourse, error) {
					require.NotNil(t, input.Name)
					assert.Equal(t, "更新コース", *input.Name)
					return &model.TrimmingCourse{ID: id}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "1",
			body:       map[string]any{"name": "test"},
			setupCtx:   func(_ *gin.Context) {},
			courseSvc:  &mockTrimmingCourseService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"name": "test"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			courseSvc:  &mockTrimmingCourseService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			body:     map[string]any{"name": "test"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			courseSvc: &mockTrimmingCourseService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateTrimmingCourseInput) (*model.TrimmingCourse, error) {
					return nil, apperrors.WrapNotFound("trimming_course", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithTrimmingCourseSvc(tt.courseSvc)
			b, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(b))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.UpdateTrimmingCourse(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- DeleteTrimmingCourse ----

func newDeleteTrimmingCourseRouter(courseSvc TrimmingCourseService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithTrimmingCourseSvc(courseSvc)
	r.DELETE("/trimming-courses/:id", func(c *gin.Context) { setClinicID(c) }, h.DeleteTrimmingCourse)
	return r
}

func TestDeleteTrimmingCourse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		courseSvc  *mockTrimmingCourseService
		wantStatus int
	}{
		{
			name:    "deletes course successfully",
			paramID: "1",
			courseSvc: &mockTrimmingCourseService{
				deleteFn: func(_ context.Context, _, _ uint64) error { return nil },
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:    "returns 404 when not found",
			paramID: "999",
			courseSvc: &mockTrimmingCourseService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("trimming_course", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "returns 409 when in use",
			paramID: "2",
			courseSvc: &mockTrimmingCourseService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapConflict("このコースは使用中のため削除できません")
				},
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteTrimmingCourseRouter(tt.courseSvc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/trimming-courses/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id missing", func(t *testing.T) {
		h := newHandlerWithTrimmingCourseSvc(&mockTrimmingCourseService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteTrimmingCourse(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- ReorderTrimmingCourses ----

func newReorderTrimmingCoursesRouter(courseSvc TrimmingCourseService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithTrimmingCourseSvc(courseSvc)
	r.POST("/trimming-courses/reorder", func(c *gin.Context) { setClinicID(c) }, h.ReorderTrimmingCourses)
	return r
}

func TestReorderTrimmingCourses(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("reorders successfully", func(t *testing.T) {
		svc := &mockTrimmingCourseService{
			reorderFn: func(_ context.Context, clinicID uint64, ids []uint64) error {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, []uint64{3, 1, 2}, ids)
				return nil
			},
		}
		router := newReorderTrimmingCoursesRouter(svc)
		body, _ := json.Marshal(map[string]any{"ids": []int{3, 1, 2}})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/trimming-courses/reorder", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("returns 401 when clinic_id missing", func(t *testing.T) {
		h := newHandlerWithTrimmingCourseSvc(&mockTrimmingCourseService{})
		body, _ := json.Marshal(map[string]any{"ids": []int{1, 2}})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		h.ReorderTrimmingCourses(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
