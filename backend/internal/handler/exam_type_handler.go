// Package handler provides HTTP handler implementations for ExaminationType entity.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

type createExaminationTypeInput struct {
	Name        string   `json:"name"        binding:"required"`
	Price       *float64 `json:"price"`
	IsActive    bool     `json:"is_active"`
	Description string   `json:"description"`
	SortOrder   int      `json:"sort_order"`
}

type updateExaminationTypeInput struct {
	Name        string   `json:"name"`
	Price       *float64 `json:"price"`
	IsActive    *bool    `json:"is_active"`
	Description string   `json:"description"`
	SortOrder   int      `json:"sort_order"`
}

// ---- ExaminationType ----

// ListExaminationTypes godoc
func (h *Handler) ListExaminationTypes(c *gin.Context) {
	exTypes, err := h.svc.ExaminationType.List(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, exTypes)
}

// CreateExaminationType godoc
func (h *Handler) CreateExaminationType(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var input createExaminationTypeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	examType := &model.ExamType{
		ClinicID:    clinicID,
		Name:        input.Name,
		Price:       input.Price,
		IsActive:    input.IsActive,
		Description: input.Description,
		SortOrder:   input.SortOrder,
	}

	if err := h.svc.ExaminationType.Create(c.Request.Context(), examType); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, examType)
}

// UpdateExaminationType godoc
func (h *Handler) UpdateExaminationType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input updateExaminationTypeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	examType := &model.ExamType{
		ID:          id,
		Name:        input.Name,
		Price:       input.Price,
		Description: input.Description,
		SortOrder:   input.SortOrder,
	}
	if input.IsActive != nil {
		examType.IsActive = *input.IsActive
	}

	if err := h.svc.ExaminationType.Update(c.Request.Context(), examType); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, examType)
}

// DeleteExaminationType godoc
func (h *Handler) DeleteExaminationType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.ExaminationType.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
