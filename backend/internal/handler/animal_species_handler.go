// Package handler provides HTTP handler implementations for AnimalSpecies entity.
package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// ListAnimalSpecies godoc
func (h *Handler) ListAnimalSpecies(c *gin.Context) {
	species, err := h.svc.AnimalSpecies.List(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(species, toAnimalSpeciesResponse))
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

	species, err := h.svc.AnimalSpecies.Create(c.Request.Context(), req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/animal-species/%d", species.ID))
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

	species, err := h.svc.AnimalSpecies.Update(c.Request.Context(), id, req.toServiceInput())
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

// ReorderAnimalSpecies は動物種マスタの表示順を更新する。
// AnimalSpecies はシステム共通マスタ（clinic_id なし）のため clinicID パラメータは不要。
func (h *Handler) ReorderAnimalSpecies(c *gin.Context) {
	var req reorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.AnimalSpecies.Reorder(c.Request.Context(), req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
