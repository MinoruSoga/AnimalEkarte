package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ListReservations godoc
func (h *Handler) ListReservations(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	filters, err := newListReservationQuery(c.Request.URL.Query()).toServiceFilters()
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}

	reservations, total, err := h.svc.Reservation.List(c.Request.Context(), clinicID, page, limit, filters.Date, filters.Status, filters.Source, filters.PetID, filters.OwnerID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(reservations, total, page, limit))
}

// GetReservation godoc
func (h *Handler) GetReservation(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	reservation, err := h.svc.Reservation.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationResponse(reservation))
}

// CreateReservation godoc
func (h *Handler) CreateReservation(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := extractStaffID(c)
	if !ok {
		return
	}
	var input createReservationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	svcInput, err := input.toServiceInput(clinicID, staffID)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}

	ctx := c.Request.Context()

	// BUG-144: staff_id のクリニック所属チェック（クロスクリニック FK 防止）
	if svcInput.DoctorID != nil {
		if err := h.checkDoctorClinicAssignment(ctx, clinicID, *svcInput.DoctorID); err != nil {
			RespondError(c, err)
			return
		}
	}

	reservation, err := h.svc.Reservation.Create(ctx, svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/reservations/%d", reservation.ID))
	c.JSON(http.StatusCreated, toReservationResponse(reservation))
}

// UpdateReservation godoc
func (h *Handler) UpdateReservation(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var input updateReservationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	svcInput, err := input.toServiceInput()
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}

	ctx := c.Request.Context()

	// BUG-144: staff_id のクリニック所属チェック（Update時も）
	if svcInput.DoctorID != nil {
		if err := h.checkDoctorClinicAssignment(ctx, clinicID, *svcInput.DoctorID); err != nil {
			RespondError(c, err)
			return
		}
	}

	reservation, err := h.svc.Reservation.Update(ctx, clinicID, id, &svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}

	// 受付済みに変更された場合は通常カルテを best-effort で自動作成する（BE-reception-auto-create-medical-record）
	if shouldAutoCreateMedicalRecordForReservation(svcInput.Status, reservation) && h.svc.MedicalRecord != nil {
		h.svc.MedicalRecord.AutoCreateFromReservation(ctx, clinicID, reservation)
	}

	c.JSON(http.StatusOK, toReservationResponse(reservation))
}

func shouldAutoCreateMedicalRecordForReservation(status *model.ReservationStatus, reservation *model.Reservation) bool {
	if status == nil || *status != model.ReservationStatusCheckedIn || reservation == nil {
		return false
	}
	if reservation.ReservationType != nil && reservation.ReservationType.Category == model.ReservationTypeCategoryTrimming {
		return false
	}
	if reservation.ReservationType != nil &&
		(strings.Contains(reservation.ReservationType.Name, "入院") ||
			strings.Contains(reservation.ReservationType.Name, "ホテル")) {
		return false
	}
	return true
}

// DeleteReservation godoc
func (h *Handler) DeleteReservation(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Reservation.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// PatchReservationReservationRoute godoc
// PATCH /reservations/:id/reservation-route — 予約経路を更新する（FEAT-381-2）。
func (h *Handler) PatchReservationReservationRoute(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req patchReservationReservationRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	reservation, err := h.svc.Reservation.UpdateReservationRoute(
		c.Request.Context(), clinicID, id,
		req.toServiceInput(),
	)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationResponse(reservation))
}

// RegisterReservationRoutes は予約関連のルートを登録する
func (h *Handler) RegisterReservationRoutes(rg *gin.RouterGroup) {
	reservations := rg.Group("/reservations")
	reservations.GET("", h.RequirePermission(string(model.ResourceReservations), "view"), h.ListReservations)
	reservations.GET("/:id", h.RequirePermission(string(model.ResourceReservations), "view"), h.GetReservation)
	reservations.POST("", h.RequirePermission(string(model.ResourceReservations), "create"), h.CreateReservation)
	reservations.PATCH("/:id", h.RequirePermission(string(model.ResourceReservations), "edit"), h.UpdateReservation)
	reservations.PATCH("/:id/reservation-route", h.RequirePermission(string(model.ResourceReservations), "edit"), h.PatchReservationReservationRoute)
	reservations.DELETE("/:id", h.RequirePermission(string(model.ResourceReservations), "delete"), h.DeleteReservation)
}
