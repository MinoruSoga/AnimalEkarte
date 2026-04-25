package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// GetLstepSettings godoc
// GET /api/v1/lstep-settings
func (h *Handler) GetLstepSettings(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	resp, err := h.svc.LstepSettings.GetSettings(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLstepSettingsResponse(resp))
}

// UpdateLstepSettings godoc
// PATCH /api/v1/lstep-settings
func (h *Handler) UpdateLstepSettings(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req updateLstepSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	resp, err := h.svc.LstepSettings.UpdateSettings(c.Request.Context(), clinicID, service.UpdateLstepSettingsInput{
		LstepAPIKey:            req.LstepAPIKey,
		LstepBaseURL:           req.LstepBaseURL,
		LineChannelAccessToken: req.LineChannelAccessToken,
		LineChannelSecret:      req.LineChannelSecret,
		LiffID:                 req.LiffID,
		LineAccountName:        req.LineAccountName,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLstepSettingsResponse(resp))
}

// DeleteLstepSettings godoc
// DELETE /api/v1/lstep-settings
func (h *Handler) DeleteLstepSettings(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	if err := h.svc.LstepSettings.DeleteSettings(c.Request.Context(), clinicID); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// TestLstepConnection godoc
// POST /api/v1/lstep-settings/test-connection
func (h *Handler) TestLstepConnection(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	result, err := h.svc.LstepSettings.TestConnection(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLstepConnectionTestResponse(result))
}

// RegisterLstepSettingsRoutes はLステップ設定のルートを登録する。
// hospital-settings リソースの権限で保護する（管理者専用）。
func (h *Handler) RegisterLstepSettingsRoutes(rg *gin.RouterGroup) {
	ls := rg.Group("/lstep-settings")
	ls.GET("", h.RequirePermission(string(model.ResourceHospitalSettings), "view"), h.GetLstepSettings)
	ls.PATCH("", h.RequirePermission(string(model.ResourceHospitalSettings), "edit"), h.UpdateLstepSettings)
	ls.DELETE("", h.RequirePermission(string(model.ResourceHospitalSettings), "delete"), h.DeleteLstepSettings)
	ls.POST("/test-connection", h.RequirePermission(string(model.ResourceHospitalSettings), "view"), h.TestLstepConnection)

	// ISSUE-003: FE が /clinics/:clinic_id/lstep-settings で呼ぶエイリアス
	alias := rg.Group("/clinics/:clinic_id/lstep-settings")
	alias.GET("", h.RequirePermission(string(model.ResourceHospitalSettings), "view"), h.GetLstepSettings)
	alias.PATCH("", h.RequirePermission(string(model.ResourceHospitalSettings), "edit"), h.UpdateLstepSettings)
	alias.DELETE("", h.RequirePermission(string(model.ResourceHospitalSettings), "delete"), h.DeleteLstepSettings)
	alias.POST("/test-connection", h.RequirePermission(string(model.ResourceHospitalSettings), "view"), h.TestLstepConnection)
}
