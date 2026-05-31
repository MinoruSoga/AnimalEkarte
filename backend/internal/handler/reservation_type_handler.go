// Package handler provides HTTP handler implementations for ReservationType entity.
package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
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
	st, err := h.svc.ReservationType.Create(c.Request.Context(), clinicID, req.toServiceInput())
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
	st, err := h.svc.ReservationType.Update(c.Request.Context(), clinicID, id, req.toServiceInput())
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
	items, err := h.svc.ReservationTypeUnavailableTime.ListUnavailableTimes(c.Request.Context(), clinicID, id)
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

	// 相互依存バリデーション
	switch req.UnavailableType {
	case "weekly":
		if req.DayOfWeek == nil {
			RespondError(c, apperrors.WrapInvalidInput("weekly タイプでは day_of_week が必要です"))
			return
		}
	case "specific":
		if req.SpecificDate == nil {
			RespondError(c, apperrors.WrapInvalidInput("specific タイプでは specific_date が必要です"))
			return
		}
	}

	input, err := req.toServiceInput()
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("specific_date は YYYY-MM-DD 形式で入力してください"))
		return
	}
	result, err := h.svc.ReservationTypeUnavailableTime.CreateUnavailableTime(c.Request.Context(), clinicID, id, input)
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
	if err := h.svc.ReservationTypeUnavailableTime.DeleteUnavailableTime(c.Request.Context(), clinicID, reservationTypeID, unavailableTimeID); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListAvailableSlots godoc
func (h *Handler) ListAvailableSlots(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	items, err := h.svc.ReservationTypeAvailableSlot.ListAvailableSlots(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(items, toAvailableSlotResponse))
}

// CreateAvailableSlot godoc
func (h *Handler) CreateAvailableSlot(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req createAvailableSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	switch req.AvailableType {
	case "weekly":
		if req.DayOfWeek == nil {
			RespondError(c, apperrors.WrapInvalidInput("weekly タイプでは day_of_week が必要です"))
			return
		}
	case "specific":
		if req.SpecificDate == nil {
			RespondError(c, apperrors.WrapInvalidInput("specific タイプでは specific_date が必要です"))
			return
		}
	}
	input, err := req.toServiceInput()
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("specific_date は YYYY-MM-DD 形式で入力してください"))
		return
	}
	result, err := h.svc.ReservationTypeAvailableSlot.CreateAvailableSlot(c.Request.Context(), clinicID, id, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	resp := toAvailableSlotResponse(result)
	c.Header("Location", fmt.Sprintf("/v1/masters/reservation-types/%d/available-slots/%d", id, result.ID))
	c.JSON(http.StatusCreated, resp)
}

// DeleteAvailableSlot godoc
func (h *Handler) DeleteAvailableSlot(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	reservationTypeID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	availableSlotID, ok := parseIDParam(c, "available_slot_id")
	if !ok {
		return
	}
	if err := h.svc.ReservationTypeAvailableSlot.DeleteAvailableSlot(c.Request.Context(), clinicID, reservationTypeID, availableSlotID); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListReservationTypeOccupations godoc
func (h *Handler) ListReservationTypeOccupations(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	items, err := h.svc.ReservationTypeOccupation.ListOccupations(c.Request.Context(), clinicID, id)
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
	result, err := h.svc.ReservationTypeOccupation.LinkOccupation(c.Request.Context(), clinicID, id, req.OccupationID)
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
	if err := h.svc.ReservationTypeOccupation.UnlinkOccupation(c.Request.Context(), clinicID, id, occupationID); err != nil {
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
