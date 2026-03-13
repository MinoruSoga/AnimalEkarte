// Package handler provides HTTP handler implementations for DiagnosisCategory and DiagnosisName entities.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

type createDiagnosisCategoryInput struct {
	Name        string `json:"name"        binding:"required"`
	IsActive    bool   `json:"is_active"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

type updateDiagnosisCategoryInput struct {
	Name        string `json:"name"`
	IsActive    *bool  `json:"is_active"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

type createDiagnosisNameInput struct {
	Name                string `json:"name"                  binding:"required"`
	DiagnosisCategoryID uint64 `json:"diagnosis_category_id" binding:"required"`
	IsActive            bool   `json:"is_active"`
	Description         string `json:"description"`
	SortOrder           int    `json:"sort_order"`
}

type updateDiagnosisNameInput struct {
	Name                string `json:"name"`
	DiagnosisCategoryID uint64 `json:"diagnosis_category_id"`
	IsActive            *bool  `json:"is_active"`
	Description         string `json:"description"`
	SortOrder           int    `json:"sort_order"`
}

// ---- DiagnosisCategory ----

// ListDiagnosisCategories godoc
func (h *Handler) ListDiagnosisCategories(c *gin.Context) {
	categories, err := h.svc.DiagnosisCategory.List(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, categories)
}

// CreateDiagnosisCategory godoc
func (h *Handler) CreateDiagnosisCategory(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var input createDiagnosisCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category := &model.DiagnosisCategory{
		ClinicID:    clinicID,
		Name:        input.Name,
		IsActive:    input.IsActive,
		Description: input.Description,
		SortOrder:   input.SortOrder,
	}

	if err := h.svc.DiagnosisCategory.Create(c.Request.Context(), category); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, category)
}

// UpdateDiagnosisCategory godoc
func (h *Handler) UpdateDiagnosisCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input updateDiagnosisCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category := &model.DiagnosisCategory{
		ID:          id,
		Name:        input.Name,
		Description: input.Description,
		SortOrder:   input.SortOrder,
	}
	if input.IsActive != nil {
		category.IsActive = *input.IsActive
	}

	if err := h.svc.DiagnosisCategory.Update(c.Request.Context(), category); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, category)
}

// DeleteDiagnosisCategory godoc
func (h *Handler) DeleteDiagnosisCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.DiagnosisCategory.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ---- DiagnosisName ----

// ListDiagnosisNames godoc
func (h *Handler) ListDiagnosisNames(c *gin.Context) {
	var names []model.DiagnosisName
	var err error

	if catIDStr := c.Query("category_id"); catIDStr != "" {
		catID, parseErr := strconv.ParseUint(catIDStr, 10, 64)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
			return
		}
		names, err = h.svc.DiagnosisName.ListByCategoryID(c.Request.Context(), catID)
	} else {
		names, err = h.svc.DiagnosisName.List(c.Request.Context())
	}

	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, names)
}

// CreateDiagnosisName godoc
func (h *Handler) CreateDiagnosisName(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var input createDiagnosisNameInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	diagnosisName := &model.DiagnosisName{
		ClinicID:            clinicID,
		Name:                input.Name,
		DiagnosisCategoryID: input.DiagnosisCategoryID,
		IsActive:            input.IsActive,
		Description:         input.Description,
		SortOrder:           input.SortOrder,
	}

	if err := h.svc.DiagnosisName.Create(c.Request.Context(), diagnosisName); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, diagnosisName)
}

// UpdateDiagnosisName godoc
func (h *Handler) UpdateDiagnosisName(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input updateDiagnosisNameInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	diagnosisName := &model.DiagnosisName{
		ID:                  id,
		Name:                input.Name,
		DiagnosisCategoryID: input.DiagnosisCategoryID,
		Description:         input.Description,
		SortOrder:           input.SortOrder,
	}
	if input.IsActive != nil {
		diagnosisName.IsActive = *input.IsActive
	}

	if err := h.svc.DiagnosisName.Update(c.Request.Context(), diagnosisName); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, diagnosisName)
}

// DeleteDiagnosisName godoc
func (h *Handler) DeleteDiagnosisName(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.DiagnosisName.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
