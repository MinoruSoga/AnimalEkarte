package support

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
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

// chatSource は回答の根拠となったマニュアル記事
type chatSource struct {
	Title    string `json:"title"`
	Category string `json:"category"`
	Slug     string `json:"slug"`
}

// chatResponse は POST /support/chat のレスポンス
type chatResponse struct {
	Reply   string       `json:"reply"`
	Sources []chatSource `json:"sources"`
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

func toChatSources(items []chatContextItem) []chatSource {
	sources := make([]chatSource, len(items))
	for i, item := range items {
		sources[i] = chatSource{Title: item.Title, Category: item.Category, Slug: item.Slug}
	}
	return sources
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
	if _, ok := httpapi.ExtractClinicID(c); !ok {
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
	c.JSON(http.StatusOK, chatResponse{Reply: reply, Sources: toChatSources(req.Context)})
}
