// Package handler provides HTTP handler implementations for Cage entity.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Cage ----

// ListCages godoc
func (h *Handler) ListCages(c *gin.Context) {
	var cageType *string
	if t := c.Query("cage_type"); t != "" {
		cageType = &t
	}
	cages, err := h.svc.Cage.List(c.Request.Context(), cageType)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, cages)
}

// GetCage godoc
func (h *Handler) GetCage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	cage, err := h.svc.Cage.GetByID(c.Request.Context(), id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, cage)
}

// CreateCage godoc
func (h *Handler) CreateCage(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var input createCageRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	cage := &model.Cage{
		ClinicID:    clinicID,
		Name:        input.Name,
		CageType:    model.CageType(input.CageType),
		CageSize:    model.CageSize(input.CageSize),
		Price:       input.Price,
		IsActive:    input.IsActive,
		Description: input.Description,
		SortOrder:   input.SortOrder,
	}

	if err := h.svc.Cage.Create(c.Request.Context(), cage); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, cage)
}

// UpdateCage godoc
func (h *Handler) UpdateCage(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	var input updateCageRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	cage := &model.Cage{
		ID:          id,
		ClinicID:    clinicID,
		Name:        input.Name,
		Price:       input.Price,
		Description: input.Description,
		SortOrder:   input.SortOrder,
	}
	if input.CageType != "" {
		cage.CageType = model.CageType(input.CageType)
	}
	if input.CageSize != "" {
		cage.CageSize = model.CageSize(input.CageSize)
	}
	if input.IsActive != nil {
		cage.IsActive = *input.IsActive
	}

	if err := h.svc.Cage.Update(c.Request.Context(), cage); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, cage)
}

// ReorderCages godoc
func (h *Handler) ReorderCages(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderCageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}
	if err := h.svc.Cage.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteCage godoc
func (h *Handler) DeleteCage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if err := h.svc.Cage.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
