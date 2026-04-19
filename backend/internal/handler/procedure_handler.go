// Package handler provides HTTP handler implementations for Procedure entity.
package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
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
	c.JSON(http.StatusOK, toProcedureResponse(procedure))
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
	c.JSON(http.StatusOK, mapSlice(procedures, toProcedureResponse))
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

	var taxType *string
	if input.TaxType != "" {
		t := input.TaxType
		taxType = &t
	}
	svcInput := &service.CreateProcedureInput{
		Name:        input.Name,
		Price:       input.Price,
		IsActive:    input.IsActive,
		Description: input.Description,
		Duration:    input.Duration,
		Anesthesia:  input.Anesthesia,
		ParentID:    input.ParentID,
		SortOrder:   input.SortOrder,
		TaxType:     taxType,
		TaxRate:     input.TaxRate,
	}

	procedure, err := h.svc.Procedure.Create(c.Request.Context(), clinicID, svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/procedures/%d", procedure.ID))
	c.JSON(http.StatusCreated, toProcedureResponse(procedure))
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
		Anesthesia:    input.Anesthesia,
		ParentID:      input.ParentID,
		ClearParentID: input.ClearParentID,
		SortOrder:     input.SortOrder,
		TaxType:       input.TaxType,
		TaxRate:       input.TaxRate,
	}

	procedure, err := h.svc.Procedure.Update(c.Request.Context(), clinicID, id, &svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toProcedureResponse(procedure))
}

// ReorderProcedures godoc
func (h *Handler) ReorderProcedures(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderRequest
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
