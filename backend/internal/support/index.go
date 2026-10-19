// index.go — support-index CLI のドメインコア。
// base コーパス + 最新 DB 上書きコーパスの embedding cache_key を全収集し、
// 既存行を除いた不足分だけ逐次 Embed→Store する。
// 外部呼出前に global(1250ms)+index(5s) の両 scope を GrantAll で取得する。
package support

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/manualarticle"
	"github.com/animal-ekarte/backend/internal/model"
)

// IndexDeps は RunIndex の注入可能な依存。
type IndexDeps struct {
	Embedder ManualEmbedder
	Cache    ManualEmbeddingCache
	Budget   LLMBudgetGranter
	Model    string
	// Wait は cancellable 待機（テストでは記録のみ実時間を消費しない）。
	Wait func(ctx context.Context, d time.Duration) error
}

// IndexResult は索引実行の集計。非ゼロ終了時も部分カウントを保持する。
type IndexResult struct {
	Chunks, Existing, Stored int
}

const indexLoadBatch = 2048

// WaitTimer は ctx 付きの実時間待機（Wait 未指定時の既定）。
func WaitTimer(ctx context.Context, d time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// RunIndex は不足 cache_key を逐次 embedding して保存する。
// 全てのエラーは構造化分類で返す（本文/キー/DB エラーの詳細は含めない）。
func RunIndex(ctx context.Context, base []manualarticle.CorpusArticle, overrides []model.ManualArticle, deps IndexDeps) (IndexResult, error) {
	var res IndexResult
	if deps.Embedder == nil || deps.Cache == nil || deps.Budget == nil || deps.Model == "" {
		return res, apperrors.WrapInvalidInput("index deps incomplete")
	}
	wait := deps.Wait
	if wait == nil {
		wait = WaitTimer
	}
	// base（無権限利用者向け）+ effective の両方を索引する。容量超過は
	// 最初の API 呼出前に検出して即失敗。
	corpus := manualarticle.BuildCorpus(base, overrides)
	need := make(map[string]manualarticle.Chunk)
	var order []string
	for _, c := range []manualarticle.Corpus{manualarticle.BuildCorpus(base, nil), corpus} {
		chunks, err := manualarticle.ChunkCorpusContext(ctx, c)
		if err != nil {
			return res, err
		}
		for _, ch := range chunks {
			key := manualarticle.ChunkCacheKey(deps.Model, manualarticle.TaskTypeRetrievalDocument, ch.EmbeddingInput)
			if _, dup := need[key]; !dup {
				need[key] = ch
				order = append(order, key)
			}
		}
	}
	res.Chunks = len(order)
	// 既存行を一度だけ <=2048 バッチで読む。
	existing := map[string]struct{}{}
	for i := 0; i < len(order); i += indexLoadBatch {
		end := i + indexLoadBatch
		if end > len(order) {
			end = len(order)
		}
		rows, err := deps.Cache.Load(ctx, order[i:end])
		if err != nil {
			return res, apperrors.WrapServiceUnavailable("manual index dependency unavailable")
		}
		for k := range rows {
			existing[k] = struct{}{}
		}
	}
	res.Existing = len(existing)
	for _, key := range order {
		if _, ok := existing[key]; ok {
			continue
		}
		if err := ctx.Err(); err != nil {
			return res, err
		}
		ch := need[key]
		vec, err := indexEmbedWithRetry(ctx, ch.EmbeddingInput, deps, wait)
		if err != nil {
			return res, err
		}
		if err := deps.Cache.StoreWithKey(ctx, key, deps.Model, manualarticle.TaskTypeRetrievalDocument, manualarticle.ChunkerVersion, manualarticle.EmbeddingDimensions, vec); err != nil {
			return res, apperrors.WrapServiceUnavailable("manual index dependency unavailable")
		}
		res.Stored++
	}
	return res, nil
}

// indexEmbedWithRetry は 1 chunk の embed を取得する。全試行前に budget grant。
// transient(transport/5xx): 2s → 5s の最大2 retry。429: 60s wait を1回だけ。
// 4xx その他・レスポンス異常: retry なし。retry 毎に fresh grant。
func indexEmbedWithRetry(ctx context.Context, input string, deps IndexDeps, wait func(context.Context, time.Duration) error) ([]float64, error) {
	retries := []time.Duration{2 * time.Second, 5 * time.Second}
	rateLimited := false
	// EMR-441: 429 待機は一時的リトライ枠を消費しない。共有カウンタだと
	// 429→500→500 で 2s 待機が飛ばされ 1 回のリトライ後に中断し、部分的な
	// インデックスで keyword-only 退避してしまう。
	transient := 0
	for attempt := 0; ; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		// 全 HTTP attempt 前に global+index grant を取得。denied は 5s 待って再取得
		// （attempt を消費しない）。grant DB エラーは即失敗（外部呼出しない）。
		for {
			granted, gerr := deps.Budget.GrantAll(ctx, []BudgetGrant{
				{Scope: LLMBudgetScopeGlobal, Model: deps.Model, Interval: LLMBudgetGlobalInterval},
				{Scope: LLMBudgetScopeIndex, Model: deps.Model, Interval: LLMBudgetIndexInterval},
			})
			if gerr != nil {
				return nil, apperrors.WrapServiceUnavailable("manual index budget unavailable")
			}
			if granted {
				break
			}
			if werr := wait(ctx, 5*time.Second); werr != nil {
				return nil, werr
			}
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		vec, err := deps.Embedder.Embed(ctx, manualarticle.TaskTypeRetrievalDocument, input)
		if err == nil {
			norm, nerr := normalizeEmbedding(vec)
			if nerr != nil {
				return nil, nerr
			}
			return norm, nil
		}
		var eerr *EmbeddingUnavailableError
		if !errors.As(err, &eerr) {
			return nil, apperrors.WrapServiceUnavailable("manual index dependency unavailable")
		}
		switch {
		case eerr.Status == http.StatusTooManyRequests && !rateLimited:
			rateLimited = true
			if werr := wait(ctx, 60*time.Second); werr != nil {
				return nil, werr
			}
		case (eerr.Reason == "transport" || eerr.Status >= 500) && transient < len(retries):
			if werr := wait(ctx, retries[transient]); werr != nil {
				return nil, werr
			}
			transient++
		default:
			slog.WarnContext(ctx, "support_index_embed_failed",
				"status", eerr.Status, "reason", eerr.Reason, "attempt", attempt)
			return nil, err
		}
	}
}
