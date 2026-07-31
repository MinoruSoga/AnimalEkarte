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
	ctx := c.Request.Context()
	excludedMap, err := h.svc.ListExcludedByStaffIDs(ctx, clinicID, staffIDs)
	if err != nil {
		respondError(c, err)
		return
	}
	capableMap, err := h.svc.ListCapableByStaffIDs(ctx, clinicID, staffIDs)
	if err != nil {
		respondError(c, err)
		return
	}
	list := make([]reservationStaffResponse, 0, len(staffs))
	for i := range staffs {
		sid := staffs[i].ID
		list = append(list, toReservationStaffResponse(&staffs[i], excludedMap[sid], capableMap[sid]))
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
	ctx := c.Request.Context()
	staff, excluded, err := h.svc.Create(ctx, clinicID, req.toServiceInput())
	if err != nil {
		respondError(c, err)
		return
	}
	capableMap, err := h.svc.ListCapableByStaffIDs(ctx, clinicID, []uint64{staff.ID})
	if err != nil {
		respondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/reservation-staffs/%d", staff.ID))
	c.JSON(http.StatusCreated, toReservationStaffResponse(staff, excluded, capableMap[staff.ID]))
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
	ctx := c.Request.Context()
	staff, excluded, err := h.svc.Update(ctx, clinicID, id, req.toServiceInput())
	if err != nil {
		respondError(c, err)
		return
	}
	capableMap, err := h.svc.ListCapableByStaffIDs(ctx, clinicID, []uint64{staff.ID})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationStaffResponse(staff, excluded, capableMap[staff.ID]))
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
	ctx := c.Request.Context()
	staff, excluded, err := h.svc.PatchStatus(ctx, clinicID, id, req.IsActive)
	if err != nil {
		respondError(c, err)
		return
	}
	capableMap, err := h.svc.ListCapableByStaffIDs(ctx, clinicID, []uint64{staff.ID})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationStaffResponse(staff, excluded, capableMap[staff.ID]))
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
