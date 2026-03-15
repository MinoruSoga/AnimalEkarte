// Package handler provides HTTP handler implementations for DiagnosisCategory and DiagnosisName entities.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- DiagnosisCategory ----

// GetDiagnosisCategory godoc
func (h *Handler) GetDiagnosisCategory(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	category, err := h.svc.DiagnosisCategory.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toDiagnosisCategoryResponse(category))
}

// ListDiagnosisCategories godoc
func (h *Handler) ListDiagnosisCategories(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	categories, err := h.svc.DiagnosisCategory.List(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toDiagnosisCategoryResponseList(categories))
}

// CreateDiagnosisCategory godoc
func (h *Handler) CreateDiagnosisCategory(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req createDiagnosisCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	category, err := h.svc.DiagnosisCategory.Create(c.Request.Context(), clinicID, &service.CreateDiagnosisCategoryInput{
		Name:        req.Name,
		IsActive:    req.IsActive,
		Description: req.Description,
		SortOrder:   req.SortOrder,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toDiagnosisCategoryResponse(category))
}

// UpdateDiagnosisCategory godoc
func (h *Handler) UpdateDiagnosisCategory(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateDiagnosisCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	category, err := h.svc.DiagnosisCategory.Update(c.Request.Context(), clinicID, id, &service.UpdateDiagnosisCategoryInput{
		Name:        req.Name,
		IsActive:    req.IsActive,
		Description: req.Description,
		SortOrder:   req.SortOrder,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toDiagnosisCategoryResponse(category))
}

// DeleteDiagnosisCategory godoc
func (h *Handler) DeleteDiagnosisCategory(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.DiagnosisCategory.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderDiagnosisCategories godoc (#019)
func (h *Handler) ReorderDiagnosisCategories(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderDiagnosisCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}
	if err := h.svc.DiagnosisCategory.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "reordered"})
}

// ---- DiagnosisName ----

// GetDiagnosisName godoc
func (h *Handler) GetDiagnosisName(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	diagnosisName, err := h.svc.DiagnosisName.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toDiagnosisNameResponse(diagnosisName))
}

// ListDiagnosisNames godoc
func (h *Handler) ListDiagnosisNames(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var names any
	var svcErr error

	if catIDStr := c.Query("category_id"); catIDStr != "" {
		catID, parseErr := strconv.ParseUint(catIDStr, 10, 64)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
			return
		}
		result, err := h.svc.DiagnosisName.ListByCategoryID(c.Request.Context(), clinicID, catID)
		names = toDiagnosisNameResponseList(result)
		svcErr = err
	} else {
		result, err := h.svc.DiagnosisName.List(c.Request.Context(), clinicID)
		names = toDiagnosisNameResponseList(result)
		svcErr = err
	}

	if svcErr != nil {
		RespondError(c, svcErr)
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

	var req createDiagnosisNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	diagnosisName, err := h.svc.DiagnosisName.Create(c.Request.Context(), clinicID, &service.CreateDiagnosisNameInput{
		Name:                req.Name,
		DiagnosisCategoryID: req.DiagnosisCategoryID,
		IsActive:            req.IsActive,
		Description:         req.Description,
		SortOrder:           req.SortOrder,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toDiagnosisNameResponse(diagnosisName))
}

// UpdateDiagnosisName godoc
func (h *Handler) UpdateDiagnosisName(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateDiagnosisNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	diagnosisName, err := h.svc.DiagnosisName.Update(c.Request.Context(), clinicID, id, &service.UpdateDiagnosisNameInput{
		Name:                req.Name,
		DiagnosisCategoryID: req.DiagnosisCategoryID,
		IsActive:            req.IsActive,
		Description:         req.Description,
		SortOrder:           req.SortOrder,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toDiagnosisNameResponse(diagnosisName))
}

// DeleteDiagnosisName godoc
func (h *Handler) DeleteDiagnosisName(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.DiagnosisName.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderDiagnosisNames godoc (#019)
func (h *Handler) ReorderDiagnosisNames(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderDiagnosisNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}
	if err := h.svc.DiagnosisName.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "reordered"})
}
