package trimming

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mock Service ----

type mockService struct {
	listFn    func(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Reservation, int64, error)
	getByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	createFn  func(ctx context.Context, clinicID uint64, input *CreateTrimmingInput) (*model.Reservation, error)
	updateFn  func(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingInput) (*model.Reservation, error)
	deleteFn  func(ctx context.Context, clinicID, id uint64, actorID *uint64) error
}

func (m *mockService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Reservation, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, clinicID, petID, ownerID, startDate, endDate, page, limit)
	}
	return nil, 0, nil
}

func (m *mockService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, clinicID, id)
	}
	return &model.Reservation{ID: id}, nil
}

func (m *mockService) Create(ctx context.Context, clinicID uint64, input *CreateTrimmingInput) (*model.Reservation, error) {
	if m.createFn != nil {
		return m.createFn(ctx, clinicID, input)
	}
	return &model.Reservation{ID: 1}, nil
}

func (m *mockService) Update(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingInput) (*model.Reservation, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, input)
	}
	return &model.Reservation{ID: id}, nil
}

func (m *mockService) Delete(ctx context.Context, clinicID, id uint64, actorID *uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id, actorID)
	}
	return nil
}

// ---- helper ----

func newHandlerWithTrimmingSvc(svc Service) *Handler {
	return &Handler{svc: &handlerServices{Trimming: svc}}
}

func setTrimmingWriteContext(c *gin.Context) {
	setClinicID(c)
	c.Set("user_id", "42")
}

// ---- ListTrimmings ----

func TestListTrimmings(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns paginated trimming list",
			query:    "page=1&limit=10",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockService{
				listFn: func(_ context.Context, clinicID uint64, _, _ *uint64, _, _ *string, page, limit int) ([]model.Reservation, int64, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, 1, page)
					assert.Equal(t, 10, limit)
					return []model.Reservation{{ID: 1}}, 1, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			// #231: 一覧レスポンスの pet オブジェクトに breed が含まれる
			name:     "includes pet breed in response",
			query:    "page=1&limit=10",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockService{
				listFn: func(_ context.Context, _ uint64, _, _ *uint64, _, _ *string, _, _ int) ([]model.Reservation, int64, error) {
					return []model.Reservation{{ID: 1, Pet: &model.Pet{ID: 5, Name: "ポチ", Breed: "トイプードル"}}}, 1, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"breed":"トイプードル"`,
		},
		{
			name:       "returns 401 when clinic_id missing",
			query:      "page=1&limit=10",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid pagination",
			query:      "page=0",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			query:    "page=1&limit=10",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockService{
				listFn: func(_ context.Context, _ uint64, _, _ *uint64, _, _ *string, _, _ int) ([]model.Reservation, int64, error) {
					return nil, 0, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithTrimmingSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)
			h.ListTrimmings(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetTrimming ----

func TestGetTrimming(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockService
		wantStatus int
	}{
		{
			name:     "returns trimming for valid id",
			paramID:  "5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), id)
					return &model.Reservation{ID: id}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
					return nil, apperrors.WrapNotFound("trimming", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithTrimmingSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.GetTrimming(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- CreateTrimming ----

func TestCreateTrimming(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Now()
	end := now.Add(time.Hour)

	validBody := func() map[string]any {
		return map[string]any{
			"reservation_type_id": 1,
			"start_time":          now.Format(time.RFC3339),
			"end_time":            end.Format(time.RFC3339),
			"pet_id":              1,
		}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockService
		wantStatus int
	}{
		{
			name:     "creates trimming successfully",
			body:     validBody(),
			setupCtx: setTrimmingWriteContext,
			svc: &mockService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateTrimmingInput) (*model.Reservation, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), input.ReservationTypeID)
					require.NotNil(t, input.ActorID)
					assert.Equal(t, uint64(42), *input.ActorID)
					return &model.Reservation{ID: 1}, nil
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "returns 401 when clinic_id missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 401 when staff actor missing",
			body:       validBody(),
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when reservation_type_id missing",
			body:       map[string]any{"start_time": now.Format(time.RFC3339), "end_time": end.Format(time.RFC3339), "pet_id": 1},
			setupCtx:   setTrimmingWriteContext,
			svc:        &mockService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when pet_id missing",
			body:       map[string]any{"reservation_type_id": 1, "start_time": now.Format(time.RFC3339), "end_time": end.Format(time.RFC3339)},
			setupCtx:   setTrimmingWriteContext,
			svc:        &mockService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: setTrimmingWriteContext,
			svc: &mockService{
				createFn: func(_ context.Context, _ uint64, _ *CreateTrimmingInput) (*model.Reservation, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithTrimmingSvc(tt.svc)
			b, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)
			h.CreateTrimming(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- UpdateTrimming ----

func TestUpdateTrimming(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockService
		wantStatus int
	}{
		{
			name:     "updates trimming successfully",
			paramID:  "1",
			body:     map[string]any{"style_request": "短めにカット"},
			setupCtx: setTrimmingWriteContext,
			svc: &mockService{
				updateFn: func(_ context.Context, _, id uint64, input *UpdateTrimmingInput) (*model.Reservation, error) {
					require.NotNil(t, input.ActorID)
					assert.Equal(t, uint64(42), *input.ActorID)
					return &model.Reservation{ID: id}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "1",
			body:       map[string]any{},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 401 when staff actor missing",
			paramID:    "1",
			body:       map[string]any{"style_request": "test"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"style_request": "test"},
			setupCtx:   setTrimmingWriteContext,
			svc:        &mockService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			body:     map[string]any{"style_request": "test"},
			setupCtx: setTrimmingWriteContext,
			svc: &mockService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateTrimmingInput) (*model.Reservation, error) {
					return nil, apperrors.WrapNotFound("trimming", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithTrimmingSvc(tt.svc)
			b, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(b))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.UpdateTrimming(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- DeleteTrimming ----

func newDeleteTrimmingRouter(svc Service) *gin.Engine {
	r := gin.New()
	h := newHandlerWithTrimmingSvc(svc)
	r.DELETE("/trimmings/:id", setTrimmingWriteContext, h.DeleteTrimming)
	return r
}

func TestDeleteTrimming(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockService
		wantStatus int
	}{
		{
			name:    "deletes trimming successfully",
			paramID: "1",
			svc: &mockService{
				deleteFn: func(_ context.Context, clinicID, id uint64, actorID *uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					require.NotNil(t, actorID)
					assert.Equal(t, uint64(42), *actorID)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			svc:        &mockService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when not found",
			paramID: "999",
			svc: &mockService{
				deleteFn: func(_ context.Context, _, _ uint64, _ *uint64) error {
					return apperrors.WrapNotFound("trimming", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteTrimmingRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/trimmings/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id missing", func(t *testing.T) {
		h := newHandlerWithTrimmingSvc(&mockService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteTrimming(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns 401 when staff actor missing", func(t *testing.T) {
		h := newHandlerWithTrimmingSvc(&mockService{
			deleteFn: func(_ context.Context, _, _ uint64, _ *uint64) error {
				t.Fatal("service must not be called without an authenticated staff actor")
				return nil
			},
		})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		setClinicID(c)
		h.DeleteTrimming(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
