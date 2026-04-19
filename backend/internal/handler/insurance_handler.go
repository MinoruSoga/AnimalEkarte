// Package handler provides HTTP handler implementations for Insurance entity.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
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
	c.JSON(http.StatusOK, mapSlice(insurances, toInsuranceResponse))
}

// GetInsurance godoc
func (h *Handler) GetInsurance(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	insurance, err := h.svc.Insurance.GetByID(c.Request.Context(), clinicID, id)
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
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	coverageRate := 0
	if req.CoverageRate != nil {
		coverageRate = *req.CoverageRate
	}
	svcInput := &service.CreateInsuranceInput{
		Name:         req.Name,
		IsActive:     req.IsActive,
		Description:  req.Description,
		CoverageRate: coverageRate,
		ContactPhone: req.ContactPhone,
		SortOrder:    req.SortOrder,
	}

	insurance, err := h.svc.Insurance.Create(c.Request.Context(), clinicID, svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toInsuranceResponse(insurance))
}

// UpdateInsurance godoc
func (h *Handler) UpdateInsurance(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateInsuranceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	svcInput := service.UpdateInsuranceInput{
		Name:         req.Name,
		IsActive:     req.IsActive,
		Description:  req.Description,
		CoverageRate: req.CoverageRate,
		ContactPhone: req.ContactPhone,
		SortOrder:    req.SortOrder,
	}

	insurance, err := h.svc.Insurance.Update(c.Request.Context(), clinicID, id, &svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toInsuranceResponse(insurance))
}

// DeleteInsurance godoc
func (h *Handler) DeleteInsurance(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Insurance.Delete(c.Request.Context(), clinicID, id); err != nil {
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
	var req reorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.Insurance.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
