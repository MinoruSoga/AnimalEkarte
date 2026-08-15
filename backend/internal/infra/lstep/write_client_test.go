package lstep

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAPIKey = "test-secret-api-key-never-log"

type capturedRequest struct {
	Method      string
	Path        string
	ContentType string
	Body        []byte
	AuthHeader  string
}

func newWriteTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *atomic.Int32, *[]capturedRequest) {
	t.Helper()
	var hits atomic.Int32
	var captured []capturedRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		body, _ := io.ReadAll(r.Body)
		captured = append(captured, capturedRequest{
			Method:      r.Method,
			Path:        r.URL.EscapedPath(),
			ContentType: r.Header.Get("Content-Type"),
			Body:        append([]byte(nil), body...),
			AuthHeader:  r.Header.Get("Authorization"),
		})
		if handler != nil {
			handler(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)
	return server, &hits, &captured
}

func enableWriteGate(t *testing.T) {
	t.Helper()
	t.Setenv(EnvWriteAPIEnabled, "true")
}

func newTestClient(baseURL string) Client {
	return NewClient(testAPIKey, baseURL)
}

func assertNoSecretLeak(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		return
	}
	msg := err.Error()
	assert.NotContains(t, msg, testAPIKey)
	assert.NotContains(t, msg, "Bearer ")
	assert.NotContains(t, strings.ToLower(msg), "authorization")
}

// ---- deploy gate ----

func TestWrite_DeployGateDisabledVariants(t *testing.T) {
	variants := []struct {
		name  string
		setup func(t *testing.T)
	}{
		{"unset", func(t *testing.T) { t.Setenv(EnvWriteAPIEnabled, "") }},
		{"empty_explicit", func(t *testing.T) { t.Setenv(EnvWriteAPIEnabled, "") }},
		{"false", func(t *testing.T) { t.Setenv(EnvWriteAPIEnabled, "false") }},
		{"True_case", func(t *testing.T) { t.Setenv(EnvWriteAPIEnabled, "True") }},
		{"TRUE", func(t *testing.T) { t.Setenv(EnvWriteAPIEnabled, "TRUE") }},
		{"1", func(t *testing.T) { t.Setenv(EnvWriteAPIEnabled, "1") }},
		{"yes", func(t *testing.T) { t.Setenv(EnvWriteAPIEnabled, "yes") }},
		{"unknown", func(t *testing.T) { t.Setenv(EnvWriteAPIEnabled, "enabled") }},
	}

	for _, tt := range variants {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)
			server, hits, _ := newWriteTestServer(t, nil)
			client := newTestClient(server.URL)
			ctx := context.Background()

			assert.ErrorIs(t, client.AddTag(ctx, "Uabc", "tag"), ErrWriteDisabled)
			assert.ErrorIs(t, client.RemoveTag(ctx, "Uabc", "tag"), ErrWriteDisabled)
			assert.ErrorIs(t, client.AddTagBulk(ctx, []string{"Uabc"}, "tag"), ErrWriteDisabled)
			assert.ErrorIs(t, client.SetProperty(ctx, "Uabc", "k", "v"), ErrWriteDisabled)
			// empty bulk still fails closed on gate (not nil success)
			assert.ErrorIs(t, client.AddTagBulk(ctx, nil, "tag"), ErrWriteDisabled)

			assert.Equal(t, int32(0), hits.Load(), "deploy gate false must issue 0 HTTP requests")
		})
	}
}

func TestWrite_DeployGateExactTrueEnables(t *testing.T) {
	enableWriteGate(t)
	server, hits, captured := newWriteTestServer(t, nil)
	client := newTestClient(server.URL)
	ctx := context.Background()

	require.NoError(t, client.AddTag(ctx, "Uabc", "health_checkup"))
	assert.Equal(t, int32(1), hits.Load())
	require.Len(t, *captured, 1)
	assert.Equal(t, http.MethodPost, (*captured)[0].Method)
	assert.Equal(t, "/contacts/Uabc/tags", (*captured)[0].Path)
}

// ---- four write methods: method/path/body/content-type/escaping ----

func TestWrite_AddTag_RequestShape(t *testing.T) {
	enableWriteGate(t)
	server, hits, captured := newWriteTestServer(t, nil)
	client := newTestClient(server.URL)

	// path segment with reserved chars must be escaped
	rawID := "U/a b?x=1"
	require.NoError(t, client.AddTag(context.Background(), rawID, "tag_a"))
	assert.Equal(t, int32(1), hits.Load())
	require.Len(t, *captured, 1)
	got := (*captured)[0]
	assert.Equal(t, http.MethodPost, got.Method)
	assert.Equal(t, "/contacts/"+url.PathEscape(rawID)+"/tags", got.Path)
	assert.Equal(t, "application/json", got.ContentType)
	var body addTagRequest
	require.NoError(t, json.Unmarshal(got.Body, &body))
	assert.Equal(t, "tag_a", body.TagName)
	assert.Equal(t, "Bearer "+testAPIKey, got.AuthHeader)
}

func TestWrite_RemoveTag_RequestShape(t *testing.T) {
	enableWriteGate(t)
	server, hits, captured := newWriteTestServer(t, nil)
	client := newTestClient(server.URL)

	rawID := "U%special/id"
	require.NoError(t, client.RemoveTag(context.Background(), rawID, "tag_b"))
	assert.Equal(t, int32(1), hits.Load())
	require.Len(t, *captured, 1)
	got := (*captured)[0]
	assert.Equal(t, http.MethodDelete, got.Method)
	assert.Equal(t, "/contacts/"+url.PathEscape(rawID)+"/tags", got.Path)
	assert.Equal(t, "application/json", got.ContentType)
	var body removeTagRequest
	require.NoError(t, json.Unmarshal(got.Body, &body))
	assert.Equal(t, "tag_b", body.TagName)
}

func TestWrite_AddTagBulk_RequestShape(t *testing.T) {
	enableWriteGate(t)
	server, hits, captured := newWriteTestServer(t, nil)
	client := newTestClient(server.URL)

	require.NoError(t, client.AddTagBulk(context.Background(), []string{"U1", "U2"}, "bulk_tag"))
	assert.Equal(t, int32(1), hits.Load())
	require.Len(t, *captured, 1)
	got := (*captured)[0]
	assert.Equal(t, http.MethodPost, got.Method)
	assert.Equal(t, "/contacts/tags/bulk", got.Path)
	assert.Equal(t, "application/json", got.ContentType)
	var body addTagBulkRequest
	require.NoError(t, json.Unmarshal(got.Body, &body))
	assert.Equal(t, []string{"U1", "U2"}, body.LineUserIDs)
	assert.Equal(t, "bulk_tag", body.TagName)
}

func TestWrite_AddTagBulk_EmptyIDsNoHTTP(t *testing.T) {
	enableWriteGate(t)
	server, hits, _ := newWriteTestServer(t, nil)
	client := newTestClient(server.URL)

	require.NoError(t, client.AddTagBulk(context.Background(), nil, "bulk_tag"))
	require.NoError(t, client.AddTagBulk(context.Background(), []string{}, "bulk_tag"))
	assert.Equal(t, int32(0), hits.Load())
}

func TestWrite_SetProperty_RequestShape(t *testing.T) {
	enableWriteGate(t)
	server, hits, captured := newWriteTestServer(t, nil)
	client := newTestClient(server.URL)

	rawID := "U path/id"
	require.NoError(t, client.SetProperty(context.Background(), rawID, "breed", "mix"))
	assert.Equal(t, int32(1), hits.Load())
	require.Len(t, *captured, 1)
	got := (*captured)[0]
	assert.Equal(t, http.MethodPost, got.Method)
	assert.Equal(t, "/contacts/"+url.PathEscape(rawID)+"/properties", got.Path)
	assert.Equal(t, "application/json", got.ContentType)
	var body setPropertyRequest
	require.NoError(t, json.Unmarshal(got.Body, &body))
	assert.Equal(t, "breed", body.Key)
	assert.Equal(t, "mix", body.Value)
}

// ---- status / sentinel / retry / cancellation ----

func TestWrite_EmptyLineUserID_NoHTTP(t *testing.T) {
	enableWriteGate(t)
	server, hits, _ := newWriteTestServer(t, nil)
	client := newTestClient(server.URL)
	ctx := context.Background()

	assert.ErrorIs(t, client.AddTag(ctx, "", "t"), ErrUserNotFound)
	assert.ErrorIs(t, client.RemoveTag(ctx, "", "t"), ErrUserNotFound)
	assert.ErrorIs(t, client.SetProperty(ctx, "", "k", "v"), ErrUserNotFound)
	assert.Equal(t, int32(0), hits.Load())
}

func TestWrite_2xxSuccess(t *testing.T) {
	enableWriteGate(t)
	for _, status := range []int{http.StatusOK, http.StatusCreated, http.StatusNoContent} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server, hits, _ := newWriteTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(status)
			})
			client := newTestClient(server.URL)
			require.NoError(t, client.AddTag(context.Background(), "Uok", "t"))
			assert.Equal(t, int32(1), hits.Load())
		})
	}
}

func TestWrite_404_UserNotFound(t *testing.T) {
	enableWriteGate(t)
	server, hits, _ := newWriteTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"secret":"` + testAPIKey + `"}`))
	})
	client := newTestClient(server.URL)

	const missingID = "Umissing"

	err := client.AddTag(context.Background(), missingID, "t")
	assert.ErrorIs(t, err, ErrUserNotFound)
	assert.NotContains(t, err.Error(), missingID, "404 error must not embed lineUserID")
	assertNoSecretLeak(t, err)
	assert.Equal(t, int32(1), hits.Load())

	err = client.RemoveTag(context.Background(), missingID, "t")
	assert.ErrorIs(t, err, ErrUserNotFound)
	assert.NotContains(t, err.Error(), missingID, "404 error must not embed lineUserID")
	assertNoSecretLeak(t, err)

	err = client.SetProperty(context.Background(), missingID, "k", "v")
	assert.ErrorIs(t, err, ErrUserNotFound)
	assert.NotContains(t, err.Error(), missingID, "404 error must not embed lineUserID")
	assertNoSecretLeak(t, err)
}

func TestGetUser_404(t *testing.T) {
	server, hits, _ := newWriteTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"secret":"` + testAPIKey + `"}`))
	})
	client := newTestClient(server.URL)

	const missingID = "U404"
	info, err := client.GetUser(context.Background(), missingID)
	assert.Nil(t, info)
	assert.ErrorIs(t, err, ErrUserNotFound)
	assert.True(t, IsUserNotFound(err))
	assert.NotContains(t, err.Error(), missingID, "404 error must not embed lineUserID")
	assertNoSecretLeak(t, err)
	assert.Equal(t, int32(1), hits.Load())

	// GetUserTags 404 も同様に redaction
	tags, err := client.GetUserTags(context.Background(), missingID)
	assert.Nil(t, tags)
	assert.ErrorIs(t, err, ErrUserNotFound)
	assert.True(t, IsUserNotFound(err))
	assert.NotContains(t, err.Error(), missingID, "404 error must not embed lineUserID")
	assertNoSecretLeak(t, err)
}

func TestWrite_Non2xx_ObservableStatus(t *testing.T) {
	enableWriteGate(t)
	server, hits, _ := newWriteTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal","token":"` + testAPIKey + `"}`))
	})
	client := newTestClient(server.URL)

	err := client.AddTag(context.Background(), "Uerr", "t")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status=500")
	assertNoSecretLeak(t, err)
	assert.Equal(t, int32(1), hits.Load())

	err = client.AddTagBulk(context.Background(), []string{"Uerr"}, "t")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status=500")
	assertNoSecretLeak(t, err)
}

func TestWrite_RateLimitRetryExhaustion(t *testing.T) {
	enableWriteGate(t)
	// maxRetries=3 → attempts 0..3 = 4 requests. Waits are real (1s+2s+4s).
	server, hits, _ := newWriteTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"retry":true,"key":"` + testAPIKey + `"}`))
	})
	client := newTestClient(server.URL)

	start := time.Now()
	err := client.AddTag(context.Background(), "Urate", "t")
	elapsed := time.Since(start)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrRateLimit), "got %v", err)
	assertNoSecretLeak(t, err)
	assert.Equal(t, int32(maxRetries+1), hits.Load(), "exhausted retries must issue maxRetries+1 requests")
	assert.GreaterOrEqual(t, elapsed, time.Second, "backoff should wait at least initial interval")
}

func TestWrite_ContextCancellation(t *testing.T) {
	enableWriteGate(t)
	server, hits, _ := newWriteTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		// Always 429 so doWithRetry waits; cancel during wait.
		w.WriteHeader(http.StatusTooManyRequests)
	})
	client := newTestClient(server.URL)

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel after first 429 is observed — use short timer.
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := client.AddTag(ctx, "Ucancel", "t")
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "canceled"),
		"expected cancellation, got %v", err)
	assertNoSecretLeak(t, err)
	// At least the first attempt happened; may be 1 (cancel during wait) .
	assert.GreaterOrEqual(t, hits.Load(), int32(1))
}

func TestWrite_ContextAlreadyCanceled(t *testing.T) {
	enableWriteGate(t)
	server, hits, _ := newWriteTestServer(t, nil)
	client := newTestClient(server.URL)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := client.AddTag(ctx, "Upre", "t")
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "canceled"),
		"expected cancellation, got %v", err)
	// Request may or may not leave the client depending on transport; assert no secret leak.
	assertNoSecretLeak(t, err)
	_ = hits
}

func TestWrite_IsWriteDisabledHelper(t *testing.T) {
	assert.True(t, IsWriteDisabled(ErrWriteDisabled))
	assert.True(t, IsWriteDisabled(errors.Join(errors.New("wrap"), ErrWriteDisabled)))
	assert.False(t, IsWriteDisabled(ErrUserNotFound))
	assert.False(t, IsWriteDisabled(nil))
}
