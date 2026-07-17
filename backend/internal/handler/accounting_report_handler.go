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
