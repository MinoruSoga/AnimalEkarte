package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ListClinicHolidays godoc
// GET /v1/clinic-holidays?year_month=YYYY-MM
func (h *Handler) ListClinicHolidays(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	yearMonth := c.Query("year_month") // YYYY-MM (optional)

	holidays, err := h.svc.ClinicHoliday.List(c.Request.Context(), clinicID, yearMonth)
	if err != nil {
		RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, mapSlice(holidays, toClinicHolidayResponse))
}

type setClinicHolidayRequest struct {
	Date   string `json:"date" binding:"required"` // YYYY-MM-DD
	Reason string `json:"reason"`
}

// SetClinicHoliday godoc
// POST /v1/clinic-holidays
func (h *Handler) SetClinicHoliday(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req setClinicHolidayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid date: use YYYY-MM-DD"))
		return
	}

	holiday, err := h.svc.ClinicHoliday.Set(c.Request.Context(), clinicID, date, req.Reason)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/clinic-holidays/%s", holiday.Date.Format("2006-01-02")))
	c.JSON(http.StatusCreated, toClinicHolidayResponse(holiday))
}

// DeleteClinicHoliday godoc
// DELETE /v1/clinic-holidays/:date
func (h *Handler) DeleteClinicHoliday(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	dateStr := c.Param("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid date: use YYYY-MM-DD"))
		return
	}

	if err := h.svc.ClinicHoliday.Remove(c.Request.Context(), clinicID, date); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterClinicHolidayRoutes は休診日関連のルートを登録する
func (h *Handler) RegisterClinicHolidayRoutes(rg *gin.RouterGroup) {
	holidays := rg.Group("/clinic-holidays")
	holidays.GET("", h.RequirePermission(string(model.ResourceShifts), "view"), h.ListClinicHolidays)
	holidays.POST("", h.RequirePermission(string(model.ResourceShifts), "create"), h.SetClinicHoliday)
	holidays.DELETE("/:date", h.RequirePermission(string(model.ResourceShifts), "delete"), h.DeleteClinicHoliday)
}
