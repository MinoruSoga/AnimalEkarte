// Package handler provides HTTP handler implementations for Consultation entity.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Consultation ----

// ListConsultations godoc
func (h *Handler) ListConsultations(c *gin.Context) {
	consultations, err := h.svc.Consultation.List(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, consultations)
}

// CreateConsultation godoc
func (h *Handler) CreateConsultation(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var input createConsultationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	consultation := &model.Consultation{
		ClinicID:      clinicID,
		Name:          input.Name,
		Price:         input.Price,
		IsActive:      input.IsActive,
		Description:   input.Description,
		TimeCondition: input.TimeCondition,
		Duration:      input.Duration,
		ParentID:      input.ParentID,
		SortOrder:     input.SortOrder,
	}

	if err := h.svc.Consultation.Create(c.Request.Context(), consultation); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, consultation)
}

// UpdateConsultation godoc
func (h *Handler) UpdateConsultation(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input updateConsultationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	consultation := &model.Consultation{
		ID:            id,
		Name:          input.Name,
		Price:         input.Price,
		Description:   input.Description,
		TimeCondition: input.TimeCondition,
		Duration:      input.Duration,
		SortOrder:     input.SortOrder,
	}
	if input.IsActive != nil {
		consultation.IsActive = *input.IsActive
	}
	if input.ClearParentID {
		consultation.ParentID = nil
	} else if input.ParentID != nil {
		consultation.ParentID = input.ParentID
	}

	if err := h.svc.Consultation.Update(c.Request.Context(), consultation); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, consultation)
}

// ReorderConsultations godoc
func (h *Handler) ReorderConsultations(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderConsultationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}
	if err := h.svc.Consultation.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "reordered"})
}

// DeleteConsultation godoc
func (h *Handler) DeleteConsultation(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Consultation.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
