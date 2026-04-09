package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
)

// GetReservationSetting godoc
func (h *Handler) GetReservationSetting(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	setting, err := h.svc.ReservationSetting.Get(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	if setting == nil {
		c.Status(http.StatusNoContent)
		return
	}
	c.JSON(http.StatusOK, toReservationSettingResponse(setting))
}

// UpsertReservationSetting godoc
func (h *Handler) UpsertReservationSetting(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req upsertReservationSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	setting, err := h.svc.ReservationSetting.Upsert(c.Request.Context(), clinicID, &service.UpsertReservationSettingInput{
		Status:                  req.Status,
		HeaderText:              req.HeaderText,
		ReservationNotice:       req.ReservationNotice,
		CancelNotice:            req.CancelNotice,
		PrivacyPolicy:           req.PrivacyPolicy,
		ClosedWeekdays:          req.ClosedWeekdays,
		ClosedDates:             req.ClosedDates,
		NationalHolidayClosed:   req.NationalHolidayClosed,
		BusinessHours:           req.BusinessHours,
		BusinessHoursByWeekday:  req.BusinessHoursByWeekday,
		BreakHours:              req.BreakHours,
		DailyLimit:              req.DailyLimit,
		MonthlyLimit:            req.MonthlyLimit,
		BookingWindowMaxDays:    req.BookingWindowMaxDays,
		BookingWindowMinDays:    req.BookingWindowMinDays,
		CalendarMonths:          req.CalendarMonths,
		PhoneNumber:             req.PhoneNumber,
		NotificationEmail:       req.NotificationEmail,
		RequestExample:          req.RequestExample,
		TimeSlotMode:            req.TimeSlotMode,
		TimeSlotIntervalMinutes: req.TimeSlotIntervalMinutes,
		NoStaffMode:             req.NoStaffMode,
		ShowNoStaffOption:       req.ShowNoStaffOption,
		AdditionalFields:        req.AdditionalFields,
		LineChannelID:           req.LineChannelID,
		LineChannelSecret:       req.LineChannelSecret,
		LiffID:                  req.LiffID,
		LineAccessToken:         req.LineAccessToken,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationSettingResponse(setting))
}
