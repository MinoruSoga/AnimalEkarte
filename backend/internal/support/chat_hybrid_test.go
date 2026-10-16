package support

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/manualarticle"
	"github.com/animal-ekarte/backend/internal/model"
)

func rev() string { return testCorpus().Revision }

func TestHybrid(t *testing.T) {
	goodBody := `{"message":"会計の締め方は？","manual_revision":"` + rev() + `","keyword_candidates":[{"category":"screens","slug":"accounting"}]}`
	tests := []struct {
		name       string
		body       string
		enabled    bool
		corpusErr  error
		perm       bool
		chat       *mockChat
		wantStatus int
		wantReply  string
	}{
		{name: "id request keyword evidence reaches completer", body: goodBody, enabled: true, chat: &mockChat{reply: "回答"}, wantStatus: 200, wantReply: "回答"},
		{name: "revision mismatch 409", body: strings.Replace(goodBody, rev(), strings.Repeat("f", 64), 1), enabled: true, chat: &mockChat{}, wantStatus: 409},
		{name: "corpus failure 503", body: goodBody, enabled: true, corpusErr: assert.AnError, chat: &mockChat{}, wantStatus: 503},
		{name: "non-hex revision 400", body: strings.Replace(goodBody, rev(), "zz", 1), enabled: true, chat: &mockChat{}, wantStatus: 400},
		{name: "missing revision 400", body: `{"message":"x","keyword_candidates":[]}`, enabled: true, chat: &mockChat{}, wantStatus: 400},
		{name: "legacy empty array accepted hybrid-disabled", body: `{"message":"x","context":[]}`, enabled: false, chat: &mockChat{reply: "ok"}, wantStatus: 200},
		{name: "legacy context rejected hybrid-enabled", body: `{"message":"x","context":[{"category":"screens","slug":"accounting"}]}`, enabled: true, chat: &mockChat{}, wantStatus: 409},
		{name: "legacy invalid category 400", body: `{"message":"x","context":[{"category":"bogus","slug":"x"}]}`, enabled: false, chat: &mockChat{}, wantStatus: 400},
		{name: "duplicate candidate ids 400", body: `{"message":"x","manual_revision":"` + rev() + `","keyword_candidates":[{"category":"screens","slug":"accounting"},{"category":"screens","slug":"accounting"}]}`, enabled: true, chat: &mockChat{}, wantStatus: 400},
		{name: "trailing json 400", body: goodBody + ` {"x":1}`, enabled: true, chat: &mockChat{}, wantStatus: 400},
		{name: "no evidence fixed reply no completer", body: `{"message":"完全に無関係な質問ですよ","manual_revision":"` + rev() + `","keyword_candidates":[]}`, enabled: true, chat: &mockChat{reply: "never"}, wantStatus: 200, wantReply: chatFixedResponse("ja")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := NewManualSearchDeps(tt.enabled, fakeCorpusSource{corpus: testCorpus(), err: tt.corpusErr},
				func(*gin.Context) (bool, error) { return tt.perm, nil }, nil, nil, "m")
			h := NewHandler(nil, nil, nil, tt.chat, nil).
				WithManualSearch(deps).
				WithLLMBudget(&fakeBudget{granted: true}, "m")
			c, rec := newRequest(t, http.MethodPost, "/api/v1/support/chat", bytes.NewBufferString(tt.body), "application/json")
			h.Chat(c)
			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantReply != "" {
				assert.Contains(t, rec.Body.String(), tt.wantReply)
			}
		})
	}
}

func TestHybridMissingDeps503(t *testing.T) {
	h := NewHandler(nil, nil, nil, &mockChat{reply: "x"}, nil)
	c, rec := newRequest(t, http.MethodPost, "/api/v1/support/chat", bytes.NewBufferString(`{"message":"x"}`), "application/json")
	h.Chat(c)
	assert.Equal(t, 503, rec.Code)
}

func TestHybridDisabledReceivesIDRequest(t *testing.T) {
	chat := &mockChat{reply: "回答"}
	h := newChatTestHandler(nil, chat, nil)
	body := `{"message":"会計の締め方は？","manual_revision":"` + rev() + `","keyword_candidates":[{"category":"screens","slug":"accounting"}]}`
	c, rec := newRequest(t, http.MethodPost, "/api/v1/support/chat", bytes.NewBufferString(body), "application/json")
	h.Chat(c)
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.NotEmpty(t, chat.messages)
	last := chat.messages[len(chat.messages)-1].Content
	assert.Contains(t, last, "画面別 会計")
	assert.Contains(t, last, "レジ締め")
}

func TestHybridLegacyClientTextNeverSent(t *testing.T) {
	chat := &mockChat{reply: "回答"}
	h := newChatTestHandler(nil, chat, nil)
	body := `{"message":"会計の締め方は？","context":[{"title":"FORGED_TITLE_SECRET","category":"screens","slug":"accounting","text":"client-injected text secret1234"}]}`
	c, rec := newRequest(t, http.MethodPost, "/api/v1/support/chat", bytes.NewBufferString(body), "application/json")
	h.Chat(c)
	require.Equal(t, 200, rec.Code, rec.Body.String())
	for _, m := range chat.messages {
		assert.NotContains(t, m.Content, "FORGED_TITLE_SECRET")
		assert.NotContains(t, m.Content, "client-injected")
	}
}

func TestHybridInflightBeforeGrant(t *testing.T) {
	b := &fakeBudget{granted: true}
	h := newChatTestHandler(nil, &mockChat{reply: "ok"}, b)
	body := `{"message":"会計の締め方は？","manual_revision":"` + rev() + `","keyword_candidates":[{"category":"screens","slug":"accounting"}]}`
	c, rec := newRequest(t, http.MethodPost, "/api/v1/support/chat", bytes.NewBufferString(body), "application/json")
	h.Chat(c)
	require.Equal(t, 200, rec.Code, rec.Body.String())
	assert.GreaterOrEqual(t, b.calls, 1)
	assert.Contains(t, b.gotScopes, LLMBudgetScopeGlobal)
}

type fakeEmbedder struct {
	calls int
	vec   []float64
	err   error
}

func (f *fakeEmbedder) Embed(_ context.Context, _, _ string) ([]float64, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	if f.vec != nil {
		return f.vec, nil
	}
	return make([]float64, 1536), nil
}

type fakeCache struct {
	vecs       map[string][]float64
	err        error
	loadCalls  int
	storeCalls int
}

func (f *fakeCache) Load(_ context.Context, keys []string) (map[string][]float64, error) {
	f.loadCalls++
	if f.err != nil {
		return nil, f.err
	}
	out := map[string][]float64{}
	for _, k := range keys {
		if v, ok := f.vecs[k]; ok {
			out[k] = v
		}
	}
	return out, nil
}
func (f *fakeCache) StoreWithKey(_ context.Context, k, m, tt string, cv, d int, e []float64) error {
	f.storeCalls++
	return nil
}

func TestSemanticQuotaZeroAPICalls(t *testing.T) {
	// cache miss => no grant, no embed call
	emb := &fakeEmbedder{}
	budget := &fakeBudget{granted: false}
	deps := NewManualSearchDeps(true, fakeCorpusSource{corpus: testCorpus()}, func(*gin.Context) (bool, error) { return true, nil }, emb, &fakeCache{vecs: map[string][]float64{}}, "m")
	_, err := deps.semanticCandidates(context.Background(), nil, "q", budget, 1, 1)
	require.Error(t, err)
	assert.Equal(t, 0, emb.calls)
	assert.Equal(t, 0, budget.calls)
	// quota denied => no embed
	chunks, _ := manualarticle.ChunkCorpus(testCorpus())
	cache := &fakeCache{vecs: map[string][]float64{}}
	for _, ch := range chunks {
		k := manualarticle.ChunkCacheKey("m", manualarticle.TaskTypeRetrievalDocument, ch.EmbeddingInput)
		cache.vecs[k] = make([]float64, 1536)
	}
	deps.Cache = cache
	_, err = deps.semanticCandidates(context.Background(), chunks, "q", budget, 1, 1)
	require.Error(t, err)
	assert.Equal(t, 1, budget.calls)
	assert.Equal(t, 0, emb.calls)
}

type recordingCorpusSource struct {
	corpus    manualarticle.Corpus
	permitted []bool
	err       error
}

func (f *recordingCorpusSource) EffectiveCorpus(_ context.Context, permitted bool) (manualarticle.Corpus, error) {
	f.permitted = append(f.permitted, permitted)
	return f.corpus, f.err
}

func TestHybridPermDeniedUsesBaseOnly(t *testing.T) {
	src := &recordingCorpusSource{corpus: testCorpus()}
	deps := NewManualSearchDeps(false, src, func(*gin.Context) (bool, error) { return false, nil }, nil, nil, "m")
	h := NewHandler(nil, nil, nil, &mockChat{reply: "ok"}, nil).
		WithManualSearch(deps).WithLLMBudget(&fakeBudget{granted: true}, "m")
	body := `{"message":"x","context":[]}`
	c, rec := newRequest(t, http.MethodPost, "/api/v1/support/chat", bytes.NewBufferString(body), "application/json")
	h.Chat(c)
	require.Equal(t, 200, rec.Code)
	require.Equal(t, []bool{false}, src.permitted)
}

func TestHybridNilPermission503(t *testing.T) {
	deps := NewManualSearchDeps(false, fakeCorpusSource{corpus: testCorpus()}, nil, nil, nil, "m")
	chat := &mockChat{reply: "x"}
	h := NewHandler(nil, nil, nil, chat, nil).WithManualSearch(deps)
	c, rec := newRequest(t, http.MethodPost, "/api/v1/support/chat", bytes.NewBufferString(`{"message":"x","context":[]}`), "application/json")
	h.Chat(c)
	assert.Equal(t, 503, rec.Code)
	assert.Empty(t, chat.messages)
}

func TestHybridCorpusDBError503NoCalls(t *testing.T) {
	chat := &mockChat{reply: "x"}
	emb := &fakeEmbedder{}
	deps := NewManualSearchDeps(true, fakeCorpusSource{err: assert.AnError}, func(*gin.Context) (bool, error) { return true, nil }, emb, &fakeCache{}, "m")
	h := NewHandler(nil, nil, nil, chat, nil).WithManualSearch(deps).WithLLMBudget(&fakeBudget{granted: true}, "m")
	c, rec := newRequest(t, http.MethodPost, "/api/v1/support/chat", bytes.NewBufferString(`{"message":"x","manual_revision":"`+strings.Repeat("a", 64)+`","keyword_candidates":[]}`), "application/json")
	h.Chat(c)
	assert.Equal(t, 503, rec.Code)
	assert.Empty(t, chat.messages)
	assert.Equal(t, 0, emb.calls)
}

func TestHybridInflightBeforeGrant429(t *testing.T) {
	b := &fakeBudget{granted: true}
	chat := &mockChat{reply: "x"}
	h := newChatTestHandler(nil, chat, b)
	// newRequest の認証コンテキストは clinic_id=1 / user_id=5。先に inflight を占有。
	require.True(t, h.acquireChatInflight(1, 5))
	defer h.releaseChatInflight(1, 5)
	body := `{"message":"会計の締め方は？","context":[{"title":"","category":"screens","slug":"accounting","text":""}]}`
	c, rec := newRequest(t, http.MethodPost, "/api/v1/support/chat", bytes.NewBufferString(body), "application/json")
	h.Chat(c)
	assert.Equal(t, 429, rec.Code)
	assert.Equal(t, 0, b.calls)
	assert.Empty(t, chat.messages)
}

func TestKeywordThresholdMinTwoGroups(t *testing.T) {
	// 2グループ抽出の質問で 1 グループのみ一致の段落は弱ヒット（採用しない）。
	art := manualarticle.CorpusArticle{
		Category: "screens", Slug: "a1", Title: "T", Section: "S",
		Body: "会計のみ含む段落。", Keywords: []string{}, KeywordsVI: []string{}, Hash: "h",
	}
	chunks, err := manualarticle.ChunkArticle(art)
	require.NoError(t, err)
	require.NotEmpty(t, chunks)
	byID := map[string]manualarticle.CorpusArticle{art.Category + "/" + art.Slug: art}
	rank, rankErr := rankKeywordChunks(context.Background(), chunks, func(chID string) (string, string, bool) {
		for _, ch := range chunks {
			if ch.ID == chID {
				a := byID[ch.ArticleID]
				n, f := articleSearchHaystack(a.Title, a.Section, ch.Heading+"\n"+ch.Body, a.Keywords, a.KeywordsVI)
				return n, f, true
			}
		}
		return "", "", false
	}, extractSearchTermGroups("会計と締めの手順"))
	require.NoError(t, rankErr)
	assert.Empty(t, rank, "1-group match must be weak (not admitted)")

	art.Body = "会計と締めの両方を含む段落。会計の締め処理。"
	chunks, _ = manualarticle.ChunkArticle(art)
	rank, rankErr = rankKeywordChunks(context.Background(), chunks, func(chID string) (string, string, bool) {
		for _, ch := range chunks {
			if ch.ID == chID {
				a := byID[ch.ArticleID]
				n, f := articleSearchHaystack(a.Title, a.Section, ch.Heading+"\n"+ch.Body, a.Keywords, a.KeywordsVI)
				return n, f, true
			}
		}
		return "", "", false
	}, extractSearchTermGroups("会計と締めの手順"))
	require.NoError(t, rankErr)
	assert.NotEmpty(t, rank)
}

func TestRRFMergeCapsAndDedupe(t *testing.T) {
	mk := func(aid string, ord int) manualarticle.Chunk {
		return manualarticle.Chunk{ID: fmt.Sprintf("%s#%d", aid, ord), ArticleID: aid, Ordinal: ord, Body: "b"}
	}
	// same chunk in both lists dedupes; max 2 per article; top3 cap
	kw := []manualarticle.Chunk{mk("a", 0), mk("a", 1), mk("a", 2), mk("b", 0)}
	se := []manualarticle.Chunk{mk("a", 0), mk("c", 0)}
	out := mergeRRF(kw, se)
	require.Len(t, out, 3)
	counts := map[string]int{}
	for _, c := range out {
		counts[c.ArticleID]++
	}
	assert.LessOrEqual(t, counts["a"], 2)
}

// 新プロトコルは keyword_candidates の非 null 配列を必須とする（EMR-410 R5）。
// missing / null は 400、空配列 [] は有効（corpus 解決まで進む）。
func TestHybridNewProtocolKeywordCandidatesRequired(t *testing.T) {
	revision := strings.Repeat("a", 64)
	for name, body := range map[string]string{
		"missing":  `{"message":"x","manual_revision":"` + revision + `"}`,
		"null":     `{"message":"x","manual_revision":"` + revision + `","keyword_candidates":null}`,
		"empty_ok": `{"message":"x","manual_revision":"` + revision + `","keyword_candidates":[]}`,
	} {
		emb := &fakeEmbedder{}
		chat := &mockChat{reply: "x"}
		deps := NewManualSearchDeps(false, fakeCorpusSource{corpus: testCorpus()}, func(*gin.Context) (bool, error) { return true, nil }, emb, &fakeCache{}, "m")
		h := NewHandler(nil, nil, nil, chat, nil).WithManualSearch(deps).WithLLMBudget(&fakeBudget{granted: true}, "m")
		c, rec := newRequest(t, http.MethodPost, "/api/v1/support/chat", bytes.NewBufferString(body), "application/json")
		h.Chat(c)
		if name == "empty_ok" {
			assert.NotEqual(t, 400, rec.Code, name)
		} else {
			assert.Equal(t, 400, rec.Code, name)
		}
		assert.Equal(t, 0, emb.calls, name)
	}
}

func TestHybridStopwordOnlyQuestionNoCalls(t *testing.T) {
	// stopword のみの質問は groups=0 → keyword ranked nil → fixed response。completer/budget 0。
	emb := &fakeEmbedder{}
	chat := &mockChat{reply: "x"}
	b := &fakeBudget{granted: true}
	deps := NewManualSearchDeps(false, fakeCorpusSource{corpus: testCorpus()}, func(*gin.Context) (bool, error) { return true, nil }, emb, &fakeCache{}, "m")
	h := NewHandler(nil, nil, nil, chat, nil).WithManualSearch(deps).WithLLMBudget(b, "m")
	c, rec := newRequest(t, http.MethodPost, "/api/v1/support/chat", bytes.NewBufferString(`{"message":"方法","context":[]}`), "application/json")
	h.Chat(c)
	assert.Equal(t, 200, rec.Code)
	assert.Empty(t, chat.messages)
	assert.Equal(t, 0, b.calls)
	assert.Equal(t, 0, emb.calls)
}

func TestHybridInflightOccupiedBeforeSemantic429(t *testing.T) {
	// inflight を事前占有した有効リクエストは semantic/chunking/quota 消費前に 429。
	emb := &fakeEmbedder{}
	chat := &mockChat{reply: "x"}
	b := &fakeBudget{granted: true}
	cache := &fakeCache{vecs: map[string][]float64{"k": {1}}}
	deps := NewManualSearchDeps(true, fakeCorpusSource{corpus: testCorpus()}, func(*gin.Context) (bool, error) { return true, nil }, emb, cache, "m")
	h := NewHandler(nil, nil, nil, chat, nil).WithManualSearch(deps).WithLLMBudget(b, "m")
	require.True(t, h.acquireChatInflight(1, 5))
	defer h.releaseChatInflight(1, 5)
	rev := testCorpus().Revision
	body := `{"message":"会計の締め方は？","manual_revision":"` + rev +
		`","keyword_candidates":[{"category":"screens","slug":"accounting"}]}`
	c, rec := newRequest(t, http.MethodPost, "/api/v1/support/chat", bytes.NewBufferString(body), "application/json")
	h.Chat(c)
	assert.Equal(t, 429, rec.Code)
	assert.Equal(t, 0, emb.calls)
	assert.Equal(t, 0, b.calls)
	assert.Empty(t, chat.messages)
}

func TestHybridGenerationDeadline504BeforeUpstream(t *testing.T) {
	chat := &mockChat{reply: "", err: context.DeadlineExceeded}
	h := newChatTestHandler(nil, chat, &fakeBudget{granted: true})
	body := `{"message":"会計の締め方は？","context":[{"title":"","category":"screens","slug":"accounting","text":""}]}`
	c, rec := newRequest(t, http.MethodPost, "/api/v1/support/chat", bytes.NewBufferString(body), "application/json")
	h.Chat(c)
	assert.Equal(t, 504, rec.Code)
}

func TestEMR410Review_StopwordRequestsSkipRetrieval(t *testing.T) {
	cases := []struct {
		name    string
		message string
		enabled bool
		reply   string
	}{
		{name: "Japanese enabled", message: "方法", enabled: true, reply: chatFixedResponse("ja")},
		{name: "Japanese disabled", message: "方法", enabled: false, reply: chatFixedResponse("ja")},
		{name: "Vietnamese enabled", message: "cách", enabled: true, reply: chatFixedResponse("vi")},
		{name: "Vietnamese disabled", message: "cách", enabled: false, reply: chatFixedResponse("vi")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := &countingCorpusSource{corpus: testCorpus()}
			permissionCalls, historyCalls, persistCalls := 0, 0, 0
			service := &mockService{
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
			embedder := &fakeEmbedder{}
			completer := &mockChat{reply: "must not be called"}
			budget := &fakeBudget{granted: true}
			deps := NewManualSearchDeps(
				tc.enabled,
				source,
				func(*gin.Context) (bool, error) {
					permissionCalls++
					return true, nil
				},
				embedder,
				cache,
				"m",
			)
			handler := NewHandler(service, nil, nil, completer, nil).
				WithManualSearch(deps).
				WithLLMBudget(budget, "m")
			body := `{"message":"` + tc.message + `","manual_revision":"` + testCorpus().Revision +
				`","keyword_candidates":[{"category":"screens","slug":"accounting"}]}`
			c, rec := newRequest(t, http.MethodPost, "/api/v1/support/chat", bytes.NewBufferString(body), "application/json")

			handler.Chat(c)

			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			var response chatResponse
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
			assert.Equal(t, tc.reply, response.Reply)
			require.NotNil(t, response.Sources)
			assert.Empty(t, response.Sources)
			require.NotNil(t, response.Retrieval)
			assert.Equal(t, RetrievalModeNone, response.Retrieval.Mode)
			assert.Equal(t, RetrievalReasonNoEvidence, response.Retrieval.Reason)
			assert.Equal(t, testCorpus().Revision, response.Retrieval.ManualRevision)
			assert.Equal(t, 1, permissionCalls)
			assert.Equal(t, 1, source.calls)
			assert.True(t, source.permitted)
			assert.Equal(t, 0, cache.loadCalls)
			assert.Equal(t, 0, cache.storeCalls)
			assert.Equal(t, 0, historyCalls)
			assert.Equal(t, 0, persistCalls)
			assert.Equal(t, 0, embedder.calls)
			assert.Equal(t, 0, budget.calls)
			assert.Equal(t, 0, completer.calls)
		})
	}
}

func TestEMR410Review_StopwordRevisionAndDeadlineGuards(t *testing.T) {
	t.Run("stale revision remains conflict", func(t *testing.T) {
		source := &countingCorpusSource{corpus: testCorpus()}
		permissionCalls := 0
		deps := NewManualSearchDeps(true, source, func(*gin.Context) (bool, error) {
			permissionCalls++
			return true, nil
		}, &fakeEmbedder{}, &fakeCache{}, "m")
		handler := NewHandler(nil, nil, nil, &mockChat{}, nil).WithManualSearch(deps)
		body := `{"message":"方法","manual_revision":"` + strings.Repeat("0", 64) +
			`","keyword_candidates":[{"category":"screens","slug":"accounting"}]}`
		c, rec := newRequest(t, http.MethodPost, "/api/v1/support/chat", bytes.NewBufferString(body), "application/json")

		handler.Chat(c)

		assert.Equal(t, http.StatusConflict, rec.Code)
		assert.Equal(t, 1, permissionCalls)
		assert.Equal(t, 1, source.calls)
	})

	t.Run("expired search context remains unavailable", func(t *testing.T) {
		source := &countingCorpusSource{corpus: testCorpus()}
		permissionCalls := 0
		deps := NewManualSearchDeps(true, source, func(*gin.Context) (bool, error) {
			permissionCalls++
			return true, nil
		}, &fakeEmbedder{}, &fakeCache{}, "m")
		deps.SearchTimeout = 0
		handler := NewHandler(nil, nil, nil, &mockChat{}, nil).WithManualSearch(deps)
		body := `{"message":"方法","manual_revision":"` + testCorpus().Revision +
			`","keyword_candidates":[{"category":"screens","slug":"accounting"}]}`
		c, rec := newRequest(t, http.MethodPost, "/api/v1/support/chat", bytes.NewBufferString(body), "application/json")

		handler.Chat(c)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
		assert.Equal(t, 1, permissionCalls)
		assert.Equal(t, 1, source.calls)
	})
}

func TestEMR410Review_MalformedResponseMaps502WithoutPersistence(t *testing.T) {
	const apiKey = "test-api-key"
	const providerMarker = "malformed-provider-marker"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"`+providerMarker)
	}))
	defer server.Close()

	persistCalls := 0
	service := &mockService{
		recordChatFn: func(context.Context, uint64, uint64, string, string, []ChatSource) error {
			persistCalls++
			return nil
		},
	}
	completer := NewChatCompleter(server.URL, apiKey, "test-model", time.Second)
	handler := newChatTestHandler(service, completer, &fakeBudget{granted: true})
	body := `{"message":"会計の締め方は？","context":[{"title":"","category":"screens","slug":"accounting","text":""}]}`
	c, rec := newRequest(t, http.MethodPost, "/api/v1/support/chat", bytes.NewBufferString(body), "application/json")

	handler.Chat(c)

	assert.Equal(t, http.StatusBadGateway, rec.Code)
	assert.NotContains(t, rec.Body.String(), apiKey)
	assert.NotContains(t, rec.Body.String(), providerMarker)
	assert.Equal(t, 0, persistCalls)
}

func TestEMR410Review_TransportFailureMaps502WithoutPersistence(t *testing.T) {
	const apiKey = "test-api-key"
	sentinel := errors.New("transport sentinel")
	completer := NewChatCompleter("https://provider.invalid/v1", apiKey, "test-model", time.Second)
	client, ok := completer.(*openAICompatibleClient)
	require.True(t, ok)
	client.httpClient = &http.Client{Transport: reviewRoundTripper(func(*http.Request) (*http.Response, error) {
		return nil, sentinel
	})}
	persistCalls := 0
	service := &mockService{
		recordChatFn: func(context.Context, uint64, uint64, string, string, []ChatSource) error {
			persistCalls++
			return nil
		},
	}
	handler := newChatTestHandler(service, completer, &fakeBudget{granted: true})
	body := `{"message":"会計の締め方は？","context":[{"title":"","category":"screens","slug":"accounting","text":""}]}`
	c, rec := newRequest(t, http.MethodPost, "/api/v1/support/chat", bytes.NewBufferString(body), "application/json")

	handler.Chat(c)

	assert.Equal(t, http.StatusBadGateway, rec.Code)
	assert.NotContains(t, rec.Body.String(), apiKey)
	assert.NotContains(t, rec.Body.String(), sentinel.Error())
	assert.Equal(t, 0, persistCalls)
}

func TestEMR410Review_DeadlineStill504(t *testing.T) {
	persistCalls := 0
	service := &mockService{
		recordChatFn: func(context.Context, uint64, uint64, string, string, []ChatSource) error {
			persistCalls++
			return nil
		},
	}
	handler := newChatTestHandler(service, &mockChat{err: context.DeadlineExceeded}, &fakeBudget{granted: true})
	body := `{"message":"会計の締め方は？","context":[{"title":"","category":"screens","slug":"accounting","text":""}]}`
	c, rec := newRequest(t, http.MethodPost, "/api/v1/support/chat", bytes.NewBufferString(body), "application/json")

	handler.Chat(c)

	assert.Equal(t, http.StatusGatewayTimeout, rec.Code)
	assert.Equal(t, 0, persistCalls)
}

func TestEMR410Review_TransportDeadlineStill504(t *testing.T) {
	const apiKey = "test-api-key"
	completer := NewChatCompleter("https://provider.invalid/v1", apiKey, "test-model", time.Second)
	client, ok := completer.(*openAICompatibleClient)
	require.True(t, ok)
	client.httpClient = &http.Client{Transport: reviewRoundTripper(func(*http.Request) (*http.Response, error) {
		return nil, context.DeadlineExceeded
	})}
	persistCalls := 0
	service := &mockService{
		recordChatFn: func(context.Context, uint64, uint64, string, string, []ChatSource) error {
			persistCalls++
			return nil
		},
	}
	handler := newChatTestHandler(service, completer, &fakeBudget{granted: true})
	body := `{"message":"会計の締め方は？","context":[{"title":"","category":"screens","slug":"accounting","text":""}]}`
	c, rec := newRequest(t, http.MethodPost, "/api/v1/support/chat", bytes.NewBufferString(body), "application/json")

	handler.Chat(c)

	assert.Equal(t, http.StatusGatewayTimeout, rec.Code)
	assert.NotContains(t, rec.Body.String(), apiKey)
	assert.Equal(t, 0, persistCalls)
}

func TestHybridSensitiveStoredHistoryOmitted(t *testing.T) {
	for _, tc := range []struct {
		name string
		role model.SupportChatMessageRole
	}{
		{name: "unsafe user history", role: model.SupportChatRoleUser},
		{name: "unsafe assistant history", role: model.SupportChatRoleAssistant},
	} {
		t.Run(tc.name, func(t *testing.T) {
			question := "会計の締め方は？"
			safeUser := "前の安全な質問"
			safeAssistant := "前の安全な回答"
			unsafeContent := "保存済みの内容 legacy.staff@example.test"
			original := []model.SupportChatMessage{
				{Role: model.SupportChatRoleUser, Content: safeUser},
				{Role: model.SupportChatRoleAssistant, Content: safeAssistant},
				{Role: tc.role, Content: unsafeContent},
			}
			originalBefore := append([]model.SupportChatMessage(nil), original...)
			historyReads, persistCalls, clearCalls := 0, 0, 0
			type persistedExchange struct {
				user      string
				assistant string
			}
			var persisted []persistedExchange
			svc := &mockService{
				listChatFn: func(context.Context, uint64, uint64) ([]model.SupportChatMessage, error) {
					historyReads++
					return original, nil
				},
				recordChatFn: func(_ context.Context, _ uint64, _ uint64, user, assistant string, _ []ChatSource) error {
					persistCalls++
					persisted = append(persisted, persistedExchange{user: user, assistant: assistant})
					return nil
				},
				clearChatFn: func(context.Context, uint64, uint64) error {
					clearCalls++
					return nil
				},
			}
			chat := &mockChat{reply: "回答"}
			handler := newChatTestHandler(svc, chat, &fakeBudget{granted: true})
			body := `{"message":"` + question + `","manual_revision":"` + testCorpus().Revision +
				`","keyword_candidates":[{"category":"screens","slug":"accounting"}]}`

			for call := 1; call <= 2; call++ {
				c, rec := newRequest(t, http.MethodPost, "/api/v1/support/chat", bytes.NewBufferString(body), "application/json")
				handler.Chat(c)

				require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
				require.Equal(t, call, chat.calls)
				require.Len(t, chat.messages, 2)
				assert.Equal(t, "system", chat.messages[0].Role)
				assert.Equal(t, "user", chat.messages[1].Role)
				assert.Contains(t, chat.messages[1].Content, question)
				assert.NotContains(t, chat.messages[1].Content, safeUser)
				assert.NotContains(t, chat.messages[1].Content, safeAssistant)
				assert.NotContains(t, chat.messages[1].Content, unsafeContent)
				assert.Equal(t, call, historyReads)
				assert.Len(t, persisted, call)
				assert.Equal(t, question, persisted[call-1].user)
				assert.Equal(t, "回答", persisted[call-1].assistant)
			}

			assert.Equal(t, 2, chat.calls)
			assert.Equal(t, 2, persistCalls)
			assert.Zero(t, clearCalls)
			assert.Equal(t, originalBefore, original)
		})
	}
}

func TestHybridSafeStoredHistoryPreserved(t *testing.T) {
	question := "会計の締め方は？"
	safeUser := "安全な前の質問"
	safeAssistant := "安全な前の回答"
	history := []model.SupportChatMessage{
		{Role: model.SupportChatRoleUser, Content: safeUser},
		{Role: model.SupportChatRoleAssistant, Content: safeAssistant},
	}
	chat := &mockChat{reply: "回答"}
	handler := newChatTestHandler(&mockService{
		listChatFn: func(context.Context, uint64, uint64) ([]model.SupportChatMessage, error) {
			return history, nil
		},
	}, chat, &fakeBudget{granted: true})
	body := `{"message":"` + question + `","manual_revision":"` + testCorpus().Revision +
		`","keyword_candidates":[{"category":"screens","slug":"accounting"}]}`
	c, rec := newRequest(t, http.MethodPost, "/api/v1/support/chat", bytes.NewBufferString(body), "application/json")

	handler.Chat(c)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Equal(t, 1, chat.calls)
	require.Len(t, chat.messages, 4)
	assert.Equal(t, "system", chat.messages[0].Role)
	assert.Equal(t, string(model.SupportChatRoleUser), chat.messages[1].Role)
	assert.Equal(t, safeUser, chat.messages[1].Content)
	assert.Equal(t, string(model.SupportChatRoleAssistant), chat.messages[2].Role)
	assert.Equal(t, safeAssistant, chat.messages[2].Content)
	assert.Equal(t, "user", chat.messages[3].Role)
	assert.Contains(t, chat.messages[3].Content, question)
}

func TestHybridOutboundScreeningStillRejectsCurrentData(t *testing.T) {
	t.Run("new message", func(t *testing.T) {
		chat := &mockChat{reply: "must not be called"}
		budget := &fakeBudget{granted: true}
		persistCalls := 0
		svc := &mockService{
			recordChatFn: func(context.Context, uint64, uint64, string, string, []ChatSource) error {
				persistCalls++
				return nil
			},
		}
		handler := newChatTestHandler(svc, chat, budget)
		body := `{"message":"会計の締め方は？ legacy.staff@example.test","manual_revision":"` + testCorpus().Revision +
			`","keyword_candidates":[{"category":"screens","slug":"accounting"}]}`
		c, rec := newRequest(t, http.MethodPost, "/api/v1/support/chat", bytes.NewBufferString(body), "application/json")

		handler.Chat(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Zero(t, chat.calls)
		assert.Zero(t, budget.calls)
		assert.Zero(t, persistCalls)
	})

	t.Run("resolved manual body", func(t *testing.T) {
		corpus := manualarticle.BuildCorpus([]manualarticle.CorpusArticle{{
			Category: "screens",
			Slug:     "accounting",
			Title:    "会計の締め手順",
			Section:  "基本",
			Body:     "会計の締め方はレジ締め画面で確認できます。 legacy.staff@example.test",
		}}, nil)
		chat := &mockChat{reply: "must not be called"}
		budget := &fakeBudget{granted: true}
		persistCalls := 0
		svc := &mockService{
			recordChatFn: func(context.Context, uint64, uint64, string, string, []ChatSource) error {
				persistCalls++
				return nil
			},
		}
		deps := NewManualSearchDeps(
			false,
			fakeCorpusSource{corpus: corpus},
			func(*gin.Context) (bool, error) { return true, nil },
			nil,
			nil,
			"m",
		)
		handler := NewHandler(svc, nil, nil, chat, nil).
			WithManualSearch(deps).
			WithLLMBudget(budget, "m")
		body := `{"message":"会計の締め方は？","manual_revision":"` + corpus.Revision +
			`","keyword_candidates":[{"category":"screens","slug":"accounting"}]}`
		c, rec := newRequest(t, http.MethodPost, "/api/v1/support/chat", bytes.NewBufferString(body), "application/json")

		handler.Chat(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Zero(t, chat.calls)
		assert.Zero(t, budget.calls)
		assert.Zero(t, persistCalls)
	})
}
