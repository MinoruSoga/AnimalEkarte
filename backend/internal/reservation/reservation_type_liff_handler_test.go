package reservation

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

// ---- mock ReservationTypeLiffService ----

type mockReservationTypeLiffService struct {
	listFn           func(ctx context.Context, clinicID uint64) ([]model.ReservationType, error)
	createFn         func(ctx context.Context, clinicID uint64, input *CreateReservationTypeLiffInput) (*model.ReservationType, error)
	updateFn         func(ctx context.Context, clinicID, id uint64, input *UpdateReservationTypeLiffInput) (*model.ReservationType, error)
	deleteFn         func(ctx context.Context, clinicID, id uint64) error
	patchStatusFn    func(ctx context.Context, clinicID, id uint64, isActive bool) (*model.ReservationType, error)
	patchSortOrderFn func(ctx context.Context, clinicID, id uint64, direction string) error
}

func (m *mockReservationTypeLiffService) List(ctx context.Context, clinicID uint64) ([]model.ReservationType, error) {
	if m.listFn != nil {
		return m.listFn(ctx, clinicID)
	}
	return nil, nil
}

func (m *mockReservationTypeLiffService) Create(ctx context.Context, clinicID uint64, input *CreateReservationTypeLiffInput) (*model.ReservationType, error) {
	if m.createFn != nil {
		return m.createFn(ctx, clinicID, input)
	}
	return &model.ReservationType{ID: 1, Name: input.Name}, nil
}

func (m *mockReservationTypeLiffService) Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationTypeLiffInput) (*model.ReservationType, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, input)
	}
	return &model.ReservationType{ID: id}, nil
}

func (m *mockReservationTypeLiffService) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockReservationTypeLiffService) PatchStatus(ctx context.Context, clinicID, id uint64, isActive bool) (*model.ReservationType, error) {
	if m.patchStatusFn != nil {
		return m.patchStatusFn(ctx, clinicID, id, isActive)
	}
	return &model.ReservationType{ID: id, IsActive: isActive}, nil
}

func (m *mockReservationTypeLiffService) PatchSortOrder(ctx context.Context, clinicID, id uint64, direction string) error {
	if m.patchSortOrderFn != nil {
		return m.patchSortOrderFn(ctx, clinicID, id, direction)
	}
	return nil
}

// ---- helper ----

func newHandlerWithReservationTypeLiffSvc(svc ReservationTypeLiffService) *ReservationTypeLiffHandler {
	return NewReservationTypeLiffHandler(svc)
}

// ---- ListReservationTypeLiffs ----

func TestListReservationTypeLiffs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(c *gin.Context)
		svc        *mockReservationTypeLiffService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of liff types",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeLiffService{
				listFn: func(_ context.Context, clinicID uint64) ([]model.ReservationType, error) {
					assert.Equal(t, uint64(1), clinicID)
					return []model.ReservationType{{ID: 1, Name: "トリミングコース"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"トリミングコース"`,
		},
		{
			name:       "returns 401 when clinic_id missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockReservationTypeLiffService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeLiffService{
				listFn: func(_ context.Context, _ uint64) ([]model.ReservationType, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithReservationTypeLiffSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			tt.setupCtx(c)
			h.ListReservationTypeLiffs(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateReservationTypeLiff ----

func TestCreateReservationTypeLiff(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{"name": "カット"}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockReservationTypeLiffService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "creates liff type successfully",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeLiffService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateReservationTypeLiffInput) (*model.ReservationType, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "カット", input.Name)
					return &model.ReservationType{ID: 1, Name: input.Name}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"カット"`,
		},
		{
			name:       "returns 401 when clinic_id missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockReservationTypeLiffService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when name is missing",
			body:       map[string]any{"color": "#FF0000"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockReservationTypeLiffService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeLiffService{
				createFn: func(_ context.Context, _ uint64, _ *CreateReservationTypeLiffInput) (*model.ReservationType, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithReservationTypeLiffSvc(tt.svc)
			b, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)
			h.CreateReservationTypeLiff(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- UpdateReservationTypeLiff ----

func TestUpdateReservationTypeLiff(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockReservationTypeLiffService
		wantStatus int
	}{
		{
			name:     "updates liff type successfully",
			paramID:  "1",
			body:     map[string]any{"name": "更新カット"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeLiffService{
				updateFn: func(_ context.Context, _, id uint64, input *UpdateReservationTypeLiffInput) (*model.ReservationType, error) {
					require.NotNil(t, input.Name)
					assert.Equal(t, "更新カット", *input.Name)
					return &model.ReservationType{ID: id}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "1",
			body:       map[string]any{"name": "test"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockReservationTypeLiffService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			body:       map[string]any{"name": "test"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockReservationTypeLiffService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on malformed body",
			paramID:    "1",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockReservationTypeLiffService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid reservation_day_option",
			paramID:    "1",
			body:       map[string]any{"reservation_day_option": "invalid-option"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockReservationTypeLiffService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			body:     map[string]any{"name": "test"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeLiffService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateReservationTypeLiffInput) (*model.ReservationType, error) {
					return nil, apperrors.WrapNotFound("reservation_type_liff", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "1",
			body:     map[string]any{"name": "test"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeLiffService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateReservationTypeLiffInput) (*model.ReservationType, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithReservationTypeLiffSvc(tt.svc)
			b, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(b))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.UpdateReservationTypeLiff(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- DeleteReservationTypeLiff ----

func newDeleteReservationTypeLiffRouter(svc ReservationTypeLiffService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithReservationTypeLiffSvc(svc)
	r.DELETE("/reservation-types-liff/:id", func(c *gin.Context) { setClinicID(c) }, h.DeleteReservationTypeLiff)
	return r
}

func TestDeleteReservationTypeLiff(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockReservationTypeLiffService
		wantStatus int
	}{
		{
			name:    "deletes successfully",
			paramID: "1",
			svc: &mockReservationTypeLiffService{
				deleteFn: func(_ context.Context, _, _ uint64) error { return nil },
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:    "returns 404 when not found",
			paramID: "999",
			svc: &mockReservationTypeLiffService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("reservation_type_liff", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteReservationTypeLiffRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/reservation-types-liff/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id missing", func(t *testing.T) {
		h := newHandlerWithReservationTypeLiffSvc(&mockReservationTypeLiffService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteReservationTypeLiff(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- UpdateReservationTypeLiffStatus ----

func TestPatchReservationTypeLiffStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockReservationTypeLiffService
		wantStatus int
	}{
		{
			name:     "patches status to active",
			paramID:  "1",
			body:     map[string]any{"is_active": true},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeLiffService{
				patchStatusFn: func(_ context.Context, _, id uint64, isActive bool) (*model.ReservationType, error) {
					assert.Equal(t, uint64(1), id)
					assert.True(t, isActive)
					return &model.ReservationType{ID: id, IsActive: true}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "1",
			body:       map[string]any{"is_active": false},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockReservationTypeLiffService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			body:     map[string]any{"is_active": true},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeLiffService{
				patchStatusFn: func(_ context.Context, _, _ uint64, _ bool) (*model.ReservationType, error) {
					return nil, apperrors.WrapNotFound("reservation_type_liff", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			body:       map[string]any{"is_active": true},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockReservationTypeLiffService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on malformed body",
			paramID:    "1",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockReservationTypeLiffService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "1",
			body:     map[string]any{"is_active": true},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeLiffService{
				patchStatusFn: func(_ context.Context, _, _ uint64, _ bool) (*model.ReservationType, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithReservationTypeLiffSvc(tt.svc)
			b, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(b))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.UpdateReservationTypeLiffStatus(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- UpdateReservationTypeLiffSortOrder ----

func newPatchLiffSortOrderRouter(svc ReservationTypeLiffService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithReservationTypeLiffSvc(svc)
	r.PATCH("/reservation-types-liff/:id/sort-order", func(c *gin.Context) { setClinicID(c) }, h.UpdateReservationTypeLiffSortOrder)
	return r
}

func TestPatchReservationTypeLiffSortOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("patches sort order successfully", func(t *testing.T) {
		svc := &mockReservationTypeLiffService{
			patchSortOrderFn: func(_ context.Context, _, id uint64, direction string) error {
				assert.Equal(t, uint64(1), id)
				assert.Equal(t, "up", direction)
				return nil
			},
		}
		router := newPatchLiffSortOrderRouter(svc)
		body, _ := json.Marshal(map[string]any{"direction": "up"})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/reservation-types-liff/1/sort-order", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("returns 400 for invalid direction", func(t *testing.T) {
		h := newHandlerWithReservationTypeLiffSvc(&mockReservationTypeLiffService{})
		body, _ := json.Marshal(map[string]any{"direction": "sideways"})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		setClinicID(c)
		h.UpdateReservationTypeLiffSortOrder(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 401 when clinic_id missing", func(t *testing.T) {
		h := newHandlerWithReservationTypeLiffSvc(&mockReservationTypeLiffService{})
		body, _ := json.Marshal(map[string]any{"direction": "up"})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.UpdateReservationTypeLiffSortOrder(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns 400 for non-numeric id", func(t *testing.T) {
		h := newHandlerWithReservationTypeLiffSvc(&mockReservationTypeLiffService{})
		body, _ := json.Marshal(map[string]any{"direction": "up"})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "abc"}}
		setClinicID(c)
		h.UpdateReservationTypeLiffSortOrder(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 400 on malformed body", func(t *testing.T) {
		h := newHandlerWithReservationTypeLiffSvc(&mockReservationTypeLiffService{})
		body, _ := json.Marshal("not-json")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		setClinicID(c)
		h.UpdateReservationTypeLiffSortOrder(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 500 on service error", func(t *testing.T) {
		svc := &mockReservationTypeLiffService{
			patchSortOrderFn: func(_ context.Context, _, _ uint64, _ string) error {
				return fmt.Errorf("db error")
			},
		}
		h := newHandlerWithReservationTypeLiffSvc(svc)
		body, _ := json.Marshal(map[string]any{"direction": "down"})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		setClinicID(c)
		h.UpdateReservationTypeLiffSortOrder(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// ---- UploadReservationTypeLiffImage ----

// TestUploadReservationTypeLiffImage は v2 スコープ未実装エンドポイントが
// 常に 501 Not Implemented を返すことを検証する（BE スコープ境界の回帰防止）。
func TestUploadReservationTypeLiffImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := newHandlerWithReservationTypeLiffSvc(&mockReservationTypeLiffService{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", http.NoBody)

	h.UploadReservationTypeLiffImage(c)

	assert.Equal(t, http.StatusNotImplemented, w.Code)
}
