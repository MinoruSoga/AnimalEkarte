// Package handler provides HTTP handler implementations for DiagnosisType and DiagnosisName entities.
package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- DiagnosisType ----

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

	categories, total, err := h.svc.DiagnosisType.List(c.Request.Context(), clinicID, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(mapSlice(categories, toDiagnosisTypeResponse), total, page, limit))
}

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
	c.Header("Location", fmt.Sprintf("/v1/masters/diagnosis-types/%d", category.ID))
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
	var req reorderRequest
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

	page, limit, paginationErr := parsePagination(c)
	if paginationErr != nil {
		RespondError(c, paginationErr)
		return
	}

	var typeID *uint64
	if s := c.Query("type_id"); s != "" {
		id, parseErr := strconv.ParseUint(s, 10, 64)
		if parseErr != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid type_id"))
			return
		}
		typeID = &id
	}

	names, total, err := h.svc.DiagnosisName.List(c.Request.Context(), clinicID, typeID, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(mapSlice(names, toDiagnosisNameResponse), total, page, limit))
}

// ListDiagnosisNamesAll godoc
// ページネーションなしで有効な診断名の一覧を返す。
// type_id クエリパラメータが指定された場合は該当カテゴリのみ、未指定の場合は全件を返す。
func (h *Handler) ListDiagnosisNamesAll(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var typeID *uint64
	if s := c.Query("type_id"); s != "" {
		id, parseErr := strconv.ParseUint(s, 10, 64)
		if parseErr != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid type_id"))
			return
		}
		typeID = &id
	}

	names, err := h.svc.DiagnosisName.ListNames(c.Request.Context(), clinicID, typeID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(names, toDiagnosisNameResponse))
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
	c.Header("Location", fmt.Sprintf("/v1/masters/diagnosis-names/%d", diagnosisName.ID))
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
	var req reorderRequest
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
