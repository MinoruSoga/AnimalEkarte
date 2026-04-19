// Package handler provides HTTP handler implementations for TrimmingCourse and TrimmingOption entities.
package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- TrimmingCourse ----

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
	c.JSON(http.StatusOK, mapSlice(courses, toTrimmingCourseResponse))
}

// GetTrimmingCourse godoc
func (h *Handler) GetTrimmingCourse(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	course, err := h.svc.TrimmingCourse.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTrimmingCourseResponse(course))
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

	svcInput := &service.CreateTrimmingCourseInput{
		Name:        req.Name,
		TargetSize:  req.TargetSize,
		Price:       req.Price,
		IsActive:    req.IsActive,
		Description: req.Description,
		Duration:    req.Duration,
		SortOrder:   req.SortOrder,
	}

	course, err := h.svc.TrimmingCourse.Create(c.Request.Context(), clinicID, svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/trimming-courses/%d", course.ID))
	c.JSON(http.StatusCreated, toTrimmingCourseResponse(course))
}

// UpdateTrimmingCourse godoc
func (h *Handler) UpdateTrimmingCourse(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
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
		TargetSize:  req.TargetSize,
		Duration:    req.Duration,
		SortOrder:   req.SortOrder,
	}

	course, err := h.svc.TrimmingCourse.Update(c.Request.Context(), clinicID, id, &svcInput)
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
	id, ok := parseIDParam(c, "id")
	if !ok {
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
	var req reorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.TrimmingCourse.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ---- TrimmingOption ----

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
	c.JSON(http.StatusOK, mapSlice(options, toTrimmingOptionResponse))
}

// GetTrimmingOption godoc
func (h *Handler) GetTrimmingOption(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	option, err := h.svc.TrimmingOption.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTrimmingOptionResponse(option))
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

	svcInput := &service.CreateTrimmingOptionInput{
		Name:         req.Name,
		Price:        req.Price,
		IsActive:     req.IsActive,
		Description:  req.Description,
		Duration:     req.Duration,
		IsCombinable: req.IsCombinable,
		SortOrder:    req.SortOrder,
	}

	option, err := h.svc.TrimmingOption.Create(c.Request.Context(), clinicID, svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/trimming-options/%d", option.ID))
	c.JSON(http.StatusCreated, toTrimmingOptionResponse(option))
}

// UpdateTrimmingOption godoc
func (h *Handler) UpdateTrimmingOption(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateTrimmingOptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	svcInput := service.UpdateTrimmingOptionInput{
		Name:         req.Name,
		Price:        req.Price,
		IsActive:     req.IsActive,
		Description:  req.Description,
		Duration:     req.Duration,
		IsCombinable: req.IsCombinable,
		SortOrder:    req.SortOrder,
	}

	option, err := h.svc.TrimmingOption.Update(c.Request.Context(), clinicID, id, &svcInput)
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
	id, ok := parseIDParam(c, "id")
	if !ok {
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
	var req reorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.TrimmingOption.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
