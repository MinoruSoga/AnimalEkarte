package support

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/manualarticle"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestHybridCapacityUsesTailEvidence(t *testing.T) {
	const fact = "末尾の会計と締めの手順を確認します。"
	corpus := manualarticle.BuildCorpus([]manualarticle.CorpusArticle{{
		Category: "screens", Slug: "tail", Title: "T", Body: strings.Repeat("あ", 100000) + fact,
	}}, nil)
	chat, embedder, budget := &mockChat{reply: "回答"}, &fakeEmbedder{}, &fakeBudget{granted: true}
	h := NewHandler(nil, nil, nil, chat, nil).
		WithManualSearch(NewManualSearchDeps(true, fakeCorpusSource{corpus: corpus}, func(*gin.Context) (bool, error) { return true, nil }, embedder, nil, "m")).
		WithLLMBudget(budget, "m")
	body, err := json.Marshal(map[string]any{
		"message": "会計と締めの手順", "manual_revision": corpus.Revision,
		"keyword_candidates": []ManualCandidate{{Category: "screens", Slug: "tail"}},
	})
	require.NoError(t, err)
	c, rec := newRequest(t, http.MethodPost, "/api/v1/support/chat", bytes.NewBuffer(body), "application/json")
	h.Chat(c)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.NotEmpty(t, chat.messages)
	require.Contains(t, chat.messages[len(chat.messages)-1].Content, fact)
	require.Contains(t, rec.Body.String(), `"reason":"capacity"`)
	require.Zero(t, embedder.calls)
	require.Equal(t, 1, budget.calls, "generation grant only; no embedding grant")
}

func TestHybridCapacityUsesOnlyLatestHistorySources(t *testing.T) {
	const fact = "最新の会計と締めの手順を確認します。"
	corpus := manualarticle.BuildCorpus([]manualarticle.CorpusArticle{
		{Category: "screens", Slug: "noise", Title: "T", Body: strings.Repeat("あ", 100000)},
		{Category: "screens", Slug: "current", Title: "T", Body: fact},
	}, nil)
	for _, latestEmpty := range []bool{false, true} {
		t.Run(fmt.Sprintf("latest_empty_%t", latestEmpty), func(t *testing.T) {
			history := []model.SupportChatMessage{{
				Role:    model.SupportChatRoleAssistant,
				Sources: sourcesPtr([]byte(`[{"category":"screens","slug":"current","title":"T"}]`)),
			}}
			if latestEmpty {
				history = append(history, model.SupportChatMessage{Role: model.SupportChatRoleAssistant})
			}
			service := &mockService{listChatFn: func(_ context.Context, clinicID, staffID uint64) ([]model.SupportChatMessage, error) {
				require.Equal(t, uint64(1), clinicID)
				require.Equal(t, uint64(5), staffID)
				return history, nil
			}}
			chat, embedder := &mockChat{reply: "回答"}, &fakeEmbedder{}
			h := NewHandler(service, nil, nil, chat, nil).
				WithManualSearch(NewManualSearchDeps(true, fakeCorpusSource{corpus: corpus}, func(*gin.Context) (bool, error) { return true, nil }, embedder, nil, "m")).
				WithLLMBudget(&fakeBudget{granted: true}, "m")
			body, err := json.Marshal(map[string]any{
				"message": "会計と締めの手順", "manual_revision": corpus.Revision, "keyword_candidates": []ManualCandidate{},
			})
			require.NoError(t, err)
			c, rec := newRequest(t, http.MethodPost, "/api/v1/support/chat", bytes.NewBuffer(body), "application/json")
			h.Chat(c)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			require.Zero(t, embedder.calls)
			if latestEmpty {
				require.Empty(t, chat.messages, "must not scan the older assistant's sources")
			} else {
				require.NotEmpty(t, chat.messages)
				require.Contains(t, chat.messages[len(chat.messages)-1].Content, fact)
			}
		})
	}
}

func TestKeywordArticleRankOrderAndLateEvidence(t *testing.T) {
	groups := [][]string{{"会計"}, {"締め"}}
	articles := []manualarticle.CorpusArticle{
		{Category: "screens", Slug: "a", Body: "## A\n会計と締め。\n## B\n会計と締め。"},
		{Category: "screens", Slug: "b", Body: "## A\n会計と締め。"},
	}
	ranked, err := rankKeywordArticles(context.Background(), articles, groups, 3)
	require.NoError(t, err)
	require.Len(t, ranked, 3)
	require.Equal(t, []string{"screens/a", "screens/b", "screens/a"}, []string{ranked[0].ArticleID, ranked[1].ArticleID, ranked[2].ArticleID})
	require.Equal(t, []int{0, 0, 1}, []int{ranked[0].Ordinal, ranked[1].Ordinal, ranked[2].Ordinal})

	var body strings.Builder
	for i := 0; i < 30; i++ {
		fmt.Fprintf(&body, "## H%d\n会計と締め。\n", i)
	}
	body.WriteString("## Last\n会計と締めと予約。")
	ranked, err = rankKeywordArticles(context.Background(), []manualarticle.CorpusArticle{
		{Category: "screens", Slug: "late", Body: body.String()},
	}, append(groups, []string{"予約"}), 20)
	require.NoError(t, err)
	require.Len(t, ranked, 20)
	require.Equal(t, 30, ranked[0].Ordinal, "later stronger evidence must displace earlier weaker chunks")
	require.Contains(t, ranked[0].Body, "予約")
}
