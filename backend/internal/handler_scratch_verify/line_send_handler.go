package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// normalizeMessageType は FE から送られる message_type エイリアスをサービス内部値に変換する（ISSUE-002）。
func normalizeMessageType(t string) string {
	switch t {
	case "pdf":
		return "pdf_url"
	case "image":
		return "image_url"
	default:
		return t
	}
}

type lineSendResponse struct {
	Sent     bool   `json:"sent"`
	SentAt   string `json:"sent_at"`
	TagAdded string `json:"tag_added,omitempty"`
}

type lineSendLogResponse struct {
	ID             uint64  `json:"id"`
	MessageType    string  `json:"message_type"`
	ContentSummary string  `json:"content_summary"`
	Status         string  `json:"status"`
	ErrorMessage   *string `json:"error_message,omitempty"`
	SentAt         string  `json:"sent_at"`
}

type lineSendLogListResponse struct {
	Items []lineSendLogResponse `json:"items"`
}

func toLineSendLogResponse(l *model.LineSendLog) lineSendLogResponse {
	return lineSendLogResponse{
		ID:             l.ID,
		MessageType:    l.MessageType,
		ContentSummary: l.ContentSummary,
		Status:         l.Status,
		ErrorMessage:   l.ErrorMessage,
		SentAt:         l.SentAt.In(time.Local).Format(time.RFC3339),
	}
}

func toLineSendLogListResponse(logs []*model.LineSendLog) []lineSendLogResponse {
	out := make([]lineSendLogResponse, 0, len(logs))
	for _, l := range logs {
		out = append(out, toLineSendLogResponse(l))
	}
	return out
}

// PostLineSend godoc
func (h *Handler) PostLineSend(c *gin.Context) {
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
		SentAt:   result.SentAt.In(time.Local).Format(time.RFC3339),
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
	owners.POST("/:id/line/send", h.RequirePermission(string(model.ResourceOwners), "edit"), h.PostLineSend)
	owners.GET("/:id/line/send-logs", h.RequirePermission(string(model.ResourceOwners), "view"), h.GetLineSendLogs)
	// ISSUE-002: FE統一エンドポイントのエイリアス
	owners.POST("/:id/lstep/send", h.RequirePermission(string(model.ResourceOwners), "edit"), h.PostLineSend)
	owners.GET("/:id/lstep/send-history", h.RequirePermission(string(model.ResourceOwners), "view"), h.GetLineSendLogs)
}
