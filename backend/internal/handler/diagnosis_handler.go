// Package handler provides HTTP handler implementations for DiagnosisCategory and DiagnosisName entities.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

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
	var input model.DiagnosisCategory
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.DiagnosisCategory.Create(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, input)
}

// UpdateDiagnosisCategory godoc
func (h *Handler) UpdateDiagnosisCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input model.DiagnosisCategory
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id
	if err := h.svc.DiagnosisCategory.Update(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, input)
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
	var input model.DiagnosisName
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.DiagnosisName.Create(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, input)
}

// UpdateDiagnosisName godoc
func (h *Handler) UpdateDiagnosisName(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input model.DiagnosisName
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id
	if err := h.svc.DiagnosisName.Update(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, input)
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
