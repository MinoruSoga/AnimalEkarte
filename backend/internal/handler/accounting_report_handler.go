package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// GetMonthlyReport godoc
// GET /v1/reports/monthly?year=2026&month=4
func (h *Handler) GetMonthlyReport(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	query := newMonthlyReportQuery(c.Request.URL.Query())
	var result *service.MonthlyReportResponse
	var err error
	if query.hasPeriod() {
		startDate, endDate, parseErr := query.toPeriod()
		if parseErr != nil {
			RespondError(c, parseErr)
			return
		}
		result, err = h.svc.AccountingReport.GetMonthlyByPeriod(c.Request.Context(), clinicID, startDate, endDate)
	} else {
		year, month, parseErr := query.toYearMonth(time.Now())
		if parseErr != nil {
			RespondError(c, parseErr)
			return
		}
		result, err = h.svc.AccountingReport.GetMonthly(c.Request.Context(), clinicID, year, month)
	}
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMonthlyReportResponse(result))
}

// ExportMonthlyCSV godoc
// GET /v1/reports/monthly/csv?year=2026&month=4
func (h *Handler) ExportMonthlyCSV(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	query := newMonthlyReportQuery(c.Request.URL.Query())
	var csvResult *service.MonthlyCSVResult
	var err error
	if query.hasPeriod() {
		startDate, endDate, parseErr := query.toPeriod()
		if parseErr != nil {
			RespondError(c, parseErr)
			return
		}
		csvResult, err = h.svc.AccountingReport.ExportMonthlyCSVByPeriod(c.Request.Context(), clinicID, startDate, endDate)
	} else {
		year, month, parseErr := query.toYearMonth(time.Now())
		if parseErr != nil {
			RespondError(c, parseErr)
			return
		}
		csvResult, err = h.svc.AccountingReport.ExportMonthlyCSV(c.Request.Context(), clinicID, year, month)
	}
	if err != nil {
		RespondError(c, err)
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", csvResult.Filename))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", csvResult.Data)
}

// RegisterAccountingReportRoutes は売上レポート関連のルートを登録する
func (h *Handler) RegisterAccountingReportRoutes(rg *gin.RouterGroup) {
	reports := rg.Group("/reports")
	reports.GET("/monthly", h.RequirePermission(string(model.ResourceAccountingReports), "view"), h.GetMonthlyReport)
	reports.GET("/monthly/csv", h.RequirePermission(string(model.ResourceAccountingReports), "view"), h.ExportMonthlyCSV)
}

// monthlyTaxBreakdownEntryResponse は税率別の課税対象額・税額
type monthlyTaxBreakdownEntryResponse struct {
	TaxableAmount int64 `json:"taxable_amount"`
	TaxAmount     int64 `json:"tax_amount"`
}

// monthlyTaxBreakdownSummaryResponse は標準・軽減税率別サマリー
type monthlyTaxBreakdownSummaryResponse struct {
	Standard monthlyTaxBreakdownEntryResponse `json:"standard"`
	Reduced  monthlyTaxBreakdownEntryResponse `json:"reduced"`
}

// monthlyReportSummaryResponse は月次サマリー情報
type monthlyReportSummaryResponse struct {
	WorkingDays     int                                `json:"working_days"`
	TotalBillings   int64                              `json:"total_billings"`
	TotalAmount     int64                              `json:"total_amount"`
	TotalRefund     int64                              `json:"total_refund"`
	NetAmount       int64                              `json:"net_amount"`
	ByPaymentMethod map[string]int64                   `json:"by_payment_method"`
	ByCategory      map[string]int64                   `json:"by_category"`
	TaxBreakdown    monthlyTaxBreakdownSummaryResponse `json:"tax_breakdown"`
}

// dailyReportDetailResponse は日別明細
type dailyReportDetailResponse struct {
	Date      string `json:"date"`
	Weekday   string `json:"weekday"`
	AMCount   int64  `json:"am_count"`
	AMNet     int64  `json:"am_net"`
	PMCount   int64  `json:"pm_count"`
	PMNet     int64  `json:"pm_net"`
	DayNet    int64  `json:"day_net"`
	Refund    int64  `json:"refund"`
	AMClosed  bool   `json:"am_closed"`
	PMClosed  bool   `json:"pm_closed"`
	IsHoliday bool   `json:"is_holiday"`
}

// monthlyReportResponse は月次売上レポートのハンドラー向けレスポンス
type monthlyReportResponse struct {
	Year         int                          `json:"year"`
	Month        int                          `json:"month"`
	StartDate    string                       `json:"start_date"`
	EndDate      string                       `json:"end_date"`
	Summary      monthlyReportSummaryResponse `json:"summary"`
	DailyDetails []dailyReportDetailResponse  `json:"daily_details"`
}

func toMonthlyReportResponse(r *service.MonthlyReportResponse) monthlyReportResponse {
	dailyDetails := make([]dailyReportDetailResponse, 0, len(r.DailyDetails))
	for _, d := range r.DailyDetails {
		dailyDetails = append(dailyDetails, dailyReportDetailResponse{
			Date:      d.Date,
			Weekday:   d.Weekday,
			AMCount:   d.AMCount,
			AMNet:     d.AMNet,
			PMCount:   d.PMCount,
			PMNet:     d.PMNet,
			DayNet:    d.DayNet,
			Refund:    d.Refund,
			AMClosed:  d.AMClosed,
			PMClosed:  d.PMClosed,
			IsHoliday: d.IsHoliday,
		})
	}
	return monthlyReportResponse{
		Year:      r.Year,
		Month:     r.Month,
		StartDate: r.StartDate,
		EndDate:   r.EndDate,
		Summary: monthlyReportSummaryResponse{
			WorkingDays:     r.Summary.WorkingDays,
			TotalBillings:   r.Summary.TotalBillings,
			TotalAmount:     r.Summary.TotalAmount,
			TotalRefund:     r.Summary.TotalRefund,
			NetAmount:       r.Summary.NetAmount,
			ByPaymentMethod: r.Summary.ByPaymentMethod,
			ByCategory:      r.Summary.ByCategory,
			TaxBreakdown: monthlyTaxBreakdownSummaryResponse{
				Standard: monthlyTaxBreakdownEntryResponse{
					TaxableAmount: r.Summary.TaxBreakdown.Standard.TaxableAmount,
					TaxAmount:     r.Summary.TaxBreakdown.Standard.TaxAmount,
				},
				Reduced: monthlyTaxBreakdownEntryResponse{
					TaxableAmount: r.Summary.TaxBreakdown.Reduced.TaxableAmount,
					TaxAmount:     r.Summary.TaxBreakdown.Reduced.TaxAmount,
				},
			},
		},
		DailyDetails: dailyDetails,
	}
}
