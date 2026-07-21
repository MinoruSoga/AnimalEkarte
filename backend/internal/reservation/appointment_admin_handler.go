package reservation

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// ReservationAdminHandler は LINE 管理用予約 CRUD の HTTP handler。
type ReservationAdminHandler struct {
	svc              ReservationAdminService
	staffAssignments staffAssignmentFinder
}

// NewReservationAdminHandler は ReservationAdminHandler を構築する。
func NewReservationAdminHandler(svc ReservationAdminService, staffAssignments staffAssignmentFinder) *ReservationAdminHandler {
	return &ReservationAdminHandler{svc: svc, staffAssignments: staffAssignments}
}

// checkDoctorClinicAssignment は ReservationHandler 側と同一の忠実移植（旧 *Handler 共有メソッド）。
func (h *ReservationAdminHandler) checkDoctorClinicAssignment(ctx context.Context, clinicID, doctorID uint64) error {
	if doctorID == 0 {
		return nil
	}
	assignments, err := h.staffAssignments.FindAllByStaffID(ctx, doctorID)
	if err != nil {
		return apperrors.Wrap(err, "failed to verify staff assignment")
	}
	for i := range assignments {
		if assignments[i].ClinicID == clinicID {
			return nil
		}
	}
	return apperrors.WrapInvalidInput("指定されたスタッフはこのクリニックに所属していません")
}

// ListReservationsAdmin godoc
func (h *ReservationAdminHandler) ListReservationsAdmin(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	query := newListReservationsAdminQuery(c.Request.URL.Query(), time.Now())

	switch query.View {
	case "month":
		items, err := h.svc.ListByMonth(c.Request.Context(), clinicID, query.Date)
		if err != nil {
			respondError(c, err)
			return
		}
		list := httpapi.MapSlice(items, toReservationSummaryResponse)
		c.JSON(http.StatusOK, list)

	case "day":
		date, err := time.ParseInLocation(time.DateOnly, query.Date, time.Local)
		if err != nil {
			respondError(c, apperrors.WrapInvalidInput("date must be YYYY-MM-DD format for day view"))
			return
		}
		items, err := h.svc.ListByDay(c.Request.Context(), clinicID, date)
		if err != nil {
			respondError(c, err)
			return
		}
		list := httpapi.MapSlice(items, toReservationDetailResponse)
		c.JSON(http.StatusOK, list)

	default:
		respondError(c, apperrors.WrapInvalidInput("view must be 'month' or 'day'"))
	}
}

// CreateReservationAdmin godoc
func (h *ReservationAdminHandler) CreateReservationAdmin(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	var req createReservationAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	// BUG-144: staff_id のクリニック所属チェック（クロスクリニック FK 防止）
	if req.DoctorID != nil {
		if err := h.checkDoctorClinicAssignment(c.Request.Context(), clinicID, *req.DoctorID); err != nil {
			respondError(c, err)
			return
		}
	}

	ra, err := h.svc.Create(c.Request.Context(), clinicID, req.toServiceInput(staffID))
	if err != nil {
		respondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/clinics/%d/reservations/%d", clinicID, ra.ID))
	c.JSON(http.StatusCreated, toReservationDetailResponse(ra))
}

// DeleteReservationAdmin godoc
func (h *ReservationAdminHandler) DeleteReservationAdmin(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "reservationId")
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), clinicID, id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
