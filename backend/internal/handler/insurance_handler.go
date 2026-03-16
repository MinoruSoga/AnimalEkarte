// Package handler provides HTTP handler implementations for Insurance entity.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Insurance ----

// ListInsurances godoc
func (h *Handler) ListInsurances(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	insurances, err := h.svc.Insurance.List(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toInsuranceResponseList(insurances))
}

// GetInsurance godoc
func (h *Handler) GetInsurance(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	insurance, err := h.svc.Insurance.GetByID(c.Request.Context(), id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toInsuranceResponse(insurance))
}

// CreateInsurance godoc
func (h *Handler) CreateInsurance(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req createInsuranceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	coverageRate := 0
	if req.CoverageRate != nil {
		coverageRate = *req.CoverageRate
	}
	insurance := &model.Insurance{
		ClinicID:     clinicID,
		Name:         req.Name,
		IsActive:     req.IsActive,
		Description:  req.Description,
		CoverageRate: coverageRate,
		ContactPhone: req.ContactPhone,
		SortOrder:    req.SortOrder,
	}

	if err := h.svc.Insurance.Create(c.Request.Context(), insurance); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toInsuranceResponse(insurance))
}

// UpdateInsurance godoc
func (h *Handler) UpdateInsurance(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req updateInsuranceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	coverageRate := 0
	if req.CoverageRate != nil {
		coverageRate = *req.CoverageRate
	}
	insurance := &model.Insurance{
		ID:           id,
		ClinicID:     clinicID,
		Name:         req.Name,
		Description:  req.Description,
		CoverageRate: coverageRate,
		ContactPhone: req.ContactPhone,
		SortOrder:    req.SortOrder,
	}
	if req.IsActive != nil {
		insurance.IsActive = *req.IsActive
	}

	if err := h.svc.Insurance.Update(c.Request.Context(), insurance); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toInsuranceResponse(insurance))
}

// DeleteInsurance godoc
func (h *Handler) DeleteInsurance(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if err := h.svc.Insurance.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderInsurances godoc
func (h *Handler) ReorderInsurances(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderInsuranceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}
	if err := h.svc.Insurance.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "reordered"})
}
