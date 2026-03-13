// Package handler provides HTTP handler implementations for ServiceType entity.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

type createServiceTypeInput struct {
	Name        string `json:"name"        binding:"required"`
	IsActive    bool   `json:"is_active"`
	Description string `json:"description"`
	Color       string `json:"color"`
	SortOrder   int    `json:"sort_order"`
}

type updateServiceTypeInput struct {
	Name        string `json:"name"`
	IsActive    *bool  `json:"is_active"`
	Description string `json:"description"`
	Color       string `json:"color"`
	SortOrder   int    `json:"sort_order"`
}

// ---- ServiceType ----

// ListServiceTypes godoc
func (h *Handler) ListServiceTypes(c *gin.Context) {
	serviceTypes, err := h.svc.ServiceType.List(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, serviceTypes)
}

// CreateServiceType godoc
func (h *Handler) CreateServiceType(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var input createServiceTypeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	serviceType := &model.ServiceType{
		ClinicID:    clinicID,
		Name:        input.Name,
		IsActive:    input.IsActive,
		Description: input.Description,
		Color:       input.Color,
		SortOrder:   input.SortOrder,
	}

	if err := h.svc.ServiceType.Create(c.Request.Context(), serviceType); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, serviceType)
}

// UpdateServiceType godoc
func (h *Handler) UpdateServiceType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input updateServiceTypeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	serviceType := &model.ServiceType{
		ID:          id,
		Name:        input.Name,
		Description: input.Description,
		Color:       input.Color,
		SortOrder:   input.SortOrder,
	}
	if input.IsActive != nil {
		serviceType.IsActive = *input.IsActive
	}

	if err := h.svc.ServiceType.Update(c.Request.Context(), serviceType); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, serviceType)
}

// DeleteServiceType godoc
func (h *Handler) DeleteServiceType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.ServiceType.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
