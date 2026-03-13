// Package handler provides HTTP handler implementations for Insurance entity.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

type createInsuranceInput struct {
	Name         string `json:"name"          binding:"required"`
	IsActive     bool   `json:"is_active"`
	Description  string `json:"description"`
	CoverageRate *int   `json:"coverage_rate"`
	ContactPhone string `json:"contact_phone"`
	SortOrder    int    `json:"sort_order"`
}

type updateInsuranceInput struct {
	Name         string `json:"name"`
	IsActive     *bool  `json:"is_active"`
	Description  string `json:"description"`
	CoverageRate *int   `json:"coverage_rate"`
	ContactPhone string `json:"contact_phone"`
	SortOrder    int    `json:"sort_order"`
}

// ---- Insurance ----

// ListInsurances godoc
func (h *Handler) ListInsurances(c *gin.Context) {
	insurances, err := h.svc.Insurance.List(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, insurances)
}

// CreateInsurance godoc
func (h *Handler) CreateInsurance(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var input createInsuranceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	insurance := &model.Insurance{
		ClinicID:     clinicID,
		Name:         input.Name,
		IsActive:     input.IsActive,
		Description:  input.Description,
		CoverageRate: input.CoverageRate,
		ContactPhone: input.ContactPhone,
		SortOrder:    input.SortOrder,
	}

	if err := h.svc.Insurance.Create(c.Request.Context(), insurance); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, insurance)
}

// UpdateInsurance godoc
func (h *Handler) UpdateInsurance(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input updateInsuranceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	insurance := &model.Insurance{
		ID:           id,
		Name:         input.Name,
		Description:  input.Description,
		CoverageRate: input.CoverageRate,
		ContactPhone: input.ContactPhone,
		SortOrder:    input.SortOrder,
	}
	if input.IsActive != nil {
		insurance.IsActive = *input.IsActive
	}

	if err := h.svc.Insurance.Update(c.Request.Context(), insurance); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, insurance)
}

// DeleteInsurance godoc
func (h *Handler) DeleteInsurance(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Insurance.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
