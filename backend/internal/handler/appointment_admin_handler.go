package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
)

// ListReservationsAdmin godoc
func (h *Handler) ListReservationsAdmin(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	view := c.DefaultQuery("view", "month")
	dateStr := c.Query("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01")
	}

	switch view {
	case "month":
		items, err := h.svc.ReservationAdmin.ListByMonth(c.Request.Context(), clinicID, dateStr)
		if err != nil {
			RespondError(c, err)
			return
		}
		list := make([]reservationSummaryResponse, 0, len(items))
		for i := range items {
			list = append(list, toReservationSummaryResponse(&items[i]))
		}
		c.JSON(http.StatusOK, list)

	case "day":
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("date must be YYYY-MM-DD format for day view"))
			return
		}
		items, err := h.svc.ReservationAdmin.ListByDay(c.Request.Context(), clinicID, date)
		if err != nil {
			RespondError(c, err)
			return
		}
		list := make([]reservationDetailResponse, 0, len(items))
		for i := range items {
			list = append(list, toReservationDetailResponse(&items[i]))
		}
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

	ra, err := h.svc.ReservationAdmin.Create(c.Request.Context(), clinicID, &service.CreateReservationAdminInput{
		StartTime:        req.StartTime,
		EndTime:          req.EndTime,
		OwnerID:          req.OwnerID,
		PetID:            req.PetID,
		VisitType:        req.VisitType,
		ReservationTypeID:    req.ReservationTypeID,
		DoctorID:         req.DoctorID,
		IsDesignated:     req.IsDesignated,
		Notes:            req.Notes,
		LineCustomerID:   req.LineCustomerID,
		IsStaffDelegated: req.IsStaffDelegated,
		CustomerFields:   req.CustomerFields,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toReservationDetailResponse(ra))
}

// DeleteReservationAdmin godoc
func (h *Handler) DeleteReservationAdmin(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if err := h.svc.ReservationAdmin.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
