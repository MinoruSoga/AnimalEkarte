// Package handler provides HTTP handler implementations for ChiefComplaintType entity.
package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// ---- ChiefComplaintType ----

// ListChiefComplaints godoc
func (h *Handler) ListChiefComplaints(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	categories, err := h.svc.ChiefComplaintType.List(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(categories, toChiefComplaintResponse))
}

// GetChiefComplaint godoc
func (h *Handler) GetChiefComplaint(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	category, err := h.svc.ChiefComplaintType.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toChiefComplaintResponse(category))
}

// CreateChiefComplaint godoc
func (h *Handler) CreateChiefComplaint(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req createChiefComplaintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	category, err := h.svc.ChiefComplaintType.Create(c.Request.Context(), clinicID, req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/chief-complaint-types/%d", category.ID))
	c.JSON(http.StatusCreated, toChiefComplaintResponse(category))
}

// UpdateChiefComplaint godoc
func (h *Handler) UpdateChiefComplaint(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req updateChiefComplaintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	updated, err := h.svc.ChiefComplaintType.Update(c.Request.Context(), clinicID, id, req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toChiefComplaintResponse(updated))
}

// ReorderChiefComplaints godoc
func (h *Handler) ReorderChiefComplaints(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.ChiefComplaintType.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteChiefComplaint godoc
func (h *Handler) DeleteChiefComplaint(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.ChiefComplaintType.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
