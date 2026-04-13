package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
)

// ListReservationTypeLiffs godoc
func (h *Handler) ListReservationTypeLiffs(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	items, err := h.svc.ReservationTypeLiff.List(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationTypeLiffResponseList(items))
}

// CreateReservationTypeLiff godoc
func (h *Handler) CreateReservationTypeLiff(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req createReservationTypeLiffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	st, err := h.svc.ReservationTypeLiff.Create(c.Request.Context(), clinicID, &service.CreateReservationTypeLiffInput{
		Name:                 req.Name,
		Color:                req.Color,
		Description:          req.Description,
		SortOrder:            req.SortOrder,
		DurationMinutes:      req.DurationMinutes,
		ShortName:            req.ShortName,
		ShowShortName:        req.ShowShortName,
		ReservationVisible:   req.ReservationVisible,
		ReservationComment:   req.ReservationComment,
		ReservationDayOption: req.ReservationDayOption,
		IsInternal:           req.IsInternal,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toReservationTypeLiffResponse(st))
}

// UpdateReservationTypeLiff godoc
func (h *Handler) UpdateReservationTypeLiff(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateReservationTypeLiffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	st, err := h.svc.ReservationTypeLiff.Update(c.Request.Context(), clinicID, id, &service.UpdateReservationTypeLiffInput{
		Name:                 req.Name,
		Color:                req.Color,
		Description:          req.Description,
		SortOrder:            req.SortOrder,
		DurationMinutes:      req.DurationMinutes,
		ShortName:            req.ShortName,
		ShowShortName:        req.ShowShortName,
		ReservationVisible:   req.ReservationVisible,
		ReservationComment:   req.ReservationComment,
		ReservationDayOption: req.ReservationDayOption,
		IsInternal:           req.IsInternal,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationTypeLiffResponse(st))
}

// DeleteReservationTypeLiff godoc
func (h *Handler) DeleteReservationTypeLiff(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.ReservationTypeLiff.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// PatchReservationTypeLiffStatus godoc
func (h *Handler) PatchReservationTypeLiffStatus(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req patchReservationTypeLiffStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	st, err := h.svc.ReservationTypeLiff.PatchStatus(c.Request.Context(), clinicID, id, req.IsActive)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationTypeLiffResponse(st))
}

// PatchReservationTypeLiffSortOrder godoc
func (h *Handler) PatchReservationTypeLiffSortOrder(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req patchReservationTypeLiffSortOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.ReservationTypeLiff.PatchSortOrder(c.Request.Context(), clinicID, id, req.Direction); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// UploadReservationTypeLiffImage godoc — v2 スコープ：未実装
func (h *Handler) UploadReservationTypeLiffImage(c *gin.Context) {
	RespondError(c, apperrors.WrapInternalServerError("この機能は未実装です"))
}
