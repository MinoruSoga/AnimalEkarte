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

// ---- mock ExamTypeService ----

type mockExamTypeService struct {
	listFn                   func(ctx context.Context, clinicID uint64) ([]model.ExaminationType, error)
	getByIDFn                func(ctx context.Context, clinicID, id uint64) (*model.ExaminationType, error)
	createFn                 func(ctx context.Context, clinicID uint64, input *CreateExamTypeInput) (*model.ExaminationType, error)
	updateFn                 func(ctx context.Context, clinicID, id uint64, input *UpdateExamTypeInput) (*model.ExaminationType, error)
	deleteFn                 func(ctx context.Context, clinicID, id uint64) error
	reorderFn                func(ctx context.Context, clinicID uint64, ids []uint64) error
	updateFieldFn            func(ctx context.Context, clinicID, examTypeID, fieldID uint64, input *UpdateExamTypeFieldInput) (*ExamTypeFieldResult, error)
	replaceReferenceRangesFn func(ctx context.Context, clinicID, examTypeID uint64, command *ReplaceReferenceRangesCommand) (*ExamTypeFieldResult, error)
}

func (m *mockExamTypeService) List(ctx context.Context, clinicID uint64) ([]model.ExaminationType, error) {
	return m.listFn(ctx, clinicID)
}
func (m *mockExamTypeService) GetByID(ctx context.Context, clinicID, id uint64) (*model.ExaminationType, error) {
	return m.getByIDFn(ctx, clinicID, id)
}
func (m *mockExamTypeService) Create(ctx context.Context, clinicID uint64, input *CreateExamTypeInput) (*model.ExaminationType, error) {
	return m.createFn(ctx, clinicID, input)
}
func (m *mockExamTypeService) Update(ctx context.Context, clinicID, id uint64, input *UpdateExamTypeInput) (*model.ExaminationType, error) {
	return m.updateFn(ctx, clinicID, id, input)
}
func (m *mockExamTypeService) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}
func (m *mockExamTypeService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return m.reorderFn(ctx, clinicID, ids)
}
func (m *mockExamTypeService) CreateField(context.Context, uint64, *CreateExamTypeFieldCommand) (*model.ExamTypeField, error) {
	return &model.ExamTypeField{}, nil
}
func (m *mockExamTypeService) UpdateField(ctx context.Context, clinicID, examTypeID, fieldID uint64, input *UpdateExamTypeFieldInput) (*ExamTypeFieldResult, error) {
	if m.updateFieldFn != nil {
		return m.updateFieldFn(ctx, clinicID, examTypeID, fieldID, input)
	}
	return &ExamTypeFieldResult{}, nil
}
func (m *mockExamTypeService) DeleteField(context.Context, uint64, uint64, uint64) error {
	return nil
}
func (m *mockExamTypeService) ReorderFields(context.Context, uint64, uint64, []uint64) error {
	return nil
}
func (m *mockExamTypeService) ReplaceReferenceRanges(ctx context.Context, clinicID, examTypeID uint64, command *ReplaceReferenceRangesCommand) (*ExamTypeFieldResult, error) {
	if m.replaceReferenceRangesFn != nil {
		return m.replaceReferenceRangesFn(ctx, clinicID, examTypeID, command)
	}
	return &ExamTypeFieldResult{}, nil
}
func (m *mockExamTypeService) ListReferenceRanges(context.Context, uint64, []uint64) (map[uint64][]model.ExamReferenceRange, error) {
	return map[uint64][]model.ExamReferenceRange{}, nil
}

// ---- helper ----

func newHandlerWithExamTypeSvc(svc ExaminationTypeService) *ExamTypeHandler {
	return NewExamTypeHandler(svc)
}

// ---- ListExaminationTypes ----

func TestListExaminationTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(c *gin.Context)
		svc        *mockExamTypeService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of exam types",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockExamTypeService{
				listFn: func(_ context.Context, clinicID uint64) ([]model.ExaminationType, error) {
					assert.Equal(t, uint64(1), clinicID)
					return []model.ExaminationType{{ID: 1, Name: "血液検査"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"血液検査"`,
		},
		{
			name:     "returns empty list",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockExamTypeService{
				listFn: func(_ context.Context, _ uint64) ([]model.ExaminationType, error) {
					return []model.ExaminationType{}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockExamTypeService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockExamTypeService{
				listFn: func(_ context.Context, _ uint64) ([]model.ExaminationType, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithExamTypeSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			tt.setupCtx(c)

			h.ListExaminationTypes(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetExaminationType ----

func TestGetExaminationType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockExamTypeService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns exam type for valid id",
			paramID:  "5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockExamTypeService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ExaminationType, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), id)
					return &model.ExaminationType{ID: 5, Name: "X線検査"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"X線検査"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockExamTypeService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockExamTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockExamTypeService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.ExaminationType, error) {
					return nil, apperrors.WrapNotFound("exam_type", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithExamTypeSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.GetExaminationType(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateExaminationType ----

func TestCreateExaminationType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{
			"name":      "超音波検査",
			"is_active": true,
		}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockExamTypeService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "creates exam type successfully",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockExamTypeService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateExamTypeInput) (*model.ExaminationType, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "超音波検査", input.Name)
					return &model.ExaminationType{ID: 1, Name: input.Name}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"超音波検査"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockExamTypeService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when name is missing",
			body:       map[string]any{"is_active": true},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockExamTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 409 on conflict",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockExamTypeService{
				createFn: func(_ context.Context, _ uint64, _ *CreateExamTypeInput) (*model.ExaminationType, error) {
					return nil, apperrors.WrapAlreadyExists("exam_type", "超音波検査")
				},
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockExamTypeService{
				createFn: func(_ context.Context, _ uint64, _ *CreateExamTypeInput) (*model.ExaminationType, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithExamTypeSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.CreateExaminationType(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- UpdateExaminationType ----

func TestUpdateExaminationType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockExamTypeService
		wantStatus int
	}{
		{
			name:     "updates exam type successfully",
			paramID:  "1",
			body:     map[string]any{"name": "更新検査"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockExamTypeService{
				updateFn: func(_ context.Context, _, _ uint64, input *UpdateExamTypeInput) (*model.ExaminationType, error) {
					require.NotNil(t, input.Name)
					assert.Equal(t, "更新検査", *input.Name)
					return &model.ExaminationType{ID: 1, Name: *input.Name}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockExamTypeService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockExamTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			body:     map[string]any{"name": "テスト"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockExamTypeService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateExamTypeInput) (*model.ExaminationType, error) {
					return nil, apperrors.WrapNotFound("exam_type", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithExamTypeSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdateExaminationType(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- DeleteExaminationType ----

func newDeleteExamTypeRouter(svc ExaminationTypeService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithExamTypeSvc(svc)
	r.DELETE("/exam-types/:id", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteExaminationType)
	return r
}

func TestDeleteExaminationType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockExamTypeService
		wantStatus int
	}{
		{
			name:    "deletes exam type successfully",
			paramID: "1",
			svc: &mockExamTypeService{
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
			svc:        &mockExamTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when not found",
			paramID: "999",
			svc: &mockExamTypeService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("exam_type", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "returns 409 when in use",
			paramID: "2",
			svc: &mockExamTypeService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapConflict("この検査種別は検査記録で使用中のため削除できません")
				},
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteExamTypeRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/exam-types/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithExamTypeSvc(&mockExamTypeService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteExaminationType(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- ReorderExaminationTypes ----

func newReorderExamTypesRouter(svc ExaminationTypeService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithExamTypeSvc(svc)
	r.PUT("/exam-types/reorder", func(c *gin.Context) {
		setClinicID(c)
	}, h.ReorderExaminationTypes)
	return r
}

func TestReorderExaminationTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("reorders exam types successfully", func(t *testing.T) {
		svc := &mockExamTypeService{
			reorderFn: func(_ context.Context, clinicID uint64, ids []uint64) error {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, []uint64{2, 1, 3}, ids)
				return nil
			},
		}
		router := newReorderExamTypesRouter(svc)
		body, _ := json.Marshal(map[string]any{"ids": []int{2, 1, 3}})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/exam-types/reorder", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithExamTypeSvc(&mockExamTypeService{})
		body, _ := json.Marshal(map[string]any{"ids": []int{1, 2}})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		h.ReorderExaminationTypes(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
