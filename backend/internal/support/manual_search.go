package support

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/manualarticle"
)

// manual_search.go — R5 ハイブリッド検索（keyword 候補 + semantic embedding の
// RRF 統合）。corpus snapshot・chunk・embedding キャッシュ・quota は
// composition root から注入され、欠落は fail-closed（503/縮退）する。

const (
	manualRetrievalTopK            = 20
	manualRRFK                     = 60
	manualSemanticMinCosine        = 0.70
	manualFinalMaxChunks           = 3
	manualFinalMaxChunksPerArticle = 2
	manualChunkBodyMaxBytes        = 1800
	manualTotalBodyMaxBytes        = 6000
	manualFallbackMaxSources       = 3
	// 検索並列上限（プロセスごと 4 件）。超過は即 keyword-only。
	manualSearchMaxInflight = 4
)

// Retrieval mode / reason の公開 enum（ticket 固定値）。
const (
	RetrievalModeHybrid  = "hybrid"
	RetrievalModeKeyword = "keyword"
	RetrievalModeNone    = "none"

	RetrievalReasonOK                   = "ok"
	RetrievalReasonDisabled             = "disabled"
	RetrievalReasonIndexMissing         = "index_missing"
	RetrievalReasonEmbeddingUnavailable = "embedding_unavailable"
	RetrievalReasonCapacity             = "capacity"
	RetrievalReasonNoEvidence           = "no_evidence"
)

// ManualCorpusSource は権限に応じた実効コーパスを供給する。
// permitted=true なら最新 DB 行を反映した snapshot、false なら base のみ。
// DB 取得失敗は error（呼出側は 503 で fail-closed、古い base で回答しない）。
type ManualCorpusSource interface {
	EffectiveCorpus(ctx context.Context, permitted bool) (manualarticle.Corpus, error)
}

// ManualSearchDeps は Chat のハイブリッド検索に必要な注入依存。
type ManualSearchDeps struct {
	// Enabled は semantic 検索を制御する（false でも server-authoritative keyword 検索）。
	Enabled bool
	// CorpusSource は権限考慮の実効コーパス供給（必須。nil は effectiveCorpusFor
	// で依存エラーにし、呼出側が 503 にマッピングする）。
	Corpus ManualCorpusSource
	// HasPermission は c.auth.Handler.CheckPermission(c,"manual-edit","view") と
	// 同じ closure（三値: true/false/error）。nil は依存欠落（503）。
	HasPermission func(c *gin.Context) (bool, error)
	// Embedder は nil 可（= embedding_unavailable で keyword 縮退）。
	Embedder ManualEmbedder
	// Cache は派生ベクトルの Load/Store。nil 可（意味検索は index_missing 縮退）。
	Cache ManualEmbeddingCache
	// EmbeddingModel は cache_key / 検証に使うモデル名。
	EmbeddingModel string
	// SearchTimeout は検索段階全体の budget（既定 2s、embedding 1.2s を内包）。
	SearchTimeout time.Duration
	// EmbedTimeout は 1 embedding 呼出の budget（既定 1.2s）。
	EmbedTimeout time.Duration
	// searchSem はプロセス内の同時検索数制限（非公開。NewManualSearchDeps で初期化）。
	searchSem chan struct{}
}

// NewManualSearchDeps は ManualSearchDeps を組み立てる。
func NewManualSearchDeps(enabled bool, corpus ManualCorpusSource, hasPermission func(*gin.Context) (bool, error),
	embedder ManualEmbedder, cache ManualEmbeddingCache, embeddingModel string) *ManualSearchDeps {
	return &ManualSearchDeps{
		Enabled: enabled, Corpus: corpus, HasPermission: hasPermission,
		Embedder: embedder, Cache: cache, EmbeddingModel: embeddingModel,
		SearchTimeout: 2 * time.Second, EmbedTimeout: 1200 * time.Millisecond,
		searchSem: make(chan struct{}, manualSearchMaxInflight),
	}
}

// ManualCandidate は FE が送る keyword 候補 ID（本文は送らない・信用しない）。
type ManualCandidate struct {
	Category string `json:"category"`
	Slug     string `json:"slug"`
}

// RetrievalInfo は response.retrieval の構造（非機密の監査分類のみ）。
type RetrievalInfo struct {
	Mode           string `json:"mode"`
	Reason         string `json:"reason"`
	ManualRevision string `json:"manual_revision"`
}

// rankedChunk は RRF 統合用の内部表現。
type rankedChunk struct {
	chunk       manualarticle.Chunk
	score       float64
	keywordRank int // 未採用は MaxInt
	chunkID     string
}

// effectiveCorpusFor は 1 要求で 1 回だけコーパスを解決する。
func (d *ManualSearchDeps) effectiveCorpusFor(ctx context.Context, c *gin.Context) (manualarticle.Corpus, error) {
	if d.Corpus == nil || d.HasPermission == nil || c == nil || c.Request == nil {
		return manualarticle.Corpus{}, errManualSearchDependency
	}
	// R2: permission 評価は search ctx を通す。元の Gin Request/Context は変更せず、
	// copy に scoped Request のみ差し替えて評価する。error は denial とは区別し
	// 呼出側で 503（corpus/external/quota/history の何も消費しない）。
	gc := c.Copy()
	gc.Request = c.Request.WithContext(ctx)
	permitted, perr := d.HasPermission(gc)
	if perr != nil {
		return manualarticle.Corpus{}, fmt.Errorf("%w: %w", errManualSearchDependency, perr)
	}
	return d.Corpus.EffectiveCorpus(ctx, permitted)
}

var errManualSearchDependency = errors.New("manual search dependency unavailable")

// rankKeywordChunks は一致意味語グループ数降順・同点 ordinal で上位20件を返す。
func rankKeywordChunks(ctx context.Context, chunks []manualarticle.Chunk, articleOf func(chunkID string) (normHaystack, foldHaystack string, ok bool), groups [][]string) ([]manualarticle.Chunk, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(groups) == 0 {
		return nil, nil
	}
	type scored struct {
		chunk   manualarticle.Chunk
		matched int
	}
	var list []scored
	for _, ch := range chunks {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		normHaystack, foldHaystack, ok := articleOf(ch.ID)
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		matched := 0
		for _, g := range groups {
			if groupHasTextEvidence(normHaystack, foldHaystack, g) {
				matched++
			}
		}
		// 記事 admission と同じ閾値: min(2, グループ数) 一致が必要。
		if matched >= min(2, len(groups)) {
			list = append(list, scored{chunk: ch, matched: matched})
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].matched != list[j].matched {
			return list[i].matched > list[j].matched
		}
		return list[i].chunk.Ordinal < list[j].chunk.Ordinal
	})
	out := make([]manualarticle.Chunk, 0, min(manualRetrievalTopK, len(list)))
	for i := 0; i < len(list) && i < manualRetrievalTopK; i++ {
		out = append(out, list[i].chunk)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// mergeRRF は keyword / semantic 両リストを RRF(score=sum(1/(60+rank)), 重み1)
// で統合し、上位3・1記事2chunks・同点は keyword 順位→chunkID で確定する。
func mergeRRF(keyword, semantic []manualarticle.Chunk) []manualarticle.Chunk {
	agg := map[string]*rankedChunk{}
	var order []string
	add := func(list []manualarticle.Chunk) {
		for rank, ch := range list {
			rc, ok := agg[ch.ID]
			if !ok {
				rc = &rankedChunk{chunk: ch, keywordRank: 1 << 30, chunkID: ch.ID}
				agg[ch.ID] = rc
				order = append(order, ch.ID)
			}
			rc.score += 1.0 / float64(manualRRFK+rank+1)
		}
	}
	add(keyword)
	add(semantic)
	for i, ch := range keyword {
		if rc, ok := agg[ch.ID]; ok && rc.keywordRank == 1<<30 {
			rc.keywordRank = i
		}
	}
	list := make([]*rankedChunk, 0, len(agg))
	for _, id := range order {
		list = append(list, agg[id])
	}
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].score != list[j].score {
			return list[i].score > list[j].score
		}
		if list[i].keywordRank != list[j].keywordRank {
			return list[i].keywordRank < list[j].keywordRank
		}
		return list[i].chunkID < list[j].chunkID
	})
	// 上限: 3 chunks、1記事あたり2chunks、各本文<=1800bytes、合計<=6000bytes。
	var out []manualarticle.Chunk
	perArticle := map[string]int{}
	totalBytes := 0
	for _, rc := range list {
		if len(out) >= manualFinalMaxChunks {
			break
		}
		if perArticle[rc.chunk.ArticleID] >= manualFinalMaxChunksPerArticle {
			continue
		}
		body := utf8SafeBytes(rc.chunk.Body, manualChunkBodyMaxBytes)
		if totalBytes+len(body) > manualTotalBodyMaxBytes {
			continue
		}
		c := rc.chunk
		c.Body = body
		out = append(out, c)
		perArticle[c.ArticleID]++
		totalBytes += len(body)
	}
	return out
}

func utf8SafeBytes(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	for maxBytes > 0 && !utf8.ValidString(s[:maxBytes]) {
		maxBytes--
	}
	return s[:maxBytes]
}

// semanticCandidates は質問の embedding を生成し cos>=0.70 の上位20を返す。
// 失敗は EmbeddingUnavailableError に包み、呼出側は keyword-only 縮退する。
// quota は外部送信前に user+global を原子 grant し、cache 欠損時は HTTP0 で戻る。
func (d *ManualSearchDeps) semanticCandidates(ctx context.Context, chunks []manualarticle.Chunk, message string, budget LLMBudgetGranter, clinicID, staffID uint64) ([]manualarticle.Chunk, error) {
	if d.Embedder == nil || d.Cache == nil || d.EmbeddingModel == "" {
		return nil, &EmbeddingUnavailableError{Reason: "index_missing"}
	}
	keys := make([]string, 0, len(chunks))
	keyOf := map[string]string{}
	for _, ch := range chunks {
		k := manualarticle.ChunkCacheKey(d.EmbeddingModel, manualarticle.TaskTypeRetrievalDocument, ch.EmbeddingInput)
		keys = append(keys, k)
		keyOf[ch.ID] = k
	}
	vectors, err := d.Cache.Load(ctx, keys)
	if err != nil {
		return nil, &EmbeddingUnavailableError{Reason: "cache_load", Err: err}
	}
	if len(vectors) == 0 {
		return nil, &EmbeddingUnavailableError{Reason: "index_missing"}
	}
	// EMR-438: 部分的なインデックス（override 追加や support-index 失敗で一部 chunk のみ
	// vector が存在する状態）を usable と扱わない。欠損 chunk を黙ってスキップすると
	// セマンティック検索が新規/更新記事を取りこぼしたまま hybrid 成功を装うため、
	// 要求キー全件を補完できなければ明示的に keyword-only へ退避する。
	for _, k := range keys {
		if _, ok := vectors[k]; !ok {
			return nil, &EmbeddingUnavailableError{Reason: "index_missing"}
		}
	}
	// 索引 hit があって初めて query embedding の quota を消費する。
	if budget == nil {
		return nil, &EmbeddingUnavailableError{Reason: "quota_unavailable"}
	}
	ok, err := budget.GrantAll(ctx, []BudgetGrant{
		{Scope: LLMUserBudgetScope(clinicID, staffID), Model: d.EmbeddingModel, Interval: LLMBudgetUserInterval},
		{Scope: LLMBudgetScopeGlobal, Model: d.EmbeddingModel, Interval: LLMBudgetGlobalInterval},
	})
	if err != nil || !ok {
		return nil, &EmbeddingUnavailableError{Reason: "quota_unavailable", Err: err}
	}
	embedCtx, cancel := context.WithTimeout(ctx, d.EmbedTimeout)
	defer cancel()
	qvec, err := d.Embedder.Embed(embedCtx, manualarticle.TaskTypeRetrievalQuery, message)
	if err != nil {
		return nil, err
	}
	type hit struct {
		ch    manualarticle.Chunk
		score float64
	}
	var hits []hit
	for _, ch := range chunks {
		vec, ok := vectors[keyOf[ch.ID]]
		if !ok {
			continue
		}
		s := CosineSimilarity(qvec, vec)
		if s >= manualSemanticMinCosine {
			hits = append(hits, hit{ch: ch, score: s})
		}
	}
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].score != hits[j].score {
			return hits[i].score > hits[j].score
		}
		return hits[i].ch.ID < hits[j].ch.ID
	})
	out := make([]manualarticle.Chunk, 0, min(manualRetrievalTopK, len(hits)))
	for i := 0; i < len(hits) && i < manualRetrievalTopK; i++ {
		out = append(out, hits[i].ch)
	}
	return out, nil
}

func (d *ManualSearchDeps) tryAcquireSearch() bool {
	select {
	case d.searchSem <- struct{}{}:
		return true
	default:
		return false
	}
}
func (d *ManualSearchDeps) releaseSearch() { <-d.searchSem }
