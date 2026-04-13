// Package handler provides HTTP handler implementations for AnimalSpecies entity.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
)

// ListAnimalSpecies godoc
func (h *Handler) ListAnimalSpecies(c *gin.Context) {
	species, err := h.svc.AnimalSpecies.List(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toAnimalSpeciesResponseList(species))
}

// GetAnimalSpecies godoc
func (h *Handler) GetAnimalSpecies(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	species, err := h.svc.AnimalSpecies.GetByID(c.Request.Context(), id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toAnimalSpeciesResponse(species))
}

// CreateAnimalSpecies godoc
func (h *Handler) CreateAnimalSpecies(c *gin.Context) {
	var req createAnimalSpeciesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	species, err := h.svc.AnimalSpecies.Create(c.Request.Context(), &service.CreateAnimalSpeciesInput{
		Name:      req.Name,
		IsActive:  req.IsActive,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toAnimalSpeciesResponse(species))
}

// UpdateAnimalSpecies godoc
func (h *Handler) UpdateAnimalSpecies(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req updateAnimalSpeciesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	species, err := h.svc.AnimalSpecies.Update(c.Request.Context(), id, &service.UpdateAnimalSpeciesInput{
		Name:      req.Name,
		IsActive:  req.IsActive,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toAnimalSpeciesResponse(species))
}

// DeleteAnimalSpecies godoc
func (h *Handler) DeleteAnimalSpecies(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.AnimalSpecies.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderAnimalSpecies godoc
func (h *Handler) ReorderAnimalSpecies(c *gin.Context) {
	var req reorderAnimalSpeciesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.AnimalSpecies.Reorder(c.Request.Context(), req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "reordered"})
}
