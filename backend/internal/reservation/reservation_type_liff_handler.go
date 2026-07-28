package reservation

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// ReservationTypeLiffHandler は LINE 予約管理用 予約区分の HTTP handler。
type ReservationTypeLiffHandler struct {
	svc ReservationTypeLiffService
}

// NewReservationTypeLiffHandler は ReservationTypeLiffHandler を構築する。
func NewReservationTypeLiffHandler(svc ReservationTypeLiffService) *ReservationTypeLiffHandler {
	return &ReservationTypeLiffHandler{svc: svc}
}

// ListReservationTypeLiffs godoc
func (h *ReservationTypeLiffHandler) ListReservationTypeLiffs(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	items, err := h.svc.List(c.Request.Context(), clinicID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(items, toReservationTypeLiffResponse))
}

// CreateReservationTypeLiff godoc
func (h *ReservationTypeLiffHandler) CreateReservationTypeLiff(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req createReservationTypeLiffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	st, err := h.svc.Create(c.Request.Context(), clinicID, req.toServiceInput())
	if err != nil {
		respondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/clinics/%d/reservation-types/%d", clinicID, st.ID))
	c.JSON(http.StatusCreated, toReservationTypeLiffResponse(st))
}

// UpdateReservationTypeLiff godoc
func (h *ReservationTypeLiffHandler) UpdateReservationTypeLiff(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateReservationTypeLiffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	st, err := h.svc.Update(c.Request.Context(), clinicID, id, req.toServiceInput())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationTypeLiffResponse(st))
}

// DeleteReservationTypeLiff godoc
func (h *ReservationTypeLiffHandler) DeleteReservationTypeLiff(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), clinicID, id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// UpdateReservationTypeLiffStatus godoc
func (h *ReservationTypeLiffHandler) UpdateReservationTypeLiffStatus(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req patchReservationTypeLiffStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	st, err := h.svc.PatchStatus(c.Request.Context(), clinicID, id, req.IsActive)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationTypeLiffResponse(st))
}

// UpdateReservationTypeLiffSortOrder godoc
func (h *ReservationTypeLiffHandler) UpdateReservationTypeLiffSortOrder(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req patchReservationTypeLiffSortOrderRequest
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

// UploadReservationTypeLiffImage godoc — v2 スコープ：未実装
func (h *ReservationTypeLiffHandler) UploadReservationTypeLiffImage(c *gin.Context) {
	respondError(c, apperrors.WrapNotImplemented("この機能は未実装です"))
}
