// Package handler provides HTTP handler implementations for HospitalizationPlan entity.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

type createHospitalizationPlanInput struct {
	Name        string   `json:"name"         binding:"required"`
	Price       *float64 `json:"price"`
	IsActive    bool     `json:"is_active"`
	Description string   `json:"description"`
	BodySize    string   `json:"body_size"`
	BillingUnit string   `json:"billing_unit"`
	SortOrder   int      `json:"sort_order"`
}

type updateHospitalizationPlanInput struct {
	Name        string   `json:"name"`
	Price       *float64 `json:"price"`
	IsActive    *bool    `json:"is_active"`
	Description string   `json:"description"`
	BodySize    string   `json:"body_size"`
	BillingUnit string   `json:"billing_unit"`
	SortOrder   int      `json:"sort_order"`
}

// ---- HospitalizationPlan ----

// ListHospitalizationPlans godoc
func (h *Handler) ListHospitalizationPlans(c *gin.Context) {
	plans, err := h.svc.HospitalizationPlan.List(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, plans)
}

// CreateHospitalizationPlan godoc
func (h *Handler) CreateHospitalizationPlan(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var input createHospitalizationPlanInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	c.JSON(http.StatusCreated, plan)
}

// UpdateHospitalizationPlan godoc
func (h *Handler) UpdateHospitalizationPlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input updateHospitalizationPlanInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	plan := &model.HospitalizationPlan{
		ID:          id,
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
	c.JSON(http.StatusOK, plan)
}

// DeleteHospitalizationPlan godoc
func (h *Handler) DeleteHospitalizationPlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.HospitalizationPlan.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
