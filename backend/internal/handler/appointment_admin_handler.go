package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// ListReservationsAdmin godoc
func (h *Handler) ListReservationsAdmin(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	query := newListReservationsAdminQuery(c.Request.URL.Query(), time.Now())

	switch query.View {
	case "month":
		items, err := h.svc.ReservationAdmin.ListByMonth(c.Request.Context(), clinicID, query.Date)
		if err != nil {
			RespondError(c, err)
			return
		}
		list := mapSlice(items, toReservationSummaryResponse)
		c.JSON(http.StatusOK, list)

	case "day":
		date, err := time.ParseInLocation(time.DateOnly, query.Date, time.Local)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("date must be YYYY-MM-DD format for day view"))
			return
		}
		items, err := h.svc.ReservationAdmin.ListByDay(c.Request.Context(), clinicID, date)
		if err != nil {
			RespondError(c, err)
			return
		}
		list := mapSlice(items, toReservationDetailResponse)
		c.JSON(http.StatusOK, list)

	default:
		RespondError(c, apperrors.WrapInvalidInput("view must be 'month' or 'day'"))
	}
}

// CreateReservationAdmin godoc
func (h *Handler) CreateReservationAdmin(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := extractStaffID(c)
	if !ok {
		return
	}
	var req createReservationAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	// BUG-144: staff_id のクリニック所属チェック（クロスクリニック FK 防止）
	if req.DoctorID != nil {
		if err := h.checkDoctorClinicAssignment(c.Request.Context(), clinicID, *req.DoctorID); err != nil {
			RespondError(c, err)
			return
		}
	}

	ra, err := h.svc.ReservationAdmin.Create(c.Request.Context(), clinicID, req.toServiceInput(staffID))
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/clinics/%d/reservations/%d", clinicID, ra.ID))
	c.JSON(http.StatusCreated, toReservationDetailResponse(ra))
}

// DeleteReservationAdmin godoc
func (h *Handler) DeleteReservationAdmin(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "reservationId")
	if !ok {
		return
	}
	if err := h.svc.ReservationAdmin.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
