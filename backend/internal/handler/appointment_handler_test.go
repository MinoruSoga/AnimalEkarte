package handler

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

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- mock ReservationService ----

type mockReservationService struct {
	listFn    func(ctx context.Context, clinicID uint64, page, limit int, date *time.Time, status, source *string, petID, ownerID *uint64) ([]model.Reservation, int64, error)
	getByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	createFn  func(ctx context.Context, r *model.Reservation) error
	updateFn  func(ctx context.Context, clinicID, id uint64, input *service.UpdateReservationInput) (*model.Reservation, error)
	deleteFn  func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockReservationService) List(ctx context.Context, clinicID uint64, page, limit int, date *time.Time, status, source *string, petID, ownerID *uint64) ([]model.Reservation, int64, error) {
	return m.listFn(ctx, clinicID, page, limit, date, status, source, petID, ownerID)
}

func (m *mockReservationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	return m.getByIDFn(ctx, clinicID, id)
}

func (m *mockReservationService) Create(ctx context.Context, r *model.Reservation) error {
	return m.createFn(ctx, r)
}

func (m *mockReservationService) Update(ctx context.Context, clinicID, id uint64, input *service.UpdateReservationInput) (*model.Reservation, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, input)
	}
	return &model.Reservation{}, nil
}

func (m *mockReservationService) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func newHandlerWithReservationSvc(svc service.ReservationService) *Handler {
	return &Handler{svc: &service.Services{
		Reservation:           svc,
		StaffClinicAssignment: &mockStaffClinicAssignmentService{},
	}}
}

// mockStaffClinicAssignmentService はテスト用モック。テストで使われるクリニックID（1, 3）すべてに所属を返す。
type mockStaffClinicAssignmentService struct{}

func (m *mockStaffClinicAssignmentService) FindAllByStaffID(_ context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
	return []model.StaffClinicAssignment{
		{StaffID: staffID, ClinicID: 1},
		{StaffID: staffID, ClinicID: 3},
	}, nil
}
func (m *mockStaffClinicAssignmentService) FindByClinicID(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
	return nil, nil
}
func (m *mockStaffClinicAssignmentService) Create(_ context.Context, _ *model.StaffClinicAssignment) error {
	return nil
}
func (m *mockStaffClinicAssignmentService) Update(_ context.Context, _ *model.StaffClinicAssignment) error {
	return nil
}
func (m *mockStaffClinicAssignmentService) Delete(_ context.Context, _, _ uint64) error {
	return nil
}

// ---- ListReservations ----

func TestListReservations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockReservationService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns paginated reservations",
			query:    "page=1&limit=10",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationService{
				listFn: func(_ context.Context, _ uint64, _, _ int, date *time.Time, status, _ *string, _, _ *uint64) ([]model.Reservation, int64, error) {
					assert.Nil(t, date)
					assert.Nil(t, status)
					return []model.Reservation{{ID: 1, Notes: "初診"}}, 1, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"notes":"初診"`,
		},
		{
			name:     "passes date filter to service",
			query:    "page=1&limit=10&date=2026-03-24",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationService{
				listFn: func(_ context.Context, _ uint64, _, _ int, date *time.Time, _, _ *string, _, _ *uint64) ([]model.Reservation, int64, error) {
					require.NotNil(t, date)
					assert.Equal(t, 2026, date.Year())
					assert.Equal(t, time.March, date.Month())
					assert.Equal(t, 24, date.Day())
					return []model.Reservation{}, 0, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 for invalid date format",
			query:      "page=1&limit=10&date=2026/03/24",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockReservationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid pet_id",
			query:      "page=1&limit=10&pet_id=abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockReservationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			query:      "page=1&limit=10",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockReservationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			query:    "page=1&limit=10",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationService{
				listFn: func(_ context.Context, _ uint64, _, _ int, _ *time.Time, _, _ *string, _, _ *uint64) ([]model.Reservation, int64, error) {
					return nil, 0, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithReservationSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)
			h.ListReservations(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetReservation ----

func TestGetReservation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockReservationService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns reservation for valid id",
			paramID:  "3",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(3), id)
					return &model.Reservation{ID: 3, Notes: "再診"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"notes":"再診"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockReservationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockReservationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
					return nil, apperrors.WrapNotFound("reservation", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithReservationSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.GetReservation(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateReservation ----

func TestCreateReservation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Now()
	validBody := func() map[string]any {
		return map[string]any{
			"start_time":          now.Format(time.RFC3339),
			"end_time":            now.Add(30 * time.Minute).Format(time.RFC3339),
			"reservation_type_id": 1,
			"notes":               "健康診断",
		}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockReservationService
		wantStatus int
	}{
		{
			name:     "creates reservation successfully",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			svc: &mockReservationService{
				createFn: func(_ context.Context, r *model.Reservation) error {
					assert.Equal(t, "健康診断", r.Notes)
					require.NotNil(t, r.CreatedBy)
					assert.Equal(t, uint64(1), *r.CreatedBy) // extractStaffID from user_id="1"
					return nil
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockReservationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when required fields are missing",
			body:       map[string]any{"notes": "テスト"}, // start_time, end_time, reservation_type_id missing
			setupCtx:   func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			svc:        &mockReservationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 400 for invalid visit_type",
			body: func() map[string]any {
				b := validBody()
				b["visit_type"] = "invalid_type"
				return b
			}(),
			setupCtx:   func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			svc:        &mockReservationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 400 for invalid status",
			body: func() map[string]any {
				b := validBody()
				b["status"] = "invalid_status"
				return b
			}(),
			setupCtx:   func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			svc:        &mockReservationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			svc: &mockReservationService{
				createFn: func(_ context.Context, _ *model.Reservation) error {
					return fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithReservationSvc(tt.svc)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)
			h.CreateReservation(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- UpdateReservation ----

func TestUpdateReservation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockReservationService
		wantStatus int
	}{
		{
			name:     "updates reservation successfully",
			paramID:  "1",
			body:     map[string]any{"notes": "更新済みメモ"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationService{
				updateFn: func(_ context.Context, _, _ uint64, input *service.UpdateReservationInput) (*model.Reservation, error) {
					require.NotNil(t, input.Notes)
					assert.Equal(t, "更新済みメモ", *input.Notes)
					return &model.Reservation{ID: 1}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 for invalid visit_type",
			paramID:    "1",
			body:       map[string]any{"visit_type": "invalid_type"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockReservationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "accepts numeric doctor_id",
			paramID:  "1",
			body:     map[string]any{"doctor_id": 42},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationService{
				updateFn: func(_ context.Context, _, _ uint64, input *service.UpdateReservationInput) (*model.Reservation, error) {
					require.NotNil(t, input.DoctorID)
					assert.Equal(t, uint64(42), *input.DoctorID)
					return &model.Reservation{ID: 1}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       map[string]any{},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockReservationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"notes": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockReservationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			body:     map[string]any{"notes": "テスト"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationService{
				updateFn: func(_ context.Context, _, _ uint64, _ *service.UpdateReservationInput) (*model.Reservation, error) {
					return nil, apperrors.WrapNotFound("reservation", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithReservationSvc(tt.svc)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.UpdateReservation(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- DeleteReservation ----

func newDeleteReservationRouter(svc service.ReservationService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithReservationSvc(svc)
	r.DELETE("/reservations/:id", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteReservation)
	return r
}

func TestDeleteReservation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockReservationService
		wantStatus int
	}{
		{
			name:    "deletes reservation successfully",
			paramID: "1",
			svc: &mockReservationService{
				deleteFn: func(_ context.Context, _, _ uint64) error { return nil },
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			svc:        &mockReservationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when not found",
			paramID: "999",
			svc: &mockReservationService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("reservation", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteReservationRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/reservations/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithReservationSvc(&mockReservationService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteReservation(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
