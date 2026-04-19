// Package handler provides HTTP handler implementations for Procedure entity.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- Procedure ----

// GetProcedure godoc
func (h *Handler) GetProcedure(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	procedure, err := h.svc.Procedure.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, procedure)
}

// ListProcedures godoc
func (h *Handler) ListProcedures(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	procedures, err := h.svc.Procedure.List(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, procedures)
}

// CreateProcedure godoc
func (h *Handler) CreateProcedure(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var input createProcedureRequest
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
	procedure := &model.Procedure{
		ClinicID:    clinicID,
		Name:        input.Name,
		Price:       input.Price,
		IsActive:    input.IsActive,
		Description: input.Description,
		Duration:    input.Duration,
		ParentID:    input.ParentID,
		SortOrder:   input.SortOrder,
		TaxType:     taxType,
		TaxRate:     taxRate,
	}
	if input.Anesthesia != "" {
		procedure.Anesthesia = model.AnesthesiaType(input.Anesthesia)
	}

	if err := h.svc.Procedure.Create(c.Request.Context(), procedure); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, procedure)
}

// UpdateProcedure godoc
func (h *Handler) UpdateProcedure(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var input updateProcedureRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	svcInput := service.UpdateProcedureInput{
		Name:          input.Name,
		Price:         input.Price,
		IsActive:      input.IsActive,
		Description:   input.Description,
		Duration:      input.Duration,
		ParentID:      input.ParentID,
		ClearParentID: input.ClearParentID,
		SortOrder:     input.SortOrder,
		TaxRate:       input.TaxRate,
	}
	if input.Anesthesia != nil {
		a := model.AnesthesiaType(*input.Anesthesia)
		svcInput.Anesthesia = &a
	}
	if input.TaxType != nil {
		t := model.TaxType(*input.TaxType)
		svcInput.TaxType = &t
	}

	procedure, err := h.svc.Procedure.Update(c.Request.Context(), clinicID, id, &svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, procedure)
}

// ReorderProcedures godoc
func (h *Handler) ReorderProcedures(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderProcedureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.Procedure.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteProcedure godoc
func (h *Handler) DeleteProcedure(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Procedure.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
