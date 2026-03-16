// Package handler provides HTTP handler implementations for HospitalizationPlan entity.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- HospitalizationPlan ----

// GetHospitalizationPlan godoc
func (h *Handler) GetHospitalizationPlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	plan, err := h.svc.HospitalizationPlan.GetByID(c.Request.Context(), id)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	plan := &model.HospitalizationPlan{
		ClinicID:    clinicID,
		Name:        input.Name,
		Price:       input.Price,
		IsActive:    input.IsActive,
		Description: input.Description,
		SortOrder:   input.SortOrder,
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
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var input updateHospitalizationPlanRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	plan := &model.HospitalizationPlan{
		ID:          id,
		ClinicID:    clinicID,
		Name:        input.Name,
		Price:       input.Price,
		Description: input.Description,
		SortOrder:   input.SortOrder,
	}
	if input.IsActive != nil {
		plan.IsActive = *input.IsActive
	}
	if input.BodySize != "" {
		bs := model.BodySize(input.BodySize)
		plan.BodySize = &bs
	}
	if input.BillingUnit != "" {
		bu := model.BillingUnit(input.BillingUnit)
		plan.BillingUnit = &bu
	}

	if err := h.svc.HospitalizationPlan.Update(c.Request.Context(), plan); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toHospitalizationPlanResponse(plan))
}

// DeleteHospitalizationPlan godoc
func (h *Handler) DeleteHospitalizationPlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if err := h.svc.HospitalizationPlan.Delete(c.Request.Context(), id); err != nil {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}
	if err := h.svc.HospitalizationPlan.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "reordered"})
}
