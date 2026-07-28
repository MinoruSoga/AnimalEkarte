package lstep

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// LstepSettingsHandler は LstepSettingsService の HTTP handler。
type LstepSettingsHandler struct {
	svc               LstepSettingsService
	requirePermission PermissionMiddleware
}

// NewLstepSettingsHandler は LstepSettingsHandler を構築する。
func NewLstepSettingsHandler(svc LstepSettingsService, requirePermission PermissionMiddleware) *LstepSettingsHandler {
	return &LstepSettingsHandler{svc: svc, requirePermission: requirePermission}
}

// GetLstepSettings godoc
// GET /api/v1/lstep-settings
func (h *LstepSettingsHandler) GetLstepSettings(c *gin.Context) {
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
func (h *LstepSettingsHandler) UpdateLstepSettings(c *gin.Context) {
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
func (h *LstepSettingsHandler) DeleteLstepSettings(c *gin.Context) {
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
func (h *LstepSettingsHandler) TestLstepConnection(c *gin.Context) {
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
