// Package handler provides HTTP handler implementations for ExaminationType entity.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- ExaminationType ----

// GetExaminationType godoc
func (h *Handler) GetExaminationType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	et, err := h.svc.ExaminationType.GetByID(c.Request.Context(), id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toExamTypeResponse(et))
}

// ListExaminationTypes godoc
func (h *Handler) ListExaminationTypes(c *gin.Context) {
	exTypes, err := h.svc.ExaminationType.List(c.Request.Context())
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
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
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
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req updateExaminationTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	examType := &model.ExaminationType{
		ID:          id,
		Name:        req.Name,
		Price:       req.Price,
		Description: req.Description,
		SortOrder:   req.SortOrder,
	}
	if req.IsActive != nil {
		examType.IsActive = *req.IsActive
	}
	if req.ClearParentID {
		examType.ParentID = nil
	} else if req.ParentID != nil {
		examType.ParentID = req.ParentID
	}

	if err := h.svc.ExaminationType.Update(c.Request.Context(), examType); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toExamTypeResponse(examType))
}

// ReorderExaminationTypes godoc
func (h *Handler) ReorderExaminationTypes(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderExaminationTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
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
