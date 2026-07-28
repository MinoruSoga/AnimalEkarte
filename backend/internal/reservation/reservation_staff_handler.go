package reservation

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// ReservationStaffHandler は ReservationStaffService の HTTP handler。
type ReservationStaffHandler struct {
	svc ReservationStaffService
}

// NewReservationStaffHandler は ReservationStaffHandler を構築する。
func NewReservationStaffHandler(svc ReservationStaffService) *ReservationStaffHandler {
	return &ReservationStaffHandler{svc: svc}
}

// ListReservationStaffs godoc
func (h *ReservationStaffHandler) ListReservationStaffs(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffs, err := h.svc.List(c.Request.Context(), clinicID)
	if err != nil {
		respondError(c, err)
		return
	}
	staffIDs := make([]uint64, len(staffs))
	for i := range staffs {
		staffIDs[i] = staffs[i].ID
	}
	excludedMap, err := h.svc.ListExcludedByStaffIDs(c.Request.Context(), clinicID, staffIDs)
	if err != nil {
		respondError(c, err)
		return
	}
	list := make([]reservationStaffResponse, 0, len(staffs))
	for i := range staffs {
		list = append(list, toReservationStaffResponse(&staffs[i], excludedMap[staffs[i].ID]))
	}
	c.JSON(http.StatusOK, list)
}

// CreateReservationStaff godoc
func (h *ReservationStaffHandler) CreateReservationStaff(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req createReservationStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	staff, excluded, err := h.svc.Create(c.Request.Context(), clinicID, req.toServiceInput())
	if err != nil {
		respondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/reservation-staffs/%d", staff.ID))
	c.JSON(http.StatusCreated, toReservationStaffResponse(staff, excluded))
}

// UpdateReservationStaff godoc
func (h *ReservationStaffHandler) UpdateReservationStaff(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "staffId")
	if !ok {
		return
	}
	var req updateReservationStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	staff, excluded, err := h.svc.Update(c.Request.Context(), clinicID, id, req.toServiceInput())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationStaffResponse(staff, excluded))
}

// DeleteReservationStaff godoc
func (h *ReservationStaffHandler) DeleteReservationStaff(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "staffId")
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), clinicID, id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// UpdateReservationStaffStatus godoc
func (h *ReservationStaffHandler) UpdateReservationStaffStatus(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "staffId")
	if !ok {
		return
	}
	var req patchReservationStaffStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	staff, excluded, err := h.svc.PatchStatus(c.Request.Context(), clinicID, id, req.IsActive)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationStaffResponse(staff, excluded))
}

// UpdateReservationStaffSortOrder godoc
func (h *ReservationStaffHandler) UpdateReservationStaffSortOrder(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "staffId")
	if !ok {
		return
	}
	var req patchReservationStaffSortOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if err := h.svc.PatchSortOrder(c.Request.Context(), clinicID, id, req.Direction); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// UploadReservationStaffImage godoc — v2 スコープ：未実装
func (h *ReservationStaffHandler) UploadReservationStaffImage(c *gin.Context) {
	respondError(c, apperrors.WrapNotImplemented("この機能は未実装です"))
}
