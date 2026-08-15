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

// ---- mock OccupationService ----

type mockOccupationService struct {
	listFn    func(ctx context.Context, clinicID uint64) ([]model.Occupation, error)
	getByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Occupation, error)
	createFn  func(ctx context.Context, clinicID uint64, input *staffdomain.CreateOccupationInput) (*model.Occupation, error)
	updateFn  func(ctx context.Context, clinicID, id uint64, input *staffdomain.UpdateOccupationInput) (*model.Occupation, error)
	deleteFn  func(ctx context.Context, clinicID, id uint64) error
	reorderFn func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockOccupationService) List(ctx context.Context, clinicID uint64) ([]model.Occupation, error) {
	return m.listFn(ctx, clinicID)
}
func (m *mockOccupationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Occupation, error) {
	return m.getByIDFn(ctx, clinicID, id)
}
func (m *mockOccupationService) Create(ctx context.Context, clinicID uint64, input *staffdomain.CreateOccupationInput) (*model.Occupation, error) {
	return m.createFn(ctx, clinicID, input)
}
func (m *mockOccupationService) Update(ctx context.Context, clinicID, id uint64, input *staffdomain.UpdateOccupationInput) (*model.Occupation, error) {
	return m.updateFn(ctx, clinicID, id, input)
}
func (m *mockOccupationService) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}
func (m *mockOccupationService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return m.reorderFn(ctx, clinicID, ids)
}

// ---- helper ----

func newHandlerWithOccupationSvc(svc staffdomain.OccupationService) *staffdomain.Handler {
	return staffdomain.NewHandler(nil, nil, svc, nil, nil, nil)
}

// ---- ListOccupations ----

func TestListOccupations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(c *gin.Context)
		svc        *mockOccupationService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of occupations",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockOccupationService{
				listFn: func(_ context.Context, clinicID uint64) ([]model.Occupation, error) {
					assert.Equal(t, uint64(1), clinicID)
					return []model.Occupation{{ID: 1, Name: "獣医師"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"獣医師"`,
		},
		{
			name:     "returns empty list",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockOccupationService{
				listFn: func(_ context.Context, _ uint64) ([]model.Occupation, error) {
					return []model.Occupation{}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockOccupationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockOccupationService{
				listFn: func(_ context.Context, _ uint64) ([]model.Occupation, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithOccupationSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			tt.setupCtx(c)

			h.ListOccupations(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetOccupation ----

func TestGetOccupation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockOccupationService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns occupation for valid id",
			paramID:  "4",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockOccupationService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Occupation, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(4), id)
					return &model.Occupation{ID: 4, Name: "看護師"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"看護師"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockOccupationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockOccupationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockOccupationService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Occupation, error) {
					return nil, apperrors.WrapNotFound("occupation", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithOccupationSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.GetOccupation(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateOccupation ----

func TestCreateOccupation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{"name": "受付", "is_active": true}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockOccupationService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "creates occupation successfully",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockOccupationService{
				createFn: func(_ context.Context, clinicID uint64, input *staffdomain.CreateOccupationInput) (*model.Occupation, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "受付", input.Name)
					return &model.Occupation{ID: 1, Name: input.Name}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"受付"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockOccupationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when name is missing",
			body:       map[string]any{"is_active": true},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockOccupationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 409 on conflict",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockOccupationService{
				createFn: func(_ context.Context, _ uint64, _ *staffdomain.CreateOccupationInput) (*model.Occupation, error) {
					return nil, apperrors.WrapAlreadyExists("occupation", "受付")
				},
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockOccupationService{
				createFn: func(_ context.Context, _ uint64, _ *staffdomain.CreateOccupationInput) (*model.Occupation, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithOccupationSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.CreateOccupation(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- UpdateOccupation ----

func TestUpdateOccupation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockOccupationService
		wantStatus int
	}{
		{
			name:     "updates occupation successfully",
			paramID:  "1",
			body:     map[string]any{"name": "トリマー"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockOccupationService{
				updateFn: func(_ context.Context, _, _ uint64, input *staffdomain.UpdateOccupationInput) (*model.Occupation, error) {
					require.NotNil(t, input.Name)
					assert.Equal(t, "トリマー", *input.Name)
					return &model.Occupation{ID: 1, Name: *input.Name}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockOccupationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockOccupationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			body:     map[string]any{"name": "テスト"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockOccupationService{
				updateFn: func(_ context.Context, _, _ uint64, _ *staffdomain.UpdateOccupationInput) (*model.Occupation, error) {
					return nil, apperrors.WrapNotFound("occupation", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithOccupationSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdateOccupation(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- DeleteOccupation ----

func newDeleteOccupationRouter(svc staffdomain.OccupationService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithOccupationSvc(svc)
	r.DELETE("/occupations/:id", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteOccupation)
	return r
}

func TestDeleteOccupation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockOccupationService
		wantStatus int
	}{
		{
			name:    "deletes occupation successfully",
			paramID: "1",
			svc: &mockOccupationService{
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
			svc:        &mockOccupationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when not found",
			paramID: "999",
			svc: &mockOccupationService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("occupation", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "returns 409 when in use",
			paramID: "2",
			svc: &mockOccupationService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapConflict("この役職はスタッフ情報で使用中のため削除できません")
				},
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteOccupationRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/occupations/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithOccupationSvc(&mockOccupationService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteOccupation(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- ReorderOccupations ----

func newReorderOccupationsRouter(svc staffdomain.OccupationService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithOccupationSvc(svc)
	r.PUT("/occupations/reorder", func(c *gin.Context) {
		setClinicID(c)
	}, h.ReorderOccupations)
	return r
}

func TestReorderOccupations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("reorders occupations successfully", func(t *testing.T) {
		svc := &mockOccupationService{
			reorderFn: func(_ context.Context, clinicID uint64, ids []uint64) error {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, []uint64{3, 1, 2}, ids)
				return nil
			},
		}
		router := newReorderOccupationsRouter(svc)
		body, _ := json.Marshal(map[string]any{"ids": []int{3, 1, 2}})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/occupations/reorder", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithOccupationSvc(&mockOccupationService{})
		body, _ := json.Marshal(map[string]any{"ids": []int{1, 2}})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		h.ReorderOccupations(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
