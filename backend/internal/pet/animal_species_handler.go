// Package handler provides HTTP handler implementations for AnimalSpecies entity.
package pet

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

// ListAnimalSpecies godoc
func (h *Handler) ListAnimalSpecies(c *gin.Context) {
	species, err := h.animalSpecies.List(c.Request.Context())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(species, toAnimalSpeciesResponse))
}

// GetAnimalSpecies godoc
func (h *Handler) GetAnimalSpecies(c *gin.Context) {
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	species, err := h.animalSpecies.GetByID(c.Request.Context(), id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toAnimalSpeciesResponse(species))
}

// CreateAnimalSpecies godoc
func (h *Handler) CreateAnimalSpecies(c *gin.Context) {
	var req createAnimalSpeciesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	species, err := h.animalSpecies.Create(c.Request.Context(), req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/masters/animal-species/%d", species.ID))
	c.JSON(http.StatusCreated, toAnimalSpeciesResponse(species))
}

// UpdateAnimalSpecies godoc
func (h *Handler) UpdateAnimalSpecies(c *gin.Context) {
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var req updateAnimalSpeciesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	species, err := h.animalSpecies.Update(c.Request.Context(), id, req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toAnimalSpeciesResponse(species))
}

// DeleteAnimalSpecies godoc
func (h *Handler) DeleteAnimalSpecies(c *gin.Context) {
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.animalSpecies.Delete(c.Request.Context(), id); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderAnimalSpecies は動物種マスタの表示順を更新する。
// AnimalSpecies はシステム共通マスタ（clinic_id なし）のため clinicID パラメータは不要。
func (h *Handler) ReorderAnimalSpecies(c *gin.Context) {
	var req httpapi.ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if err := h.animalSpecies.Reorder(c.Request.Context(), req.IDs); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
