package billing

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"
)

// AccountingReportHandler は AccountingReportService の HTTP handler。
type AccountingReportHandler struct {
	svc               AccountingReportService
	requirePermission PermissionMiddleware
}

// NewAccountingReportHandler は AccountingReportHandler を構築する。
func NewAccountingReportHandler(svc AccountingReportService, requirePermission PermissionMiddleware) *AccountingReportHandler {
	return &AccountingReportHandler{svc: svc, requirePermission: requirePermission}
}

// GetMonthlyReport godoc
// GET /v1/reports/monthly?year=2026&month=4
func (h *AccountingReportHandler) GetMonthlyReport(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	query := newMonthlyReportQuery(c.Request.URL.Query())
	var result *MonthlyReportResponse
	var err error
	if query.hasPeriod() {
		startDate, endDate, parseErr := query.toPeriod()
		if parseErr != nil {
			httpapi.RespondError(c, parseErr)
			return
		}
		result, err = h.svc.GetMonthlyByPeriod(c.Request.Context(), clinicID, startDate, endDate)
	} else {
		year, month, parseErr := query.toYearMonth(time.Now())
		if parseErr != nil {
			httpapi.RespondError(c, parseErr)
			return
		}
		result, err = h.svc.GetMonthly(c.Request.Context(), clinicID, year, month)
	}
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMonthlyReportResponse(result))
}

// ExportMonthlyCSV godoc
// GET /v1/reports/monthly/csv?year=2026&month=4
func (h *AccountingReportHandler) ExportMonthlyCSV(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	query := newMonthlyReportQuery(c.Request.URL.Query())
	var csvResult *MonthlyCSVResult
	var err error
	if query.hasPeriod() {
		startDate, endDate, parseErr := query.toPeriod()
		if parseErr != nil {
			httpapi.RespondError(c, parseErr)
			return
		}
		csvResult, err = h.svc.ExportMonthlyCSVByPeriod(c.Request.Context(), clinicID, startDate, endDate)
	} else {
		year, month, parseErr := query.toYearMonth(time.Now())
		if parseErr != nil {
			httpapi.RespondError(c, parseErr)
			return
		}
		csvResult, err = h.svc.ExportMonthlyCSV(c.Request.Context(), clinicID, year, month)
	}
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", csvResult.Filename))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", csvResult.Data)
}
