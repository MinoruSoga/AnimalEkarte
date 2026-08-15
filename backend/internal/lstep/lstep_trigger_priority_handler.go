package lstep

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// triggerPriorityItemResponse はトリガー優先順位1件のJSONレスポンス。
type triggerPriorityItemResponse struct {
	TriggerType string `json:"trigger_type"`
	Priority    int    `json:"priority"`
}

// triggerPriorityListResponse は GET /lstep/trigger-priorities のJSONレスポンス。
type triggerPriorityListResponse struct {
	ClinicID string                        `json:"clinic_id"`
	Items    []triggerPriorityItemResponse `json:"items"`
}

func toTriggerPriorityListResponse(clinicID uint64, items []model.LstepTriggerPriority) triggerPriorityListResponse {
	resp := triggerPriorityListResponse{
		ClinicID: strconv.FormatUint(clinicID, 10),
		Items:    make([]triggerPriorityItemResponse, len(items)),
	}
	for i, it := range items {
		resp.Items[i] = triggerPriorityItemResponse{
			TriggerType: it.TriggerType,
			Priority:    it.Priority,
		}
	}
	return resp
}

// GetLstepTriggerPriorities godoc
// GET /api/v1/lstep/trigger-priorities — クリニックのトリガー優先順位一覧を返す（Q23）。
func (h *Handler) GetLstepTriggerPriorities(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	items, err := h.triggerPriority.GetByClinicID(c.Request.Context(), clinicID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTriggerPriorityListResponse(clinicID, items))
}

// UpdateLstepTriggerPriorities godoc
// PATCH /api/v1/lstep/trigger-priorities — トリガー優先順位を一括更新する（Q23）。
func (h *Handler) UpdateLstepTriggerPriorities(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req updateTriggerPrioritiesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if err := h.triggerPriority.UpdatePriorities(c.Request.Context(), clinicID, req.toServiceInput()); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusOK)
}

// RegisterLstepTriggerPriorityRoutes は Q23 のルートを登録する。
func (h *Handler) RegisterLstepTriggerPriorityRoutes(rg *gin.RouterGroup) {
	lstep := rg.Group("/lstep")
	lstep.GET("/trigger-priorities", h.requirePermission(string(model.ResourceHospitalSettings), "view"), h.GetLstepTriggerPriorities)
	lstep.PATCH("/trigger-priorities", h.requirePermission(string(model.ResourceHospitalSettings), "edit"), h.UpdateLstepTriggerPriorities)

	// FE が /clinics/:clinic_id/lstep/... で呼ぶエイリアス
	clinicLstep := rg.Group("/clinics/:clinic_id/lstep")
	clinicLstep.GET("/trigger-priorities", h.requirePermission(string(model.ResourceHospitalSettings), "view"), h.GetLstepTriggerPriorities)
	clinicLstep.PATCH("/trigger-priorities", h.requirePermission(string(model.ResourceHospitalSettings), "edit"), h.UpdateLstepTriggerPriorities)
}
