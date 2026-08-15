package clinic

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// GetClosingSettings godoc
// GET /v1/closing-settings
func (h *Handler) GetClosingSettings(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	resp, err := h.closingSettingsSvc.Get(c.Request.Context(), clinicID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, ToClosingSettingsFullResponse(resp.Settings, resp.SpecialPeriods))
}

// UpdateClosingSettings godoc
// PATCH /v1/closing-settings
func (h *Handler) UpdateClosingSettings(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	actorID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	var req UpdateClinicSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	result, err := h.closingSettingsSvc.UpdateStandard(c.Request.Context(), clinicID, actorID, req.ToServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, ToClinicSettingsResponse(result))
}

// ListSpecialPeriods godoc
// GET /v1/closing-settings/special-periods
func (h *Handler) ListSpecialPeriods(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	periods, err := h.closingSettingsSvc.ListSpecialPeriods(c.Request.Context(), clinicID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(periods, ToClosingSpecialPeriodResponse))
}

// CreateSpecialPeriod godoc
// POST /v1/closing-settings/special-periods
func (h *Handler) CreateSpecialPeriod(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req CreateSpecialPeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	period, err := h.closingSettingsSvc.CreateSpecialPeriod(c.Request.Context(), clinicID, req.ToServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/closing-settings/special-periods/%d", period.ID))
	c.JSON(http.StatusCreated, ToClosingSpecialPeriodResponse(period))
}

// UpdateSpecialPeriod godoc
// PATCH /v1/closing-settings/special-periods/:id
func (h *Handler) UpdateSpecialPeriod(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req UpdateSpecialPeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	period, err := h.closingSettingsSvc.UpdateSpecialPeriod(c.Request.Context(), clinicID, id, req.ToServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, ToClosingSpecialPeriodResponse(period))
}

// DeleteSpecialPeriod godoc
// DELETE /v1/closing-settings/special-periods/:id
func (h *Handler) DeleteSpecialPeriod(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.closingSettingsSvc.DeleteSpecialPeriod(c.Request.Context(), clinicID, id); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterClosingSettingsRoutes は締め設定関連のルートを登録する
func (h *Handler) RegisterClosingSettingsRoutes(rg *gin.RouterGroup) {
	cs := rg.Group("/closing-settings")
	cs.GET("", h.requirePermission(string(model.ResourceClosingSettings), "view"), h.GetClosingSettings)
	cs.PATCH("", h.requirePermission(string(model.ResourceClosingSettings), "edit"), h.UpdateClosingSettings)

	sp := cs.Group("/special-periods")
	sp.GET("", h.requirePermission(string(model.ResourceClosingSettings), "view"), h.ListSpecialPeriods)
	sp.POST("", h.requirePermission(string(model.ResourceClosingSettings), "create"), h.CreateSpecialPeriod)
	sp.PATCH("/:id", h.requirePermission(string(model.ResourceClosingSettings), "edit"), h.UpdateSpecialPeriod)
	sp.DELETE("/:id", h.requirePermission(string(model.ResourceClosingSettings), "delete"), h.DeleteSpecialPeriod)

	// 休診日は既存 ClinicHoliday ハンドラに委譲
	holidays := cs.Group("/holidays")
	holidays.GET("", h.requirePermission(string(model.ResourceClosingSettings), "view"), h.ListClinicHolidays)
	holidays.POST("", h.requirePermission(string(model.ResourceClosingSettings), "create"), h.SetClinicHoliday)
	holidays.DELETE("/:date", h.requirePermission(string(model.ResourceClosingSettings), "delete"), h.DeleteClinicHoliday)
}
