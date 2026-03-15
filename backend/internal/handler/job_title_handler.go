// Package handler provides HTTP handler implementations for JobTitle entity.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- JobTitle ----

// ListJobTitles godoc
func (h *Handler) ListJobTitles(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	jobTitles, err := h.svc.JobTitle.List(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toJobTitleResponseList(jobTitles))
}

// CreateJobTitle godoc
func (h *Handler) CreateJobTitle(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req createJobTitleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	jt := &model.JobTitle{
		ClinicID:  clinicID,
		Name:      req.Name,
		IsActive:  req.IsActive,
		SortOrder: req.SortOrder,
	}
	if err := h.svc.JobTitle.Create(c.Request.Context(), jt); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toJobTitleResponse(jt))
}

// UpdateJobTitle godoc
func (h *Handler) UpdateJobTitle(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}

	var req updateJobTitleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	input := &service.UpdateJobTitleInput{
		Name:        req.Name,
		Description: req.Description,
		SortOrder:   req.SortOrder,
		IsActive:    req.IsActive,
	}

	updated, err := h.svc.JobTitle.Update(c.Request.Context(), clinicID, id, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toJobTitleResponse(updated))
}

// DeleteJobTitle godoc
func (h *Handler) DeleteJobTitle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if err := h.svc.JobTitle.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderJobTitles godoc
func (h *Handler) ReorderJobTitles(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderJobTitleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}
	if err := h.svc.JobTitle.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "reordered"})
}
