// Package handler provides HTTP handler implementations for ExaminationType entity.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- ExaminationType ----

// GetExaminationType godoc
func (h *Handler) GetExaminationType(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	et, err := h.svc.ExaminationType.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toExamTypeResponse(et))
}

// ListExaminationTypes godoc
func (h *Handler) ListExaminationTypes(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	exTypes, err := h.svc.ExaminationType.List(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toExamTypeResponseList(exTypes))
}

// CreateExaminationType godoc
func (h *Handler) CreateExaminationType(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req createExaminationTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	examType := &model.ExaminationType{
		ClinicID:    clinicID,
		Name:        req.Name,
		Price:       req.Price,
		IsActive:    req.IsActive,
		Description: req.Description,
		ParentID:    req.ParentID,
		SortOrder:   req.SortOrder,
	}

	if err := h.svc.ExaminationType.Create(c.Request.Context(), examType); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toExamTypeResponse(examType))
}

// UpdateExaminationType godoc
func (h *Handler) UpdateExaminationType(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateExaminationTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	svcInput := service.UpdateExamTypeInput{
		Name:          req.Name,
		Price:         req.Price,
		IsActive:      req.IsActive,
		Description:   req.Description,
		ParentID:      req.ParentID,
		ClearParentID: req.ClearParentID,
		SortOrder:     req.SortOrder,
	}

	exType, err := h.svc.ExaminationType.Update(c.Request.Context(), clinicID, id, svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toExamTypeResponse(exType))
}

// ReorderExaminationTypes godoc
func (h *Handler) ReorderExaminationTypes(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderExaminationTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.ExaminationType.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "reordered"})
}

// DeleteExaminationType godoc
func (h *Handler) DeleteExaminationType(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.ExaminationType.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
