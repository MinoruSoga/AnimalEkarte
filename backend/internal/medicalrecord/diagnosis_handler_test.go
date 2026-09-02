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
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// setClinicID は gin.Context に clinic_id を設定するヘルパー（internal/handler/owner_handler_test.go
// 由来・同一実装。medicalrecord パッケージのハンドラテスト全体で共有）。
func setClinicID(c *gin.Context) {
	c.Set("clinic_id", "1")
	httpapi.SetClinicPermissionChecker(c, func(_ *gin.Context, _ uint64, _, _ string) bool {
		return true
	})
}

func setResourcePermissionOnlyClinic(c *gin.Context, clinicID uint64, resource, action string) {
	httpapi.SetClinicPermissionChecker(c, func(_ *gin.Context, id uint64, res, act string) bool {
		return id == clinicID && res == resource && act == action
	})
}

// ---- mock DiagnosisTypeService ----

type mockDiagnosisTypeService struct {
	listFn    func(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisType, int64, error)
	getByIDFn func(ctx context.Context, clinicID, id uint64) (*model.DiagnosisType, error)
	createFn  func(ctx context.Context, clinicID uint64, input *CreateDiagnosisTypeInput) (*model.DiagnosisType, error)
	updateFn  func(ctx context.Context, clinicID, id uint64, input *UpdateDiagnosisTypeInput) (*model.DiagnosisType, error)
	deleteFn  func(ctx context.Context, clinicID, id uint64) error
	reorderFn func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockDiagnosisTypeService) List(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisType, int64, error) {
	return m.listFn(ctx, clinicID, page, limit)
}
func (m *mockDiagnosisTypeService) GetByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisType, error) {
	return m.getByIDFn(ctx, clinicID, id)
}
func (m *mockDiagnosisTypeService) Create(ctx context.Context, clinicID uint64, input *CreateDiagnosisTypeInput) (*model.DiagnosisType, error) {
	return m.createFn(ctx, clinicID, input)
}
func (m *mockDiagnosisTypeService) Update(ctx context.Context, clinicID, id uint64, input *UpdateDiagnosisTypeInput) (*model.DiagnosisType, error) {
	return m.updateFn(ctx, clinicID, id, input)
}
func (m *mockDiagnosisTypeService) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}
func (m *mockDiagnosisTypeService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return m.reorderFn(ctx, clinicID, ids)
}

// ---- mock DiagnosisNameService ----

type mockDiagnosisNameService struct {
	listFn      func(ctx context.Context, clinicID uint64, typeID *uint64, page, limit int) ([]model.DiagnosisName, int64, error)
	listNamesFn func(ctx context.Context, clinicID uint64, typeID *uint64) ([]model.DiagnosisName, error)
	getByIDFn   func(ctx context.Context, clinicID, id uint64) (*model.DiagnosisName, error)
	createFn    func(ctx context.Context, clinicID uint64, input *CreateDiagnosisNameInput) (*model.DiagnosisName, error)
	updateFn    func(ctx context.Context, clinicID, id uint64, input *UpdateDiagnosisNameInput) (*model.DiagnosisName, error)
	deleteFn    func(ctx context.Context, clinicID, id uint64) error
	reorderFn   func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockDiagnosisNameService) List(ctx context.Context, clinicID uint64, typeID *uint64, page, limit int) ([]model.DiagnosisName, int64, error) {
	return m.listFn(ctx, clinicID, typeID, page, limit)
}
func (m *mockDiagnosisNameService) ListNames(ctx context.Context, clinicID uint64, typeID *uint64) ([]model.DiagnosisName, error) {
	if m.listNamesFn != nil {
		return m.listNamesFn(ctx, clinicID, typeID)
	}
	return nil, nil
}
func (m *mockDiagnosisNameService) GetByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisName, error) {
	return m.getByIDFn(ctx, clinicID, id)
}
func (m *mockDiagnosisNameService) Create(ctx context.Context, clinicID uint64, input *CreateDiagnosisNameInput) (*model.DiagnosisName, error) {
	return m.createFn(ctx, clinicID, input)
}
func (m *mockDiagnosisNameService) Update(ctx context.Context, clinicID, id uint64, input *UpdateDiagnosisNameInput) (*model.DiagnosisName, error) {
	return m.updateFn(ctx, clinicID, id, input)
}
func (m *mockDiagnosisNameService) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}
func (m *mockDiagnosisNameService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return m.reorderFn(ctx, clinicID, ids)
}

// ---- helper ----

func newHandlerWithDiagnosisSvc(typeSvc DiagnosisTypeService, nameSvc DiagnosisNameService) *DiagnosisHandler {
	return NewDiagnosisHandler(typeSvc, nameSvc)
}

// ---- ListDiagnosisTypes ----

func TestListDiagnosisTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		typeSvc    *mockDiagnosisTypeService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns paginated diagnosis types",
			query:    "page=1&limit=10",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			typeSvc: &mockDiagnosisTypeService{
				listFn: func(_ context.Context, _ uint64, _, _ int) ([]model.DiagnosisType, int64, error) {
					return []model.DiagnosisType{{ID: 1, Name: "皮膚疾患"}}, 1, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"皮膚疾患"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			query:      "page=1&limit=10",
			setupCtx:   func(_ *gin.Context) {},
			typeSvc:    &mockDiagnosisTypeService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid pagination",
			query:      "page=0",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			typeSvc:    &mockDiagnosisTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			query:    "page=1&limit=10",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			typeSvc: &mockDiagnosisTypeService{
				listFn: func(_ context.Context, _ uint64, _, _ int) ([]model.DiagnosisType, int64, error) {
					return nil, 0, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithDiagnosisSvc(tt.typeSvc, &mockDiagnosisNameService{})

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)

			h.ListDiagnosisTypes(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetDiagnosisType ----

func TestGetDiagnosisType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		typeSvc    *mockDiagnosisTypeService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns diagnosis type for valid id",
			paramID:  "3",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			typeSvc: &mockDiagnosisTypeService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.DiagnosisType, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(3), id)
					return &model.DiagnosisType{ID: 3, Name: "感染症"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"感染症"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			typeSvc:    &mockDiagnosisTypeService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			typeSvc:    &mockDiagnosisTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			typeSvc: &mockDiagnosisTypeService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.DiagnosisType, error) {
					return nil, apperrors.WrapNotFound("diagnosis_type", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithDiagnosisSvc(tt.typeSvc, &mockDiagnosisNameService{})

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.GetDiagnosisType(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateDiagnosisType ----

func TestCreateDiagnosisType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{"name": "新カテゴリ", "is_active": true}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		typeSvc    *mockDiagnosisTypeService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "creates diagnosis type successfully",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			typeSvc: &mockDiagnosisTypeService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateDiagnosisTypeInput) (*model.DiagnosisType, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "新カテゴリ", input.Name)
					return &model.DiagnosisType{ID: 1, Name: input.Name}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"新カテゴリ"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			typeSvc:    &mockDiagnosisTypeService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when name is missing",
			body:       map[string]any{"is_active": true},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			typeSvc:    &mockDiagnosisTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 409 on conflict",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			typeSvc: &mockDiagnosisTypeService{
				createFn: func(_ context.Context, _ uint64, _ *CreateDiagnosisTypeInput) (*model.DiagnosisType, error) {
					return nil, apperrors.WrapAlreadyExists("diagnosis_type", "新カテゴリ")
				},
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			typeSvc: &mockDiagnosisTypeService{
				createFn: func(_ context.Context, _ uint64, _ *CreateDiagnosisTypeInput) (*model.DiagnosisType, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithDiagnosisSvc(tt.typeSvc, &mockDiagnosisNameService{})

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.CreateDiagnosisType(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- UpdateDiagnosisType ----

func TestUpdateDiagnosisType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		typeSvc    *mockDiagnosisTypeService
		wantStatus int
	}{
		{
			name:     "updates diagnosis type successfully",
			paramID:  "1",
			body:     map[string]any{"name": "更新カテゴリ"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			typeSvc: &mockDiagnosisTypeService{
				updateFn: func(_ context.Context, _, _ uint64, input *UpdateDiagnosisTypeInput) (*model.DiagnosisType, error) {
					require.NotNil(t, input.Name)
					assert.Equal(t, "更新カテゴリ", *input.Name)
					return &model.DiagnosisType{ID: 1, Name: *input.Name}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(_ *gin.Context) {},
			typeSvc:    &mockDiagnosisTypeService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			typeSvc:    &mockDiagnosisTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON",
			paramID:    "1",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			typeSvc:    &mockDiagnosisTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			body:     map[string]any{"name": "テスト"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			typeSvc: &mockDiagnosisTypeService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateDiagnosisTypeInput) (*model.DiagnosisType, error) {
					return nil, apperrors.WrapNotFound("diagnosis_type", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithDiagnosisSvc(tt.typeSvc, &mockDiagnosisNameService{})

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdateDiagnosisType(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- DeleteDiagnosisType ----

func newDeleteDiagnosisTypeRouter(typeSvc DiagnosisTypeService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithDiagnosisSvc(typeSvc, &mockDiagnosisNameService{})
	r.DELETE("/diagnosis-types/:id", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteDiagnosisType)
	return r
}

func TestDeleteDiagnosisType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		typeSvc    *mockDiagnosisTypeService
		wantStatus int
	}{
		{
			name:    "deletes diagnosis type successfully",
			paramID: "1",
			typeSvc: &mockDiagnosisTypeService{
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
			typeSvc:    &mockDiagnosisTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when not found",
			paramID: "999",
			typeSvc: &mockDiagnosisTypeService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("diagnosis_type", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "returns 409 when in use",
			paramID: "2",
			typeSvc: &mockDiagnosisTypeService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapConflict("この診断カテゴリには診断名が登録されているため削除できません")
				},
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteDiagnosisTypeRouter(tt.typeSvc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/diagnosis-types/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithDiagnosisSvc(&mockDiagnosisTypeService{}, &mockDiagnosisNameService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteDiagnosisType(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- ReorderDiagnosisTypes ----

func newReorderDiagnosisTypesRouter(typeSvc DiagnosisTypeService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithDiagnosisSvc(typeSvc, &mockDiagnosisNameService{})
	r.PUT("/diagnosis-types/reorder", func(c *gin.Context) {
		setClinicID(c)
	}, h.ReorderDiagnosisTypes)
	return r
}

func TestReorderDiagnosisTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("reorders diagnosis types successfully", func(t *testing.T) {
		typeSvc := &mockDiagnosisTypeService{
			reorderFn: func(_ context.Context, clinicID uint64, ids []uint64) error {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, []uint64{3, 1, 2}, ids)
				return nil
			},
		}
		router := newReorderDiagnosisTypesRouter(typeSvc)
		body, _ := json.Marshal(map[string]any{"ids": []int{3, 1, 2}})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/diagnosis-types/reorder", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithDiagnosisSvc(&mockDiagnosisTypeService{}, &mockDiagnosisNameService{})
		body, _ := json.Marshal(map[string]any{"ids": []int{1, 2}})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		h.ReorderDiagnosisTypes(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		router := newReorderDiagnosisTypesRouter(&mockDiagnosisTypeService{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/diagnosis-types/reorder", bytes.NewReader([]byte(`{invalid}`)))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 500 on service error", func(t *testing.T) {
		typeSvc := &mockDiagnosisTypeService{
			reorderFn: func(_ context.Context, _ uint64, _ []uint64) error {
				return fmt.Errorf("db failure")
			},
		}
		router := newReorderDiagnosisTypesRouter(typeSvc)
		body, _ := json.Marshal(map[string]any{"ids": []int{1, 2}})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/diagnosis-types/reorder", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// ---- ListDiagnosisNames ----

func TestListDiagnosisNames(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		nameSvc    *mockDiagnosisNameService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns all diagnosis names",
			query:    "page=1&limit=10",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			nameSvc: &mockDiagnosisNameService{
				listFn: func(_ context.Context, _ uint64, typeID *uint64, _, _ int) ([]model.DiagnosisName, int64, error) {
					assert.Nil(t, typeID)
					return []model.DiagnosisName{{ID: 1, Name: "急性胃炎"}}, 1, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"急性胃炎"`,
		},
		{
			name:     "filters by type_id when provided",
			query:    "page=1&limit=10&type_id=5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			nameSvc: &mockDiagnosisNameService{
				listFn: func(_ context.Context, _ uint64, typeID *uint64, _, _ int) ([]model.DiagnosisName, int64, error) {
					require.NotNil(t, typeID)
					assert.Equal(t, uint64(5), *typeID)
					return []model.DiagnosisName{{ID: 2, Name: "慢性胃炎"}}, 1, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"慢性胃炎"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			query:      "page=1&limit=10",
			setupCtx:   func(_ *gin.Context) {},
			nameSvc:    &mockDiagnosisNameService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid pagination",
			query:      "page=0",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			nameSvc:    &mockDiagnosisNameService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			query:    "page=1&limit=10",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			nameSvc: &mockDiagnosisNameService{
				listFn: func(_ context.Context, _ uint64, _ *uint64, _, _ int) ([]model.DiagnosisName, int64, error) {
					return nil, 0, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithDiagnosisSvc(&mockDiagnosisTypeService{}, tt.nameSvc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)

			h.ListDiagnosisNames(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetDiagnosisName ----

func TestGetDiagnosisName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		nameSvc    *mockDiagnosisNameService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns diagnosis name for valid id",
			paramID:  "7",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			nameSvc: &mockDiagnosisNameService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.DiagnosisName, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(7), id)
					return &model.DiagnosisName{ID: 7, Name: "肺炎"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"肺炎"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			nameSvc:    &mockDiagnosisNameService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			nameSvc:    &mockDiagnosisNameService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			nameSvc: &mockDiagnosisNameService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.DiagnosisName, error) {
					return nil, apperrors.WrapNotFound("diagnosis_name", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithDiagnosisSvc(&mockDiagnosisTypeService{}, tt.nameSvc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.GetDiagnosisName(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateDiagnosisName ----

func TestCreateDiagnosisName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{
			"name":              "腸炎",
			"diagnosis_type_id": 2,
			"is_active":         true,
		}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		nameSvc    *mockDiagnosisNameService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "creates diagnosis name successfully",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			nameSvc: &mockDiagnosisNameService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateDiagnosisNameInput) (*model.DiagnosisName, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "腸炎", input.Name)
					return &model.DiagnosisName{ID: 1, Name: input.Name}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"腸炎"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			nameSvc:    &mockDiagnosisNameService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when name is missing",
			body:       map[string]any{"diagnosis_type_id": 1},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			nameSvc:    &mockDiagnosisNameService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 409 on conflict",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			nameSvc: &mockDiagnosisNameService{
				createFn: func(_ context.Context, _ uint64, _ *CreateDiagnosisNameInput) (*model.DiagnosisName, error) {
					return nil, apperrors.WrapAlreadyExists("diagnosis_name", "腸炎")
				},
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			nameSvc: &mockDiagnosisNameService{
				createFn: func(_ context.Context, _ uint64, _ *CreateDiagnosisNameInput) (*model.DiagnosisName, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithDiagnosisSvc(&mockDiagnosisTypeService{}, tt.nameSvc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.CreateDiagnosisName(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- UpdateDiagnosisName ----

func TestUpdateDiagnosisName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		nameSvc    *mockDiagnosisNameService
		wantStatus int
	}{
		{
			name:     "updates diagnosis name successfully",
			paramID:  "1",
			body:     map[string]any{"name": "更新診断名"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			nameSvc: &mockDiagnosisNameService{
				updateFn: func(_ context.Context, _, _ uint64, input *UpdateDiagnosisNameInput) (*model.DiagnosisName, error) {
					require.NotNil(t, input.Name)
					assert.Equal(t, "更新診断名", *input.Name)
					return &model.DiagnosisName{ID: 1, Name: *input.Name}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(_ *gin.Context) {},
			nameSvc:    &mockDiagnosisNameService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			nameSvc:    &mockDiagnosisNameService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON",
			paramID:    "1",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			nameSvc:    &mockDiagnosisNameService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			body:     map[string]any{"name": "テスト"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			nameSvc: &mockDiagnosisNameService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateDiagnosisNameInput) (*model.DiagnosisName, error) {
					return nil, apperrors.WrapNotFound("diagnosis_name", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithDiagnosisSvc(&mockDiagnosisTypeService{}, tt.nameSvc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdateDiagnosisName(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- DeleteDiagnosisName ----

func newDeleteDiagnosisNameRouter(nameSvc DiagnosisNameService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithDiagnosisSvc(&mockDiagnosisTypeService{}, nameSvc)
	r.DELETE("/diagnosis-names/:id", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteDiagnosisName)
	return r
}

func TestDeleteDiagnosisName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		nameSvc    *mockDiagnosisNameService
		wantStatus int
	}{
		{
			name:    "deletes diagnosis name successfully",
			paramID: "1",
			nameSvc: &mockDiagnosisNameService{
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
			nameSvc:    &mockDiagnosisNameService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when not found",
			paramID: "999",
			nameSvc: &mockDiagnosisNameService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("diagnosis_name", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "returns 409 when in use",
			paramID: "2",
			nameSvc: &mockDiagnosisNameService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapConflict("この診断名は診療記録で使用中のため削除できません")
				},
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteDiagnosisNameRouter(tt.nameSvc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/diagnosis-names/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithDiagnosisSvc(&mockDiagnosisTypeService{}, &mockDiagnosisNameService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteDiagnosisName(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- ReorderDiagnosisNames ----

func newReorderDiagnosisNamesRouter(nameSvc DiagnosisNameService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithDiagnosisSvc(&mockDiagnosisTypeService{}, nameSvc)
	r.PUT("/diagnosis-names/reorder", func(c *gin.Context) {
		setClinicID(c)
	}, h.ReorderDiagnosisNames)
	return r
}

func TestReorderDiagnosisNames(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("reorders diagnosis names successfully", func(t *testing.T) {
		nameSvc := &mockDiagnosisNameService{
			reorderFn: func(_ context.Context, clinicID uint64, ids []uint64) error {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, []uint64{2, 1, 3}, ids)
				return nil
			},
		}
		router := newReorderDiagnosisNamesRouter(nameSvc)
		body, _ := json.Marshal(map[string]any{"ids": []int{2, 1, 3}})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/diagnosis-names/reorder", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithDiagnosisSvc(&mockDiagnosisTypeService{}, &mockDiagnosisNameService{})
		body, _ := json.Marshal(map[string]any{"ids": []int{1, 2}})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		h.ReorderDiagnosisNames(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		router := newReorderDiagnosisNamesRouter(&mockDiagnosisNameService{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/diagnosis-names/reorder", bytes.NewReader([]byte(`{invalid}`)))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 500 on service error", func(t *testing.T) {
		nameSvc := &mockDiagnosisNameService{
			reorderFn: func(_ context.Context, _ uint64, _ []uint64) error {
				return fmt.Errorf("db failure")
			},
		}
		router := newReorderDiagnosisNamesRouter(nameSvc)
		body, _ := json.Marshal(map[string]any{"ids": []int{1, 2}})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/diagnosis-names/reorder", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// ---- ListDiagnosisNamesAll ----

func TestListDiagnosisNamesAll(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		nameSvc    *mockDiagnosisNameService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns all diagnosis names without pagination",
			query:    "",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			nameSvc: &mockDiagnosisNameService{
				listNamesFn: func(_ context.Context, clinicID uint64, typeID *uint64) ([]model.DiagnosisName, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Nil(t, typeID)
					return []model.DiagnosisName{{ID: 1, Name: "急性胃炎"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"急性胃炎"`,
		},
		{
			name:     "filters by type_id when provided",
			query:    "type_id=5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			nameSvc: &mockDiagnosisNameService{
				listNamesFn: func(_ context.Context, _ uint64, typeID *uint64) ([]model.DiagnosisName, error) {
					require.NotNil(t, typeID)
					assert.Equal(t, uint64(5), *typeID)
					return []model.DiagnosisName{{ID: 2, Name: "慢性胃炎"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"慢性胃炎"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			query:      "",
			setupCtx:   func(_ *gin.Context) {},
			nameSvc:    &mockDiagnosisNameService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for malformed type_id",
			query:      "type_id=abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			nameSvc:    &mockDiagnosisNameService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			query:    "",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			nameSvc: &mockDiagnosisNameService{
				listNamesFn: func(_ context.Context, _ uint64, _ *uint64) ([]model.DiagnosisName, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithDiagnosisSvc(&mockDiagnosisTypeService{}, tt.nameSvc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)

			h.ListDiagnosisNamesAll(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}
