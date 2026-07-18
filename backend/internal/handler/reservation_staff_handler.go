package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// ListReservationStaffs godoc
func (h *Handler) ListReservationStaffs(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	staffs, err := h.svc.ReservationStaff.List(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	staffIDs := make([]uint64, len(staffs))
	for i := range staffs {
		staffIDs[i] = staffs[i].ID
	}
	excludedMap, err := h.svc.ReservationStaff.ListExcludedByStaffIDs(c.Request.Context(), staffIDs)
	if err != nil {
		RespondError(c, err)
		return
	}
	list := make([]reservationStaffResponse, 0, len(staffs))
	for i := range staffs {
		list = append(list, toReservationStaffResponse(&staffs[i], excludedMap[staffs[i].ID]))
	}
	c.JSON(http.StatusOK, list)
}

// CreateReservationStaff godoc
func (h *Handler) CreateReservationStaff(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req createReservationStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	staff, excluded, err := h.svc.ReservationStaff.Create(c.Request.Context(), clinicID, req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/reservation-staffs/%d", staff.ID))
	c.JSON(http.StatusCreated, toReservationStaffResponse(staff, excluded))
}

// UpdateReservationStaff godoc
func (h *Handler) UpdateReservationStaff(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "staffId")
	if !ok {
		return
	}
	var req updateReservationStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	staff, excluded, err := h.svc.ReservationStaff.Update(c.Request.Context(), clinicID, id, req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationStaffResponse(staff, excluded))
}

// DeleteReservationStaff godoc
func (h *Handler) DeleteReservationStaff(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "staffId")
	if !ok {
		return
	}
	if err := h.svc.ReservationStaff.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// UpdateReservationStaffStatus godoc
func (h *Handler) UpdateReservationStaffStatus(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "staffId")
	if !ok {
		return
	}
	var req patchReservationStaffStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	staff, excluded, err := h.svc.ReservationStaff.PatchStatus(c.Request.Context(), clinicID, id, req.IsActive)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationStaffResponse(staff, excluded))
}

// UpdateReservationStaffSortOrder godoc
func (h *Handler) UpdateReservationStaffSortOrder(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "staffId")
	if !ok {
		return
	}
	var req patchReservationStaffSortOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.ReservationStaff.PatchSortOrder(c.Request.Context(), clinicID, id, req.Direction); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// UploadReservationStaffImage godoc — v2 スコープ：未実装
func (h *Handler) UploadReservationStaffImage(c *gin.Context) {
	RespondError(c, apperrors.WrapNotImplemented("この機能は未実装です"))
}
