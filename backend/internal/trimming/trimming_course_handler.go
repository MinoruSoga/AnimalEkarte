package trimming

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// ListTrimmingCourses godoc
func (h *Handler) ListTrimmingCourses(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceMasterTrimming), "view") {
		return
	}
	courses, err := h.svc.TrimmingCourse.List(c.Request.Context(), clinicID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	responses := make([]trimmingCourseResponse, 0, len(courses))
	for i := range courses {
		responses = append(responses, h.signedTrimmingCourseResponse(c.Request.Context(), clinicID, &courses[i]))
	}
	c.JSON(http.StatusOK, responses)
}

// GetTrimmingCourse godoc
func (h *Handler) GetTrimmingCourse(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceMasterTrimming), "view") {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	course, err := h.svc.TrimmingCourse.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, h.signedTrimmingCourseResponse(c.Request.Context(), clinicID, course))
}

// CreateTrimmingCourse godoc
func (h *Handler) CreateTrimmingCourse(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	if !protectCourseImageBody(c, clinicID, h.courseImageDrafts, "image_object_key") {
		return
	}
	var req createTrimmingCourseRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	input := req.toServiceInput()
	if req.ImageObjectKey != "" {
		staffID, ok := httpapi.ExtractStaffID(c)
		if !ok {
			return
		}
		input.ActorStaffID = staffID
	}
	course, err := h.svc.TrimmingCourse.Create(c.Request.Context(), clinicID, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/trimming-courses/%d", course.ID))
	c.JSON(http.StatusCreated, h.signedTrimmingCourseResponse(c.Request.Context(), clinicID, course))
}

// UpdateTrimmingCourse godoc
func (h *Handler) UpdateTrimmingCourse(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if !protectCourseImageBody(c, clinicID, h.courseImageDrafts, "image_object_key") {
		return
	}
	var req updateTrimmingCourseRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	// 原本参照の同一TX監査を残すため authenticated staff actor が必須
	//（ExtractStaffID は失敗時に応答を書く）
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}

	input := req.toServiceInput()
	input.ActorStaffID = staffID
	course, err := h.svc.TrimmingCourse.Update(c.Request.Context(), clinicID, id, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, h.signedTrimmingCourseResponse(c.Request.Context(), clinicID, course))
}

// signedTrimmingCourseResponse adds a display-only signed URL for the stored
// image key. Signing failures degrade to no URL (the key stays in the
// response) so one image cannot hide the master list; the signed URL is never
// persisted.
func (h *Handler) signedTrimmingCourseResponse(ctx context.Context, clinicID uint64, course *model.TrimmingCourse) trimmingCourseResponse {
	resp := toTrimmingCourseResponse(course)
	if resp.ImageObjectKey == nil {
		return resp
	}
	signed, err := h.signStoredTrimmingImageURL(ctx, clinicID, *resp.ImageObjectKey)
	if err != nil {
		slog.WarnContext(ctx, "trimming course image signing failed",
			slog.String("error", err.Error()),
			slog.Uint64("clinic_id", clinicID))
		return resp
	}
	resp.ImageURL = &signed
	return resp
}

// DeleteTrimmingCourse godoc
func (h *Handler) DeleteTrimmingCourse(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.TrimmingCourse.Delete(c.Request.Context(), clinicID, id); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderTrimmingCourses godoc
func (h *Handler) ReorderTrimmingCourses(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req httpapi.ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if err := h.svc.TrimmingCourse.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
