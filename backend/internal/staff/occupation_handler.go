// Package handler provides HTTP handler implementations for Occupation entity.
package staff

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"
)

// ---- Occupation ----

// ListOccupations godoc
func (h *Handler) ListOccupations(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	occupations, err := h.svc.Occupation.List(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(occupations, toOccupationResponse))
}

// GetOccupation godoc
func (h *Handler) GetOccupation(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	occ, err := h.svc.Occupation.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toOccupationResponse(occ))
}

// CreateOccupation godoc
func (h *Handler) CreateOccupation(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req createOccupationRequest
	if err := bindStaffJSON(c, &req); err != nil {
		RespondError(c, err)
		return
	}

	occ, err := h.svc.Occupation.Create(c.Request.Context(), clinicID, req.toServiceInput())
	if err != nil {
		httpapi.RespondErrorPreferringConflictCode(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/occupations/%d", occ.ID))
	c.JSON(http.StatusCreated, toOccupationResponse(occ))
}

// UpdateOccupation godoc
func (h *Handler) UpdateOccupation(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req updateOccupationRequest
	if err := bindStaffJSON(c, &req); err != nil {
		RespondError(c, err)
		return
	}

	updated, err := h.svc.Occupation.Update(c.Request.Context(), clinicID, id, req.toServiceInput())
	if err != nil {
		httpapi.RespondErrorPreferringConflictCode(c, err)
		return
	}
	c.JSON(http.StatusOK, toOccupationResponse(updated))
}

// DeleteOccupation godoc
func (h *Handler) DeleteOccupation(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Occupation.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderOccupations godoc
func (h *Handler) ReorderOccupations(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderRequest
	if err := bindStaffJSON(c, &req); err != nil {
		RespondError(c, err)
		return
	}
	if err := h.svc.Occupation.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
