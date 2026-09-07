package reservation

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// LineReservationSettingHandler は LINE 予約基本設定の HTTP handler。
type LineReservationSettingHandler struct {
	svc LineReservationSettingService
}

// NewLineReservationSettingHandler は LineReservationSettingHandler を構築する。
func NewLineReservationSettingHandler(svc LineReservationSettingService) *LineReservationSettingHandler {
	return &LineReservationSettingHandler{svc: svc}
}

func pathAuthorizedClinicID(c *gin.Context) (uint64, bool) {
	clinicID, ok := httpapi.ParseIDParam(c, "clinic_id")
	if !ok {
		return 0, false
	}
	if !httpapi.AuthorizeClinicIDs(c, []uint64{clinicID}) {
		return 0, false
	}
	return clinicID, true
}

// GetLineReservationSetting godoc
func (h *LineReservationSettingHandler) GetLineReservationSetting(c *gin.Context) {
	clinicID, ok := pathAuthorizedClinicID(c)
	if !ok {
		return
	}
	setting, err := h.svc.Get(c.Request.Context(), clinicID)
	if err != nil {
		respondError(c, err)
		return
	}
	if setting == nil {
		c.Status(http.StatusNoContent)
		return
	}
	c.JSON(http.StatusOK, toLineReservationSettingResponse(setting))
}

// SaveLineReservationSetting godoc
func (h *LineReservationSettingHandler) SaveLineReservationSetting(c *gin.Context) {
	clinicID, ok := pathAuthorizedClinicID(c)
	if !ok {
		return
	}
	var req upsertLineReservationSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	setting, isNew, err := h.svc.Save(c.Request.Context(), clinicID, req.toServiceInput())
	if err != nil {
		respondError(c, err)
		return
	}
	if isNew {
		c.Header("Location", fmt.Sprintf("/v1/clinics/%d/line-reservation-settings", clinicID))
		c.JSON(http.StatusCreated, toLineReservationSettingResponse(setting))
		return
	}
	c.JSON(http.StatusOK, toLineReservationSettingResponse(setting))
}
