package clinic

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

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mock ClinicHolidayService ----

type mockClinicHolidayService struct {
	listFn   func(ctx context.Context, clinicID uint64, yearMonth string) ([]model.ClinicHoliday, error)
	setFn    func(ctx context.Context, clinicID uint64, date time.Time, reason string) (*model.ClinicHoliday, error)
	removeFn func(ctx context.Context, clinicID uint64, date time.Time) error
}

func (m *mockClinicHolidayService) List(ctx context.Context, clinicID uint64, yearMonth string) ([]model.ClinicHoliday, error) {
	return m.listFn(ctx, clinicID, yearMonth)
}

func (m *mockClinicHolidayService) Set(ctx context.Context, clinicID uint64, date time.Time, reason string) (*model.ClinicHoliday, error) {
	return m.setFn(ctx, clinicID, date, reason)
}

func (m *mockClinicHolidayService) Remove(ctx context.Context, clinicID uint64, date time.Time) error {
	return m.removeFn(ctx, clinicID, date)
}

func newHandlerWithClinicHolidaySvc(svc ClinicHolidayService) *Handler {
	return &Handler{holidaySvc: svc}
}

// ---- ListClinicHolidays ----

func TestListClinicHolidays(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockClinicHolidayService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of clinic holidays",
			query:    "year_month=2026-05",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockClinicHolidayService{
				listFn: func(_ context.Context, clinicID uint64, yearMonth string) ([]model.ClinicHoliday, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "2026-05", yearMonth)
					return []model.ClinicHoliday{{ID: 1, ClinicID: clinicID, Reason: "祝日"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"reason":"祝日"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockClinicHolidayService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockClinicHolidayService{
				listFn: func(_ context.Context, _ uint64, _ string) ([]model.ClinicHoliday, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithClinicHolidaySvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)
			h.ListClinicHolidays(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- SetClinicHoliday ----

func TestSetClinicHoliday(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockClinicHolidayService
		wantStatus int
		wantBody   string
		wantHeader string
	}{
		{
			name:     "sets clinic holiday successfully",
			body:     map[string]any{"date": "2026-05-05", "reason": "こどもの日"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockClinicHolidayService{
				setFn: func(_ context.Context, clinicID uint64, date time.Time, reason string) (*model.ClinicHoliday, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "2026-05-05", date.Format("2006-01-02"))
					assert.Equal(t, "こどもの日", reason)
					return &model.ClinicHoliday{ID: 1, ClinicID: clinicID, Date: date, Reason: reason}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"reason":"こどもの日"`,
			wantHeader: "/v1/clinic-holidays/2026-05-05",
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       map[string]any{"date": "2026-05-05"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockClinicHolidayService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when date is missing",
			body:       map[string]any{"reason": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockClinicHolidayService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid date format",
			body:       map[string]any{"date": "2026/05/05"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockClinicHolidayService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockClinicHolidayService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			body:     map[string]any{"date": "2026-05-05"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockClinicHolidayService{
				setFn: func(_ context.Context, _ uint64, _ time.Time, _ string) (*model.ClinicHoliday, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithClinicHolidaySvc(tt.svc)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.SetClinicHoliday(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
			if tt.wantHeader != "" {
				assert.Equal(t, tt.wantHeader, w.Header().Get("Location"))
			}
		})
	}
}

// ---- DeleteClinicHoliday ----

func TestDeleteClinicHoliday(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramDate  string
		setupCtx   func(c *gin.Context)
		svc        *mockClinicHolidayService
		wantStatus int
	}{
		{
			name:      "deletes clinic holiday successfully",
			paramDate: "2026-05-05",
			setupCtx:  func(c *gin.Context) { setClinicID(c) },
			svc: &mockClinicHolidayService{
				removeFn: func(_ context.Context, clinicID uint64, date time.Time) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "2026-05-05", date.Format("2006-01-02"))
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramDate:  "2026-05-05",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockClinicHolidayService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for invalid date format",
			paramDate:  "invalid-date",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockClinicHolidayService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "returns 500 on service error",
			paramDate: "2026-05-05",
			setupCtx:  func(c *gin.Context) { setClinicID(c) },
			svc: &mockClinicHolidayService{
				removeFn: func(_ context.Context, _ uint64, _ time.Time) error {
					return fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithClinicHolidaySvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
			c.Params = gin.Params{{Key: "date", Value: tt.paramDate}}
			tt.setupCtx(c)

			h.DeleteClinicHoliday(c)
			c.Writer.WriteHeaderNow() // flush a bare c.Status() (no body) to the recorder

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
