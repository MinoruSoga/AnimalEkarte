// Package handler provides HTTP handler implementations for Consultation entity.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

type createConsultationInput struct {
	Name          string   `json:"name"           binding:"required"`
	Price         *float64 `json:"price"`
	IsActive      bool     `json:"is_active"`
	Description   string   `json:"description"`
	TimeCondition string   `json:"time_condition"`
	Duration      *int     `json:"duration"`
	SortOrder     int      `json:"sort_order"`
}

type updateConsultationInput struct {
	Name          string   `json:"name"`
	Price         *float64 `json:"price"`
	IsActive      *bool    `json:"is_active"`
	Description   string   `json:"description"`
	TimeCondition string   `json:"time_condition"`
	Duration      *int     `json:"duration"`
	SortOrder     int      `json:"sort_order"`
}

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

	var input createConsultationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	var input updateConsultationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	if err := h.svc.Consultation.Update(c.Request.Context(), consultation); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, consultation)
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
