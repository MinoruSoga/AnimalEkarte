// Package handler provides HTTP handler implementations for ReservationType entity.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- ReservationType ----

// GetReservationType godoc
func (h *Handler) GetReservationType(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	st, err := h.svc.ReservationType.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationTypeResponse(st))
}

// ListReservationTypes godoc
func (h *Handler) ListReservationTypes(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	reservationTypes, err := h.svc.ReservationType.List(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationTypeResponseList(reservationTypes))
}

// CreateReservationType godoc
func (h *Handler) CreateReservationType(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req createReservationTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	reservationVisible := true
	if req.ReservationVisible != nil {
		reservationVisible = *req.ReservationVisible
	}
	durationMinutes := 15
	if req.DurationMinutes != nil {
		durationMinutes = *req.DurationMinutes
	}
	st, err := h.svc.ReservationType.Create(c.Request.Context(), clinicID, &service.CreateReservationTypeInput{
		Name:                   req.Name,
		Color:                  req.Color,
		IsActive:               true,
		Description:            req.Description,
		SortOrder:              req.SortOrder,
		ReservationDisplayName: req.ReservationDisplayName,
		DurationMinutes:        durationMinutes,
		ShortName:              req.ShortName,
		ShowShortName:          req.ShowShortName,
		ReservationVisible:     reservationVisible,
		ReservationComment:     req.ReservationComment,
		ReservationImageURL:    req.ReservationImageURL,
		ReservationDayOption:   req.ReservationDayOption,
		IsInternal:             req.IsInternal,
		GroupID:                req.GroupID,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toReservationTypeResponse(st))
}

// UpdateReservationType godoc
func (h *Handler) UpdateReservationType(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateReservationTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	st, err := h.svc.ReservationType.Update(c.Request.Context(), clinicID, id, &service.UpdateReservationTypeInput{
		Name:                   req.Name,
		Color:                  req.Color,
		IsActive:               req.IsActive,
		Description:            req.Description,
		SortOrder:              req.SortOrder,
		ReservationDisplayName: req.ReservationDisplayName,
		DurationMinutes:        req.DurationMinutes,
		ShortName:              req.ShortName,
		ShowShortName:          req.ShowShortName,
		ReservationVisible:     req.ReservationVisible,
		ReservationComment:     req.ReservationComment,
		ReservationImageURL:    req.ReservationImageURL,
		ReservationDayOption:   req.ReservationDayOption,
		IsInternal:             req.IsInternal,
		GroupID:                req.GroupID,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationTypeResponse(st))
}

// DeleteReservationType godoc
func (h *Handler) DeleteReservationType(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.ReservationType.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderReservationTypes godoc
func (h *Handler) ReorderReservationTypes(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderReservationTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.ReservationType.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
