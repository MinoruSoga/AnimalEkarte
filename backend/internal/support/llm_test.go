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
		var gotBody chatCompletionRequest
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
		assert.Equal(t, "gpt-5-nano", gotBody.Model)
		assert.Equal(t, chatMaxCompletionTokens, gotBody.MaxCompletionTokens)
		require.Len(t, gotBody.Messages, 2)
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
