package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// ListReservationSchedules godoc
func (h *Handler) ListReservationSchedules(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := parseIDParam(c, "staffId")
	if !ok {
		return
	}
	query := newListReservationSchedulesQuery(c.Request.URL.Query(), time.Now())
	entries, err := h.svc.ReservationSchedule.ListByMonth(c.Request.Context(), clinicID, staffID, query.Month)
	if err != nil {
		RespondError(c, err)
		return
	}
	list := mapSlice(entries, toScheduleEntryResponse)
	c.JSON(http.StatusOK, list)
}

// UpsertReservationSchedule godoc
func (h *Handler) UpsertReservationSchedule(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := parseIDParam(c, "staffId")
	if !ok {
		return
	}
	dateStr := c.Param("date")
	date, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("date must be YYYY-MM-DD format"))
		return
	}

	var req upsertReservationScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	entry, isNew, err := h.svc.ReservationSchedule.Save(c.Request.Context(), clinicID, staffID, date, req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	if isNew {
		c.Header("Location", fmt.Sprintf("/v1/clinics/%d/reservation-staffs/%d/schedules/%s", clinicID, staffID, dateStr))
		c.JSON(http.StatusCreated, toScheduleEntryResponse(entry))
		return
	}
	c.JSON(http.StatusOK, toScheduleEntryResponse(entry))
}

// DeleteReservationSchedule godoc
func (h *Handler) DeleteReservationSchedule(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := parseIDParam(c, "staffId")
	if !ok {
		return
	}
	dateStr := c.Param("date")
	date, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("date must be YYYY-MM-DD format"))
		return
	}
	if err := h.svc.ReservationSchedule.Delete(c.Request.Context(), clinicID, staffID, date); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
