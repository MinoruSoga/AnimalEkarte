// Package handler provides HTTP handler implementations for HospitalizationPlan entity.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
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

	svcInput := &service.CreateHospitalizationPlanInput{
		Name:        req.Name,
		Price:       req.Price,
		IsActive:    req.IsActive,
		Description: req.Description,
		SortOrder:   req.SortOrder,
		TaxType:     req.TaxType,
		TaxRate:     req.TaxRate,
		BodySize:    req.BodySize,
		BillingUnit: req.BillingUnit,
	}
	plan, err := h.svc.HospitalizationPlan.Create(c.Request.Context(), clinicID, svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
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
	var input updateHospitalizationPlanRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	svcInput := service.UpdateHospitalizationPlanInput{
		Name:        input.Name,
		Price:       input.Price,
		IsActive:    input.IsActive,
		Description: input.Description,
		BodySize:    input.BodySize,
		BillingUnit: input.BillingUnit,
		SortOrder:   input.SortOrder,
		TaxType:     input.TaxType,
		TaxRate:     input.TaxRate,
	}

	plan, err := h.svc.HospitalizationPlan.Update(c.Request.Context(), clinicID, id, svcInput)
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
	var req reorderHospitalizationPlanRequest
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
