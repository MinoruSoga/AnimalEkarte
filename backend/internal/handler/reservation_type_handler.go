// Package handler provides HTTP handler implementations for ReservationType entity.
package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- ReservationType ----

// ListReservationTypes godoc
func (h *Handler) ListReservationTypes(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	reservationTypes, err := h.svc.ReservationType.List(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(reservationTypes, toReservationTypeResponse))
}

// GetReservationType godoc
func (h *Handler) GetReservationType(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	st, err := h.svc.ReservationType.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationTypeResponse(st))
}

// CreateReservationType godoc
func (h *Handler) CreateReservationType(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req createReservationTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	st, err := h.svc.ReservationType.Create(c.Request.Context(), clinicID, &service.CreateReservationTypeInput{
		Name:                   req.Name,
		Color:                  req.Color,
		IsActive:               req.IsActive,
		Description:            req.Description,
		SortOrder:              req.SortOrder,
		Category:               req.Category,
		ReservationDisplayName: req.ReservationDisplayName,
		DurationMinutes:        req.DurationMinutes,
		ShortName:              req.ShortName,
		ShowShortName:          req.ShowShortName,
		ReservationVisible:     req.ReservationVisible,
		ReservationComment:     req.ReservationComment,
		ReservationImageURL:    req.ReservationImageURL,
		ReservationDayOption:   req.ReservationDayOption,
		IsInternal:             req.IsInternal,
		GroupID:                req.GroupID,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/reservation-types/%d", st.ID))
	c.JSON(http.StatusCreated, toReservationTypeResponse(st))
}

// UpdateReservationType godoc
func (h *Handler) UpdateReservationType(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateReservationTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	st, err := h.svc.ReservationType.Update(c.Request.Context(), clinicID, id, &service.UpdateReservationTypeInput{
		Name:                   req.Name,
		Color:                  req.Color,
		IsActive:               req.IsActive,
		Description:            req.Description,
		SortOrder:              req.SortOrder,
		Category:               req.Category,
		ReservationDisplayName: req.ReservationDisplayName,
		DurationMinutes:        req.DurationMinutes,
		ShortName:              req.ShortName,
		ShowShortName:          req.ShowShortName,
		ReservationVisible:     req.ReservationVisible,
		ReservationComment:     req.ReservationComment,
		ReservationImageURL:    req.ReservationImageURL,
		ReservationDayOption:   req.ReservationDayOption,
		IsInternal:             req.IsInternal,
		GroupID:                req.GroupID,
		ClearGroupID:           req.ClearGroupID,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationTypeResponse(st))
}

// DeleteReservationType godoc
func (h *Handler) DeleteReservationType(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.ReservationType.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListUnavailableTimes godoc
func (h *Handler) ListUnavailableTimes(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	items, err := h.svc.ReservationType.ListUnavailableTimes(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(items, toUnavailableTimeResponse))
}

// CreateUnavailableTime godoc
func (h *Handler) CreateUnavailableTime(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req createUnavailableTimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	input := service.CreateUnavailableTimeInput{
		UnavailableType: req.UnavailableType,
		DayOfWeek:       req.DayOfWeek,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
	}
	if req.SpecificDate != nil {
		t, err := time.Parse("2006-01-02", *req.SpecificDate)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("specific_date must be YYYY-MM-DD"))
			return
		}
		input.SpecificDate = &t
	}
	result, err := h.svc.ReservationType.CreateUnavailableTime(c.Request.Context(), clinicID, id, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	resp := toUnavailableTimeResponse(result)
	c.Header("Location", fmt.Sprintf("/v1/masters/reservation-types/%d/unavailable-times/%d", id, result.ID))
	c.JSON(http.StatusCreated, resp)
}

// DeleteUnavailableTime godoc
func (h *Handler) DeleteUnavailableTime(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	reservationTypeID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	unavailableTimeID, ok := parseIDParam(c, "unavailable_time_id")
	if !ok {
		return
	}
	if err := h.svc.ReservationType.DeleteUnavailableTime(c.Request.Context(), clinicID, reservationTypeID, unavailableTimeID); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListOccupations godoc
func (h *Handler) ListReservationTypeOccupations(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	items, err := h.svc.ReservationType.ListOccupations(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(items, toReservationTypeOccupationResponse))
}

// LinkOccupation godoc
func (h *Handler) LinkReservationTypeOccupation(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req linkOccupationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	result, err := h.svc.ReservationType.LinkOccupation(c.Request.Context(), clinicID, id, req.OccupationID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/reservation-types/%d/occupations/%d", id, result.ID))
	c.JSON(http.StatusCreated, toReservationTypeOccupationResponse(result))
}

// UnlinkOccupation godoc
func (h *Handler) UnlinkReservationTypeOccupation(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	occupationID, ok := parseIDParam(c, "occupation_id")
	if !ok {
		return
	}
	if err := h.svc.ReservationType.UnlinkOccupation(c.Request.Context(), clinicID, id, occupationID); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderReservationTypes godoc
func (h *Handler) ReorderReservationTypes(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.ReservationType.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
