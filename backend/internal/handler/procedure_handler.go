// Package handler provides HTTP handler implementations for Procedure entity.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Procedure ----

// GetProcedure godoc
func (h *Handler) GetProcedure(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	procedure, err := h.svc.Procedure.GetByID(c.Request.Context(), id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, procedure)
}

// ListProcedures godoc
func (h *Handler) ListProcedures(c *gin.Context) {
	procedures, err := h.svc.Procedure.List(c.Request.Context())
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
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
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
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input updateProcedureRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	procedure := &model.Procedure{
		ID:          id,
		Name:        input.Name,
		Price:       input.Price,
		Description: input.Description,
		Duration:    input.Duration,
		SortOrder:   input.SortOrder,
	}
	if input.IsActive != nil {
		procedure.IsActive = *input.IsActive
	}
	if input.Anesthesia != "" {
		procedure.Anesthesia = model.AnesthesiaType(input.Anesthesia)
	}
	if input.ClearParentID {
		procedure.ParentID = nil
	} else if input.ParentID != nil {
		procedure.ParentID = input.ParentID
	}

	if err := h.svc.Procedure.Update(c.Request.Context(), procedure); err != nil {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}
	if err := h.svc.Procedure.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "reordered"})
}

// DeleteProcedure godoc
func (h *Handler) DeleteProcedure(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Procedure.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
