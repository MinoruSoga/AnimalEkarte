// Package handler provides HTTP handler implementations for ServiceType entity.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- ServiceType ----

// GetServiceType godoc
func (h *Handler) GetServiceType(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	st, err := h.svc.ServiceType.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toServiceTypeResponse(st))
}

// ListServiceTypes godoc
func (h *Handler) ListServiceTypes(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	serviceTypes, err := h.svc.ServiceType.List(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toServiceTypeResponseList(serviceTypes))
}

// CreateServiceType godoc
func (h *Handler) CreateServiceType(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req createServiceTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	st, err := h.svc.ServiceType.Create(c.Request.Context(), clinicID, &service.CreateServiceTypeInput{
		Name:        req.Name,
		Color:       req.Color,
		IsActive:    req.IsActive,
		Description: req.Description,
		SortOrder:   req.SortOrder,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toServiceTypeResponse(st))
}

// UpdateServiceType godoc
func (h *Handler) UpdateServiceType(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	var req updateServiceTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	st, err := h.svc.ServiceType.Update(c.Request.Context(), clinicID, id, &service.UpdateServiceTypeInput{
		Name:        req.Name,
		Color:       req.Color,
		IsActive:    req.IsActive,
		Description: req.Description,
		SortOrder:   req.SortOrder,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toServiceTypeResponse(st))
}

// DeleteServiceType godoc
func (h *Handler) DeleteServiceType(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if err := h.svc.ServiceType.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderServiceTypes godoc
func (h *Handler) ReorderServiceTypes(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderServiceTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.ServiceType.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
