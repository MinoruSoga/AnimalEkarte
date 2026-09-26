package support

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewChatCompleter(t *testing.T) {
	t.Run("returns nil without api key", func(t *testing.T) {
		assert.Nil(t, NewChatCompleter("https://api.openai.com/v1", "", "gpt-5-nano", time.Second))
		assert.Nil(t, NewChatCompleter("https://api.openai.com/v1", "  ", "gpt-5-nano", time.Second))
	})

	t.Run("returns client with api key", func(t *testing.T) {
		assert.NotNil(t, NewChatCompleter("https://api.openai.com/v1", "sk-test", "gpt-5-nano", time.Second))
	})
}

func TestOpenAICompatibleClient_Complete(t *testing.T) {
	t.Run("sends auth header and model, parses reply", func(t *testing.T) {
		var gotAuth, gotContentType string
		var gotBody map[string]any
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v1/chat/completions", r.URL.Path)
			gotAuth = r.Header.Get("Authorization")
			gotContentType = r.Header.Get("Content-Type")
			require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"choices":[{"message":{"role":"assistant","content":"  回答です  "}}]}`)
		}))
		defer srv.Close()

		client := NewChatCompleter(srv.URL+"/v1/", "sk-secret", "gpt-5-nano", 5*time.Second)
		reply, err := client.Complete(context.Background(), []ChatMessage{
			{Role: "system", Content: "rules"},
			{Role: "user", Content: "質問"},
		})

		require.NoError(t, err)
		assert.Equal(t, "回答です", reply)
		assert.Equal(t, "Bearer sk-secret", gotAuth)
		assert.Equal(t, "application/json", gotContentType)
		assert.Equal(t, "gpt-5-nano", gotBody["model"])
		assert.Equal(t, float64(chatMaxCompletionTokens), gotBody["max_completion_tokens"])
		assert.NotContains(t, gotBody, "max_tokens")
		assert.NotContains(t, gotBody, "thinking")
		require.Len(t, gotBody["messages"], 2)
	})

	t.Run("maps non-2xx to upstreamError without leaking body", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = io.WriteString(w, `{"error":"rate limited: sk-secret-echo"}`)
		}))
		defer srv.Close()

		client := NewChatCompleter(srv.URL, "sk-secret", "gpt-5-nano", 5*time.Second)
		_, err := client.Complete(context.Background(), []ChatMessage{{Role: "user", Content: "x"}})

		var upstream *upstreamError
		require.ErrorAs(t, err, &upstream)
		assert.Equal(t, http.StatusTooManyRequests, upstream.status)
		assert.NotContains(t, err.Error(), "sk-secret-echo")
	})

	t.Run("returns errUpstreamEmpty on empty choices", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, `{"choices":[]}`)
		}))
		defer srv.Close()

		client := NewChatCompleter(srv.URL, "k", "m", 5*time.Second)
		_, err := client.Complete(context.Background(), []ChatMessage{{Role: "user", Content: "x"}})
		assert.ErrorIs(t, err, errUpstreamEmpty)
	})

	t.Run("returns errUpstreamEmpty on blank content", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, `{"choices":[{"message":{"role":"assistant","content":"  "}}]}`)
		}))
		defer srv.Close()

		client := NewChatCompleter(srv.URL, "k", "m", 5*time.Second)
		_, err := client.Complete(context.Background(), []ChatMessage{{Role: "user", Content: "x"}})
		assert.ErrorIs(t, err, errUpstreamEmpty)
	})

	t.Run("propagates timeout", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			time.Sleep(200 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		client := NewChatCompleter(srv.URL, "k", "m", 30*time.Millisecond)
		_, err := client.Complete(context.Background(), []ChatMessage{{Role: "user", Content: "x"}})
		require.Error(t, err)
	})
}

func TestLLMProfileFor(t *testing.T) {
	cases := []struct {
		baseURL         string
		maxTokensField  string
		disableThinking bool
	}{
		{"https://api.openai.com/v1", "max_completion_tokens", false},
		{"https://api.x.ai/v1", "max_completion_tokens", false},
		{"https://api.z.ai/api/paas/v4", "max_tokens", true},
		{"https://open.bigmodel.cn/api/paas/v4", "max_tokens", true},
		{"http://localhost:18080/v1", "max_completion_tokens", false},
	}
	for _, tc := range cases {
		p := llmProfileFor(tc.baseURL)
		assert.Equal(t, tc.maxTokensField, p.maxTokensField, tc.baseURL)
		assert.Equal(t, tc.disableThinking, p.disableThinking, tc.baseURL)
	}
}

func TestBuildPayload_ZAI(t *testing.T) {
	completer := NewChatCompleter("https://api.z.ai/api/paas/v4", "k", "glm-4.5-flash", time.Second)
	require.NotNil(t, completer)
	client, ok := completer.(*openAICompatibleClient)
	require.True(t, ok)

	payload, err := client.buildPayload([]ChatMessage{{Role: "user", Content: "x"}})
	require.NoError(t, err)

	var body map[string]any
	require.NoError(t, json.Unmarshal(payload, &body))
	assert.Equal(t, "glm-4.5-flash", body["model"])
	// Z.AI は max_completion_tokens を黙殺するため max_tokens を使う（実測で確認済み）
	assert.Equal(t, float64(chatMaxCompletionTokens), body["max_tokens"])
	assert.NotContains(t, body, "max_completion_tokens")
	assert.Equal(t, map[string]any{"type": "disabled"}, body["thinking"])
}
