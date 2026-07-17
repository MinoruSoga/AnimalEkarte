package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// SendLineMessage godoc
func (h *Handler) SendLineMessage(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	ownerID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	staffID, ok := extractStaffID(c)
	if !ok {
		return
	}

	var req lineSendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	if req.MessageType == "text" && len([]rune(req.Text)) > 500 {
		RespondError(c, apperrors.WrapInvalidInput("テキストは500文字以内で入力してください"))
		return
	}

	result, err := h.svc.LineSend.Send(c.Request.Context(), clinicID, req.toServiceInput(ownerID, staffID))
	if err != nil {
		RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, lineSendResponse{
		Sent:     true,
		SentAt:   localTimeRFC3339(result.SentAt),
		TagAdded: result.TagAdded,
	})
}

// GetLineSendLogs godoc
func (h *Handler) GetLineSendLogs(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	ownerID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	logs, err := h.svc.LineSend.GetSendLogs(c.Request.Context(), clinicID, ownerID)
	if err != nil {
		RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, lineSendLogListResponse{Items: toLineSendLogListResponse(logs)})
}

// RegisterLineSendRoutes は LINE 個別送信ルートを /owners/:id 以下に登録する
func (h *Handler) RegisterLineSendRoutes(owners *gin.RouterGroup) {
	owners.POST("/:id/line/send", h.RequirePermission(string(model.ResourceOwners), "edit"), h.SendLineMessage)
	owners.GET("/:id/line/send-logs", h.RequirePermission(string(model.ResourceOwners), "view"), h.GetLineSendLogs)
	// ISSUE-002: FE統一エンドポイントのエイリアス
	owners.POST("/:id/lstep/send", h.RequirePermission(string(model.ResourceOwners), "edit"), h.SendLineMessage)
	owners.GET("/:id/lstep/send-history", h.RequirePermission(string(model.ResourceOwners), "view"), h.GetLineSendLogs)
}
