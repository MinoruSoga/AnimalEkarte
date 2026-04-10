// Package handler provides HTTP handler implementations for ReservationCategory entity.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- ReservationCategory ----

// GetReservationCategory godoc
func (h *Handler) GetReservationCategory(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	st, err := h.svc.ReservationCategory.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationCategoryResponse(st))
}

// ListReservationCategories godoc
func (h *Handler) ListReservationCategories(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	reservationCategories, err := h.svc.ReservationCategory.List(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationCategoryResponseList(reservationCategories))
}

// CreateReservationCategory godoc
func (h *Handler) CreateReservationCategory(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req createReservationCategoryRequest
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
	st, err := h.svc.ReservationCategory.Create(c.Request.Context(), clinicID, &service.CreateReservationCategoryInput{
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
	c.JSON(http.StatusCreated, toReservationCategoryResponse(st))
}

// UpdateReservationCategory godoc
func (h *Handler) UpdateReservationCategory(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	var req updateReservationCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	st, err := h.svc.ReservationCategory.Update(c.Request.Context(), clinicID, id, &service.UpdateReservationCategoryInput{
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
	c.JSON(http.StatusOK, toReservationCategoryResponse(st))
}

// DeleteReservationCategory godoc
func (h *Handler) DeleteReservationCategory(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if err := h.svc.ReservationCategory.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderReservationCategories godoc
func (h *Handler) ReorderReservationCategories(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderReservationCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.ReservationCategory.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
