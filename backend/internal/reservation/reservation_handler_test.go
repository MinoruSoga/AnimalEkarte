package reservation

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

// ---- mock ReservationService ----

type mockReservationService struct {
	listFn                   func(ctx context.Context, clinicIDs []uint64, page, limit int, date, startDate, endDate *time.Time, status, source *string, petID, ownerID *uint64) ([]model.Reservation, int64, error)
	getByIDFn                func(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	getByIDForClinicsFn      func(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Reservation, error)
	createFn                 func(ctx context.Context, input *CreateManualReservationInput) (*model.Reservation, error)
	updateFn                 func(ctx context.Context, clinicID, id uint64, input *UpdateReservationInput) (*model.Reservation, error)
	deleteFn                 func(ctx context.Context, clinicID, id uint64) error
	updateReservationRouteFn func(ctx context.Context, clinicID, id uint64, input UpdateReservationRouteInput) (*model.Reservation, error)
}

func (m *mockReservationService) List(ctx context.Context, clinicIDs []uint64, page, limit int, date, startDate, endDate *time.Time, status, source *string, petID, ownerID *uint64) ([]model.Reservation, int64, error) {
	return m.listFn(ctx, clinicIDs, page, limit, date, startDate, endDate, status, source, petID, ownerID)
}

func (m *mockReservationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	return m.getByIDFn(ctx, clinicID, id)
}

func (m *mockReservationService) GetByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Reservation, error) {
	if m.getByIDForClinicsFn != nil {
		return m.getByIDForClinicsFn(ctx, clinicIDs, id)
	}
	return nil, nil
}

func (m *mockReservationService) Create(ctx context.Context, input *CreateManualReservationInput) (*model.Reservation, error) {
	return m.createFn(ctx, input)
}

func (m *mockReservationService) Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationInput) (*model.Reservation, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, input)
	}
	return &model.Reservation{}, nil
}

func (m *mockReservationService) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockReservationService) UpdateReservationRoute(ctx context.Context, clinicID, id uint64, input UpdateReservationRouteInput) (*model.Reservation, error) {
	if m.updateReservationRouteFn != nil {
		return m.updateReservationRouteFn(ctx, clinicID, id, input)
	}
	return nil, nil
}

func newHandlerWithReservationSvc(svc ReservationService) *ReservationHandler {
	return NewReservationHandler(svc, nil, nil, &mockStaffClinicAssignmentService{})
}

func newHandlerWithReservationAndMedicalRecordSvc(reservationSvc ReservationService, medicalRecordSvc medicalRecordAutoCreator) *ReservationHandler {
	return NewReservationHandler(reservationSvc, medicalRecordSvc, nil, &mockStaffClinicAssignmentService{})
}

func newHandlerWithLiffSvc(liffSvc liffAvailability) *ReservationHandler {
	return NewReservationHandler(nil, nil, liffSvc, &mockStaffClinicAssignmentService{})
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
				listFn: func(_ context.Context, _ []uint64, _, _ int, date, _, _ *time.Time, status, _ *string, _, _ *uint64) ([]model.Reservation, int64, error) {
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
				listFn: func(_ context.Context, _ []uint64, _, _ int, date, _, _ *time.Time, _, _ *string, _, _ *uint64) ([]model.Reservation, int64, error) {
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
			name:     "allows large limit for calendar range",
			query:    "page=1&limit=1000&start_date=2026-03-22&end_date=2026-03-28",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationService{
				listFn: func(_ context.Context, _ []uint64, _ int, limit int, _, startDate, endDate *time.Time, _, _ *string, _, _ *uint64) ([]model.Reservation, int64, error) {
					assert.Equal(t, 1000, limit)
					require.NotNil(t, startDate)
					require.NotNil(t, endDate)
					assert.Equal(t, 2026, startDate.Year())
					assert.Equal(t, time.March, startDate.Month())
					assert.Equal(t, 22, startDate.Day())
					assert.Equal(t, 2026, endDate.Year())
					assert.Equal(t, time.March, endDate.Month())
					assert.Equal(t, 29, endDate.Day())
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
				listFn: func(_ context.Context, _ []uint64, _, _ int, _, _, _ *time.Time, _, _ *string, _, _ *uint64) ([]model.Reservation, int64, error) {
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
				getByIDForClinicsFn: func(_ context.Context, clinicIDs []uint64, id uint64) (*model.Reservation, error) {
					assert.Equal(t, []uint64{1}, clinicIDs)
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
				getByIDForClinicsFn: func(_ context.Context, _ []uint64, _ uint64) (*model.Reservation, error) {
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

func TestGetReservationAvailableTimes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        liffAvailability
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns available time slots",
			query:    "reservation_type_id=5&staff_id=10&date=2026-06-01",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockLiffService{
				getAvailableTimesFn: func(_ context.Context, clinicID, typeID, staffID uint64, date time.Time) ([]TimeSlot, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), typeID)
					assert.Equal(t, uint64(10), staffID)
					assert.Equal(t, 2026, date.Year())
					assert.Equal(t, time.June, date.Month())
					assert.Equal(t, 1, date.Day())
					return []TimeSlot{
						{StartTime: "0945", EndTime: "1045"},
						{StartTime: "1230", EndTime: "1330"},
					}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"start_time":"0945"`,
		},
		{
			name:       "returns 400 when reservation_type_id is missing",
			query:      "date=2026-06-01",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockLiffService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when date is invalid",
			query:      "reservation_type_id=5&date=20260601",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockLiffService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 401 when clinic id is missing",
			query:      "reservation_type_id=5&date=2026-06-01",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockLiffService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 501 when availability service is not configured",
			query:      "reservation_type_id=5&date=2026-06-01",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        nil,
			wantStatus: http.StatusNotImplemented,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithLiffSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)
			h.GetReservationAvailableTimes(c)
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
				createFn: func(_ context.Context, input *CreateManualReservationInput) (*model.Reservation, error) {
					assert.Equal(t, "健康診断", input.Notes)
					require.NotNil(t, input.CreatedBy)
					assert.Equal(t, uint64(1), *input.CreatedBy) // extractStaffID from user_id="1"
					return &model.Reservation{ID: 1, Notes: input.Notes}, nil
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "accepts record shortcut route on create",
			body: func() map[string]any {
				b := validBody()
				b["status"] = "in_consultation"
				b["reservation_route"] = "record_shortcut"
				return b
			}(),
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			svc: &mockReservationService{
				createFn: func(_ context.Context, input *CreateManualReservationInput) (*model.Reservation, error) {
					require.NotNil(t, input.ReservationRoute)
					assert.Equal(t, "record_shortcut", *input.ReservationRoute)
					assert.Equal(t, model.ReservationStatusInConsultation, input.Status)
					return &model.Reservation{ID: 1, ReservationRoute: input.ReservationRoute, Status: input.Status}, nil
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
				createFn: func(_ context.Context, _ *CreateManualReservationInput) (*model.Reservation, error) {
					return nil, fmt.Errorf("db error")
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
				updateFn: func(_ context.Context, _, _ uint64, input *UpdateReservationInput) (*model.Reservation, error) {
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
				updateFn: func(_ context.Context, _, _ uint64, input *UpdateReservationInput) (*model.Reservation, error) {
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
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateReservationInput) (*model.Reservation, error) {
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

func TestUpdateReservation_AutoCreateMedicalRecord(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		reservation    *model.Reservation
		wantAutoCreate bool
	}{
		{
			name: "creates medical record for general checked-in reservation",
			reservation: &model.Reservation{
				ID:              1,
				ReservationType: &model.ReservationType{Category: model.ReservationTypeCategoryGeneral},
			},
			wantAutoCreate: true,
		},
		{
			name: "skips medical record for trimming checked-in reservation",
			reservation: &model.Reservation{
				ID:              2,
				ReservationType: &model.ReservationType{Category: model.ReservationTypeCategoryTrimming},
			},
			wantAutoCreate: false,
		},
		{
			name: "skips medical record for hotel checked-in reservation",
			reservation: &model.Reservation{
				ID:              4,
				ReservationType: &model.ReservationType{Category: model.ReservationTypeCategoryGeneral, Name: "ペットホテル"},
			},
			wantAutoCreate: false,
		},
		{
			name:           "keeps legacy behavior when reservation type is not loaded",
			reservation:    &model.Reservation{ID: 3},
			wantAutoCreate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			autoCreateCalls := 0
			reservationSvc := &mockReservationService{
				updateFn: func(_ context.Context, clinicID, id uint64, input *UpdateReservationInput) (*model.Reservation, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, tt.reservation.ID, id)
					require.NotNil(t, input.Status)
					assert.Equal(t, model.ReservationStatusCheckedIn, *input.Status)
					// #83 Q9: shouldAutoCreate は更新後の reservation.Status で判定するため、
					// 更新後ステータス(checked_in)を反映した予約を返す。
					updated := *tt.reservation
					updated.Status = *input.Status
					return &updated, nil
				},
			}
			medicalRecordSvc := &mockMedicalRecordService{
				autoCreateFromReservationFn: func(_ context.Context, clinicID uint64, reservation *model.Reservation) {
					autoCreateCalls++
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, tt.reservation.ID, reservation.ID)
				},
			}
			h := newHandlerWithReservationAndMedicalRecordSvc(reservationSvc, medicalRecordSvc)

			bodyBytes, err := json.Marshal(map[string]any{"status": string(model.ReservationStatusCheckedIn)})
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", tt.reservation.ID)}}
			setClinicID(c)

			h.UpdateReservation(c)

			assert.Equal(t, http.StatusOK, w.Code)
			if tt.wantAutoCreate {
				assert.Equal(t, 1, autoCreateCalls)
			} else {
				assert.Zero(t, autoCreateCalls)
			}
		})
	}
}

// ---- DeleteReservation ----

func newDeleteReservationRouter(svc ReservationService) *gin.Engine {
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
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
					return &model.Reservation{ID: 1, OwnerID: nil}, nil
				},
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

// ---- UpdateReservationReservationRoute ----

func newPatchReservationRouteRouter(rSvc ReservationService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithReservationSvc(rSvc)
	r.PATCH("/reservations/:id/reservation-route", func(c *gin.Context) { setClinicID(c) }, h.UpdateReservationReservationRoute)
	return r
}

func TestPatchReservationReservationRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	route := func(s string) string { return "/reservations/1/reservation-route" }

	tests := []struct {
		name       string
		body       any
		svc        *mockReservationService
		wantStatus int
	}{
		{
			name: "200 success with valid route",
			body: map[string]any{"route": "line"},
			svc: &mockReservationService{
				updateReservationRouteFn: func(_ context.Context, _ uint64, _ uint64, input UpdateReservationRouteInput) (*model.Reservation, error) {
					assert.Equal(t, "line", input.Route)
					return &model.Reservation{ID: 1}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "200 success with empty route (clears to NULL)",
			body: map[string]any{"route": ""},
			svc: &mockReservationService{
				updateReservationRouteFn: func(_ context.Context, _ uint64, _ uint64, input UpdateReservationRouteInput) (*model.Reservation, error) {
					assert.Equal(t, "", input.Route)
					return &model.Reservation{ID: 1}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "400 malformed JSON",
			body:       "not-json",
			svc:        &mockReservationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "400 route exceeds max=20",
			body:       map[string]any{"route": "this_value_is_way_too_long_for_varchar20"},
			svc:        &mockReservationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "404 reservation not found",
			body: map[string]any{"route": "phone"},
			svc: &mockReservationService{
				updateReservationRouteFn: func(_ context.Context, _ uint64, _ uint64, _ UpdateReservationRouteInput) (*model.Reservation, error) {
					return nil, apperrors.WrapNotFound("reservation", "1")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "500 service internal error",
			body: map[string]any{"route": "reception"},
			svc: &mockReservationService{
				updateReservationRouteFn: func(_ context.Context, _ uint64, _ uint64, _ UpdateReservationRouteInput) (*model.Reservation, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newPatchReservationRouteRouter(tt.svc)
			w := httptest.NewRecorder()
			var bodyBytes []byte
			if s, ok := tt.body.(string); ok {
				bodyBytes = []byte(s)
			} else {
				bodyBytes, _ = json.Marshal(tt.body)
			}
			req := httptest.NewRequest(http.MethodPatch, route("1"), bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("401 when clinic_id is missing", func(t *testing.T) {
		r := gin.New()
		h := newHandlerWithReservationSvc(&mockReservationService{})
		r.PATCH("/reservations/:id/reservation-route", h.UpdateReservationReservationRoute)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/reservations/1/reservation-route", bytes.NewReader([]byte(`{"route":"line"}`)))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// mockMedicalRecordService — medicalRecordAutoCreator view の最小モック
// （handler/appointment_medical_record_mock_test.go の full 版とは独立・view 2メソッドのみ）。
type mockMedicalRecordService struct {
	autoCreateFromReservationFn  func(ctx context.Context, clinicID uint64, reservation *model.Reservation)
	deleteDraftFromReservationFn func(ctx context.Context, clinicID, reservationID uint64)
}

func (m *mockMedicalRecordService) AutoCreateFromReservation(ctx context.Context, clinicID uint64, reservation *model.Reservation) {
	if m.autoCreateFromReservationFn != nil {
		m.autoCreateFromReservationFn(ctx, clinicID, reservation)
	}
}

func (m *mockMedicalRecordService) DeleteDraftFromReservation(ctx context.Context, clinicID, reservationID uint64) {
	if m.deleteDraftFromReservationFn != nil {
		m.deleteDraftFromReservationFn(ctx, clinicID, reservationID)
	}
}
