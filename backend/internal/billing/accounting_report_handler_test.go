package billing

import (
	"context"
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

// ---- mock AccountingReportService ----

type mockAccountingReportService struct {
	getMonthlyFn               func(ctx context.Context, clinicID uint64, year, month int) (*MonthlyReportResponse, error)
	getMonthlyByPeriodFn       func(ctx context.Context, clinicID uint64, startDate, endDate time.Time) (*MonthlyReportResponse, error)
	exportMonthlyCSVFn         func(ctx context.Context, clinicID uint64, year, month int) (*MonthlyCSVResult, error)
	exportMonthlyCSVByPeriodFn func(ctx context.Context, clinicID uint64, startDate, endDate time.Time) (*MonthlyCSVResult, error)
}

func (m *mockAccountingReportService) GetMonthly(ctx context.Context, clinicID uint64, year, month int) (*MonthlyReportResponse, error) {
	return m.getMonthlyFn(ctx, clinicID, year, month)
}

func (m *mockAccountingReportService) GetMonthlyByPeriod(ctx context.Context, clinicID uint64, startDate, endDate time.Time) (*MonthlyReportResponse, error) {
	return m.getMonthlyByPeriodFn(ctx, clinicID, startDate, endDate)
}

func (m *mockAccountingReportService) ExportMonthlyCSV(ctx context.Context, clinicID uint64, year, month int) (*MonthlyCSVResult, error) {
	return m.exportMonthlyCSVFn(ctx, clinicID, year, month)
}

func (m *mockAccountingReportService) ExportMonthlyCSVByPeriod(ctx context.Context, clinicID uint64, startDate, endDate time.Time) (*MonthlyCSVResult, error) {
	return m.exportMonthlyCSVByPeriodFn(ctx, clinicID, startDate, endDate)
}

var _ AccountingReportService = (*mockAccountingReportService)(nil)

func newHandlerWithAccountingReportSvc(svc AccountingReportService) *AccountingReportHandler {
	return NewAccountingReportHandler(svc, func(_, _ string) gin.HandlerFunc { return func(_ *gin.Context) {} })
}

func sampleMonthlyReportResponse() *MonthlyReportResponse {
	return &MonthlyReportResponse{
		Year:      2026,
		Month:     4,
		StartDate: "2026-04-01",
		EndDate:   "2026-04-30",
		Summary: MonthlyReportSummary{
			WorkingDays:   20,
			TotalBillings: 100,
			TotalAmount:   500000,
			NetAmount:     490000,
		},
		DailyDetails: []DailyReportDetail{
			{Date: "2026-04-01", Weekday: "水", AMCount: 1, DayNet: 1000},
		},
	}
}

// ---- GetMonthlyReport ----

func TestGetMonthlyReport(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockAccountingReportService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns report for explicit year/month",
			query:    "year=2026&month=4",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockAccountingReportService{
				getMonthlyFn: func(_ context.Context, clinicID uint64, year, month int) (*MonthlyReportResponse, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, 2026, year)
					assert.Equal(t, 4, month)
					return sampleMonthlyReportResponse(), nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"year":2026`,
		},
		{
			name:     "returns report for explicit period",
			query:    "start_date=2026-04-01&end_date=2026-04-30",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockAccountingReportService{
				getMonthlyByPeriodFn: func(_ context.Context, clinicID uint64, startDate, endDate time.Time) (*MonthlyReportResponse, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "2026-04-01", startDate.Format("2006-01-02"))
					assert.Equal(t, "2026-04-30", endDate.Format("2006-01-02"))
					return sampleMonthlyReportResponse(), nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"start_date":"2026-04-01"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			query:      "year=2026&month=4",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockAccountingReportService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when month is invalid",
			query:      "year=2026&month=13",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockAccountingReportService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when period is incomplete",
			query:      "start_date=2026-04-01",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockAccountingReportService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error for year/month",
			query:    "year=2026&month=4",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockAccountingReportService{
				getMonthlyFn: func(_ context.Context, _ uint64, _, _ int) (*MonthlyReportResponse, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:     "returns 500 on service error for period",
			query:    "start_date=2026-04-01&end_date=2026-04-30",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockAccountingReportService{
				getMonthlyByPeriodFn: func(_ context.Context, _ uint64, _, _ time.Time) (*MonthlyReportResponse, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithAccountingReportSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/reports/monthly?"+tt.query, http.NoBody)
			tt.setupCtx(c)
			h.GetMonthlyReport(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- ExportMonthlyCSV ----

func TestExportMonthlyCSV(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		query        string
		setupCtx     func(c *gin.Context)
		svc          *mockAccountingReportService
		wantStatus   int
		wantFilename string
	}{
		{
			name:     "returns csv for explicit year/month",
			query:    "year=2026&month=4",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockAccountingReportService{
				exportMonthlyCSVFn: func(_ context.Context, clinicID uint64, year, month int) (*MonthlyCSVResult, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, 2026, year)
					assert.Equal(t, 4, month)
					return &MonthlyCSVResult{Filename: "monthly_report_202604.csv", Data: []byte("a,b\n1,2\n")}, nil
				},
			},
			wantStatus:   http.StatusOK,
			wantFilename: "monthly_report_202604.csv",
		},
		{
			name:     "returns csv for explicit period",
			query:    "start_date=2026-04-01&end_date=2026-04-30",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockAccountingReportService{
				exportMonthlyCSVByPeriodFn: func(_ context.Context, clinicID uint64, startDate, endDate time.Time) (*MonthlyCSVResult, error) {
					assert.Equal(t, uint64(1), clinicID)
					return &MonthlyCSVResult{Filename: "monthly_report_20260401_20260430.csv", Data: []byte("a,b\n1,2\n")}, nil
				},
			},
			wantStatus:   http.StatusOK,
			wantFilename: "monthly_report_20260401_20260430.csv",
		},
		{
			name:       "returns 401 when clinic_id is missing",
			query:      "year=2026&month=4",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockAccountingReportService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when month is invalid",
			query:      "year=2026&month=13",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockAccountingReportService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when period is incomplete",
			query:      "start_date=2026-04-01",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockAccountingReportService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error for year/month",
			query:    "year=2026&month=4",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockAccountingReportService{
				exportMonthlyCSVFn: func(_ context.Context, _ uint64, _, _ int) (*MonthlyCSVResult, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:     "returns 500 on service error for period",
			query:    "start_date=2026-04-01&end_date=2026-04-30",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockAccountingReportService{
				exportMonthlyCSVByPeriodFn: func(_ context.Context, _ uint64, _, _ time.Time) (*MonthlyCSVResult, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithAccountingReportSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/reports/monthly/csv?"+tt.query, http.NoBody)
			tt.setupCtx(c)
			h.ExportMonthlyCSV(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantFilename != "" {
				assert.Contains(t, w.Header().Get("Content-Disposition"), tt.wantFilename)
				assert.Contains(t, w.Header().Get("Content-Type"), "text/csv")
			}
		})
	}
}

// ---- toMonthlyReportResponse ----

func TestToMonthlyReportResponse(t *testing.T) {
	t.Run("maps all fields including tax breakdown and daily details", func(t *testing.T) {
		resp := toMonthlyReportResponse(&MonthlyReportResponse{
			Year:      2026,
			Month:     4,
			StartDate: "2026-04-01",
			EndDate:   "2026-04-30",
			Summary: MonthlyReportSummary{
				WorkingDays:     20,
				TotalBillings:   100,
				TotalAmount:     500000,
				TotalRefund:     1000,
				NetAmount:       499000,
				ByPaymentMethod: map[string]int64{"現金": 100000},
				ByCategory:      map[string]int64{string(model.ItemCategoryExamination): 200000},
				TaxBreakdown: TaxBreakdownSummary{
					Standard: TaxBreakdownEntry{TaxableAmount: 400000, TaxAmount: 40000},
					Reduced:  TaxBreakdownEntry{TaxableAmount: 100000, TaxAmount: 8000},
				},
			},
			DailyDetails: []DailyReportDetail{
				{
					Date:      "2026-04-01",
					Weekday:   "水",
					AMCount:   1,
					AMNet:     1000,
					PMCount:   2,
					PMNet:     2000,
					DayNet:    3000,
					Refund:    100,
					AMClosed:  true,
					PMClosed:  false,
					IsHoliday: false,
				},
			},
		})

		assert.Equal(t, 2026, resp.Year)
		assert.Equal(t, 4, resp.Month)
		assert.Equal(t, "2026-04-01", resp.StartDate)
		assert.Equal(t, "2026-04-30", resp.EndDate)
		assert.Equal(t, 20, resp.Summary.WorkingDays)
		assert.Equal(t, int64(100), resp.Summary.TotalBillings)
		assert.Equal(t, int64(500000), resp.Summary.TotalAmount)
		assert.Equal(t, int64(1000), resp.Summary.TotalRefund)
		assert.Equal(t, int64(499000), resp.Summary.NetAmount)
		assert.Equal(t, int64(100000), resp.Summary.ByPaymentMethod["現金"])
		assert.Equal(t, int64(200000), resp.Summary.ByCategory[string(model.ItemCategoryExamination)])
		assert.Equal(t, int64(400000), resp.Summary.TaxBreakdown.Standard.TaxableAmount)
		assert.Equal(t, int64(40000), resp.Summary.TaxBreakdown.Standard.TaxAmount)
		assert.Equal(t, int64(100000), resp.Summary.TaxBreakdown.Reduced.TaxableAmount)
		assert.Equal(t, int64(8000), resp.Summary.TaxBreakdown.Reduced.TaxAmount)
		require.Len(t, resp.DailyDetails, 1)
		d := resp.DailyDetails[0]
		assert.Equal(t, "2026-04-01", d.Date)
		assert.Equal(t, "水", d.Weekday)
		assert.Equal(t, int64(1), d.AMCount)
		assert.Equal(t, int64(1000), d.AMNet)
		assert.Equal(t, int64(2), d.PMCount)
		assert.Equal(t, int64(2000), d.PMNet)
		assert.Equal(t, int64(3000), d.DayNet)
		assert.Equal(t, int64(100), d.Refund)
		assert.True(t, d.AMClosed)
		assert.False(t, d.PMClosed)
		assert.False(t, d.IsHoliday)
	})

	t.Run("handles empty daily details", func(t *testing.T) {
		resp := toMonthlyReportResponse(&MonthlyReportResponse{
			Year:         2026,
			Month:        4,
			StartDate:    "2026-04-01",
			EndDate:      "2026-04-30",
			DailyDetails: nil,
		})

		assert.Empty(t, resp.DailyDetails)
		assert.NotNil(t, resp.DailyDetails)
	})
}
