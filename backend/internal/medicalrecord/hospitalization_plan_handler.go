// Package handler provides HTTP handler implementations for HospitalizationPlan entity.
package medicalrecord

import (
	"fmt"
	"net/http"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// HospitalizationPlanHandler serves the hospitalization-plan HTTP boundary. Moved from internal/handler
// (BE9-2D ⑤ Batch C).
type HospitalizationPlanHandler struct {
	service HospitalizationPlanService
}

// NewHospitalizationPlanHandler initializes a HospitalizationPlanHandler.
func NewHospitalizationPlanHandler(service HospitalizationPlanService) *HospitalizationPlanHandler {
	return &HospitalizationPlanHandler{service: service}
}

// ---- HospitalizationPlan ----

// GetHospitalizationPlan godoc
func (h *HospitalizationPlanHandler) GetHospitalizationPlan(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	plan, err := h.service.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toHospitalizationPlanResponse(plan))
}

// ListHospitalizationPlans godoc
func (h *HospitalizationPlanHandler) ListHospitalizationPlans(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	plans, err := h.service.List(c.Request.Context(), clinicID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(plans, toHospitalizationPlanResponse))
}

// CreateHospitalizationPlan godoc
func (h *HospitalizationPlanHandler) CreateHospitalizationPlan(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	var req createHospitalizationPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	plan, err := h.service.Create(c.Request.Context(), clinicID, req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/hospitalization-plans/%d", plan.ID))
	c.JSON(http.StatusCreated, toHospitalizationPlanResponse(plan))
}

// UpdateHospitalizationPlan godoc
func (h *HospitalizationPlanHandler) UpdateHospitalizationPlan(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateHospitalizationPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	plan, err := h.service.Update(c.Request.Context(), clinicID, id, req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toHospitalizationPlanResponse(plan))
}

// DeleteHospitalizationPlan godoc
func (h *HospitalizationPlanHandler) DeleteHospitalizationPlan(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), clinicID, id); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderHospitalizationPlans godoc
func (h *HospitalizationPlanHandler) ReorderHospitalizationPlans(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req httpapi.ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if err := h.service.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
