package support

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	chatMessageMaxLength    = 2000
	chatHistoryMaxMessages  = 16
	chatHistoryMaxLength    = 4000
	chatContextMaxItems     = 6
	chatMaxCompletionTokens = 1024
)

// chatSystemPrompt はマニュアル抜粋のみを根拠に回答させるグラウンディング用プロンプト。
// 引用元の明示は構造化 sources で返すため、本文内の記事名出力は求めない。
const chatSystemPrompt = `あなたは動物病院向け電子カルテ「Animal Ekarte」の操作サポートアシスタントです。

ルール:
- 回答は質問と同じ言語で、簡潔に（300字以内を目安）
- <manual> タグ内のマニュアル抜粋だけを根拠に回答する
- 抜粋に記載がない場合は質問の言語でその旨を述べ（日本語なら「マニュアルに記載がありません」、ベトナム語なら "Không có thông tin trong tài liệu hướng dẫn"）、取扱説明書ページか「バグを報告」タブの利用を案内する
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
	// EMR-410 ハイブリッド検索契約。
	ManualRevision    string            `json:"manual_revision,omitempty"`
	KeywordCandidates []ManualCandidate `json:"keyword_candidates,omitempty"`
}

// ChatSource は回答の根拠となったマニュアル記事（レスポンス・履歴保存の双方で使う）
type ChatSource struct {
	Title    string `json:"title"`
	Category string `json:"category"`
	Slug     string `json:"slug"`
}

// chatResponse は POST /support/chat のレスポンス
type chatResponse struct {
	Reply     string         `json:"reply"`
	Sources   []ChatSource   `json:"sources"`
	Retrieval *RetrievalInfo `json:"retrieval,omitempty"`
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

// chatRoomTurnItem は GET /support/chat/rooms のルーム内1発言
type chatRoomTurnItem struct {
	ID        uint64                       `json:"id"`
	Role      model.SupportChatMessageRole `json:"role"`
	Content   string                       `json:"content"`
	Sources   []ChatSource                 `json:"sources,omitempty"`
	CreatedAt time.Time                    `json:"created_at"`
}

// chatRoomItem は GET /support/chat/rooms の1ルーム（疑似セッション + provenance）
type chatRoomItem struct {
	ID            string             `json:"id"`
	ClinicName    string             `json:"clinic_name"`
	StaffName     string             `json:"staff_name"`
	StartedAt     time.Time          `json:"started_at"`
	LastMessageAt time.Time          `json:"last_message_at"`
	Turns         []chatRoomTurnItem `json:"turns"`
}

// chatRoomListResponse は GET /support/chat/rooms のレスポンス
type chatRoomListResponse struct {
	Data []chatRoomItem `json:"data"`
}

// toChatRoomItem はサービス層のルームをレスポンス形に変換する。
func toChatRoomItem(r ChatRoom) chatRoomItem {
	turns := make([]chatRoomTurnItem, len(r.Turns))
	for i, t := range r.Turns {
		turns[i] = chatRoomTurnItem{
			ID:        t.ID,
			Role:      t.Role,
			Content:   t.Content,
			CreatedAt: t.CreatedAt,
		}
		if len(t.Sources) > 0 {
			var sources []ChatSource
			if err := json.Unmarshal(t.Sources, &sources); err == nil {
				turns[i].Sources = sources
			}
		}
	}
	return chatRoomItem{
		ID:            r.ID,
		ClinicName:    r.ClinicName,
		StaffName:     r.StaffName,
		StartedAt:     r.StartedAt,
		LastMessageAt: r.LastMessageAt,
		Turns:         turns,
	}
}

// ListChatRooms は全医院のチャット履歴を疑似会話ルーム単位で返す。
//
// GET /api/v1/support/chat/rooms
// 認証済みスタッフ全員が利用できる（ListChatExchanges と同じ共有ボード方針）。
// ルーム化は時間ギャップによるヒューリスティック — DB に会話境界は無い。
func (h *Handler) ListChatRooms(c *gin.Context) {
	if h.service == nil {
		c.JSON(http.StatusOK, chatRoomListResponse{Data: []chatRoomItem{}})
		return
	}
	rooms, err := h.service.ListChatRooms(c.Request.Context())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	items := make([]chatRoomItem, len(rooms))
	for i, r := range rooms {
		items[i] = toChatRoomItem(r)
	}
	c.JSON(http.StatusOK, chatRoomListResponse{Data: items})
}

// chatOutboundScreeningPatterns は外部 LLM へ送信する前に拒否する機微情報パターン。
// ヒューリスティックな best-effort 分類（完全な検出は保証しない）— 方針として
// 個人情報・認証情報らしき入力は外部 LLM へ送らず invalid input で拒否する。
var chatOutboundScreeningPatterns = []*regexp.Regexp{
	// メールアドレス
	regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9-]+(\.[A-Za-z0-9-]+)+`),
	// 日本の電話番号（0X0-XXXX-XXXX などハイフン区切り）
	regexp.MustCompile(`0\d{1,4}-\d{1,4}-\d{4}`),
	// 携帯電話（ハイフンなし 070/080/090 の 11 桁）
	regexp.MustCompile(`0[789]0\d{8}`),
	// クレジットカード・マイナンバー相当の長い数字列（12 桁以上）
	regexp.MustCompile(`\d{12,}`),
	// 認証情報らしき代入形式（password=..., api_key: ... 等）
	regexp.MustCompile(`(?i)(password|passwd|api[_-]?key|secret|token|authorization)\s*[:=]\s*\S+`),
	// Bearer トークン
	regexp.MustCompile(`(?i)bearer\s+[a-z0-9._\-]{8,}`),
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
	seen := map[string]struct{}{}
	sources := make([]ChatSource, 0, len(items))
	for _, item := range items {
		key := item.Category + "/" + item.Slug
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		sources = append(sources, ChatSource{Title: item.Title, Category: item.Category, Slug: item.Slug})
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
	// EMR-410: 本番経路は manualSearch 依存必須。欠落は 503。
	if h.manualSearch == nil {
		httpapi.RespondError(c, apperrors.WrapServiceUnavailable("manual search is unavailable"))
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), chatRequestDeadline)
	defer cancel()
	c.Request = c.Request.WithContext(ctx)
	h.runHybridChat(c, clinicID, staffID)
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
