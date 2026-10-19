package support

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/manualarticle"
	"github.com/animal-ekarte/backend/internal/model"
)

type indexEmbedder struct {
	calls  int32
	fail   *EmbeddingUnavailableError
	malvec bool
	inputs []string
	tasks  []string
}

func (e *indexEmbedder) Embed(ctx context.Context, taskType, input string) ([]float64, error) {
	atomic.AddInt32(&e.calls, 1)
	e.inputs = append(e.inputs, input)
	e.tasks = append(e.tasks, taskType)
	if e.fail != nil {
		return nil, e.fail
	}
	v := make([]float64, manualarticle.EmbeddingDimensions)
	if e.malvec {
		return v, nil // zero norm → reject
	}
	v[0] = 1
	return v, nil
}

type indexCache struct {
	stored    map[string][]float64
	keys      []string
	loadCalls int
}

func newIndexCache() *indexCache { return &indexCache{stored: map[string][]float64{}} }

func (c *indexCache) Load(_ context.Context, keys []string) (map[string][]float64, error) {
	c.loadCalls++
	out := map[string][]float64{}
	for _, k := range keys {
		if v, ok := c.stored[k]; ok {
			out[k] = v
		}
	}
	return out, nil
}

func (c *indexCache) StoreWithKey(_ context.Context, key, _, _ string, _, _ int, vec []float64) error {
	c.keys = append(c.keys, key)
	c.stored[key] = vec
	return nil
}

type indexBudget struct {
	calls   int32
	granted []bool
	err     error
}

func (b *indexBudget) GrantAll(_ context.Context, _ []BudgetGrant) (bool, error) {
	i := int(atomic.AddInt32(&b.calls, 1)) - 1
	if b.err != nil {
		return false, b.err
	}
	if i < len(b.granted) {
		return b.granted[i], nil
	}
	return true, nil
}

func indexDeps(e *indexEmbedder, c *indexCache, b *indexBudget) (IndexDeps, *[]time.Duration) {
	waits := &[]time.Duration{}
	return IndexDeps{
		Embedder: e, Cache: c, Budget: b, Model: "sarashina3-embedding",
		Wait: func(ctx context.Context, d time.Duration) error {
			*waits = append(*waits, d)
			return ctx.Err()
		},
	}, waits
}

var indexBase = []manualarticle.CorpusArticle{
	{Category: "screens", Slug: "a1", Title: "A1", Section: "S", Body: "本文一", Hash: "h1"},
	{Category: "screens", Slug: "a2", Title: "A2", Section: "S", Body: "本文二", Hash: "h2"},
}

func TestEMR410Review_IndexCancelledBeforeCache(t *testing.T) {
	e, c, b := &indexEmbedder{}, newIndexCache(), &indexBudget{}
	deps, _ := indexDeps(e, c, b)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	res, err := RunIndex(ctx, indexBase, nil, deps)

	require.ErrorIs(t, err, context.Canceled)
	assert.Zero(t, res.Chunks)
	assert.Zero(t, c.loadCalls)
	assert.Zero(t, atomic.LoadInt32(&e.calls))
	assert.Zero(t, atomic.LoadInt32(&b.calls))
	assert.Empty(t, c.keys)
	assert.Empty(t, c.stored)
}

func TestIndexStoresMissingKeys(t *testing.T) {
	e, c, b := &indexEmbedder{}, newIndexCache(), &indexBudget{}
	deps, _ := indexDeps(e, c, b)
	res, err := RunIndex(context.Background(), indexBase, nil, deps)
	require.NoError(t, err)
	require.Greater(t, res.Chunks, 0)
	assert.Equal(t, res.Chunks, res.Stored)
	assert.Equal(t, int32(res.Stored), atomic.LoadInt32(&e.calls))
	assert.Equal(t, int32(res.Stored), atomic.LoadInt32(&b.calls))
}

func TestIndexUsesRetrievalCacheKeyAndEmbeddingInput(t *testing.T) {
	e, c, b := &indexEmbedder{}, newIndexCache(), &indexBudget{}
	deps, _ := indexDeps(e, c, b)
	chunks, err := manualarticle.ChunkCorpus(manualarticle.BuildCorpus(indexBase[:1], nil))
	require.NoError(t, err)
	require.Len(t, chunks, 1)
	wantKey := manualarticle.ChunkCacheKey(deps.Model, manualarticle.TaskTypeRetrievalDocument, chunks[0].EmbeddingInput)

	_, err = RunIndex(context.Background(), indexBase[:1], nil, deps)
	require.NoError(t, err)
	require.Contains(t, c.stored, wantKey, "index and online retrieval must address the same vector")
	assert.Equal(t, []string{chunks[0].EmbeddingInput}, e.inputs)
	assert.Equal(t, []string{manualarticle.TaskTypeRetrievalDocument}, e.tasks)
}

func TestIndexSkipExistingSecondRun(t *testing.T) {
	e, c, b := &indexEmbedder{}, newIndexCache(), &indexBudget{}
	deps, _ := indexDeps(e, c, b)
	_, err := RunIndex(context.Background(), indexBase, nil, deps)
	require.NoError(t, err)
	res2, err := RunIndex(context.Background(), indexBase, nil, deps)
	require.NoError(t, err)
	assert.Equal(t, 0, res2.Stored)
	assert.Equal(t, res2.Chunks, res2.Existing)
}

func TestIndexUnionBaseAndEffectiveKeys(t *testing.T) {
	e, c, b := &indexEmbedder{}, newIndexCache(), &indexBudget{}
	deps, _ := indexDeps(e, c, b)
	ov := []model.ManualArticle{{Category: "screens", Slug: "a1", Title: "A1x", Section: "S", BodyMarkdown: "別本文"}}
	res, err := RunIndex(context.Background(), indexBase, ov, deps)
	require.NoError(t, err)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, res.Stored, res.Chunks)
	assert.True(t, res.Chunks >= 2)
}

func TestIndexBudgetDeniedWaitsAndRegrants(t *testing.T) {
	e, c := &indexEmbedder{}, newIndexCache()
	b := &indexBudget{granted: []bool{false, true}}
	deps, waits := indexDeps(e, c, b)
	_, err := RunIndex(context.Background(), indexBase[:1], nil, deps)
	require.NoError(t, err)
	assert.Contains(t, *waits, 5*time.Second)
	assert.True(t, atomic.LoadInt32(&b.calls) >= 2)
}

func TestIndexBudgetDBErrorNoHTTP(t *testing.T) {
	e, c := &indexEmbedder{}, newIndexCache()
	b := &indexBudget{err: fmt.Errorf("db down")}
	deps, _ := indexDeps(e, c, b)
	_, err := RunIndex(context.Background(), indexBase[:1], nil, deps)
	require.Error(t, err)
	assert.Equal(t, int32(0), atomic.LoadInt32(&e.calls))
}

func TestIndex400NoRetryNoWait(t *testing.T) {
	e, c := &indexEmbedder{fail: &EmbeddingUnavailableError{Status: http.StatusBadRequest, Reason: "status"}}, newIndexCache()
	b := &indexBudget{}
	deps, waits := indexDeps(e, c, b)
	_, err := RunIndex(context.Background(), indexBase[:1], nil, deps)
	require.Error(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&e.calls))
	assert.Empty(t, *waits)
}

func TestIndex500RetriesTwice(t *testing.T) {
	e, c := &indexEmbedder{fail: &EmbeddingUnavailableError{Status: http.StatusInternalServerError, Reason: "status"}}, newIndexCache()
	b := &indexBudget{}
	deps, waits := indexDeps(e, c, b)
	_, err := RunIndex(context.Background(), indexBase[:1], nil, deps)
	require.Error(t, err)
	assert.Equal(t, int32(3), atomic.LoadInt32(&e.calls))
	assert.Equal(t, []time.Duration{2 * time.Second, 5 * time.Second}, *waits)
}

func TestIndex429Waits60Once(t *testing.T) {
	e, c := &indexEmbedder{fail: &EmbeddingUnavailableError{Status: http.StatusTooManyRequests, Reason: "status"}}, newIndexCache()
	b := &indexBudget{}
	deps, waits := indexDeps(e, c, b)
	_, err := RunIndex(context.Background(), indexBase[:1], nil, deps)
	require.Error(t, err)
	assert.Equal(t, []time.Duration{60 * time.Second}, *waits)
	assert.Equal(t, int32(2), atomic.LoadInt32(&e.calls))
}

func TestIndexMalformedVectorNotStored(t *testing.T) {
	e, c := &indexEmbedder{malvec: true}, newIndexCache()
	deps, _ := indexDeps(e, c, &indexBudget{})
	_, err := RunIndex(context.Background(), indexBase[:1], nil, deps)
	require.Error(t, err)
	assert.Empty(t, c.stored)
}

func TestIndexCorpusOverflowNoHTTP(t *testing.T) {
	e, c := &indexEmbedder{}, newIndexCache()
	deps, _ := indexDeps(e, c, &indexBudget{})
	big := make([]manualarticle.CorpusArticle, 0, 2050)
	for i := 0; i < 2050; i++ {
		big = append(big, manualarticle.CorpusArticle{Category: "screens", Slug: fmt.Sprintf("a%d", i), Title: "T", Section: "S", Body: "x", Hash: fmt.Sprintf("h%d", i)})
	}
	_, err := RunIndex(context.Background(), big, nil, deps)
	require.Error(t, err)
	assert.Equal(t, int32(0), atomic.LoadInt32(&e.calls))
}

func TestIndexCancelWaitStops(t *testing.T) {
	e, c := &indexEmbedder{fail: &EmbeddingUnavailableError{Reason: "transport"}}, newIndexCache()
	b := &indexBudget{}
	ctx, cancel := context.WithCancel(context.Background())
	deps := IndexDeps{
		Embedder: e, Cache: c, Budget: b, Model: "m",
		Wait: func(ctx context.Context, d time.Duration) error { cancel(); return ctx.Err() },
	}
	_, err := RunIndex(ctx, indexBase[:1], nil, deps)
	require.Error(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&e.calls))
}

// EMR-441: 429 待機は一時的リトライ枠を消費しない。429→500→500→500 でも
// 2s/5s の両リトライを実施してから失敗する（旧実装は 429 が枠を消費し 5s のみ）。
func TestIndex429DoesNotConsumeTransientRetries(t *testing.T) {
	failures := []*EmbeddingUnavailableError{
		{Status: http.StatusTooManyRequests, Reason: "status"},
		{Status: http.StatusInternalServerError, Reason: "status"},
		{Status: http.StatusInternalServerError, Reason: "status"},
		{Status: http.StatusInternalServerError, Reason: "status"},
	}
	var calls int32
	e := embedderFunc(func(ctx context.Context, taskType, input string) ([]float64, error) {
		i := int(atomic.AddInt32(&calls, 1)) - 1
		if i < len(failures) {
			return nil, failures[i]
		}
		v := make([]float64, manualarticle.EmbeddingDimensions)
		v[0] = 1
		return v, nil
	})
	waits := &[]time.Duration{}
	deps := IndexDeps{
		Embedder: e, Cache: newIndexCache(), Budget: &indexBudget{}, Model: "sarashina3-embedding",
		Wait: func(ctx context.Context, d time.Duration) error {
			*waits = append(*waits, d)
			return ctx.Err()
		},
	}
	_, err := RunIndex(context.Background(), indexBase[:1], nil, deps)
	require.Error(t, err)
	assert.Equal(t, []time.Duration{60 * time.Second, 2 * time.Second, 5 * time.Second}, *waits)
	assert.Equal(t, int32(4), atomic.LoadInt32(&calls))
}

type embedderFunc func(ctx context.Context, taskType, input string) ([]float64, error)

func (f embedderFunc) Embed(ctx context.Context, taskType, input string) ([]float64, error) {
	return f(ctx, taskType, input)
}
