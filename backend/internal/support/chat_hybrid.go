package support

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/manualarticle"
	"github.com/animal-ekarte/backend/internal/model"
)

// chat_hybrid.go — POST /support/chat の権威検索経路（R5/R6/R7）。
// production では手動検索依存の配線が必須（欠落=503）。Enabled は semantic
// 検索と legacy context 受理可否のみを制御する。クライアントの title/text は
// 一切信用せず、ID を権限済みコーパスで再解決する。

const (
	chatRequestBodyMaxBytes  = 32 << 10 // 32KiB（trailing 含む全体）
	chatCandidateMaxItems    = 20
	chatCandidateFieldMaxLen = 100
	chatLegacyMaxItems       = 6
	chatRequestDeadline      = 22 * time.Second
	chatGenerationTimeout    = 20 * time.Second
	chatPromptMaxBytes       = 16 << 10 // 16KiB UTF-8
	chatFallbackSourceLimit  = 3
)

var candidateSlugRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,99}$`)
var revisionHexRe = regexp.MustCompile(`^[0-9a-f]{64}$`)

var errNotGranted = errors.New("llm budget not granted")

func chatFixedResponse(lang string) string {
	if lang == "vi" {
		return "Không có thông tin trong tài liệu hướng dẫn. Vui lòng xem trang hướng dẫn sử dụng hoặc tab \"Báo cáo lỗi\"."
	}
	return "マニュアルに記載がありません。取扱説明書ページか「バグを報告」タブをご利用ください。"
}

func detectChatLanguage(message string) string {
	for _, r := range message {
		if (r >= 'ぁ' && r <= 'ん') || (r >= 'ァ' && r <= 'ン') || (r >= '一' && r <= '鿿') {
			return "ja"
		}
	}
	return "vi"
}

// screenOutboundTexts は外部送信する全ての文字列に機微情報パターンを適用する。
func screenOutboundTexts(texts ...string) error {
	for _, text := range texts {
		for _, pattern := range chatOutboundScreeningPatterns {
			if pattern.MatchString(text) {
				return apperrors.WrapInvalidInput("sensitive information is not allowed in support chat")
			}
		}
	}
	return nil
}

// hybridChatRequest は raw JSON から context フィールドの有無を区別して解析する。
type hybridChatRequest struct {
	chatRequest
	hasContextField bool
}

func parseHybridChatRequest(c *gin.Context) (*hybridChatRequest, error) {
	// 32KiB+1 で読み切り、超過は 413。trailing JSON/whitespace は Unmarshal が拒否。
	raw, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, chatRequestBodyMaxBytes))
	if err != nil {
		return nil, apperrors.WrapPayloadTooLarge("request body exceeds size limit")
	}
	var probe struct {
		Context           json.RawMessage `json:"context"`
		KeywordCandidates json.RawMessage `json:"keyword_candidates"`
	}
	var req chatRequest
	if err := json.Unmarshal(raw, &probe); err != nil {
		return nil, apperrors.WrapInvalidInput("invalid request body")
	}
	if err := json.Unmarshal(raw, &req); err != nil {
		return nil, apperrors.WrapInvalidInput("invalid request body")
	}
	out := &hybridChatRequest{chatRequest: req, hasContextField: probe.Context != nil}
	// 新プロトコル（context 無し）は keyword_candidates の非 null 配列が必須。
	// missing / null は 400。空配列 [] は有効（全 semantic/keyword フォールバックへ）。
	if !out.hasContextField && (probe.KeywordCandidates == nil || string(probe.KeywordCandidates) == "null") {
		return nil, apperrors.WrapInvalidInput("keyword_candidates is required")
	}
	out.Message = strings.TrimSpace(out.Message)
	if out.Message == "" {
		return nil, apperrors.WrapInvalidInput("message is required")
	}
	if len(out.Message) > chatMessageMaxLength {
		return nil, apperrors.WrapInvalidInput("message is too long")
	}
	if len(out.KeywordCandidates) > chatCandidateMaxItems {
		return nil, apperrors.WrapInvalidInput("keyword_candidates has too many items")
	}
	seen := map[string]struct{}{}
	for _, cnd := range out.KeywordCandidates {
		if !model.IsValidManualCategory(cnd.Category) || len(cnd.Category) > chatCandidateFieldMaxLen ||
			len(cnd.Slug) > chatCandidateFieldMaxLen || !candidateSlugRe.MatchString(cnd.Slug) {
			return nil, apperrors.WrapInvalidInput("invalid keyword candidate")
		}
		key := cnd.Category + "/" + cnd.Slug
		if _, dup := seen[key]; dup {
			return nil, apperrors.WrapInvalidInput("duplicate keyword candidate")
		}
		seen[key] = struct{}{}
	}
	// 質問のみ parse 時点で screen（client title/text は破棄されるので対象外）。
	if err := screenOutboundTexts(out.Message); err != nil {
		return nil, err
	}
	return out, nil
}

// WithManualSearch はハイブリッド検索依存を Handler へ注入する（既存 NewHandler
// シグネチャを維持する focused option）。
func (h *Handler) WithManualSearch(deps *ManualSearchDeps) *Handler {
	h.manualSearch = deps
	return h
}

func (h *Handler) runHybridChat(c *gin.Context, clinicID, staffID uint64) {
	deps := h.manualSearch
	if deps == nil || deps.Corpus == nil || deps.HasPermission == nil {
		httpapi.RespondError(c, apperrors.WrapServiceUnavailable("manual search is unavailable"))
		return
	}
	req, err := parseHybridChatRequest(c)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	var legacyIDs []ManualCandidate
	if req.hasContextField {
		// legacy context: hybrid=true → 409（空配列でも新規約へ移行させる）。
		if deps.Enabled {
			httpapi.RespondError(c, apperrors.WrapConflict("legacy chat request is not accepted in hybrid mode"))
			return
		}
		if len(req.Context) > chatLegacyMaxItems {
			httpapi.RespondError(c, apperrors.WrapInvalidInput("context has too many items"))
			return
		}
		for _, item := range req.Context {
			if !model.IsValidManualCategory(item.Category) || len(item.Category) > chatCandidateFieldMaxLen ||
				len(item.Slug) > chatCandidateFieldMaxLen || !candidateSlugRe.MatchString(item.Slug) {
				httpapi.RespondError(c, apperrors.WrapInvalidInput("invalid context item"))
				return
			}
			legacyIDs = append(legacyIDs, ManualCandidate{Category: item.Category, Slug: item.Slug})
		}
	} else if !revisionHexRe.MatchString(req.ManualRevision) {
		// 新規約は 64hex revision 必須。
		httpapi.RespondError(c, apperrors.WrapInvalidInput("manual_revision is required"))
		return
	}
	ctx := c.Request.Context()
	searchStart := time.Now()
	searchCtx, cancel := context.WithTimeout(ctx, deps.SearchTimeout)
	defer cancel()
	corpus, err := deps.effectiveCorpusFor(searchCtx, c)
	if err != nil {
		httpapi.RespondError(c, apperrors.WrapServiceUnavailable("manual search is unavailable"))
		return
	}
	if !req.hasContextField && req.ManualRevision != corpus.Revision {
		httpapi.RespondError(c, apperrors.WrapConflict("manual revision mismatch"))
		return
	}
	groups := extractSearchTermGroups(req.Message)
	if len(groups) == 0 {
		if searchCtx.Err() != nil {
			httpapi.RespondError(c, apperrors.WrapServiceUnavailable("manual search timed out"))
			return
		}
		c.JSON(http.StatusOK, chatResponse{
			Reply:   chatFixedResponse(detectChatLanguage(req.Message)),
			Sources: []ChatSource{},
			Retrieval: &RetrievalInfo{
				Mode: RetrievalModeNone, Reason: RetrievalReasonNoEvidence, ManualRevision: corpus.Revision,
			},
		})
		return
	}
	// inflight gate は corpus/revision 確定後・chunking/semantic/quota 消費前に取る。
	if !h.acquireChatInflight(clinicID, staffID) {
		httpapi.RespondError(c, apperrors.WrapTooManyRequests("another message is in flight"))
		return
	}
	defer h.releaseChatInflight(clinicID, staffID)

	chunks, err := manualarticle.ChunkCorpusContext(searchCtx, corpus)
	capacity := errors.Is(err, manualarticle.ErrCorpusTooLarge)
	if err != nil && !capacity {
		httpapi.RespondError(c, apperrors.WrapServiceUnavailable("manual search is unavailable"))
		return
	}
	candidateList := req.KeywordCandidates
	if len(legacyIDs) > 0 {
		candidateList = legacyIDs
	}
	candidateIDs := map[string]struct{}{}
	for _, cnd := range candidateList {
		candidateIDs[cnd.Category+"/"+cnd.Slug] = struct{}{}
	}
	articleByID := map[string]manualarticle.CorpusArticle{}
	for _, a := range corpus.Articles {
		articleByID[a.Category+"/"+a.Slug] = a
	}
	var candidateChunks, corpusChunks []manualarticle.Chunk
	var capacityArticles []manualarticle.CorpusArticle
	if !capacity {
		corpusChunks = chunks
		for _, ch := range chunks {
			if _, ok := candidateIDs[ch.ArticleID]; ok {
				candidateChunks = append(candidateChunks, ch)
			}
		}
	} else {
		// Only authorized candidate articles enter the bounded keyword scan.
		for _, cnd := range candidateList {
			a, ok := articleByID[cnd.Category+"/"+cnd.Slug]
			if !ok {
				continue
			}
			capacityArticles = append(capacityArticles, a)
		}
	}
	if searchCtx.Err() != nil {
		httpapi.RespondError(c, apperrors.WrapServiceUnavailable("manual search timed out"))
		return
	}
	haystackOf := func(chID string) (string, string, bool) {
		for _, ch := range candidateChunks {
			if ch.ID == chID {
				a := articleByID[ch.ArticleID]
				n, f := articleSearchHaystack(a.Title, a.Section, ch.Heading+"\n"+ch.Body, a.Keywords, a.KeywordsVI)
				return n, f, true
			}
		}
		return "", "", false
	}
	var keywordRanked []manualarticle.Chunk
	if capacity {
		keywordRanked, err = rankKeywordArticles(searchCtx, capacityArticles, groups, manualRetrievalTopK)
	} else {
		keywordRanked, err = rankKeywordChunks(searchCtx, candidateChunks, haystackOf, groups)
	}
	if err != nil {
		httpapi.RespondError(c, apperrors.WrapServiceUnavailable("manual search timed out"))
		return
	}

	var semantic []manualarticle.Chunk
	var semanticErr error
	reason := RetrievalReasonOK
	switch {
	case capacity:
		reason = RetrievalReasonCapacity
	case !deps.Enabled:
		reason = RetrievalReasonDisabled // semantic は呼ばない
	case !deps.tryAcquireSearch():
		reason = RetrievalReasonEmbeddingUnavailable
	default:
		semantic, semanticErr = deps.semanticCandidates(searchCtx, corpusChunks, req.Message, h.chatBudget, clinicID, staffID)
		deps.releaseSearch()
		if semanticErr != nil {
			var ue *EmbeddingUnavailableError
			if errors.As(semanticErr, &ue) && ue.Reason == "index_missing" {
				reason = RetrievalReasonIndexMissing
			} else {
				reason = RetrievalReasonEmbeddingUnavailable
			}
		}
	}
	mode := RetrievalModeKeyword
	if deps.Enabled && len(semantic) > 0 {
		mode = RetrievalModeHybrid
	}
	selected := mergeRRF(keywordRanked, semantic)
	if len(selected) == 0 {
		selected = h.fallbackChunksFromHistory(searchCtx, clinicID, staffID, corpus, groups)
	}
	logManualRetrieval(ctx, mode, reason, searchStart, deps.EmbeddingModel, len(selected))
	if len(selected) == 0 {
		// 有効な keyword fallback 自体は拒否しないが、証拠も履歴も無い状態で
		// search ctx が切れている場合は no_evidence 200 ではなく 503。
		if searchCtx.Err() != nil {
			httpapi.RespondError(c, apperrors.WrapServiceUnavailable("manual search timed out"))
			return
		}
		sources := weakLinkSources(candidateList, articleByID)
		c.JSON(http.StatusOK, chatResponse{
			Reply:   chatFixedResponse(detectChatLanguage(req.Message)),
			Sources: sources,
			Retrieval: &RetrievalInfo{
				Mode: RetrievalModeNone, Reason: RetrievalReasonNoEvidence, ManualRevision: corpus.Revision,
			},
		})
		return
	}

	contextItems := make([]chatContextItem, 0, len(selected))
	for _, ch := range selected {
		a := articleByID[ch.ArticleID]
		contextItems = append(contextItems, chatContextItem{
			Title: a.Title, Category: a.Category, Slug: a.Slug, Text: ch.Body,
		})
	}
	// 実際に外部へ送る本文（解決済みチャンク）のみ screen — 送信はしない。
	req.Context = contextItems // サーバー解決済みの本文のみ
	screenTexts := make([]string, 0, len(contextItems)*2)
	for _, it := range contextItems {
		screenTexts = append(screenTexts, it.Title, it.Text)
	}
	if err := screenOutboundTexts(screenTexts...); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	hist := h.loadChatHistory(ctx, clinicID, staffID)
	for _, m := range hist {
		if screenOutboundTexts(m.Content) != nil {
			slog.WarnContext(ctx, "support_chat_history_omitted", "reason", "sensitive_persisted_content")
			hist = nil
			break
		}
	}

	if h.chat == nil {
		httpapi.RespondError(c, apperrors.WrapNotImplemented("support chat is not enabled"))
		return
	}
	if err := h.grantGenerationBudget(ctx, clinicID, staffID); err != nil {
		if errors.Is(err, errNotGranted) {
			httpapi.RespondError(c, apperrors.WrapTooManyRequests("generation quota exceeded"))
		} else {
			httpapi.RespondError(c, apperrors.WrapServiceUnavailable("generation quota unavailable"))
		}
		return
	}
	genCtx, genCancel := context.WithTimeout(ctx, chatGenerationTimeout)
	defer genCancel()
	messages, err := buildChatMessagesBounded(&req.chatRequest, hist)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	reply, err := h.chat.Complete(genCtx, messages)
	if err != nil {
		var upstream *upstreamError
		switch {
		case errors.Is(err, context.DeadlineExceeded), errors.Is(genCtx.Err(), context.DeadlineExceeded):
			httpapi.RespondError(c, apperrors.WrapGatewayTimeout("generation deadline exceeded"))
		case errors.As(err, &upstream) && upstream.status == http.StatusTooManyRequests:
			httpapi.RespondError(c, apperrors.WrapTooManyRequests("support chat quota exceeded"))
		case errors.As(err, &upstream):
			httpapi.RespondError(c, apperrors.WrapBadGateway("support chat upstream error"))
		case errors.Is(err, errUpstreamEmpty):
			httpapi.RespondError(c, apperrors.WrapBadGateway("support chat returned empty response"))
		default:
			httpapi.RespondError(c, apperrors.WrapInternalServerError("support chat failed"))
		}
		return
	}
	sources := toChatSources(contextItems)
	if h.service != nil {
		if len(reply) > chatHistoryMaxLength {
			slog.WarnContext(ctx, "support_chat_persist_skipped", "reason", "oversized_reply")
		} else if err := h.service.RecordChatExchange(ctx, clinicID, staffID, req.Message, reply, sources); err != nil {
			slog.WarnContext(ctx, "support_chat_persist_failed", "reason", "write_failed")
		}
	}
	c.JSON(http.StatusOK, chatResponse{
		Reply:   reply,
		Sources: sources,
		Retrieval: &RetrievalInfo{
			Mode: mode, Reason: reason, ManualRevision: corpus.Revision,
		},
	})
}

func weakLinkSources(cands []ManualCandidate, byID map[string]manualarticle.CorpusArticle) []ChatSource {
	out := make([]ChatSource, 0, chatFallbackSourceLimit)
	seen := map[string]struct{}{}
	for _, cnd := range cands {
		if len(out) >= chatFallbackSourceLimit {
			break
		}
		key := cnd.Category + "/" + cnd.Slug
		if _, dup := seen[key]; dup {
			continue
		}
		if a, ok := byID[key]; ok {
			seen[key] = struct{}{}
			out = append(out, ChatSource{Title: a.Title, Category: a.Category, Slug: a.Slug})
		}
	}
	return out
}

func (h *Handler) fallbackChunksFromHistory(ctx context.Context, clinicID, staffID uint64, corpus manualarticle.Corpus, groups [][]string) []manualarticle.Chunk {
	if h.service == nil || len(groups) == 0 {
		return nil
	}
	stored, err := h.service.ListChatHistory(ctx, clinicID, staffID)
	if err != nil || len(stored) == 0 {
		return nil
	}
	byID := map[string]manualarticle.CorpusArticle{}
	for _, a := range corpus.Articles {
		byID[a.Category+"/"+a.Slug] = a
	}
	wanted := map[string]struct{}{}
	// 直近 assistant ターンの sources のみ（最大3・一意）。
	for i := len(stored) - 1; i >= 0; i-- {
		if stored[i].Role != model.SupportChatRoleAssistant {
			continue
		}
		// 最新 assistant の sources のみ使う。sources 無し/不正でも古いターンは辿らない。
		if len(stored[i].Sources) == 0 {
			break
		}
		var ss []ChatSource
		if err := json.Unmarshal(stored[i].Sources, &ss); err != nil {
			break
		}
		for _, s := range ss {
			key := s.Category + "/" + s.Slug
			if len(wanted) < chatFallbackSourceLimit {
				if _, ok := byID[key]; ok {
					wanted[key] = struct{}{}
				}
			}
		}
		break
	}
	if len(wanted) == 0 {
		return nil
	}
	var poolArticles []manualarticle.CorpusArticle
	for _, a := range corpus.Articles {
		if _, ok := wanted[a.Category+"/"+a.Slug]; ok {
			poolArticles = append(poolArticles, a)
		}
	}
	pool, perr := rankKeywordArticles(ctx, poolArticles, groups, manualRetrievalTopK)
	if perr != nil {
		return nil
	}
	return mergeRRF(pool, nil) // 最終キャップ（3件・2/article・bytes）を適用
}

// buildChatMessagesBounded は buildChatMessages の結果を 16KiB UTF-8 内に収める。
// system と最新 user は削らず古い履歴から削除する。必須2件だけで超過なら error。
func buildChatMessagesBounded(req *chatRequest, history []ChatMessage) ([]ChatMessage, error) {
	full := buildChatMessages(req, history)
	total := 0
	for _, m := range full {
		total += len(m.Content)
	}
	if total <= chatPromptMaxBytes {
		return full, nil
	}
	mandatory := len(full[0].Content) + len(full[len(full)-1].Content)
	if mandatory > chatPromptMaxBytes {
		return nil, apperrors.WrapInvalidInput("prompt budget exceeded")
	}
	keep := total
	for i := 1; i < len(full)-1 && keep > chatPromptMaxBytes; i++ {
		keep -= len(full[i].Content)
		full[i].Content = ""
	}
	out := []ChatMessage{full[0]}
	for _, m := range full[1 : len(full)-1] {
		if m.Content != "" {
			out = append(out, m)
		}
	}
	return append(out, full[len(full)-1]), nil
}

// logManualRetrieval は retrieval 結果の分類ログ（本文/キー/ハッシュ/エラー詳細は出さない）。
func logManualRetrieval(ctx context.Context, mode, reason string, start time.Time, modelName string, chunkCount int) {
	attrs := []any{
		"mode", mode,
		"reason", reason,
		"elapsed_ms", time.Since(start).Milliseconds(),
		"model", modelName,
		"chunk_count", chunkCount,
	}
	switch reason {
	case RetrievalReasonIndexMissing, RetrievalReasonEmbeddingUnavailable, RetrievalReasonCapacity:
		slog.WarnContext(ctx, "support_manual_retrieval", attrs...)
	default:
		slog.InfoContext(ctx, "support_manual_retrieval", attrs...)
	}
}
