package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
)

// linkTokenResponse は GenerateLineLinkToken のレスポンス。
type linkTokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
	LiffURL   string `json:"liff_url,omitempty"`
}

func toLinkTokenResponse(r *service.LinkTokenResult) linkTokenResponse {
	return linkTokenResponse{
		Token:     r.Token,
		ExpiresAt: localTimeRFC3339(r.ExpiresAt),
		LiffURL:   r.LiffURL,
	}
}

// ReceiveLineWebhook は LINE Webhook を受信する。
// POST /api/line/webhook
func (h *Handler) ReceiveLineWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("failed to read request body"))
		return
	}
	signature := c.GetHeader("X-Line-Signature")
	if signature == "" {
		RespondError(c, apperrors.WrapInvalidInput("missing X-Line-Signature header"))
		return
	}
	if err := h.svc.LineLink.HandleWebhook(c.Request.Context(), body, signature); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusOK)
}

// GenerateLineLinkToken は飼い主向け LINE 連携トークンを発行する。
// POST /api/v1/owners/:id/line/link-token
func (h *Handler) GenerateLineLinkToken(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	ownerID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	result, err := h.svc.LineLink.GenerateLinkToken(c.Request.Context(), clinicID, ownerID)
	if err != nil {
		RespondError(c, err)
		return
	}
	// P15例外: link token は単回使用で GET エンドポイントを持たないため Location ヘッダなし
	c.JSON(http.StatusCreated, toLinkTokenResponse(result))
}

// LinkLiffAccount は LIFF アプリから link_token + line_id_token で飼い主に LINE User ID を紐付ける。
// POST /api/liff/:clinicId/link
func (h *Handler) LinkLiffAccount(c *gin.Context) {
	clinicID, ok := parseIDParam(c, "clinicId")
	if !ok {
		return
	}
	var req linkAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	owner, err := h.svc.LineLink.LinkAccount(c.Request.Context(), clinicID, req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toOwnerResponse(owner))
}
