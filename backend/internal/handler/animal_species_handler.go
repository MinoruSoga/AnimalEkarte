// Package handler provides HTTP handler implementations for AnimalSpecies entity.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListAnimalSpecies godoc
func (h *Handler) ListAnimalSpecies(c *gin.Context) {
	species, err := h.svc.AnimalSpecies.List(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, species)
}
