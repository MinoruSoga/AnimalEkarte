package lstep

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

const (
	// maxLineWebhookRequestBytes bounds webhook body size before signature
	// verification. follow/unfollow payloads are small; a large body amplifies
	// per-clinic HMAC work on invalid signatures (SEC-CS-F05).
	maxLineWebhookRequestBytes int64 = 256 * 1024
	maxLineLinkRequestBytes    int64 = 16 * 1024
	maxLineLinkTokenChars            = 128
	maxLineIDTokenChars              = 12 * 1024
)

// LineLinkHandler は LineLinkService の HTTP handler。
type LineLinkHandler struct {
	svc               LineLinkService
	requirePermission PermissionMiddleware
}

// NewLineLinkHandler は LineLinkHandler を構築する。
func NewLineLinkHandler(svc LineLinkService, requirePermission PermissionMiddleware) *LineLinkHandler {
	return &LineLinkHandler{svc: svc, requirePermission: requirePermission}
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
	if err := decodeLineLinkAccountRequest(c, &req); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	_, err := h.svc.LinkAccount(c.Request.Context(), clinicID, req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func decodeLineLinkAccountRequest(c *gin.Context, req *linkAccountRequest) error {
	if c.Request.ContentLength > maxLineLinkRequestBytes {
		return apperrors.WrapPayloadTooLarge("LINE link request exceeds size limit")
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxLineLinkRequestBytes)

	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(req); err != nil {
		return lineLinkDecodeError(err)
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return apperrors.WrapInvalidInput("request body must contain one JSON object")
		}
		return lineLinkDecodeError(err)
	}
	if req.LinkToken == "" {
		return apperrors.WrapInvalidInput("link_token is required")
	}
	if len(req.LinkToken) > maxLineLinkTokenChars {
		return apperrors.WrapInvalidInput("link_token exceeds size limit")
	}
	if req.LineIDToken == "" {
		return apperrors.WrapInvalidInput("line_id_token is required")
	}
	if len(req.LineIDToken) > maxLineIDTokenChars {
		return apperrors.WrapInvalidInput("line_id_token exceeds size limit")
	}
	return nil
}

func lineLinkDecodeError(err error) error {
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		return apperrors.WrapPayloadTooLarge("LINE link request exceeds size limit")
	}
	return apperrors.WrapInvalidInput("invalid LINE link request body")
}
