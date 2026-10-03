package support

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
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

// chatRequest は POST /support/chat のボディ。
// history は意図的に受け付けない — LLM に送る会話履歴はクライアント入力ではなく
// サーバー保存済みのやり取り（送信時にスクリーニング済み）だけを使う。
type chatRequest struct {
	Message string            `json:"message"`
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

// chatExchangeItem は GET /support/chat/exchanges の1行（質問+回答ペア + provenance）
type chatExchangeItem struct {
	ID         uint64       `json:"id"`
	ClinicName string       `json:"clinic_name"`
	StaffName  string       `json:"staff_name"`
	Question   string       `json:"question"`
	Answer     string       `json:"answer"`
	Sources    []ChatSource `json:"sources,omitempty"`
	CreatedAt  time.Time    `json:"created_at"`
}

// chatExchangeListResponse は GET /support/chat/exchanges のレスポンス
type chatExchangeListResponse struct {
	Data []chatExchangeItem `json:"data"`
}

// toChatExchangeItem はペア化済み履歴をレスポンス形に変換する。
// created_at は質問送信時刻（回答はほぼ即時に続く）。
func toChatExchangeItem(e ChatExchange) chatExchangeItem {
	item := chatExchangeItem{
		ID:         e.AssistantMessage.ID,
		ClinicName: e.ClinicName,
		StaffName:  e.StaffName,
		Question:   e.UserMessage.Content,
		Answer:     e.AssistantMessage.Content,
		CreatedAt:  e.UserMessage.CreatedAt,
	}
	if len(e.AssistantMessage.Sources) > 0 {
		var sources []ChatSource
		if err := json.Unmarshal(e.AssistantMessage.Sources, &sources); err == nil {
			item.Sources = sources
		}
	}
	return item
}

// ListChatExchanges は全医院の質問+回答ペアを新しい順で返す。
//
// GET /api/v1/support/chat/exchanges
// 認証済みスタッフ全員が利用できる（権限ゲート・医院絞りなし — バグ報告ボードと
// 同じ共有ボード方針。質問傾向の横断分析が目的で、個人情報を含み得る点は product 承認済み）。
func (h *Handler) ListChatExchanges(c *gin.Context) {
	if h.service == nil {
		c.JSON(http.StatusOK, chatExchangeListResponse{Data: []chatExchangeItem{}})
		return
	}
	exchanges, err := h.service.ListChatExchanges(c.Request.Context())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	items := make([]chatExchangeItem, len(exchanges))
	for i, e := range exchanges {
		items[i] = toChatExchangeItem(e)
	}
	c.JSON(http.StatusOK, chatExchangeListResponse{Data: items})
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
	if err := screenChatOutbound(&req); err != nil {
		return nil, err
	}
	return &req, nil
}

// chatOutboundScreeningPatterns は外部 LLM へ送信する前に拒否する機微情報パターン。
// ヒューリスティックな best-effort 分類（完全な検出は保証しない）— 方針として
// 個人情報・認証情報らしき入力は外部 LLM へ送らず invalid input で拒否する。
var chatOutboundScreeningPatterns = []*regexp.Regexp{
	// メールアドレス
	regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9-]+(\.[A-Za-z0-9-]+)+`),
	// 日本の電話番号（0X0-XXXX-XXXX などハイフン区切り）
	regexp.MustCompile(`0[0-9]{1,4}-[0-9]{1,4}-[0-9]{4}`),
	// 携帯電話（ハイフンなし 070/080/090 の 11 桁）
	regexp.MustCompile(`0[789]0[0-9]{8}`),
	// クレジットカード・マイナンバー相当の長い数字列（12 桁以上）
	regexp.MustCompile(`[0-9]{12,}`),
	// 認証情報らしき代入形式（password=..., api_key: ... 等）
	regexp.MustCompile(`(?i)(password|passwd|api[_-]?key|secret|token|authorization)\s*[:=]\s*\S+`),
	// Bearer トークン
	regexp.MustCompile(`(?i)bearer\s+[a-z0-9._\-]{8,}`),
}

// screenChatOutbound は外部 LLM へ送信する新規入力（message・context 抜粋）に
// 機微情報らしき内容が含まれないか検査する。該当時は汎用 invalid_input を返し
// （どのパターンに当たったかは開示しない）、LLM 呼び出し自体を行わない。
func screenChatOutbound(req *chatRequest) error {
	texts := make([]string, 0, len(req.Context)*2+1)
	texts = append(texts, req.Message)
	for _, item := range req.Context {
		texts = append(texts, item.Title, item.Text)
	}
	for _, text := range texts {
		for _, pattern := range chatOutboundScreeningPatterns {
			if pattern.MatchString(text) {
				return apperrors.WrapInvalidInput("sensitive information is not allowed in support chat")
			}
		}
	}
	return nil
}

// buildChatMessages はシステムプロンプト + 会話履歴 + 最新質問（マニュアル抜粋注入）を組み立てる。
// history はサーバー保存済みの履歴（クライアント入力ではない）。抜粋は最終
// ユーザーメッセージに埋め込み、履歴は会話だけを保持する形にする。
func buildChatMessages(req *chatRequest, history []ChatMessage) []ChatMessage {
	messages := make([]ChatMessage, 0, len(history)+2)
	messages = append(messages, ChatMessage{Role: "system", Content: chatSystemPrompt})
	messages = append(messages, history...)

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

// loadChatHistory は保存済みの会話履歴を LLM コンテキスト用に整形して返す。
// クライアント送信の history は信頼しない — 外部 LLM へ送るのはサーバー保存済み
// （送信時にスクリーニング済み）のやり取りだけに限定する。
// 読み取り失敗・service 未配線時は履歴なしで継続する（回答自体は成立する）。
func (h *Handler) loadChatHistory(ctx context.Context, clinicID, staffID uint64) []ChatMessage {
	if h.service == nil {
		return nil
	}
	stored, err := h.service.ListChatHistory(ctx, clinicID, staffID)
	if err != nil {
		slog.WarnContext(ctx, "failed to load support chat history; continuing without it",
			"error", err, "clinic_id", clinicID, "staff_id", staffID)
		return nil
	}
	if len(stored) > chatHistoryMaxMessages {
		stored = stored[len(stored)-chatHistoryMaxMessages:]
	}
	messages := make([]ChatMessage, 0, len(stored))
	for _, m := range stored {
		messages = append(messages, ChatMessage{Role: string(m.Role), Content: m.Content})
	}
	return messages
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
	reply, err := h.chat.Complete(c.Request.Context(), buildChatMessages(req, h.loadChatHistory(c.Request.Context(), clinicID, staffID)))
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
