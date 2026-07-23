package lstep

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const maxLineWebhookRequestBytes int64 = 2 * 1024 * 1024

// OwnerResponder は紐付け成功時の飼主レスポンスを書き出す注入 closure。
// owner の公開 DTO は internal/handler が単一正本として保持しており、owner domain 移行
// （BE9-2E）まで lstep 側に複製しないため注入で解決する。
type OwnerResponder func(c *gin.Context, o *model.Owner)

// LineLinkHandler は LineLinkService の HTTP handler。
type LineLinkHandler struct {
	svc               LineLinkService
	respondOwner      OwnerResponder
	requirePermission PermissionMiddleware
}

// NewLineLinkHandler は LineLinkHandler を構築する。
func NewLineLinkHandler(svc LineLinkService, respondOwner OwnerResponder, requirePermission PermissionMiddleware) *LineLinkHandler {
	return &LineLinkHandler{svc: svc, respondOwner: respondOwner, requirePermission: requirePermission}
}

// linkTokenResponse は GenerateLineLinkToken のレスポンス。
type linkTokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
	LiffURL   string `json:"liff_url,omitempty"`
}

func toLinkTokenResponse(r *LinkTokenResult) linkTokenResponse {
	return linkTokenResponse{
		Token:     r.Token,
		ExpiresAt: httpapi.LocalTimeRFC3339(r.ExpiresAt),
		LiffURL:   r.LiffURL,
	}
}

// ReceiveLineWebhook は LINE Webhook を受信する。
// POST /api/line/webhook
func (h *LineLinkHandler) ReceiveLineWebhook(c *gin.Context) {
	signature := c.GetHeader("X-Line-Signature")
	if signature == "" {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("missing X-Line-Signature header"))
		return
	}
	if c.Request.ContentLength > maxLineWebhookRequestBytes {
		httpapi.RespondError(c, apperrors.WrapPayloadTooLarge("LINE webhook request exceeds size limit"))
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxLineWebhookRequestBytes)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httpapi.RespondError(c, apperrors.WrapPayloadTooLarge("LINE webhook request exceeds size limit"))
			return
		}
		httpapi.RespondError(c, apperrors.WrapInvalidInput("failed to read request body"))
		return
	}
	if err := h.svc.HandleWebhook(c.Request.Context(), body, signature); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusOK)
}

// GenerateLineLinkToken は飼い主向け LINE 連携トークンを発行する。
// POST /api/v1/owners/:id/line/link-token
func (h *LineLinkHandler) GenerateLineLinkToken(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	ownerID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	result, err := h.svc.GenerateLinkToken(c.Request.Context(), clinicID, ownerID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	// P15例外: link token は単回使用で GET エンドポイントを持たないため Location ヘッダなし
	c.JSON(http.StatusCreated, toLinkTokenResponse(result))
}

// LinkLiffAccount は LIFF アプリから link_token + line_id_token で飼い主に LINE User ID を紐付ける。
// POST /api/liff/:clinicId/link
func (h *LineLinkHandler) LinkLiffAccount(c *gin.Context) {
	clinicID, ok := httpapi.ParseIDParam(c, "clinicId")
	if !ok {
		return
	}
	var req linkAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	owner, err := h.svc.LinkAccount(c.Request.Context(), clinicID, req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	h.respondOwner(c, owner)
}
