package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
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
	staff, err := h.svc.ReservationStaff.Create(c.Request.Context(), clinicID, &service.CreateReservationStaffInput{
		Name:               req.Name,
		StaffType:          req.StaffType,
		ReservationVisible: req.ReservationVisible,
		ReservationComment: req.ReservationComment,
		SortOrder:          req.SortOrder,
		ExcludedTypeIDs:    req.ExcludedTypeIDs,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	excluded, err := h.svc.ReservationStaff.GetExcludedReservationTypes(c.Request.Context(), staff.ID)
	if err != nil {
		RespondError(c, err)
		return
	}
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
	staff, err := h.svc.ReservationStaff.Update(c.Request.Context(), clinicID, id, &service.UpdateReservationStaffInput{
		Name:               req.Name,
		StaffType:          req.StaffType,
		ReservationVisible: req.ReservationVisible,
		ReservationComment: req.ReservationComment,
		SortOrder:          req.SortOrder,
		ExcludedTypeIDs:    req.ExcludedTypeIDs,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	excluded, err := h.svc.ReservationStaff.GetExcludedReservationTypes(c.Request.Context(), staff.ID)
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

// PatchReservationStaffStatus godoc
func (h *Handler) PatchReservationStaffStatus(c *gin.Context) {
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
	staff, err := h.svc.ReservationStaff.PatchStatus(c.Request.Context(), clinicID, id, req.IsActive)
	if err != nil {
		RespondError(c, err)
		return
	}
	excluded, err := h.svc.ReservationStaff.GetExcludedReservationTypes(c.Request.Context(), staff.ID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationStaffResponse(staff, excluded))
}

// PatchReservationStaffSortOrder godoc
func (h *Handler) PatchReservationStaffSortOrder(c *gin.Context) {
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
	RespondError(c, apperrors.WrapInternalServerError("この機能は未実装です"))
}
