package reservation

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// ReservationScheduleHandler は ReservationScheduleService の HTTP handler。
type ReservationScheduleHandler struct {
	svc ReservationScheduleService
}

// NewReservationScheduleHandler は ReservationScheduleHandler を構築する。
func NewReservationScheduleHandler(svc ReservationScheduleService) *ReservationScheduleHandler {
	return &ReservationScheduleHandler{svc: svc}
}

// ListReservationSchedules godoc
func (h *ReservationScheduleHandler) ListReservationSchedules(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ParseIDParam(c, "staffId")
	if !ok {
		return
	}
	query := newListReservationSchedulesQuery(c.Request.URL.Query(), time.Now())
	entries, err := h.svc.ListByMonth(c.Request.Context(), clinicID, staffID, query.Month)
	if err != nil {
		respondError(c, err)
		return
	}
	list := httpapi.MapSlice(entries, toScheduleEntryResponse)
	c.JSON(http.StatusOK, list)
}

// UpsertReservationSchedule godoc
func (h *ReservationScheduleHandler) UpsertReservationSchedule(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ParseIDParam(c, "staffId")
	if !ok {
		return
	}
	dateStr := c.Param("date")
	date, err := time.ParseInLocation(time.DateOnly, dateStr, time.Local)
	if err != nil {
		respondError(c, apperrors.WrapInvalidInput("date must be YYYY-MM-DD format"))
		return
	}

	var req upsertReservationScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	entry, isNew, err := h.svc.Save(c.Request.Context(), clinicID, staffID, date, req.toServiceInput())
	if err != nil {
		respondError(c, err)
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
func (h *ReservationScheduleHandler) DeleteReservationSchedule(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ParseIDParam(c, "staffId")
	if !ok {
		return
	}
	dateStr := c.Param("date")
	date, err := time.ParseInLocation(time.DateOnly, dateStr, time.Local)
	if err != nil {
		respondError(c, apperrors.WrapInvalidInput("date must be YYYY-MM-DD format"))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), clinicID, staffID, date); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
