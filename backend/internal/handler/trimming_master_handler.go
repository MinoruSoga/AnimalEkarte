// Package handler provides HTTP handler implementations for TrimmingCourse and TrimmingOption entities.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- TrimmingCourse ----

// GetTrimmingCourse godoc
func (h *Handler) GetTrimmingCourse(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	course, err := h.svc.TrimmingCourse.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTrimmingCourseResponse(course))
}

// ListTrimmingCourses godoc
func (h *Handler) ListTrimmingCourses(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	courses, err := h.svc.TrimmingCourse.List(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTrimmingCourseResponseList(courses))
}

// CreateTrimmingCourse godoc
func (h *Handler) CreateTrimmingCourse(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req createTrimmingCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	course := &model.TrimmingCourse{
		ClinicID:    clinicID,
		Name:        req.Name,
		Price:       req.Price,
		IsActive:    req.IsActive,
		Description: req.Description,
		Duration:    req.Duration,
		SortOrder:   req.SortOrder,
	}
	if req.TargetSize != "" {
		ts := model.TargetSize(req.TargetSize)
		course.TargetSize = &ts
	}

	if err := h.svc.TrimmingCourse.Create(c.Request.Context(), course); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toTrimmingCourseResponse(course))
}

// UpdateTrimmingCourse godoc
func (h *Handler) UpdateTrimmingCourse(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req updateTrimmingCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	svcInput := service.UpdateTrimmingCourseInput{
		Name:        req.Name,
		Price:       req.Price,
		IsActive:    req.IsActive,
		Description: req.Description,
		Duration:    req.Duration,
		SortOrder:   req.SortOrder,
	}
	if req.TargetSize != nil {
		ts := model.TargetSize(*req.TargetSize)
		svcInput.TargetSize = &ts
	}

	course, err := h.svc.TrimmingCourse.Update(c.Request.Context(), clinicID, id, svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTrimmingCourseResponse(course))
}

// DeleteTrimmingCourse godoc
func (h *Handler) DeleteTrimmingCourse(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if err := h.svc.TrimmingCourse.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderTrimmingCourses godoc
func (h *Handler) ReorderTrimmingCourses(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderTrimmingCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.TrimmingCourse.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "reordered"})
}

// ---- TrimmingOption ----

// GetTrimmingOption godoc
func (h *Handler) GetTrimmingOption(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	option, err := h.svc.TrimmingOption.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTrimmingOptionResponse(option))
}

// ListTrimmingOptions godoc
func (h *Handler) ListTrimmingOptions(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	options, err := h.svc.TrimmingOption.List(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTrimmingOptionResponseList(options))
}

// CreateTrimmingOption godoc
func (h *Handler) CreateTrimmingOption(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req createTrimmingOptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	option := &model.TrimmingOption{
		ClinicID:    clinicID,
		Name:        req.Name,
		Price:       req.Price,
		IsActive:    req.IsActive,
		Description: req.Description,
		Duration:    req.Duration,
		Combinable:  req.Combinable,
		SortOrder:   req.SortOrder,
	}

	if err := h.svc.TrimmingOption.Create(c.Request.Context(), option); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toTrimmingOptionResponse(option))
}

// UpdateTrimmingOption godoc
func (h *Handler) UpdateTrimmingOption(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req updateTrimmingOptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	svcInput := service.UpdateTrimmingOptionInput{
		Name:        req.Name,
		Price:       req.Price,
		IsActive:    req.IsActive,
		Description: req.Description,
		Duration:    req.Duration,
		Combinable:  req.Combinable,
		SortOrder:   req.SortOrder,
	}

	option, err := h.svc.TrimmingOption.Update(c.Request.Context(), clinicID, id, svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTrimmingOptionResponse(option))
}

// DeleteTrimmingOption godoc
func (h *Handler) DeleteTrimmingOption(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if err := h.svc.TrimmingOption.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderTrimmingOptions godoc
func (h *Handler) ReorderTrimmingOptions(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderTrimmingOptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.TrimmingOption.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "reordered"})
}
