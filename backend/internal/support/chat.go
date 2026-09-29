package support

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	chatMessageMaxLength    = 2000
	chatHistoryMaxMessages  = 16
	chatHistoryMaxLength    = 4000
	chatContextMaxItems     = 6
	chatContextTitleMaxLen  = 200
	chatContextFieldMaxLen  = 100
	chatContextTextMaxLen   = 6000
	chatMaxCompletionTokens = 1024
)

// chatSystemPrompt はマニュアル抜粋のみを根拠に回答させるグラウンディング用プロンプト。
// 引用元の明示は構造化 sources で返すため、本文内の記事名出力は求めない。
const chatSystemPrompt = `あなたは動物病院向け電子カルテ「Animal Ekarte」の操作サポートアシスタントです。

ルール:
- 回答は必ず日本語で、簡潔に（300字以内を目安）
- <manual> タグ内のマニュアル抜粋だけを根拠に回答する
- 抜粋に記載がない場合は「マニュアルに記載がありません」と述べ、取扱説明書ページか「バグを報告」タブの利用を案内する
- 医療判断・診断・料金・薬に関する質問には回答せず、システム操作の質問のみ扱う
- 手順は箇条書きで示す`

// chatContextItem はフロントエンドが検索で取得したマニュアル抜粋。
// 回答の根拠に加え、sources としてそのままエコーバックする。
type chatContextItem struct {
	Title    string `json:"title"`
	Category string `json:"category"`
	Slug     string `json:"slug"`
	Text     string `json:"text"`
}

// chatRequest は POST /support/chat のボディ
type chatRequest struct {
	Message string            `json:"message"`
	History []ChatMessage     `json:"history"`
	Context []chatContextItem `json:"context"`
}

// ChatSource は回答の根拠となったマニュアル記事（レスポンス・履歴保存の双方で使う）
type ChatSource struct {
	Title    string `json:"title"`
	Category string `json:"category"`
	Slug     string `json:"slug"`
}

// chatResponse は POST /support/chat のレスポンス
type chatResponse struct {
	Reply   string       `json:"reply"`
	Sources []ChatSource `json:"sources"`
}

// chatHistoryItem は GET /support/chat/history の履歴1件
type chatHistoryItem struct {
	ID        uint64       `json:"id"`
	Role      string       `json:"role"`
	Content   string       `json:"content"`
	Sources   []ChatSource `json:"sources,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
}

// chatHistoryResponse は GET /support/chat/history のレスポンス
type chatHistoryResponse struct {
	Data []chatHistoryItem `json:"data"`
}

// parseChatRequest はチャットリクエストをバインド・検証する。
func parseChatRequest(c *gin.Context) (*chatRequest, error) {
	var req chatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, apperrors.WrapInvalidInput("invalid request body")
	}
	req.Message = strings.TrimSpace(req.Message)
	if req.Message == "" {
		return nil, apperrors.WrapInvalidInput("message is required")
	}
	if len(req.Message) > chatMessageMaxLength {
		return nil, apperrors.WrapInvalidInput("message is too long")
	}
	if len(req.History) > chatHistoryMaxMessages {
		return nil, apperrors.WrapInvalidInput("history is too long")
	}
	for _, m := range req.History {
		if m.Role != "user" && m.Role != "assistant" {
			return nil, apperrors.WrapInvalidInput("invalid history role")
		}
		if len(m.Content) > chatHistoryMaxLength {
			return nil, apperrors.WrapInvalidInput("history message is too long")
		}
	}
	if len(req.Context) > chatContextMaxItems {
		return nil, apperrors.WrapInvalidInput("context has too many items")
	}
	for _, item := range req.Context {
		if len(item.Title) > chatContextTitleMaxLen ||
			len(item.Category) > chatContextFieldMaxLen ||
			len(item.Slug) > chatContextFieldMaxLen ||
			len(item.Text) > chatContextTextMaxLen {
			return nil, apperrors.WrapInvalidInput("context item is too long")
		}
	}
	return &req, nil
}

// buildChatMessages はシステムプロンプト + 会話履歴 + 最新質問（マニュアル抜粋注入）を組み立てる。
// 抜粋は最終ユーザーメッセージに埋め込み、履歴は会話だけを保持する形にする。
func buildChatMessages(req *chatRequest) []ChatMessage {
	messages := make([]ChatMessage, 0, len(req.History)+2)
	messages = append(messages, ChatMessage{Role: "system", Content: chatSystemPrompt})
	messages = append(messages, req.History...)

	content := "質問: " + req.Message
	if len(req.Context) > 0 {
		var excerpts strings.Builder
		for _, item := range req.Context {
			fmt.Fprintf(&excerpts, "<manual title=%q>\n%s\n</manual>\n", item.Title, item.Text)
		}
		content = "参考となるマニュアル抜粋:\n" + excerpts.String() + "\n質問: " + req.Message
	}
	return append(messages, ChatMessage{Role: "user", Content: content})
}

func toChatSources(items []chatContextItem) []ChatSource {
	sources := make([]ChatSource, len(items))
	for i, item := range items {
		sources[i] = ChatSource{Title: item.Title, Category: item.Category, Slug: item.Slug}
	}
	return sources
}

// toChatHistoryItem は保存済みメッセージをレスポンス形に変換する。
// sources は保存時に必ず有効な JSON（jsonb）なので、壊れていた場合は空にして件名だけ返す。
func toChatHistoryItem(m model.SupportChatMessage) chatHistoryItem {
	item := chatHistoryItem{
		ID:        m.ID,
		Role:      string(m.Role),
		Content:   m.Content,
		CreatedAt: httpapi.LocalTime(m.CreatedAt),
	}

	if len(m.Sources) > 0 {
		var sources []ChatSource
		if err := json.Unmarshal(m.Sources, &sources); err == nil {
			item.Sources = sources
		}
	}
	return item
}

// ChatStatus はチャット機能の有効/無効を返す。
//
// GET /api/v1/support/chat/status
func (h *Handler) ChatStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"enabled": h.chat != nil})
}

// Chat はマニュアル抜粋を根拠に LLM へ質問し回答を返す。
//
// POST /api/v1/support/chat
// 認証済みスタッフ全員が利用できる（権限ゲートなし）。
func (h *Handler) Chat(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	if h.chat == nil {
		httpapi.RespondError(c, apperrors.WrapNotImplemented("support chat is not configured"))
		return
	}
	req, err := parseChatRequest(c)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	reply, err := h.chat.Complete(c.Request.Context(), buildChatMessages(req))
	if err != nil {
		var upstream *upstreamError
		switch {
		case errors.As(err, &upstream):
			slog.WarnContext(c.Request.Context(), "support chat upstream error", "status", upstream.status)
			httpapi.RespondError(c, apperrors.WrapBadGateway("support chat upstream error"))
		case errors.Is(err, errUpstreamEmpty):
			slog.WarnContext(c.Request.Context(), "support chat upstream returned empty content")
			httpapi.RespondError(c, apperrors.WrapBadGateway("support chat returned empty response"))
		default:
			slog.WarnContext(c.Request.Context(), "support chat failed", "error", err)
			httpapi.RespondError(c, apperrors.WrapInternalServerError("support chat failed"))
		}
		return
	}
	sources := toChatSources(req.Context)
	// 履歴保存は回答返却が主契約のため best-effort とする。保存失敗時は当該やり取りが
	// 次回セッションの履歴に出ないだけなので、WARN を残し回答は通常どおり返す。
	// service はテストで nil にできるため nil ガードを挟む。
	if h.service != nil {
		// プロバイダ側の退行で過大な応答が返ってきても text 列へ無制限には書き込まない。
		if len(reply) > chatHistoryMaxLength {
			slog.WarnContext(c.Request.Context(), "support chat reply too long; skipping persistence", "reply_length", len(reply), "clinic_id", clinicID, "staff_id", staffID)
		} else if err := h.service.RecordChatExchange(c.Request.Context(), clinicID, staffID, req.Message, reply, sources); err != nil {
			logChatPersistFailure(c.Request.Context(), err, clinicID, staffID)
		}
	}
	c.JSON(http.StatusOK, chatResponse{Reply: reply, Sources: sources})
}

// logChatPersistFailure は履歴保存失敗を WARN で残す。pg エラーの DETAIL/WHERE は
// 行内容を含み得るため、Message と SQLSTATE のみ記録する。
func logChatPersistFailure(ctx context.Context, err error, clinicID, staffID uint64) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		slog.WarnContext(ctx, "failed to persist support chat history", "error", pgErr.Message, "pg_code", pgErr.Code, "clinic_id", clinicID, "staff_id", staffID)
		return
	}
	slog.WarnContext(ctx, "failed to persist support chat history", "error", err, "clinic_id", clinicID, "staff_id", staffID)
}

// ChatHistory はログイン中スタッフ（選択clinic内）の会話履歴を古い順で返す。
//
// GET /api/v1/support/chat/history
// 認証済みスタッフ全員が利用できる（権限ゲートなし）。他人・他院の履歴は scope 外。
func (h *Handler) ChatHistory(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	// 永続化未配線（テスト等）では履歴は空が正しい値。Chat の 501 とは意味が違う。
	if h.service == nil {
		c.JSON(http.StatusOK, chatHistoryResponse{Data: []chatHistoryItem{}})
		return
	}
	messages, err := h.service.ListChatHistory(c.Request.Context(), clinicID, staffID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	items := make([]chatHistoryItem, len(messages))
	for i, m := range messages {
		items[i] = toChatHistoryItem(m)
	}
	c.JSON(http.StatusOK, chatHistoryResponse{Data: items})
}

// ClearChatHistory はログイン中スタッフ（選択clinic内）の会話履歴をすべて削除する。
//
// DELETE /api/v1/support/chat/history
func (h *Handler) ClearChatHistory(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	// 永続化未配線なら消すものがないので no-op で成功扱いにする。
	if h.service == nil {
		c.Status(http.StatusNoContent)
		return
	}
	if err := h.service.ClearChatHistory(c.Request.Context(), clinicID, staffID); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
