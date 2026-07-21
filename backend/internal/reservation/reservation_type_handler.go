// Package handler provides HTTP handler implementations for ReservationType entity.
package reservation

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// ReservationTypeHandler は予約区分マスタ（本体+不可時間+可能枠+職種リンク）の HTTP handler。
type ReservationTypeHandler struct {
	core            ReservationTypeCoreService
	unavailableTime ReservationTypeUnavailableTimeService
	availableSlot   ReservationTypeAvailableSlotService
	occupation      ReservationTypeOccupationService
}

// NewReservationTypeHandler は ReservationTypeHandler を構築する。
func NewReservationTypeHandler(core ReservationTypeCoreService, unavailableTime ReservationTypeUnavailableTimeService, availableSlot ReservationTypeAvailableSlotService, occupation ReservationTypeOccupationService) *ReservationTypeHandler {
	return &ReservationTypeHandler{core: core, unavailableTime: unavailableTime, availableSlot: availableSlot, occupation: occupation}
}

// ---- ReservationType ----

// ListReservationTypes godoc
func (h *ReservationTypeHandler) ListReservationTypes(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	reservationTypes, err := h.core.List(c.Request.Context(), clinicID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(reservationTypes, ToReservationTypeResponse))
}

// GetReservationType godoc
func (h *ReservationTypeHandler) GetReservationType(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	st, err := h.core.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, ToReservationTypeResponse(st))
}

// CreateReservationType godoc
func (h *ReservationTypeHandler) CreateReservationType(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req createReservationTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	st, err := h.core.Create(c.Request.Context(), clinicID, req.toServiceInput())
	if err != nil {
		respondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/reservation-types/%d", st.ID))
	c.JSON(http.StatusCreated, ToReservationTypeResponse(st))
}

// UpdateReservationType godoc
func (h *ReservationTypeHandler) UpdateReservationType(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateReservationTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	st, err := h.core.Update(c.Request.Context(), clinicID, id, req.toServiceInput())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, ToReservationTypeResponse(st))
}

// DeleteReservationType godoc
func (h *ReservationTypeHandler) DeleteReservationType(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.core.Delete(c.Request.Context(), clinicID, id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListUnavailableTimes godoc
func (h *ReservationTypeHandler) ListUnavailableTimes(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	items, err := h.unavailableTime.ListUnavailableTimes(c.Request.Context(), clinicID, id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(items, toUnavailableTimeResponse))
}

// CreateUnavailableTime godoc
func (h *ReservationTypeHandler) CreateUnavailableTime(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req createUnavailableTimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	// 相互依存バリデーション
	switch req.UnavailableType {
	case "weekly":
		if req.DayOfWeek == nil {
			respondError(c, apperrors.WrapInvalidInput("weekly タイプでは day_of_week が必要です"))
			return
		}
	case "specific":
		if req.SpecificDate == nil {
			respondError(c, apperrors.WrapInvalidInput("specific タイプでは specific_date が必要です"))
			return
		}
	}

	input, err := req.toServiceInput()
	if err != nil {
		respondError(c, apperrors.WrapInvalidInput("specific_date は YYYY-MM-DD 形式で入力してください"))
		return
	}
	result, err := h.unavailableTime.CreateUnavailableTime(c.Request.Context(), clinicID, id, input)
	if err != nil {
		respondError(c, err)
		return
	}
	resp := toUnavailableTimeResponse(result)
	c.Header("Location", fmt.Sprintf("/v1/masters/reservation-types/%d/unavailable-times/%d", id, result.ID))
	c.JSON(http.StatusCreated, resp)
}

// DeleteUnavailableTime godoc
func (h *ReservationTypeHandler) DeleteUnavailableTime(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	reservationTypeID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	unavailableTimeID, ok := httpapi.ParseIDParam(c, "unavailable_time_id")
	if !ok {
		return
	}
	if err := h.unavailableTime.DeleteUnavailableTime(c.Request.Context(), clinicID, reservationTypeID, unavailableTimeID); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListAvailableSlots godoc
func (h *ReservationTypeHandler) ListAvailableSlots(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	items, err := h.availableSlot.ListAvailableSlots(c.Request.Context(), clinicID, id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(items, toAvailableSlotResponse))
}

// CreateAvailableSlot godoc
func (h *ReservationTypeHandler) CreateAvailableSlot(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req createAvailableSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	switch req.AvailableType {
	case "weekly":
		if req.DayOfWeek == nil {
			respondError(c, apperrors.WrapInvalidInput("weekly タイプでは day_of_week が必要です"))
			return
		}
	case "specific":
		if req.SpecificDate == nil {
			respondError(c, apperrors.WrapInvalidInput("specific タイプでは specific_date が必要です"))
			return
		}
	}
	input, err := req.toServiceInput()
	if err != nil {
		respondError(c, apperrors.WrapInvalidInput("specific_date は YYYY-MM-DD 形式で入力してください"))
		return
	}
	result, err := h.availableSlot.CreateAvailableSlot(c.Request.Context(), clinicID, id, input)
	if err != nil {
		respondError(c, err)
		return
	}
	resp := toAvailableSlotResponse(result)
	c.Header("Location", fmt.Sprintf("/v1/masters/reservation-types/%d/available-slots/%d", id, result.ID))
	c.JSON(http.StatusCreated, resp)
}

// DeleteAvailableSlot godoc
func (h *ReservationTypeHandler) DeleteAvailableSlot(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	reservationTypeID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	availableSlotID, ok := httpapi.ParseIDParam(c, "available_slot_id")
	if !ok {
		return
	}
	if err := h.availableSlot.DeleteAvailableSlot(c.Request.Context(), clinicID, reservationTypeID, availableSlotID); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListReservationTypeOccupations godoc
func (h *ReservationTypeHandler) ListReservationTypeOccupations(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	items, err := h.occupation.ListOccupations(c.Request.Context(), clinicID, id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(items, toReservationTypeOccupationResponse))
}

// LinkOccupation godoc
func (h *ReservationTypeHandler) LinkReservationTypeOccupation(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req linkOccupationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	result, err := h.occupation.LinkOccupation(c.Request.Context(), clinicID, id, req.OccupationID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/reservation-types/%d/occupations/%d", id, result.ID))
	c.JSON(http.StatusCreated, toReservationTypeOccupationResponse(result))
}

// UnlinkOccupation godoc
func (h *ReservationTypeHandler) UnlinkReservationTypeOccupation(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	occupationID, ok := httpapi.ParseIDParam(c, "occupation_id")
	if !ok {
		return
	}
	if err := h.occupation.UnlinkOccupation(c.Request.Context(), clinicID, id, occupationID); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderReservationTypes godoc
func (h *ReservationTypeHandler) ReorderReservationTypes(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req httpapi.ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if err := h.core.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
