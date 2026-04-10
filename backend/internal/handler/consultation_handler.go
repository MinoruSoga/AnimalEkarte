// Package handler provides HTTP handler implementations for Consultation entity.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- Consultation ----

// GetConsultation godoc
func (h *Handler) GetConsultation(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	consultation, err := h.svc.Consultation.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, consultation)
}

// ListConsultations godoc
func (h *Handler) ListConsultations(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	consultations, err := h.svc.Consultation.List(c.Request.Context(), clinicID)
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
		TaxType:       taxType,
		TaxRate:       taxRate,
	}

	if err := h.svc.Consultation.Create(c.Request.Context(), consultation); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, consultation)
}

// UpdateConsultation godoc
func (h *Handler) UpdateConsultation(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var input updateConsultationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	var taxType *model.TaxType
	if input.TaxType != nil {
		tt := model.TaxType(*input.TaxType)
		taxType = &tt
	}

	svcInput := service.UpdateConsultationInput{
		Name:          input.Name,
		Price:         input.Price,
		IsActive:      input.IsActive,
		Description:   input.Description,
		TimeCondition: input.TimeCondition,
		Duration:      input.Duration,
		ParentID:      input.ParentID,
		ClearParentID: input.ClearParentID,
		SortOrder:     input.SortOrder,
		TaxType:       taxType,
		TaxRate:       input.TaxRate,
	}

	consultation, err := h.svc.Consultation.Update(c.Request.Context(), clinicID, id, &svcInput)
	if err != nil {
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
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
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
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Consultation.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
