package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
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
	month := c.Query("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	entries, err := h.svc.ReservationSchedule.ListByMonth(c.Request.Context(), clinicID, staffID, month)
	if err != nil {
		RespondError(c, err)
		return
	}
	list := make([]scheduleEntryResponse, 0, len(entries))
	for i := range entries {
		list = append(list, toScheduleEntryResponse(&entries[i]))
	}
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
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("date must be YYYY-MM-DD format"))
		return
	}

	var req upsertReservationScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	breaks := make([]service.ReservationScheduleBreakInput, 0, len(req.Breaks))
	for _, b := range req.Breaks {
		breaks = append(breaks, service.ReservationScheduleBreakInput{Start: b.Start, End: b.End})
	}

	entry, isNew, err := h.svc.ReservationSchedule.Upsert(c.Request.Context(), clinicID, staffID, date, &service.CreateReservationScheduleInput{
		ShiftType: req.ShiftType,
		WorkStart: req.WorkStart,
		WorkEnd:   req.WorkEnd,
		Breaks:    breaks,
	})
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
	date, err := time.Parse("2006-01-02", dateStr)
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
