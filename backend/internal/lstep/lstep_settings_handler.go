package lstep

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// SettingsHandler は LstepSettingsService の HTTP handler。
type SettingsHandler struct {
	svc               LstepSettingsService
	requirePermission PermissionMiddleware
}

// NewSettingsHandler は SettingsHandler を構築する。
func NewSettingsHandler(svc LstepSettingsService, requirePermission PermissionMiddleware) *SettingsHandler {
	return &SettingsHandler{svc: svc, requirePermission: requirePermission}
}

// GetLstepSettings godoc
// GET /api/v1/lstep-settings
func (h *SettingsHandler) GetLstepSettings(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	resp, err := h.svc.GetSettings(c.Request.Context(), clinicID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLstepSettingsResponse(resp))
}

// UpdateLstepSettings godoc
// PATCH /api/v1/lstep-settings
func (h *SettingsHandler) UpdateLstepSettings(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req updateLstepSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	var actorID *uint64
	if staffID, ok := httpapi.ExtractStaffID(c); ok {
		actorID = &staffID
	}
	resp, err := h.svc.UpdateSettings(c.Request.Context(), clinicID, req.toServiceInput(), actorID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLstepSettingsResponse(resp))
}

// DeleteLstepSettings godoc
// DELETE /api/v1/lstep-settings
func (h *SettingsHandler) DeleteLstepSettings(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var actorID *uint64
	if staffID, ok := httpapi.ExtractStaffID(c); ok {
		actorID = &staffID
	}
	if err := h.svc.DeleteSettings(c.Request.Context(), clinicID, actorID); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// TestLstepConnection godoc
// POST /api/v1/lstep-settings/test-connection
func (h *SettingsHandler) TestLstepConnection(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	result, err := h.svc.TestConnection(c.Request.Context(), clinicID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLstepConnectionTestResponse(result))
}
