// Package handler provides HTTP handler implementations for DiagnosisType and DiagnosisName entities.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- DiagnosisType ----

// GetDiagnosisType godoc
func (h *Handler) GetDiagnosisType(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	category, err := h.svc.DiagnosisType.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toDiagnosisTypeResponse(category))
}

// ListDiagnosisTypes godoc
func (h *Handler) ListDiagnosisTypes(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	categories, err := h.svc.DiagnosisType.List(c.Request.Context(), clinicID, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(categories, toDiagnosisTypeResponse))
}

// CreateDiagnosisType godoc
func (h *Handler) CreateDiagnosisType(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req createDiagnosisTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	category, err := h.svc.DiagnosisType.Create(c.Request.Context(), clinicID, &service.CreateDiagnosisTypeInput{
		Name:        req.Name,
		IsActive:    req.IsActive,
		Description: req.Description,
		SortOrder:   req.SortOrder,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toDiagnosisTypeResponse(category))
}

// UpdateDiagnosisType godoc
func (h *Handler) UpdateDiagnosisType(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req updateDiagnosisTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	category, err := h.svc.DiagnosisType.Update(c.Request.Context(), clinicID, id, &service.UpdateDiagnosisTypeInput{
		Name:        req.Name,
		IsActive:    req.IsActive,
		Description: req.Description,
		SortOrder:   req.SortOrder,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toDiagnosisTypeResponse(category))
}

// DeleteDiagnosisType godoc
func (h *Handler) DeleteDiagnosisType(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.DiagnosisType.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderDiagnosisTypes godoc (#019)
func (h *Handler) ReorderDiagnosisTypes(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderDiagnosisTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.DiagnosisType.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ---- DiagnosisName ----

// GetDiagnosisName godoc
func (h *Handler) GetDiagnosisName(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
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

	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	var resp any
	if typeIDStr := c.Query("type_id"); typeIDStr != "" {
		catID, parseErr := strconv.ParseUint(typeIDStr, 10, 64)
		if parseErr != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid type_id"))
			return
		}
		names, _, svcErr := h.svc.DiagnosisName.ListByCategoryID(c.Request.Context(), clinicID, catID, page, limit)
		if svcErr != nil {
			RespondError(c, svcErr)
			return
		}
		resp = mapSlice(names, toDiagnosisNameResponse)
	} else {
		names, _, svcErr := h.svc.DiagnosisName.List(c.Request.Context(), clinicID, page, limit)
		if svcErr != nil {
			RespondError(c, svcErr)
			return
		}
		resp = mapSlice(names, toDiagnosisNameResponse)
	}

	c.JSON(http.StatusOK, resp)
}

// CreateDiagnosisName godoc
func (h *Handler) CreateDiagnosisName(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req createDiagnosisNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	diagnosisName, err := h.svc.DiagnosisName.Create(c.Request.Context(), clinicID, &service.CreateDiagnosisNameInput{
		Name:            req.Name,
		DiagnosisTypeID: req.DiagnosisTypeID,
		IsActive:        req.IsActive,
		Description:     req.Description,
		SortOrder:       req.SortOrder,
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
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req updateDiagnosisNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	diagnosisName, err := h.svc.DiagnosisName.Update(c.Request.Context(), clinicID, id, &service.UpdateDiagnosisNameInput{
		Name:            req.Name,
		DiagnosisTypeID: req.DiagnosisTypeID,
		IsActive:        req.IsActive,
		Description:     req.Description,
		SortOrder:       req.SortOrder,
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
	id, ok := parseIDParam(c, "id")
	if !ok {
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
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.DiagnosisName.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
