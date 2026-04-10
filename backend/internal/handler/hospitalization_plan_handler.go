// Package handler provides HTTP handler implementations for HospitalizationPlan entity.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
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
	c.JSON(http.StatusOK, toHospitalizationPlanResponseList(plans))
}

// CreateHospitalizationPlan godoc
func (h *Handler) CreateHospitalizationPlan(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var input createHospitalizationPlanRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	taxType := model.TaxTypeExcluded
	if input.TaxType != "" {
		taxType = model.TaxType(input.TaxType)
	}
	taxRate := 0.10
	if input.TaxRate != nil {
		taxRate = *input.TaxRate
	}
	plan := &model.HospitalizationPlan{
		ClinicID:    clinicID,
		Name:        input.Name,
		Price:       input.Price,
		IsActive:    input.IsActive,
		Description: input.Description,
		SortOrder:   input.SortOrder,
		TaxType:     taxType,
		TaxRate:     taxRate,
	}
	if input.BodySize != "" {
		bs := model.BodySize(input.BodySize)
		plan.BodySize = &bs
	}
	if input.BillingUnit != "" {
		bu := model.BillingUnit(input.BillingUnit)
		plan.BillingUnit = &bu
	}

	if err := h.svc.HospitalizationPlan.Create(c.Request.Context(), plan); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toHospitalizationPlanResponse(plan))
}

// UpdateHospitalizationPlan godoc
func (h *Handler) UpdateHospitalizationPlan(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var input updateHospitalizationPlanRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	var bodySize *model.BodySize
	if input.BodySize != nil {
		bs := model.BodySize(*input.BodySize)
		bodySize = &bs
	}
	var billingUnit *model.BillingUnit
	if input.BillingUnit != nil {
		bu := model.BillingUnit(*input.BillingUnit)
		billingUnit = &bu
	}
	var taxType *model.TaxType
	if input.TaxType != nil {
		tt := model.TaxType(*input.TaxType)
		taxType = &tt
	}

	svcInput := service.UpdateHospitalizationPlanInput{
		Name:        input.Name,
		Price:       input.Price,
		IsActive:    input.IsActive,
		Description: input.Description,
		BodySize:    bodySize,
		BillingUnit: billingUnit,
		SortOrder:   input.SortOrder,
		TaxType:     taxType,
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
	c.JSON(http.StatusOK, gin.H{"message": "reordered"})
}
