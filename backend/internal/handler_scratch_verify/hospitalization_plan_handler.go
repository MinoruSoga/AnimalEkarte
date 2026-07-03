// Package handler provides HTTP handler implementations for HospitalizationPlan entity.
package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// ---- HospitalizationPlan ----

// GetHospitalizationPlan godoc
func (h *Handler) GetHospitalizationPlan(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	plan, err := h.svc.HospitalizationPlan.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toHospitalizationPlanResponse(plan))
}

// ListHospitalizationPlans godoc
func (h *Handler) ListHospitalizationPlans(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	plans, err := h.svc.HospitalizationPlan.List(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(plans, toHospitalizationPlanResponse))
}

// CreateHospitalizationPlan godoc
func (h *Handler) CreateHospitalizationPlan(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req createHospitalizationPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	plan, err := h.svc.HospitalizationPlan.Create(c.Request.Context(), clinicID, req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/hospitalization-plans/%d", plan.ID))
	c.JSON(http.StatusCreated, toHospitalizationPlanResponse(plan))
}

// UpdateHospitalizationPlan godoc
func (h *Handler) UpdateHospitalizationPlan(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateHospitalizationPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	plan, err := h.svc.HospitalizationPlan.Update(c.Request.Context(), clinicID, id, req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toHospitalizationPlanResponse(plan))
}

// DeleteHospitalizationPlan godoc
func (h *Handler) DeleteHospitalizationPlan(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.HospitalizationPlan.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderHospitalizationPlans godoc
func (h *Handler) ReorderHospitalizationPlans(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.HospitalizationPlan.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
