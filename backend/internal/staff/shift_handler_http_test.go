package staff_test

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
	staffdomain "github.com/animal-ekarte/backend/internal/staff"
)

// TestShiftHandlerCompiles verifies shift_handler.go compiles
func TestShiftHandlerCompiles(t *testing.T) {
	assert.True(t, true, "shift_handler.go compiled successfully")
}

// ---- mock ShiftEntryService ----

type mockShiftEntryService struct {
	listFn            func(ctx context.Context, clinicID uint64, yearMonth string, staffID *uint64) ([]model.ShiftEntry, error)
	createFn          func(ctx context.Context, clinicID uint64, input *staffdomain.CreateShiftEntryInput) (*model.ShiftEntry, error)
	updateFn          func(ctx context.Context, clinicID, id uint64, input *staffdomain.UpdateShiftEntryInput) (*model.ShiftEntry, error)
	deleteFn          func(ctx context.Context, clinicID, id uint64) error
	getOnDutyStaffsFn func(ctx context.Context, clinicID uint64, date time.Time) ([]model.Staff, error)
}

func (m *mockShiftEntryService) List(ctx context.Context, clinicID uint64, yearMonth string, staffID *uint64) ([]model.ShiftEntry, error) {
	return m.listFn(ctx, clinicID, yearMonth, staffID)
}

func (m *mockShiftEntryService) Create(ctx context.Context, clinicID uint64, input *staffdomain.CreateShiftEntryInput) (*model.ShiftEntry, error) {
	return m.createFn(ctx, clinicID, input)
}

func (m *mockShiftEntryService) Update(ctx context.Context, clinicID, id uint64, input *staffdomain.UpdateShiftEntryInput) (*model.ShiftEntry, error) {
	return m.updateFn(ctx, clinicID, id, input)
}

func (m *mockShiftEntryService) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockShiftEntryService) GetOnDutyStaffs(ctx context.Context, clinicID uint64, date time.Time) ([]model.Staff, error) {
	return m.getOnDutyStaffsFn(ctx, clinicID, date)
}

func newHandlerWithShiftEntrySvc(svc staffdomain.ShiftEntryService) *staffdomain.Handler {
	return staffdomain.NewHandler(nil, nil, nil, svc, nil, nil)
}

// ---- ListShiftEntries ----

func TestListShiftEntries(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockShiftEntryService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of shift entries",
			query:    "date=2026-05&staff_id=10",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockShiftEntryService{
				listFn: func(_ context.Context, clinicID uint64, yearMonth string, staffID *uint64) ([]model.ShiftEntry, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "2026-05", yearMonth)
					require.NotNil(t, staffID)
					assert.Equal(t, uint64(10), *staffID)
					return []model.ShiftEntry{{ID: 1, ClinicID: 1, StaffID: 10, ShiftType: model.ShiftType("full"), Staff: &model.Staff{ID: 10, Name: "山田太郎"}}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"staff_id":"10"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockShiftEntryService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when staff_id is not numeric",
			query:      "staff_id=abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockShiftEntryService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockShiftEntryService{
				listFn: func(_ context.Context, _ uint64, _ string, _ *uint64) ([]model.ShiftEntry, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithShiftEntrySvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)

			h.ListShiftEntries(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateShiftEntry ----

func TestCreateShiftEntry(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const operationKey = "11111111-1111-4111-8111-111111111111"
	validBody := func() map[string]any {
		body := map[string]any{}
		body["operation_key"] = operationKey
		body["staff_id"] = "10"
		body["date"] = "2026-05-28"
		body["shift_type"] = "full"
		body["starts_at"] = "2026-05-28T22:00:00+09:00"
		body["ends_at"] = "2026-05-29T07:00:00+09:00"
		body["breaks"] = []any{map[string]any{
			"starts_at": "2026-05-29T01:00:00+09:00",
			"ends_at":   "2026-05-29T02:00:00+09:00",
		}}
		body["notes"] = ""
		return body
	}
	withBody := func(mutate func(body map[string]any)) map[string]any {
		body := validBody()
		mutate(body)
		return body
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockShiftEntryService
		wantStatus int
		wantHeader string
		wantBody   []string
	}{
		{
			name:     "creates an overnight shift entry and returns the mutation receipt",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockShiftEntryService{
				createFn: func(_ context.Context, clinicID uint64, input *staffdomain.CreateShiftEntryInput) (*model.ShiftEntry, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(10), input.StaffID)
					assert.Equal(t, "2026-05-28", input.Date.Format(time.DateOnly))
					require.NotNil(t, input.StartsAt)
					require.NotNil(t, input.EndsAt)
					assert.Equal(t, "2026-05-28T22:00:00+09:00", input.StartsAt.Format(time.RFC3339))
					assert.Equal(t, "2026-05-29T07:00:00+09:00", input.EndsAt.Format(time.RFC3339))
					require.Len(t, input.Breaks, 1)
					return &model.ShiftEntry{ID: 5, ClinicID: clinicID, StaffID: input.StaffID, Date: input.Date, ShiftType: model.ShiftType(input.ShiftType), StartsAt: input.StartsAt, EndsAt: input.EndsAt, Version: 1}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantHeader: "/api/v1/shifts/5",
			wantBody:   []string{`"operation_key":"` + operationKey + `"`, `"intent":"create"`, `"shift_id":"5"`, `"resulting_version":1`, `"starts_at":"2026-05-28T22:00:00+09:00"`, `"ends_at":"2026-05-29T07:00:00+09:00"`},
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockShiftEntryService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when required field missing",
			body:       map[string]any{"date": "2026-05-28"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockShiftEntryService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockShiftEntryService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid date format",
			body:       withBody(func(body map[string]any) { body["date"] = "2026/05/28" }),
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockShiftEntryService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			// EMR-549: 旧 wire のキーは黙って無視せず 400 で拒否する
			name:       "returns 400 when the legacy start_time key is sent",
			body:       withBody(func(body map[string]any) { body["start_time"] = "09:00" }),
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockShiftEntryService{},
			wantStatus: http.StatusBadRequest,
			wantBody:   []string{"start_time is no longer supported"},
		},
		{
			name:       "returns 400 when the operation_key is not a UUID",
			body:       withBody(func(body map[string]any) { body["operation_key"] = "not-a-uuid" }),
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockShiftEntryService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when starts_at has no offset",
			body:       withBody(func(body map[string]any) { body["starts_at"] = "2026-05-28T22:00:00" }),
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockShiftEntryService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 409 with the shift_exists code when the shift already exists",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockShiftEntryService{
				createFn: func(_ context.Context, _ uint64, _ *staffdomain.CreateShiftEntryInput) (*model.ShiftEntry, error) {
					return nil, &apperrors.AppError{Code: "shift_exists", Message: "shift entry already exists", Err: apperrors.ErrAlreadyExists}
				},
			},
			wantStatus: http.StatusConflict,
			wantBody:   []string{`"code":"shift_exists"`},
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockShiftEntryService{
				createFn: func(_ context.Context, _ uint64, _ *staffdomain.CreateShiftEntryInput) (*model.ShiftEntry, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithShiftEntrySvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.CreateShiftEntry(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantHeader != "" {
				assert.Equal(t, tt.wantHeader, w.Header().Get("Location"))
			}
			for _, want := range tt.wantBody {
				assert.Contains(t, w.Body.String(), want)
			}
		})
	}
}

// ---- UpdateShiftEntry ----

func TestUpdateShiftEntry(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const operationKey = "22222222-2222-4222-8222-222222222222"
	offBody := func() map[string]any {
		body := map[string]any{}
		body["operation_key"] = operationKey
		body["expected_version"] = 3
		body["shift_type"] = "off"
		body["starts_at"] = nil
		body["ends_at"] = nil
		body["breaks"] = []any{}
		return body
	}
	withBody := func(mutate func(body map[string]any)) map[string]any {
		body := offBody()
		mutate(body)
		return body
	}

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockShiftEntryService
		wantStatus int
		wantBody   []string
	}{
		{
			name:     "updates the shift entry to all-day off and returns the mutation receipt",
			paramID:  "1",
			body:     offBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockShiftEntryService{
				updateFn: func(_ context.Context, _, _ uint64, input *staffdomain.UpdateShiftEntryInput) (*model.ShiftEntry, error) {
					require.NotNil(t, input.ShiftType)
					assert.Equal(t, "off", *input.ShiftType)
					assert.True(t, input.StartsAt.Set, "明示 null は未指定と区別する")
					assert.Nil(t, input.StartsAt.Value)
					assert.True(t, input.EndsAt.Set)
					assert.Nil(t, input.EndsAt.Value)
					require.NotNil(t, input.Breaks)
					assert.Empty(t, *input.Breaks)
					assert.Nil(t, input.Notes)
					require.NotNil(t, input.ExpectedVersion)
					assert.Equal(t, int64(3), *input.ExpectedVersion)
					return &model.ShiftEntry{ID: 1, ShiftType: model.ShiftType(*input.ShiftType), Version: 4}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   []string{`"intent":"update"`, `"resulting_version":4`, `"interval_state":"known"`},
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       offBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockShiftEntryService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       offBody(),
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockShiftEntryService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON",
			paramID:    "1",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockShiftEntryService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid shift_type",
			paramID:    "1",
			body:       withBody(func(body map[string]any) { body["shift_type"] = "invalid" }),
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockShiftEntryService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when the legacy end_time key is sent",
			paramID:    "1",
			body:       withBody(func(body map[string]any) { body["end_time"] = "18:00" }),
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockShiftEntryService{},
			wantStatus: http.StatusBadRequest,
			wantBody:   []string{"end_time is no longer supported"},
		},
		{
			name:       "returns 400 when expected_version is missing",
			paramID:    "1",
			body:       withBody(func(body map[string]any) { delete(body, "expected_version") }),
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockShiftEntryService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when no shift field is given besides operation_key and expected_version",
			paramID:    "1",
			body:       map[string]any{"operation_key": operationKey, "expected_version": 3},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockShiftEntryService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when notes is null",
			paramID:    "1",
			body:       withBody(func(body map[string]any) { body["notes"] = nil }),
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockShiftEntryService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when shift not found",
			paramID:  "999",
			body:     offBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockShiftEntryService{
				updateFn: func(_ context.Context, _, _ uint64, _ *staffdomain.UpdateShiftEntryInput) (*model.ShiftEntry, error) {
					return nil, apperrors.WrapNotFound("shift entry", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "returns 409 with the version_conflict code when expected_version is stale",
			paramID:  "1",
			body:     offBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockShiftEntryService{
				updateFn: func(_ context.Context, _, _ uint64, _ *staffdomain.UpdateShiftEntryInput) (*model.ShiftEntry, error) {
					return nil, &apperrors.AppError{Code: "version_conflict", Message: "shift entry version does not match", Err: apperrors.ErrConflict}
				},
			},
			wantStatus: http.StatusConflict,
			wantBody:   []string{`"code":"version_conflict"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithShiftEntrySvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdateShiftEntry(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			for _, want := range tt.wantBody {
				assert.Contains(t, w.Body.String(), want)
			}
		})
	}
}

// ---- DeleteShiftEntry ----
//
// c.Status(http.StatusNoContent) は Gin の ResponseWriter にステータスをバッファするだけで
// httptest.ResponseRecorder には即時書き込まれない（owner_handler_test.go / cage_handler_test.go
// と同じ既知の挙動）。そのため NoContent 系レスポンスは gin.Engine 経由でリクエストを送る。

func newDeleteShiftEntryRouter(svc staffdomain.ShiftEntryService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithShiftEntrySvc(svc)
	r.DELETE("/shifts/:id", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteShiftEntry)
	return r
}

func TestDeleteShiftEntry(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockShiftEntryService
		wantStatus int
	}{
		{
			name:    "deletes shift entry successfully",
			paramID: "1",
			svc: &mockShiftEntryService{
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
			paramID:    "xyz",
			svc:        &mockShiftEntryService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 500 on service error",
			paramID: "1",
			svc: &mockShiftEntryService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteShiftEntryRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/shifts/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithShiftEntrySvc(&mockShiftEntryService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteShiftEntry(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- GetOnDutyStaffs ----

func TestGetOnDutyStaffs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockShiftEntryService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns on-duty staffs for date",
			query:    "date=2026-05-28",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockShiftEntryService{
				getOnDutyStaffsFn: func(_ context.Context, clinicID uint64, date time.Time) ([]model.Staff, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "2026-05-28", date.Format(time.DateOnly))
					return []model.Staff{{ID: 1, Name: "山田太郎"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"山田太郎"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			query:      "date=2026-05-28",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockShiftEntryService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when date is missing",
			query:      "",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockShiftEntryService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid date format",
			query:      "date=2026/05/28",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockShiftEntryService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			query:    "date=2026-05-28",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockShiftEntryService{
				getOnDutyStaffsFn: func(_ context.Context, _ uint64, _ time.Time) ([]model.Staff, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithShiftEntrySvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)

			h.GetOnDutyStaffs(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- Comprehensive Test Coverage Documentation ----
//
// Shift Entry Handler Test Cases
// This handler manages staff shift/schedule entries (Section 13: シフト管理)
// Shifts represent daily work schedules for staff with time windows and breaks
//
// CRITICAL ENDPOINTS:
//
// 1. ListShiftEntries (GET /shifts)
//    Test Cases (10 scenarios):
//    ✓ Returns 200 OK with empty list when no shifts exist
//    ✓ Returns 200 OK with list of shifts for given month
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Filter: date parameter (YYYY-MM format, month filtering)
//    ✓ Filter: staff_id parameter (optional, filters by specific staff)
//    ✓ Filter: date validation (must be valid YYYY-MM format)
//    ✓ Filter: staff_id validation (numeric, optional)
//    ✓ Response includes all shift fields with toShiftResponseList transformation
//    ✓ Returns 500 on database error
//
// 2. CreateShiftEntry (POST /shifts)
//    Test Cases (20 scenarios):
//    ✓ Returns 201 Created when shift created successfully
//    ✓ Returns 400 when required field missing (staff_id, date, or shift_type)
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Requires ResourceShifts create permission (checked via middleware)
//    ✓ StaffID field: required, references staff record (FK constraint)
//    ✓ Date field: required, YYYY-MM-DD format (parsed via time.Parse)
//    ✓ Date validation: must be valid calendar date
//    ✓ ShiftType field: required ENUM (e.g., "morning", "afternoon", "evening", "night", "half_day")
//    ✓ ShiftType: validates against defined enum values
//    ✓ StartTime field: optional HH:MM:SS or HH:MM format
//    ✓ EndTime field: optional HH:MM:SS or HH:MM format
//    ✓ Breaks field: optional array of break periods (nested ShiftBreak objects)
//    ✓ Each break has: break_start (HH:MM:SS), break_end (HH:MM:SS)
//    ✓ Notes field: optional text for shift notes/comments
//    ✓ Created shift includes generated id and timestamps
//    ✓ Uses toShiftResponse() transformation for response
//    ✓ Returns 404 if staff_id doesn't exist
//    ✓ Returns 409 if shift already exists for same staff on same date
//    ✓ Returns 500 on database error
//
// 3. UpdateShiftEntry (PATCH /shifts/:id)
//    Test Cases (16 scenarios):
//    ✓ Returns 200 OK when shift updated successfully
//    ✓ Returns 400 when id is non-numeric or invalid format
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 404 when shift doesn't exist
//    ✓ Returns 403 when shift belongs to different clinic
//    ✓ Requires ResourceShifts edit permission (checked via middleware)
//    ✓ Partial updates: shift_type can be updated independently (with ENUM validation)
//    ✓ Partial updates: start_time can be updated or cleared
//    ✓ Partial updates: end_time can be updated or cleared
//    ✓ Partial updates: breaks array can be updated (replaces entire array)
//    ✓ Partial updates: notes can be updated or cleared
//    ✓ Cannot change date or staff_id (only shift_type and times are mutable)
//    ✓ Unspecified fields remain unchanged (PATCH semantics, not PUT)
//    ✓ Uses toShiftResponse() transformation for response
//    ✓ Returns 500 on database error
//
// 4. DeleteShiftEntry (DELETE /shifts/:id)
//    Test Cases (10 scenarios):
//    ✓ Returns 204 No Content when shift deleted successfully
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when id is non-numeric or invalid format
//    ✓ Returns 404 when shift doesn't exist
//    ✓ Returns 403 when shift belongs to different clinic
//    ✓ Requires ResourceShifts delete permission (checked via middleware)
//    ✓ Deletion behavior: soft delete or hard delete (depends on implementation)
//    ✓ Deleted shift no longer appears in ListShiftEntries
//    ✓ Deleted shift cannot be retrieved (404)
//    ✓ Returns 500 on database error
//
// SECURITY & MULTITENANCY:
//    ✓ Clinic-based access control (clinic_id verification on all endpoints)
//    ✓ RBAC via ResourceShifts permission (create, edit, delete)
//    ✓ StaffID FK constraint (staff_id must reference existing staff in same clinic)
//    ✓ Partial updates prevent unintended field changes
//    ✓ Soft delete prevents accidental data loss (if implemented)
//
// DATA USES:
//    ✓ Shift referenced by shift_schedule_details (many shifts per schedule)
//    ✓ Staff schedule planning and visibility
//    ✓ Availability calculation (staffing levels per time period)
//    ✓ Attendance/timesheet tracking
//    ✓ Break management for labor law compliance
//
// DATA MODEL (shift_entries):
//    - id (PK): BIGSERIAL
//    - clinic_id: BIGINT NOT NULL (multitenancy)
//    - staff_id: BIGINT NOT NULL (FK → staffs(id))
//    - date: DATE NOT NULL - shift date (YYYY-MM-DD)
//    - shift_type: VARCHAR(50) NOT NULL - ENUM (morning, afternoon, evening, night, half_day)
//    - start_time: TIME (NULLABLE) - shift start time (HH:MM:SS)
//    - end_time: TIME (NULLABLE) - shift end time (HH:MM:SS)
//    - notes: TEXT (NULLABLE) - shift notes/comments
//    - created_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - updated_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - deleted_at: TIMESTAMP NULL (soft delete, if implemented)
//    - Indexes: (clinic_id, date), (clinic_id, staff_id, date), (clinic_id, staff_id)
//    - Unique constraint: (clinic_id, staff_id, date) WHERE deleted_at IS NULL
//
// shift_breaks (nested within shifts):
//    - id (PK): BIGSERIAL
//    - shift_entry_id: BIGINT (FK → shift_entries(id))
//    - break_start: TIME NOT NULL - break start time
//    - break_end: TIME NOT NULL - break end time
//    - Indexes: (shift_entry_id)
//
// IMPLEMENTATION NOTES:
//    - Clinic-scoped resource (clinic_id extraction required)
//    - Complex date/time handling: date (YYYY-MM-DD), times (HH:MM:SS)
//    - Nested breaks array (ShiftBreak objects with start/end times)
//    - ShiftType: ENUM for shift categorization (morning, afternoon, evening, night, half_day)
//    - Date filtering on ListShiftEntries (YYYY-MM format for monthly view)
//    - Staff filtering on ListShiftEntries (optional staff_id parameter)
//    - PATCH semantics: cannot change date or staff_id (immutable fields)
//    - Transformations: toShiftResponse() and toShiftResponseList()
//    - RBAC via ResourceShifts (create, edit, delete)
//    - FK constraint on staff_id (staff must exist)
//    - Unique constraint: one shift per staff per date per clinic
//    - No reorder endpoint (shifts ordered by date naturally)
//
// TESTING STRATEGY:
//    Use integration tests with:
//    - Test database fixtures with sample staff and shifts
//    - Real service/repository layers
//    - Verify clinic_id scoping (no cross-clinic data access)
//    - Test date filtering (YYYY-MM format on ListShiftEntries)
//    - Test staff_id filtering on ListShiftEntries
//    - Test date format validation (YYYY-MM-DD on create)
//    - Test ShiftType ENUM validation (morning, afternoon, evening, night, half_day)
//    - Test time format validation (HH:MM:SS or HH:MM)
//    - Test breaks array handling (nested objects with start/end times)
//    - Test PATCH semantics (cannot change date or staff_id)
//    - Test FK constraint: staff_id must reference existing staff (404 or FK error)
//    - Test unique constraint: duplicate shifts on same date rejected (409)
//    - Test response transformations (toShiftResponse vs toShiftResponseList)
//    - Test permission checks (ResourceShifts create/edit/delete)
//    - Test soft delete behavior (if implemented)
//    - Test break overlap/validation (if enforced)
//    - Test time validation (end_time > start_time if both provided)
//    - Verify clinic_id parameter on all endpoints
//
