package handler

import (
	"encoding/csv"
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

	result, err := h.svc.AccountingReport.GetMonthly(c.Request.Context(), clinicID, year, month)
	if err != nil {
		RespondError(c, err)
		return
	}

	filename := fmt.Sprintf("monthly_report_%04d%02d.csv", year, month)
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Status(http.StatusOK)

	// BOM（Excel の UTF-8 認識用）
	_, _ = c.Writer.Write([]byte("\xEF\xBB\xBF"))

	w := csv.NewWriter(c.Writer)
	_ = w.Write([]string{
		"日付", "曜日", "AM件数", "AM純売上(円)", "PM件数", "PM純売上(円)", "日計(円)", "返金(円)", "AM締め", "PM締め", "休診",
		"標準税率課税対象額(円)", "標準税率消費税額(円)", "軽減税率課税対象額(円)", "軽減税率消費税額(円)",
	})

	for _, d := range result.DailyDetails {
		amClosed := "未"
		if d.AMClosed {
			amClosed = "済"
		}
		pmClosed := "未"
		if d.PMClosed {
			pmClosed = "済"
		}
		holiday := ""
		if d.IsHoliday {
			holiday = "休"
		}
		_ = w.Write([]string{
			d.Date,
			d.Weekday,
			fmt.Sprintf("%d", d.AMCount),
			fmt.Sprintf("%d", d.AMNet),
			fmt.Sprintf("%d", d.PMCount),
			fmt.Sprintf("%d", d.PMNet),
			fmt.Sprintf("%d", d.DayNet),
			fmt.Sprintf("%d", d.Refund),
			amClosed,
			pmClosed,
			holiday,
			"", "", "", "", // 税率別内訳は月次サマリのみ（日別内訳なし）
		})
	}

	tb := result.Summary.TaxBreakdown
	_ = w.Write([]string{
		"合計", "", "", "", "", "",
		fmt.Sprintf("%d", result.Summary.NetAmount),
		fmt.Sprintf("%d", result.Summary.TotalRefund),
		"", "", "",
		fmt.Sprintf("%d", tb.Standard.TaxableAmount),
		fmt.Sprintf("%d", tb.Standard.TaxAmount),
		fmt.Sprintf("%d", tb.Reduced.TaxableAmount),
		fmt.Sprintf("%d", tb.Reduced.TaxAmount),
	})
	w.Flush()
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
