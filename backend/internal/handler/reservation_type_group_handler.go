package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
)

// ListReservationTypeGroups godoc
func (h *Handler) ListReservationTypeGroups(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	groups, err := h.svc.ReservationTypeGroup.List(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(groups, toReservationTypeGroupResponse))
}

// GetReservationTypeGroup godoc
func (h *Handler) GetReservationTypeGroup(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	g, err := h.svc.ReservationTypeGroup.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationTypeGroupResponse(g))
}

// CreateReservationTypeGroup godoc
func (h *Handler) CreateReservationTypeGroup(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req createReservationTypeGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	g, err := h.svc.ReservationTypeGroup.Create(c.Request.Context(), clinicID, &service.CreateReservationTypeGroupInput{
		Name:      req.Name,
		Color:     req.Color,
		SortOrder: req.SortOrder,
		IsActive:  req.IsActive,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toReservationTypeGroupResponse(g))
}

// UpdateReservationTypeGroup godoc
func (h *Handler) UpdateReservationTypeGroup(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateReservationTypeGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	g, err := h.svc.ReservationTypeGroup.Update(c.Request.Context(), clinicID, id, &service.UpdateReservationTypeGroupInput{
		Name:      req.Name,
		Color:     req.Color,
		SortOrder: req.SortOrder,
		IsActive:  req.IsActive,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationTypeGroupResponse(g))
}

// DeleteReservationTypeGroup godoc
func (h *Handler) DeleteReservationTypeGroup(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.ReservationTypeGroup.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderReservationTypeGroups godoc
func (h *Handler) ReorderReservationTypeGroups(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.ReservationTypeGroup.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
