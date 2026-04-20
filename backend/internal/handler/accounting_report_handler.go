package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// GetMonthlyReport godoc
// GET /v1/reports/monthly?year=2026&month=4
func (h *Handler) GetMonthlyReport(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	year, month, err := parseYearMonth(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	result, err := h.svc.AccountingReport.GetMonthly(c.Request.Context(), clinicID, year, month)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// ExportMonthlyCSV godoc
// GET /v1/reports/monthly/csv?year=2026&month=4
func (h *Handler) ExportMonthlyCSV(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	year, month, err := parseYearMonth(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	reader, err := h.svc.AccountingReport.ExportCSV(c.Request.Context(), clinicID, year, month)
	if err != nil {
		RespondError(c, err)
		return
	}

	filename := fmt.Sprintf("monthly_report_%04d%02d.csv", year, month)
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Status(http.StatusOK)

	// ストリーム書き出し
	buf := make([]byte, 32*1024)
	for {
		n, readErr := reader.Read(buf)
		if n > 0 {
			if _, writeErr := c.Writer.Write(buf[:n]); writeErr != nil {
				return
			}
		}
		if readErr != nil {
			break
		}
	}
}

// RegisterAccountingReportRoutes は売上レポート関連のルートを登録する
func (h *Handler) RegisterAccountingReportRoutes(rg *gin.RouterGroup) {
	reports := rg.Group("/reports")
	reports.GET("/monthly", h.RequirePermission(string(model.ResourceAccountingReports), "view"), h.GetMonthlyReport)
	reports.GET("/monthly/csv", h.RequirePermission(string(model.ResourceAccountingReports), "view"), h.ExportMonthlyCSV)
}

// parseYearMonth はクエリパラメータ year/month を検証してパースする
func parseYearMonth(c *gin.Context) (year, month int, err error) {
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	monthStr := c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month())))

	year, err = strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		return 0, 0, apperrors.WrapInvalidInput("year は 2000〜2100 の整数で指定してください")
	}
	month, err = strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		return 0, 0, apperrors.WrapInvalidInput("month は 1〜12 の整数で指定してください")
	}
	return year, month, nil
}
