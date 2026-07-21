package billing

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

// ---- mock CashRegisterService ----

type mockCashRegisterService struct {
	getPreviewFn   func(ctx context.Context, clinicID uint64, dateStr, period string) (*CashRegisterPreview, error)
	closeFn        func(ctx context.Context, clinicID uint64, input CloseRegisterInput) (*model.CashRegisterClose, error)
	listFn         func(ctx context.Context, clinicID uint64, startDate, endDate *time.Time, page, limit int) ([]model.CashRegisterClose, int64, error)
	getByIDFn      func(ctx context.Context, clinicID, id uint64) (*model.CashRegisterClose, error)
	isDateClosedFn func(ctx context.Context, clinicID uint64, date time.Time) (bool, error)
}

func (m *mockCashRegisterService) GetPreview(ctx context.Context, clinicID uint64, dateStr, period string) (*CashRegisterPreview, error) {
	return m.getPreviewFn(ctx, clinicID, dateStr, period)
}

func (m *mockCashRegisterService) Close(ctx context.Context, clinicID uint64, input CloseRegisterInput) (*model.CashRegisterClose, error) {
	return m.closeFn(ctx, clinicID, input)
}

func (m *mockCashRegisterService) List(ctx context.Context, clinicID uint64, startDate, endDate *time.Time, page, limit int) ([]model.CashRegisterClose, int64, error) {
	return m.listFn(ctx, clinicID, startDate, endDate, page, limit)
}

func (m *mockCashRegisterService) GetByID(ctx context.Context, clinicID, id uint64) (*model.CashRegisterClose, error) {
	return m.getByIDFn(ctx, clinicID, id)
}

func (m *mockCashRegisterService) IsDateClosed(ctx context.Context, clinicID uint64, date time.Time) (bool, error) {
	if m.isDateClosedFn != nil {
		return m.isDateClosedFn(ctx, clinicID, date)
	}
	return false, nil
}

// ---- helper ----

func newHandlerWithCashRegisterSvc(svc CashRegisterService) *CashRegisterHandler {
	return NewCashRegisterHandler(svc, func(_, _ string) gin.HandlerFunc { return func(_ *gin.Context) {} })
}

// ---- GetCashRegisterPreview ----

func TestGetCashRegisterPreview(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockCashRegisterService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns preview",
			query:    "date=2026-05-28&period=am",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCashRegisterService{
				getPreviewFn: func(_ context.Context, clinicID uint64, dateStr, period string) (*CashRegisterPreview, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "2026-05-28", dateStr)
					assert.Equal(t, "am", period)
					return &CashRegisterPreview{Date: dateStr, Period: period}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"date":"2026-05-28"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCashRegisterService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			query:    "date=2026-05-28&period=am",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCashRegisterService{
				getPreviewFn: func(_ context.Context, _ uint64, _, _ string) (*CashRegisterPreview, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCashRegisterSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)
			h.GetCashRegisterPreview(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CloseCashRegister ----

func TestCloseCashRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{
			"date":        "2026-05-28",
			"period":      "am",
			"actual_cash": 12000,
			"memo":        "memo",
		}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockCashRegisterService
		wantStatus int
		wantHeader bool
	}{
		{
			name:     "closes cash register successfully",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc: &mockCashRegisterService{
				closeFn: func(_ context.Context, clinicID uint64, input CloseRegisterInput) (*model.CashRegisterClose, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "am", input.Period)
					assert.NotNil(t, input.ClosedBy)
					assert.Equal(t, uint64(1), *input.ClosedBy)
					return &model.CashRegisterClose{ID: 5, ClinicID: clinicID, Period: input.Period}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantHeader: true,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCashRegisterService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 401 when staff/user_id is missing",
			body:       validBody(),
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCashRegisterService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when date is missing",
			body:       map[string]any{"period": "am"},
			setupCtx:   func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc:        &mockCashRegisterService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when date format is invalid",
			body:       map[string]any{"date": "2026/05/28", "period": "am"},
			setupCtx:   func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc:        &mockCashRegisterService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc: &mockCashRegisterService{
				closeFn: func(_ context.Context, _ uint64, _ CloseRegisterInput) (*model.CashRegisterClose, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCashRegisterSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.CloseCashRegister(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantHeader {
				assert.NotEmpty(t, w.Header().Get("Location"))
			}
		})
	}
}

// ---- ListCashRegisterCloses ----

func TestListCashRegisterCloses(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockCashRegisterService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns paginated closes",
			query:    "start_date=2026-05-01&end_date=2026-05-31&page=1&limit=10",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCashRegisterService{
				listFn: func(_ context.Context, clinicID uint64, startDate, endDate *time.Time, page, limit int) ([]model.CashRegisterClose, int64, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.NotNil(t, startDate)
					assert.NotNil(t, endDate)
					assert.Equal(t, 1, page)
					assert.Equal(t, 10, limit)
					return []model.CashRegisterClose{{ID: 1, ClinicID: clinicID, Period: "am"}}, 1, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"period":"am"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCashRegisterService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid start_date filter",
			query:      "start_date=2026/05/01",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCashRegisterService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on invalid pagination",
			query:      "page=0",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCashRegisterService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCashRegisterService{
				listFn: func(_ context.Context, _ uint64, _, _ *time.Time, _, _ int) ([]model.CashRegisterClose, int64, error) {
					return nil, 0, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCashRegisterSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)
			h.ListCashRegisterCloses(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetCashRegisterClose ----

func TestGetCashRegisterClose(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockCashRegisterService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns close record",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCashRegisterService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.CashRegisterClose, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					return &model.CashRegisterClose{ID: id, ClinicID: clinicID, Period: "pm"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"period":"pm"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCashRegisterService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCashRegisterService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 on service error",
			paramID:  "99",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCashRegisterService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.CashRegisterClose, error) {
					return nil, apperrors.WrapNotFound("cash_register_close", "99")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCashRegisterSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.GetCashRegisterClose(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}
