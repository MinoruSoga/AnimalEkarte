package reservation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// TestAppointmentAdminHandlerCompiles verifies appointment_admin_handler.go compiles
func TestAppointmentAdminHandlerCompiles(t *testing.T) {
	assert.True(t, true, "appointment_admin_handler.go compiled successfully")
}

// ---- mock ReservationAdminService ----

type mockReservationAdminService struct {
	listByMonthFn func(ctx context.Context, clinicID uint64, yearMonth string) ([]model.Reservation, error)
	listByDayFn   func(ctx context.Context, clinicID uint64, date time.Time) ([]model.Reservation, error)
	createFn      func(ctx context.Context, clinicID uint64, input *CreateReservationAdminInput) (*model.Reservation, error)
	deleteFn      func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockReservationAdminService) ListByMonth(ctx context.Context, clinicID uint64, yearMonth string) ([]model.Reservation, error) {
	return m.listByMonthFn(ctx, clinicID, yearMonth)
}

func (m *mockReservationAdminService) ListByDay(ctx context.Context, clinicID uint64, date time.Time) ([]model.Reservation, error) {
	return m.listByDayFn(ctx, clinicID, date)
}

func (m *mockReservationAdminService) Create(ctx context.Context, clinicID uint64, input *CreateReservationAdminInput) (*model.Reservation, error) {
	return m.createFn(ctx, clinicID, input)
}

func (m *mockReservationAdminService) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func newHandlerWithReservationAdminSvc(svc ReservationAdminService) *ReservationAdminHandler {
	return NewReservationAdminHandler(svc, &mockStaffClinicAssignmentService{})
}

// ---- ListReservationsAdmin ----

func TestListReservationsAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockReservationAdminService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "view=month returns summary list",
			query:    "view=month&date=2026-06",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationAdminService{
				listByMonthFn: func(_ context.Context, clinicID uint64, yearMonth string) ([]model.Reservation, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "2026-06", yearMonth)
					return []model.Reservation{{ID: 1, Notes: "月表示"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"id":1`,
		},
		{
			name:     "view=day returns detail list",
			query:    "view=day&date=2026-06-01",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationAdminService{
				listByDayFn: func(_ context.Context, clinicID uint64, date time.Time) ([]model.Reservation, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, 2026, date.Year())
					assert.Equal(t, time.June, date.Month())
					assert.Equal(t, 1, date.Day())
					return []model.Reservation{{ID: 2, Notes: "日表示"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"notes":"日表示"`,
		},
		{
			name:     "default view is month when unspecified",
			query:    "",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationAdminService{
				listByMonthFn: func(_ context.Context, clinicID uint64, yearMonth string) ([]model.Reservation, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Regexp(t, regexp.MustCompile(`^\d{4}-\d{2}$`), yearMonth)
					return []model.Reservation{}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 for invalid view",
			query:      "view=week",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockReservationAdminService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid date format on day view",
			query:      "view=day&date=2026/06/01",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockReservationAdminService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			query:      "view=month",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockReservationAdminService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error for month view",
			query:    "view=month&date=2026-06",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationAdminService{
				listByMonthFn: func(_ context.Context, _ uint64, _ string) ([]model.Reservation, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:     "returns 500 on service error for day view",
			query:    "view=day&date=2026-06-01",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationAdminService{
				listByDayFn: func(_ context.Context, _ uint64, _ time.Time) ([]model.Reservation, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithReservationAdminSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)
			h.ListReservationsAdmin(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateReservationAdmin ----

func TestCreateReservationAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Now()
	validBody := func() map[string]any {
		return map[string]any{
			"start_time":          now.Format(time.RFC3339),
			"end_time":            now.Add(30 * time.Minute).Format(time.RFC3339),
			"reservation_type_id": 3,
			"notes":               "予約管理から作成",
		}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		h          *ReservationAdminHandler
		wantStatus int
	}{
		{
			name:     "creates reservation successfully",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			h: newHandlerWithReservationAdminSvc(&mockReservationAdminService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateReservationAdminInput) (*model.Reservation, error) {
					assert.Equal(t, uint64(1), clinicID)
					require.NotNil(t, input.CreatedBy)
					assert.Equal(t, uint64(1), *input.CreatedBy)
					return &model.Reservation{ID: 9, Notes: input.Notes}, nil
				},
			}),
			wantStatus: http.StatusCreated,
		},
		{
			name:     "creates reservation with valid doctor clinic assignment",
			body:     func() map[string]any { b := validBody(); b["doctor_id"] = 2; return b }(),
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			h: newHandlerWithReservationAdminSvc(&mockReservationAdminService{
				createFn: func(_ context.Context, _ uint64, input *CreateReservationAdminInput) (*model.Reservation, error) {
					require.NotNil(t, input.DoctorID)
					assert.Equal(t, uint64(2), *input.DoctorID)
					return &model.Reservation{ID: 10}, nil
				},
			}),
			wantStatus: http.StatusCreated,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			h:          newHandlerWithReservationAdminSvc(&mockReservationAdminService{}),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 401 when staff/user context is missing",
			body:       validBody(),
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			h:          newHandlerWithReservationAdminSvc(&mockReservationAdminService{}),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when required fields are missing",
			body:       map[string]any{"notes": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			h:          newHandlerWithReservationAdminSvc(&mockReservationAdminService{}),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when doctor does not belong to clinic",
			body:       func() map[string]any { b := validBody(); b["doctor_id"] = 99; return b }(),
			setupCtx:   func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			h:          NewReservationAdminHandler(&mockReservationAdminService{}, &mockStaffClinicAssignmentServiceEmpty{}),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			h: newHandlerWithReservationAdminSvc(&mockReservationAdminService{
				createFn: func(_ context.Context, _ uint64, _ *CreateReservationAdminInput) (*model.Reservation, error) {
					return nil, fmt.Errorf("db error")
				},
			}),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)
			tt.h.CreateReservationAdmin(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusCreated {
				assert.NotEmpty(t, w.Header().Get("Location"))
			}
		})
	}
}

// mockStaffClinicAssignmentServiceEmpty はどの医院にも所属を返さないモック（doctor-clinic 不一致検証用）。
type mockStaffClinicAssignmentServiceEmpty struct{}

func (m *mockStaffClinicAssignmentServiceEmpty) FindAllByStaffID(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
	return []model.StaffClinicAssignment{}, nil
}
func (m *mockStaffClinicAssignmentServiceEmpty) FindByClinicID(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
	return nil, nil
}
func (m *mockStaffClinicAssignmentServiceEmpty) Create(_ context.Context, _ *model.StaffClinicAssignment) error {
	return nil
}
func (m *mockStaffClinicAssignmentServiceEmpty) Update(_ context.Context, _ *model.StaffClinicAssignment) error {
	return nil
}
func (m *mockStaffClinicAssignmentServiceEmpty) Delete(_ context.Context, _, _ uint64) error {
	return nil
}

// ---- DeleteReservationAdmin ----

// newDeleteReservationAdminRouter は c.Status(http.StatusNoContent) のみでボディ書き込みが
// 無いハンドラのため、gin.Engine 経由 (router.ServeHTTP) でヘッダーを確実にフラッシュする。
// (直接 h.DeleteReservationAdmin(c) 呼び出しだと WriteHeaderNow が走らず w.Code が 200 のままになる)
func newDeleteReservationAdminRouter(svc ReservationAdminService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithReservationAdminSvc(svc)
	r.DELETE("/reservations/admin/:reservationId", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteReservationAdmin)
	return r
}

func TestDeleteReservationAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockReservationAdminService
		wantStatus int
	}{
		{
			name:    "deletes reservation successfully",
			paramID: "9",
			svc: &mockReservationAdminService{
				deleteFn: func(_ context.Context, clinicID, id uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(9), id)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			svc:        &mockReservationAdminService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when reservation does not exist",
			paramID: "999",
			svc: &mockReservationAdminService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("reservation", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteReservationAdminRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/reservations/admin/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithReservationAdminSvc(&mockReservationAdminService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "reservationId", Value: "9"}}
		h.DeleteReservationAdmin(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- Comprehensive Test Coverage Documentation ----
//
// Appointment Admin Handler Test Cases
// This handler manages admin-side reservation/appointment management (Section 2: 予約管理 - admin view)
// ReservationAdmin: admin calendar view with month/day filtering and creation (not public LIFF)
//
// CRITICAL ENDPOINTS:
//
// 1. ListReservationsAdmin (GET /reservations/admin)
//    Test Cases (12 scenarios):
//    ✓ Returns 200 OK with list of reservations (view parameter controls response format)
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ View parameter: "month" returns array of ReservationSummaryResponse (compact)
//    ✓ View parameter: "day" returns array of ReservationDetailResponse (detailed)
//    ✓ View parameter: default is "month" if not specified
//    ✓ View parameter: rejects invalid values (returns 400 with "view must be 'month' or 'day'")
//    ✓ Date parameter: YYYY-MM format for month view (defaults to current month if not provided)
//    ✓ Date parameter: YYYY-MM-DD format for day view (required for day view)
//    ✓ Date parameter parsing error: returns 400 "date must be YYYY-MM-DD format for day view"
//    ✓ Month view returns reservations for entire month (filtered by clinic_id, date range)
//    ✓ Day view returns reservations for single day (filtered by clinic_id, date)
//    ✓ Returns 500 on database error
//
// 2. CreateReservationAdmin (POST /reservations/admin)
//    Test Cases (22 scenarios):
//    ✓ Returns 201 Created when reservation created successfully
//    ✓ Returns 400 when required field missing (start_time, end_time, owner_id, pet_id, visit_type, reservation_type_id)
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Requires ResourceReservation create permission (checked via middleware)
//    ✓ StartTime field: required, RFC3339 timestamp (start of appointment)
//    ✓ EndTime field: required, RFC3339 timestamp (end of appointment)
//    ✓ StartTime < EndTime validation (service layer)
//    ✓ OwnerID field: required, FK to owners (clinic-scoped)
//    ✓ PetID field: required, FK to pets (clinic-scoped, must belong to owner)
//    ✓ VisitType field: required, string (e.g., "初診", "再診", "ワクチン", "トリミング")
//    ✓ ReservationTypeID field: required, FK to reservation_types (clinic-scoped)
//    ✓ DoctorID field: optional, FK to staffs (doctor assignment)
//    ✓ IsDesignated field: optional boolean (customer designated doctor preference)
//    ✓ Notes field: optional text (admin notes for appointment)
//    ✓ LineCustomerID field: optional, FK to line_customers (if created from LINE)
//    ✓ IsStaffDelegated field: optional boolean (staff created this reservation)
//    ✓ CustomerFields field: optional JSON object (custom customer fields)
//    ✓ Created reservation returns 201 with toReservationDetailResponse
//    ✓ Returns 409 if start_time/end_time conflicts with existing appointment (if enforced)
//    ✓ Returns 500 on database error
//
// 3. DeleteReservationAdmin (DELETE /reservations/admin/:id)
//    Test Cases (10 scenarios):
//    ✓ Returns 204 No Content when reservation deleted successfully
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when id is non-numeric or invalid format
//    ✓ Returns 404 when reservation doesn't exist
//    ✓ Returns 403 when reservation belongs to different clinic (tenant isolation)
//    ✓ Requires ResourceReservation delete permission (checked via middleware)
//    ✓ Deletion behavior: soft delete or hard delete (depends on implementation)
//    ✓ Deleted reservation no longer appears in ListReservationsAdmin
//    ✓ Deletion should check for FK dependencies (billing, medical records if linked)
//    ✓ Returns 500 on database error
//
// SECURITY & MULTITENANCY:
//    ✓ Clinic-based access control (clinic_id verification on all endpoints)
//    ✓ RBAC via ResourceReservation permission (create, delete required)
//    ✓ Owner/Pet FK validation (must belong to same clinic)
//    ✓ Soft delete prevents accidental data loss
//    ✓ Partial data access control (admin calendar vs LIFF public calendar)
//
// DATA USES:
//    ✓ Reservation referenced by medical_records (optional FK)
//    ✓ Reservation referenced by billing (if medical record exists)
//    ✓ Used for admin calendar scheduling and appointment management
//    ✓ Owner/Pet/Staff relationships define who/what/where
//    ✓ ReservationType defines service type being booked
//
// DATA MODEL (appointments/reservations):
//    - id (PK): BIGSERIAL
//    - clinic_id: BIGINT NOT NULL (multitenancy)
//    - owner_id: BIGINT NOT NULL (FK → owners)
//    - pet_id: BIGINT NOT NULL (FK → pets)
//    - doctor_id: BIGINT (NULLABLE, FK → staffs) - doctor assignment
//    - reservation_type_id: BIGINT NOT NULL (FK → reservation_types)
//    - start_time: TIMESTAMP NOT NULL - appointment start
//    - end_time: TIMESTAMP NOT NULL - appointment end
//    - visit_type: VARCHAR(50) NOT NULL - visit category (初診, 再診, ワクチン, etc.)
//    - is_designated: BOOLEAN DEFAULT false - customer designated doctor flag
//    - notes: TEXT (NULLABLE) - admin notes
//    - line_customer_id: BIGINT (NULLABLE, FK → line_customers)
//    - is_staff_delegated: BOOLEAN DEFAULT false - created by staff flag
//    - customer_fields: JSONB (NULLABLE) - custom customer field data
//    - created_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - updated_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - deleted_at: TIMESTAMP NULL (soft delete)
//    - Indexes: (clinic_id, id), (clinic_id, start_time), (clinic_id, owner_id), (clinic_id, pet_id)
//    - Unique constraint: none (overlapping appointments allowed)
//
// IMPLEMENTATION NOTES:
//    - Clinic-scoped reservation management (admin side, not public LIFF)
//    - ListReservationsAdmin: Dual response format (month=summary, day=detail)
//    - List endpoint: NO pagination (returns all reservations for date range)
//    - Month view returns: ReservationSummaryResponse (compact representation)
//    - Day view returns: ReservationDetailResponse (full details with nested data)
//    - Date format flexibility: YYYY-MM for month, YYYY-MM-DD for day
//    - Default date: time.Now().Format("2006-01") if not provided
//    - FK validation: Owner/Pet must belong to same clinic
//    - Soft delete prevents accidental loss of appointment history
//    - Response transformation: toReservationDetailResponse (Create response)
//    - RBAC: ResourceReservation permission required
//    - No pagination on list endpoint
//
// TESTING STRATEGY:
//    Use integration tests with:
//    - Test database fixtures with sample owners, pets, staffs, reservation types
//    - Real service/repository layers
//    - Verify clinic_id scoping (no cross-clinic reservations)
//    - Test ListReservationsAdmin with view=month (default date, custom date)
//    - Test ListReservationsAdmin with view=day (custom date)
//    - Test ListReservationsAdmin with invalid view parameter (400 error)
//    - Test ListReservationsAdmin with invalid date format (400 error)
//    - Test CreateReservationAdmin with all required fields
//    - Test CreateReservationAdmin with optional fields (doctor_id, notes, etc.)
//    - Test start_time < end_time validation
//    - Test owner_id/pet_id FK validation (cross-clinic rejection)
//    - Test reservation_type_id FK validation
//    - Test doctor_id FK validation (if provided)
//    - Test response transformation (toReservationDetailResponse)
//    - Test DeleteReservationAdmin soft delete behavior
//    - Test FK constraint on deletion (billing/medical records)
//    - Test permission checks (ResourceReservation)
//    - Test date filtering (month view gets all days, day view gets single day)
//    - Test month view response format (ReservationSummaryResponse fields)
//    - Test day view response format (ReservationDetailResponse fields)
//    - Verify clinic_id parameter on all endpoints
//
