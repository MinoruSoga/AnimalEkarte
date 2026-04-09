package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/service"
)

// ---- LINE Webhook Event types ----

type lineWebhookBody struct {
	Events []lineWebhookEvent `json:"events"`
}

type lineWebhookEvent struct {
	Type       string          `json:"type"`
	ReplyToken string          `json:"replyToken"`
	Source     lineEventSource `json:"source"`
}

type lineEventSource struct {
	Type   string `json:"type"`
	UserID string `json:"userId"`
}

// HandleLineWebhook はLINE Messaging APIのWebhookを受け取る。
// 友だち追加（follow）イベントで予約ページURLを自動送信する。
// POST /api/line-webhook/:clinicId
func (h *Handler) HandleLineWebhook(c *gin.Context) {
	clinicID, err := strconv.ParseUint(c.Param("clinicId"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// リクエストボディを読み取り（署名検証に生バイトが必要）
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	ctx := c.Request.Context()

	// 設定取得
	setting, err := h.repos.ReservationSetting.FindByClinicID(ctx, clinicID)
	if err != nil || setting == nil || setting.LineChannelSecret == "" {
		// LINE未設定の医院 → 200を返す（LINEは200以外でリトライする）
		c.Status(http.StatusOK)
		return
	}

	// 署名検証
	signature := c.GetHeader("X-Line-Signature")
	if !verifyLineSignature(body, signature, setting.LineChannelSecret) {
		slog.WarnContext(ctx, "LINE webhook signature mismatch", "clinic_id", clinicID)
		c.Status(http.StatusForbidden)
		return
	}

	// イベント解析
	var webhook lineWebhookBody
	if err := json.Unmarshal(body, &webhook); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// 各イベントを処理
	for _, event := range webhook.Events {
		if event.Type == "follow" && event.Source.UserID != "" {
			h.handleFollowEvent(ctx, clinicID, event.Source.UserID, setting.LiffID, setting.LineAccessToken)
		}
	}

	c.Status(http.StatusOK)
}

// handleFollowEvent は友だち追加イベントを処理する。
// 顧客レコードの作成と予約ページURLの送信を行う。
func (h *Handler) handleFollowEvent(ctx context.Context, clinicID uint64, lineUserID, liffID, accessToken string) {
	// 顧客レコードを作成（既存なら取得）
	_, err := h.repos.ReservationCustomerMgr.FindOrCreateByLineUserID(ctx, clinicID, lineUserID, "")
	if err != nil {
		slog.WarnContext(ctx, "webhook: failed to create customer", "error", err, "line_user_id", lineUserID)
		return
	}

	// 予約ページURLを送信
	url := buildReservationPageURL(liffID)
	if url == "" {
		return
	}

	msg := fmt.Sprintf("友だち追加ありがとうございます！\n\nこちらからオンライン予約ができます。\n%s", url)
	msgSvc := service.NewLineMessagingService(accessToken)
	if err := msgSvc.PushText(ctx, lineUserID, msg); err != nil {
		slog.WarnContext(ctx, "webhook: failed to send welcome message", "error", err, "line_user_id", lineUserID)
	}
}

// verifyLineSignature はLINE Webhookの署名を検証する。
func verifyLineSignature(body []byte, signature, channelSecret string) bool {
	if signature == "" || channelSecret == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(channelSecret))
	mac.Write(body)
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

// buildReservationPageURL は予約ページのLIFF URLを生成する。
func buildReservationPageURL(liffID string) string {
	if liffID == "" {
		return ""
	}
	return fmt.Sprintf("https://liff.line.me/%s", liffID)
}
