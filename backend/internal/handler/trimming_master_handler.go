// Package handler provides HTTP handler implementations for TrimmingCourse and TrimmingOption entities.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- TrimmingCourse ----

// ListTrimmingCourses godoc
func (h *Handler) ListTrimmingCourses(c *gin.Context) {
	courses, err := h.svc.TrimmingCourse.List(c.Request.Context())
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
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	course := &model.TrimmingCourse{
		ID:          id,
		ClinicID:    clinicID,
		Name:        req.Name,
		Price:       req.Price,
		Description: req.Description,
		Duration:    req.Duration,
		SortOrder:   req.SortOrder,
	}
	if req.IsActive != nil {
		course.IsActive = *req.IsActive
	}
	if req.TargetSize != "" {
		ts := model.TargetSize(req.TargetSize)
		course.TargetSize = &ts
	}

	if err := h.svc.TrimmingCourse.Update(c.Request.Context(), course); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTrimmingCourseResponse(course))
}

// DeleteTrimmingCourse godoc
func (h *Handler) DeleteTrimmingCourse(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if err := h.svc.TrimmingCourse.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ---- TrimmingOption ----

// ListTrimmingOptions godoc
func (h *Handler) ListTrimmingOptions(c *gin.Context) {
	options, err := h.svc.TrimmingOption.List(c.Request.Context())
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
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	option := &model.TrimmingOption{
		ID:          id,
		ClinicID:    clinicID,
		Name:        req.Name,
		Price:       req.Price,
		Description: req.Description,
		Duration:    req.Duration,
		SortOrder:   req.SortOrder,
	}
	if req.IsActive != nil {
		option.IsActive = *req.IsActive
	}
	if req.Combinable != nil {
		option.Combinable = *req.Combinable
	}

	if err := h.svc.TrimmingOption.Update(c.Request.Context(), option); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTrimmingOptionResponse(option))
}

// DeleteTrimmingOption godoc
func (h *Handler) DeleteTrimmingOption(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if err := h.svc.TrimmingOption.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
