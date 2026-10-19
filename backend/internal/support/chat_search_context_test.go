package support

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/manualarticle"
	"github.com/animal-ekarte/backend/internal/model"
)

// countingCorpusSource は EffectiveCorpus 呼出回数を数える。
type countingCorpusSource struct {
	corpus    manualarticle.Corpus
	calls     int
	permitted bool
}

func (f *countingCorpusSource) EffectiveCorpus(_ context.Context, permitted bool) (manualarticle.Corpus, error) {
	f.calls++
	f.permitted = permitted
	return f.corpus, nil
}

// TestHybridPermissionLookupError503NoCalls は R2: permission lookup error は
// denial ではなく 503。corpus/external/quota/history/persist は何も消費しない。
func TestHybridPermissionLookupError503NoCalls(t *testing.T) {
	sentinel := errors.New("permission repository unavailable")
	src := &countingCorpusSource{corpus: testCorpus()}
	emb := &fakeEmbedder{}
	chat := &mockChat{reply: "x"}
	b := &fakeBudget{granted: true}
	historyCalls, persistCalls := 0, 0
	svc := &mockService{
		listChatFn: func(context.Context, uint64, uint64) ([]model.SupportChatMessage, error) {
			historyCalls++
			return nil, nil
		},
		recordChatFn: func(context.Context, uint64, uint64, string, string, []ChatSource) error {
			persistCalls++
			return nil
		},
	}
	cache := &fakeCache{vecs: map[string][]float64{}}
	deps := NewManualSearchDeps(true, src, func(*gin.Context) (bool, error) { return false, sentinel }, emb, cache, "m")
	h := NewHandler(svc, nil, nil, chat, nil).WithManualSearch(deps).WithLLMBudget(b, "m")
	rev := testCorpus().Revision
	body := `{"message":"会計の締め方は？","manual_revision":"` + rev + `","keyword_candidates":[{"category":"screens","slug":"accounting"}]}`
	c, rec := newRequest(t, http.MethodPost, "/api/v1/support/chat", bytes.NewBufferString(body), "application/json")
	h.Chat(c)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.NotContains(t, rec.Body.String(), sentinel.Error())
	assert.Equal(t, 0, src.calls, "corpus must not be resolved on permission error")
	assert.Equal(t, 0, cache.loadCalls)
	assert.Equal(t, 0, cache.storeCalls)
	assert.Equal(t, 0, historyCalls)
	assert.Equal(t, 0, persistCalls)
	assert.Equal(t, 0, emb.calls)
	assert.Equal(t, 0, b.calls)
	assert.Equal(t, 0, chat.calls)
}

// TestHybridPermissionDenyUsesBase は対照: 正規の denial (false,nil) は base corpus で継続する。
func TestHybridPermissionDenyUsesBase(t *testing.T) {
	src := &countingCorpusSource{corpus: testCorpus()}
	deps := NewManualSearchDeps(true, src, func(*gin.Context) (bool, error) { return false, nil }, &fakeEmbedder{}, &fakeCache{}, "m")
	h := NewHandler(nil, nil, nil, &mockChat{reply: "ok"}, nil).WithManualSearch(deps).WithLLMBudget(&fakeBudget{granted: true}, "m")
	rev := testCorpus().Revision
	body := `{"message":"会計の締め方は？","manual_revision":"` + rev + `","keyword_candidates":[]}`
	c, rec := newRequest(t, http.MethodPost, "/api/v1/support/chat", bytes.NewBufferString(body), "application/json")
	h.Chat(c)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, src.calls)
	assert.False(t, src.permitted)
}

// TestManualPermissionUsesSearchContext は permission 評価が元 Request を汚染せず、
// scoped ctx を経由して cancel を検知できることを証明する。
func TestManualPermissionUsesSearchContext(t *testing.T) {
	var seenCtx context.Context
	var seenCopy *gin.Context
	src := &countingCorpusSource{corpus: testCorpus()}
	checker := func(gc *gin.Context) (bool, error) {
		seenCopy = gc
		seenCtx = gc.Request.Context()
		return false, gc.Request.Context().Err()
	}
	deps := NewManualSearchDeps(true, src, checker, nil, nil, "m")

	c, _ := newRequest(t, http.MethodPost, "/", bytes.NewBufferString("{}"), "application/json")
	origReq := c.Request
	origCtx := c.Request.Context()
	scoped, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := deps.effectiveCorpusFor(scoped, c)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.NotSame(t, c, seenCopy, "checker must receive a copied Gin context")
	assert.NotSame(t, origReq, seenCopy.Request, "copy must carry scoped request")
	assert.Same(t, origReq, c.Request, "original gin request must be unchanged")
	assert.Equal(t, origCtx, c.Request.Context(), "original request context must be unchanged")
	assert.ErrorIs(t, seenCtx.Err(), context.Canceled)
	assert.Equal(t, 0, src.calls)
}

// TestKeywordRankingCancellation は callback cancel 後も ranking が継続しないことを証明する。
func TestKeywordRankingCancellation(t *testing.T) {
	art := manualarticle.CorpusArticle{
		Category: "screens", Slug: "a", Title: "会計 締め",
		Body: "会計と締めの手順。", Keywords: []string{}, KeywordsVI: []string{}, Hash: "h",
	}
	chunks, err := manualarticle.ChunkArticle(art)
	require.NoError(t, err)
	require.NotEmpty(t, chunks)
	second := chunks[0]
	second.ID += "-second"
	second.Ordinal++
	chunks = append(chunks, second)
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	out, err := rankKeywordChunks(ctx, chunks, func(string) (string, string, bool) {
		calls++
		cancel()
		return "会計 締め 会計 締め", "kaikei shime", true
	}, extractSearchTermGroups("会計と締め"))
	assert.Nil(t, out)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 1, calls, "no second callback after cancellation")

	cancelled, cancel2 := context.WithCancel(context.Background())
	cancel2()
	out, err = rankKeywordChunks(cancelled, nil, func(string) (string, string, bool) {
		return "", "", false
	}, extractSearchTermGroups("会計"))
	assert.Nil(t, out)
	assert.ErrorIs(t, err, context.Canceled)
}

// TestHybridHistoryCancellation503NoCalls は履歴スキャン中の取消が false no_evidence
// ではなく 503 に分類され、以降の外部/quota/persist を消費しないことを証明する。
func TestHybridHistoryCancellation503NoCalls(t *testing.T) {
	src := &countingCorpusSource{corpus: testCorpus()}
	chat := &mockChat{reply: "x"}
	b := &fakeBudget{granted: true}
	emb := &fakeEmbedder{}
	var cancelParent context.CancelFunc
	historyCalls, persistCalls := 0, 0
	svc := &mockService{
		listChatFn: func(ctx context.Context, _, _ uint64) ([]model.SupportChatMessage, error) {
			historyCalls++
			cancelParent()
			return []model.SupportChatMessage{{
				Role:    model.SupportChatRoleAssistant,
				Sources: json.RawMessage(`[{"title":"x","category":"screens","slug":"accounting"}]`),
			}}, nil
		},
		recordChatFn: func(context.Context, uint64, uint64, string, string, []ChatSource) error {
			persistCalls++
			return nil
		},
	}
	deps := NewManualSearchDeps(true, src, func(*gin.Context) (bool, error) { return true, nil }, emb, &fakeCache{}, "m")
	h := NewHandler(svc, nil, nil, chat, nil).WithManualSearch(deps).WithLLMBudget(b, "m")
	rev := testCorpus().Revision
	body := `{"message":"会計の締め方は？","manual_revision":"` + rev + `","keyword_candidates":[{"category":"screens","slug":"nonexistent"}]}`
	scopedParent, parentCancel := context.WithCancel(context.Background())
	cancelParent = parentCancel
	c, rec := newRequest(t, http.MethodPost, "/api/v1/support/chat", bytes.NewBufferString(body), "application/json")
	c.Request = c.Request.WithContext(scopedParent)
	h.Chat(c)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Equal(t, 1, historyCalls)
	assert.Equal(t, 0, persistCalls)
	assert.Equal(t, 0, emb.calls)
	assert.Equal(t, 0, b.calls)
	assert.Equal(t, 0, chat.calls)
}
