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

// ---- mock ChiefComplaintTypeService ----

type mockChiefComplaintTypeService struct {
	listFn    func(ctx context.Context, clinicID uint64) ([]model.ChiefComplaintType, error)
	getByIDFn func(ctx context.Context, clinicID, id uint64) (*model.ChiefComplaintType, error)
	createFn  func(ctx context.Context, clinicID uint64, input *CreateChiefComplaintTypeInput) (*model.ChiefComplaintType, error)
	updateFn  func(ctx context.Context, clinicID, id uint64, input *UpdateChiefComplaintTypeInput) (*model.ChiefComplaintType, error)
	deleteFn  func(ctx context.Context, clinicID, id uint64) error
	reorderFn func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockChiefComplaintTypeService) List(ctx context.Context, clinicID uint64) ([]model.ChiefComplaintType, error) {
	return m.listFn(ctx, clinicID)
}

func (m *mockChiefComplaintTypeService) GetByID(ctx context.Context, clinicID, id uint64) (*model.ChiefComplaintType, error) {
	return m.getByIDFn(ctx, clinicID, id)
}

func (m *mockChiefComplaintTypeService) Create(ctx context.Context, clinicID uint64, input *CreateChiefComplaintTypeInput) (*model.ChiefComplaintType, error) {
	return m.createFn(ctx, clinicID, input)
}

func (m *mockChiefComplaintTypeService) Update(ctx context.Context, clinicID, id uint64, input *UpdateChiefComplaintTypeInput) (*model.ChiefComplaintType, error) {
	return m.updateFn(ctx, clinicID, id, input)
}

func (m *mockChiefComplaintTypeService) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockChiefComplaintTypeService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return m.reorderFn(ctx, clinicID, ids)
}

// ---- test helper ----

func newHandlerWithChiefComplaintSvc(svc ChiefComplaintTypeService) *ChiefComplaintHandler {
	return NewChiefComplaintHandler(svc)
}

// ---- ListChiefComplaints ----

func TestListChiefComplaints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(c *gin.Context)
		svc        *mockChiefComplaintTypeService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of chief complaint types",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockChiefComplaintTypeService{
				listFn: func(_ context.Context, clinicID uint64) ([]model.ChiefComplaintType, error) {
					assert.Equal(t, uint64(1), clinicID)
					return []model.ChiefComplaintType{{ID: 1, ClinicID: clinicID, Name: "嘔吐"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"嘔吐"`,
		},
		{
			name:     "returns empty list when no types exist",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockChiefComplaintTypeService{
				listFn: func(_ context.Context, _ uint64) ([]model.ChiefComplaintType, error) {
					return []model.ChiefComplaintType{}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `[]`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockChiefComplaintTypeService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockChiefComplaintTypeService{
				listFn: func(_ context.Context, _ uint64) ([]model.ChiefComplaintType, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithChiefComplaintSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			tt.setupCtx(c)

			h.ListChiefComplaints(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetChiefComplaint ----

func TestGetChiefComplaint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockChiefComplaintTypeService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns chief complaint type for valid id",
			paramID:  "4",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockChiefComplaintTypeService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ChiefComplaintType, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(4), id)
					return &model.ChiefComplaintType{ID: 4, ClinicID: clinicID, Name: "下痢"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"下痢"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockChiefComplaintTypeService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockChiefComplaintTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when chief complaint type not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockChiefComplaintTypeService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.ChiefComplaintType, error) {
					return nil, apperrors.WrapNotFound("chief_complaint_type", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithChiefComplaintSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.GetChiefComplaint(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateChiefComplaint ----

func TestCreateChiefComplaint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{
			"name":      "食欲不振",
			"is_active": true,
		}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockChiefComplaintTypeService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "creates chief complaint type successfully",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockChiefComplaintTypeService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateChiefComplaintTypeInput) (*model.ChiefComplaintType, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "食欲不振", input.Name)
					return &model.ChiefComplaintType{ID: 8, ClinicID: clinicID, Name: input.Name}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"食欲不振"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockChiefComplaintTypeService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when name is missing",
			body:       map[string]any{"is_active": true},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockChiefComplaintTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockChiefComplaintTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockChiefComplaintTypeService{
				createFn: func(_ context.Context, _ uint64, _ *CreateChiefComplaintTypeInput) (*model.ChiefComplaintType, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithChiefComplaintSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.CreateChiefComplaint(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- UpdateChiefComplaint ----

func TestUpdateChiefComplaint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockChiefComplaintTypeService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "updates chief complaint type successfully",
			paramID:  "1",
			body:     map[string]any{"name": "体重減少"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockChiefComplaintTypeService{
				updateFn: func(_ context.Context, clinicID, id uint64, input *UpdateChiefComplaintTypeInput) (*model.ChiefComplaintType, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					require.NotNil(t, input.Name)
					assert.Equal(t, "体重減少", *input.Name)
					return &model.ChiefComplaintType{ID: 1, ClinicID: clinicID, Name: *input.Name}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"体重減少"`,
		},
		{
			name:     "partial update with description",
			paramID:  "1",
			body:     map[string]any{"description": "急激な体重減少"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockChiefComplaintTypeService{
				updateFn: func(_ context.Context, _, _ uint64, input *UpdateChiefComplaintTypeInput) (*model.ChiefComplaintType, error) {
					require.NotNil(t, input.Description)
					assert.Equal(t, "急激な体重減少", *input.Description)
					assert.Nil(t, input.Name)
					return &model.ChiefComplaintType{ID: 1, Description: *input.Description}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockChiefComplaintTypeService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockChiefComplaintTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when chief complaint type not found",
			paramID:  "999",
			body:     map[string]any{"name": "テスト"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockChiefComplaintTypeService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateChiefComplaintTypeInput) (*model.ChiefComplaintType, error) {
					return nil, apperrors.WrapNotFound("chief_complaint_type", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithChiefComplaintSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdateChiefComplaint(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- DeleteChiefComplaint ----

func newDeleteChiefComplaintRouter(svc ChiefComplaintTypeService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithChiefComplaintSvc(svc)
	r.DELETE("/chief-complaints/:id", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteChiefComplaint)
	return r
}

func TestDeleteChiefComplaint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockChiefComplaintTypeService
		wantStatus int
	}{
		{
			name:    "deletes chief complaint type successfully",
			paramID: "1",
			svc: &mockChiefComplaintTypeService{
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
			svc:        &mockChiefComplaintTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when chief complaint type not found",
			paramID: "999",
			svc: &mockChiefComplaintTypeService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("chief_complaint_type", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "returns 409 when chief complaint type is in use",
			paramID: "2",
			svc: &mockChiefComplaintTypeService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapConflict("この主訴カテゴリは問診記録で使用中のため削除できません")
				},
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteChiefComplaintRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/chief-complaints/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithChiefComplaintSvc(&mockChiefComplaintTypeService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteChiefComplaint(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- ReorderChiefComplaints ----

func newReorderChiefComplaintsRouter(svc ChiefComplaintTypeService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithChiefComplaintSvc(svc)
	r.POST("/chief-complaints/reorder", func(c *gin.Context) {
		setClinicID(c)
	}, h.ReorderChiefComplaints)
	return r
}

func TestReorderChiefComplaints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("reorders chief complaint types successfully", func(t *testing.T) {
		svc := &mockChiefComplaintTypeService{
			reorderFn: func(_ context.Context, clinicID uint64, ids []uint64) error {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, []uint64{2, 1, 3}, ids)
				return nil
			},
		}
		router := newReorderChiefComplaintsRouter(svc)
		bodyBytes, err := json.Marshal(map[string]any{"ids": []int{2, 1, 3}})
		require.NoError(t, err)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/chief-complaints/reorder", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithChiefComplaintSvc(&mockChiefComplaintTypeService{})
		bodyBytes, err := json.Marshal(map[string]any{"ids": []int{1, 2}})
		require.NoError(t, err)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		h.ReorderChiefComplaints(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		router := newReorderChiefComplaintsRouter(&mockChiefComplaintTypeService{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/chief-complaints/reorder", bytes.NewBufferString("not-json"))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 500 on service error", func(t *testing.T) {
		svc := &mockChiefComplaintTypeService{
			reorderFn: func(_ context.Context, _ uint64, _ []uint64) error {
				return fmt.Errorf("db error")
			},
		}
		router := newReorderChiefComplaintsRouter(svc)
		bodyBytes, err := json.Marshal(map[string]any{"ids": []int{1, 2}})
		require.NoError(t, err)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/chief-complaints/reorder", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
