// Package pet provides HTTP handler implementations for AnimalSpecies entity.
package pet

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

// requireSystemAdminForGlobalMaster closes clinic-scoped RBAC writes against
// global animal_species facts (DEC-31 / POC-07 / U-X03-PET-SPECIES-AUDIT).
func requireSystemAdminForGlobalMaster(c *gin.Context) bool {
	isSystemAdmin, ok := httpapi.ExtractIsSystemAdmin(c)
	if !ok {
		return false
	}
	if !isSystemAdmin {
		httpapi.RespondError(c, apperrors.WrapForbidden("system administrator access required"))
		return false
	}
	return true
}

func animalSpeciesMutationMetaFromContext(c *gin.Context) (AnimalSpeciesMutationMeta, bool) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return AnimalSpeciesMutationMeta{}, false
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return AnimalSpeciesMutationMeta{}, false
	}
	actorID := staffID
	return AnimalSpeciesMutationMeta{
		ClinicID:  clinicID,
		ActorID:   &actorID,
		ActorType: "staff",
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}, true
}

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
	if !requireSystemAdminForGlobalMaster(c) {
		return
	}
	meta, ok := animalSpeciesMutationMetaFromContext(c)
	if !ok {
		return
	}
	var req createAnimalSpeciesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	species, err := h.animalSpecies.Create(c.Request.Context(), req.toServiceInput(), meta)
	if err != nil {
		httpapi.RespondErrorPreferringConflictCode(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/masters/animal-species/%d", species.ID))
	c.JSON(http.StatusCreated, toAnimalSpeciesResponse(species))
}

// UpdateAnimalSpecies godoc
func (h *Handler) UpdateAnimalSpecies(c *gin.Context) {
	if !requireSystemAdminForGlobalMaster(c) {
		return
	}
	meta, ok := animalSpeciesMutationMetaFromContext(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var req updateAnimalSpeciesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	species, err := h.animalSpecies.Update(c.Request.Context(), id, req.toServiceInput(), meta)
	if err != nil {
		httpapi.RespondErrorPreferringConflictCode(c, err)
		return
	}
	c.JSON(http.StatusOK, toAnimalSpeciesResponse(species))
}

// DeleteAnimalSpecies godoc
func (h *Handler) DeleteAnimalSpecies(c *gin.Context) {
	if !requireSystemAdminForGlobalMaster(c) {
		return
	}
	meta, ok := animalSpeciesMutationMetaFromContext(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.animalSpecies.Delete(c.Request.Context(), id, meta); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderAnimalSpecies は動物種マスタの表示順を更新する。
// AnimalSpecies はシステム共通マスタ（clinic_id なし）のため clinicID パラメータは不要。
// 変更は system_admin のみ（DEC-31）。
func (h *Handler) ReorderAnimalSpecies(c *gin.Context) {
	if !requireSystemAdminForGlobalMaster(c) {
		return
	}
	meta, ok := animalSpeciesMutationMetaFromContext(c)
	if !ok {
		return
	}
	var req httpapi.ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if err := h.animalSpecies.Reorder(c.Request.Context(), req.IDs, meta); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
