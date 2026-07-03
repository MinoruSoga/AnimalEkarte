package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// GetLineReservationSetting godoc
func (h *Handler) GetLineReservationSetting(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	setting, err := h.svc.LineReservationSetting.Get(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	if setting == nil {
		c.Status(http.StatusNoContent)
		return
	}
	c.JSON(http.StatusOK, toLineReservationSettingResponse(setting))
}

// SaveLineReservationSetting godoc
func (h *Handler) SaveLineReservationSetting(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req upsertLineReservationSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	setting, isNew, err := h.svc.LineReservationSetting.Save(c.Request.Context(), clinicID, req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	if isNew {
		c.Header("Location", fmt.Sprintf("/v1/clinics/%d/line-reservation-settings", clinicID))
		c.JSON(http.StatusCreated, toLineReservationSettingResponse(setting))
		return
	}
	c.JSON(http.StatusOK, toLineReservationSettingResponse(setting))
}
