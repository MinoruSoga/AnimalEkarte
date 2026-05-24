package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
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

// updateTriggerPriorityItemRequest は PATCH /lstep/trigger-priorities のリクエスト内1件。
type updateTriggerPriorityItemRequest struct {
	TriggerType string `json:"trigger_type" binding:"required"`
	Priority    int    `json:"priority"     binding:"required,min=1"`
}

// updateTriggerPrioritiesRequest は PATCH /lstep/trigger-priorities のリクエストボディ。
type updateTriggerPrioritiesRequest struct {
	Items []updateTriggerPriorityItemRequest `json:"items" binding:"required,min=1,dive"`
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
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	items, err := h.svc.LstepTriggerPriority.GetByClinicID(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTriggerPriorityListResponse(clinicID, items))
}

// UpdateLstepTriggerPriorities godoc
// PATCH /api/v1/lstep/trigger-priorities — トリガー優先順位を一括更新する（Q23）。
func (h *Handler) UpdateLstepTriggerPriorities(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req updateTriggerPrioritiesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	input := service.UpdateTriggerPrioritiesInput{
		Items: make([]service.TriggerPriorityItem, len(req.Items)),
	}
	for i, it := range req.Items {
		input.Items[i] = service.TriggerPriorityItem{
			TriggerType: it.TriggerType,
			Priority:    it.Priority,
		}
	}
	if err := h.svc.LstepTriggerPriority.UpdatePriorities(c.Request.Context(), clinicID, input); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusOK)
}

// RegisterLstepTriggerPriorityRoutes は Q23 のルートを登録する。
func (h *Handler) RegisterLstepTriggerPriorityRoutes(rg *gin.RouterGroup) {
	lstep := rg.Group("/lstep")
	lstep.GET("/trigger-priorities", h.RequirePermission(string(model.ResourceHospitalSettings), "view"), h.GetLstepTriggerPriorities)
	lstep.PATCH("/trigger-priorities", h.RequirePermission(string(model.ResourceHospitalSettings), "edit"), h.UpdateLstepTriggerPriorities)

	// FE が /clinics/:clinic_id/lstep/... で呼ぶエイリアス
	clinicLstep := rg.Group("/clinics/:clinic_id/lstep")
	clinicLstep.GET("/trigger-priorities", h.RequirePermission(string(model.ResourceHospitalSettings), "view"), h.GetLstepTriggerPriorities)
	clinicLstep.PATCH("/trigger-priorities", h.RequirePermission(string(model.ResourceHospitalSettings), "edit"), h.UpdateLstepTriggerPriorities)
}
