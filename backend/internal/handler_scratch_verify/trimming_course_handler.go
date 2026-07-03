package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

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

	course, err := h.svc.TrimmingCourse.Create(c.Request.Context(), clinicID, req.toServiceInput())
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

	course, err := h.svc.TrimmingCourse.Update(c.Request.Context(), clinicID, id, req.toServiceInput())
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
