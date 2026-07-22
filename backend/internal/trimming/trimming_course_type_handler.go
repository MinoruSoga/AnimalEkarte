package trimming

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

// ListTrimmingCourseTypes godoc
// GET /v1/masters/trimming-course-types
func (h *Handler) ListTrimmingCourseTypes(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	ms, err := h.svc.TrimmingCourseType.List(c.Request.Context(), clinicID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(ms, toTrimmingCourseTypeResponse))
}

// GetTrimmingCourseType godoc
// GET /v1/masters/trimming-course-types/:id
func (h *Handler) GetTrimmingCourseType(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	m, err := h.svc.TrimmingCourseType.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTrimmingCourseTypeResponse(m))
}

// CreateTrimmingCourseType godoc
// POST /v1/masters/trimming-course-types
func (h *Handler) CreateTrimmingCourseType(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req createTrimmingCourseTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	m, err := h.svc.TrimmingCourseType.Create(c.Request.Context(), clinicID, req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/trimming-course-types/%d", m.ID))
	c.JSON(http.StatusCreated, toTrimmingCourseTypeResponse(m))
}

// UpdateTrimmingCourseType godoc
// PATCH /v1/masters/trimming-course-types/:id
func (h *Handler) UpdateTrimmingCourseType(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateTrimmingCourseTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	m, err := h.svc.TrimmingCourseType.Update(c.Request.Context(), clinicID, id, req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTrimmingCourseTypeResponse(m))
}

// ReorderTrimmingCourseTypes godoc
// PATCH /v1/masters/trimming-course-types/reorder
func (h *Handler) ReorderTrimmingCourseTypes(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req httpapi.ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if err := h.svc.TrimmingCourseType.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteTrimmingCourseType godoc
// DELETE /v1/masters/trimming-course-types/:id
func (h *Handler) DeleteTrimmingCourseType(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.TrimmingCourseType.Delete(c.Request.Context(), clinicID, id); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
