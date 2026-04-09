package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
)

// ListReservationCourses godoc
func (h *Handler) ListReservationCourses(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	items, err := h.svc.ReservationCourse.List(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationCourseResponseList(items))
}

// CreateReservationCourse godoc
func (h *Handler) CreateReservationCourse(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req createReservationCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	st, err := h.svc.ReservationCourse.Create(c.Request.Context(), clinicID, &service.CreateReservationCourseInput{
		Name:                 req.Name,
		Color:                req.Color,
		Description:          req.Description,
		SortOrder:            req.SortOrder,
		DurationMinutes:      req.DurationMinutes,
		ShortName:            req.ShortName,
		ShowShortName:        req.ShowShortName,
		ReservationVisible:   req.ReservationVisible,
		ReservationComment:   req.ReservationComment,
		ReservationDayOption: req.ReservationDayOption,
		IsInternal:           req.IsInternal,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toReservationCourseResponse(st))
}

// UpdateReservationCourse godoc
func (h *Handler) UpdateReservationCourse(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	var req updateReservationCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	st, err := h.svc.ReservationCourse.Update(c.Request.Context(), clinicID, id, &service.UpdateReservationCourseInput{
		Name:                 req.Name,
		Color:                req.Color,
		Description:          req.Description,
		SortOrder:            req.SortOrder,
		DurationMinutes:      req.DurationMinutes,
		ShortName:            req.ShortName,
		ShowShortName:        req.ShowShortName,
		ReservationVisible:   req.ReservationVisible,
		ReservationComment:   req.ReservationComment,
		ReservationDayOption: req.ReservationDayOption,
		IsInternal:           req.IsInternal,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationCourseResponse(st))
}

// DeleteReservationCourse godoc
func (h *Handler) DeleteReservationCourse(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if err := h.svc.ReservationCourse.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// PatchReservationCourseStatus godoc
func (h *Handler) PatchReservationCourseStatus(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	var req patchReservationCourseStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	st, err := h.svc.ReservationCourse.PatchStatus(c.Request.Context(), clinicID, id, req.IsActive)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationCourseResponse(st))
}

// PatchReservationCourseSortOrder godoc
func (h *Handler) PatchReservationCourseSortOrder(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	var req patchReservationCourseSortOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.ReservationCourse.PatchSortOrder(c.Request.Context(), clinicID, id, req.Direction); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// UploadReservationCourseImage godoc — v2 スコープ：未実装
func (h *Handler) UploadReservationCourseImage(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}
