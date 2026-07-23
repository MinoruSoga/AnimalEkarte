package clinic

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// ListClinicHolidays godoc
// GET /v1/clinic-holidays?year_month=YYYY-MM
func (h *Handler) ListClinicHolidays(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	query := NewListClinicHolidaysQuery(c.Request.URL.Query())
	holidays, err := h.holidaySvc.List(c.Request.Context(), clinicID, query.YearMonth)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, httpapi.MapSlice(holidays, ToClinicHolidayResponse))
}

type setClinicHolidayRequest struct {
	Date   string `json:"date" binding:"required"` // YYYY-MM-DD
	Reason string `json:"reason"`
}

// SetClinicHoliday godoc
// POST /v1/clinic-holidays
func (h *Handler) SetClinicHoliday(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	var req setClinicHolidayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	date, err := time.ParseInLocation(time.DateOnly, req.Date, time.Local)
	if err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("invalid date: use YYYY-MM-DD"))
		return
	}

	holiday, err := h.holidaySvc.Set(c.Request.Context(), clinicID, date, req.Reason)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/clinic-holidays/%s", holiday.Date.In(time.Local).Format(time.DateOnly)))
	c.JSON(http.StatusCreated, ToClinicHolidayResponse(holiday))
}

// DeleteClinicHoliday godoc
// DELETE /v1/clinic-holidays/:date
func (h *Handler) DeleteClinicHoliday(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	dateStr := c.Param("date")
	date, err := time.ParseInLocation(time.DateOnly, dateStr, time.Local)
	if err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("invalid date: use YYYY-MM-DD"))
		return
	}

	if err := h.holidaySvc.Remove(c.Request.Context(), clinicID, date); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterClinicHolidayRoutes は休診日関連のルートを登録する
func (h *Handler) RegisterClinicHolidayRoutes(rg *gin.RouterGroup) {
	holidays := rg.Group("/clinic-holidays")
	holidays.GET("", h.requirePermission(string(model.ResourceShifts), "view"), h.ListClinicHolidays)
	holidays.POST("", h.requirePermission(string(model.ResourceShifts), "create"), h.SetClinicHoliday)
	holidays.DELETE("/:date", h.requirePermission(string(model.ResourceShifts), "delete"), h.DeleteClinicHoliday)
}
