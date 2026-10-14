// Package middleware — EMR-332 C18 evidence tests
// (docs/work/emr/emr-332-security/workflow-audit/c18-logging-audit.md).
//
// Scope: deterministic, single-threaded, in-memory verification of
// middleware.RequestID + middleware.RequestLoggingMiddleware. slog output is
// captured in a bytes.Buffer (no persistent log). Marker convention:
// SECCHK20261009-style alphanumeric sentinels — request IDs, bearer values,
// query tokens — which must appear only where the field contract allows.
package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"

	"github.com/animal-ekarte/backend/internal/middleware"
)

func captureRequestLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return buf
}

// lastLogRecord returns the final JSON slog record in buf as a key→value map.
func lastLogRecord(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
	require.NotEmpty(t, lines)
	rec := map[string]any{}
	require.NoError(t, json.Unmarshal(lines[len(lines)-1], &rec))
	return rec
}

// ── C18: request_id sanitization ─────────────────────────────────────────────

func TestC18_RequestIDSanitization(t *testing.T) {
	cases := []struct {
		name        string
		header      string
		wantEcho    bool   // caller-supplied value is preserved verbatim
		wantPattern string // regex-free shape check for generated IDs
	}{
		{"valid marker preserved", "SECCHK20261009ALPHA1", true, ""},
		{"hyphen+underscore preserved", "SECCHK20261009-B_2", true, ""},
		{"missing header generates", "", false, "hex8"},
		{"space rejected", "SECCHK 20261009", false, "hex8"},
		{"semicolon rejected", "SECCHK20261009;X", false, "hex8"},
		{"quote rejected", "SECCHK20261009\"X", false, "hex8"},
		{"slash rejected", "SECCHK20261009/X", false, "hex8"},
		{"65 chars rejected", strings.Repeat("A", 65), false, "hex8"},
		{"64 chars accepted", strings.Repeat("B", 64), true, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var contextRequestID any
			r := newEngine()
			r.GET("/rid", middleware.RequestID(), func(c *gin.Context) {
				contextRequestID, _ = c.Get("request_id")
				c.Status(http.StatusNoContent)
			})

			req := httptest.NewRequest(http.MethodGet, "/rid", nil)
			if tc.header != "" {
				req.Header.Set("X-Request-ID", tc.header)
			}
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			got := rec.Header().Get("X-Request-ID")
			require.NotEmpty(t, got)
			assert.Equal(t, got, contextRequestID, "response header and context key must agree")

			if tc.wantEcho {
				assert.Equal(t, tc.header, got)
				return
			}
			// generated: 8 lowercase hex chars from 4 random bytes
			assert.Len(t, got, 8)
			for _, ch := range got {
				assert.True(t, (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f'),
					"generated request_id must be hex, got %q", got)
			}
		})
	}
}

// ── C18: structured fields + severity mapping + secret/PII exclusion ─────────

func TestC18_RequestLoggingFieldsAndSecretExclusion(t *testing.T) {
	cases := []struct {
		name       string
		status     int
		wantLevel  string
		wantMsg    string
		handlerErr bool
	}{
		{"2xx -> INFO request completed", http.StatusOK, "INFO", "request completed", false},
		{"4xx -> WARN client error", http.StatusNotFound, "WARN", "client error", false},
		{"5xx -> ERROR server error + error attr", http.StatusInternalServerError, "ERROR", "server error", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := captureRequestLogs(t)
			r := newEngine()
			r.Use(middleware.RequestID(), middleware.RequestLoggingMiddleware())
			r.GET("/sekck-check", func(c *gin.Context) {
				if tc.handlerErr {
					_ = c.Error(assert.AnError)
				}
				c.Status(tc.status)
			})

			req := httptest.NewRequest(http.MethodGet, "/sekck-check?token=SECCHK20261009QRY&owner_name=SECCHK20261009PII", nil)
			req.RemoteAddr = testClientIP + ":443"
			// Sentinels that must NEVER reach the log record.
			req.Header.Set("X-Request-ID", "SECCHK20261009REQID1")
			req.Header.Set("Authorization", "Bearer SECCHK20261009BEARER")
			req.Header.Set("Cookie", "session=SECCHK20261009COOKIE")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			record := lastLogRecord(t, buf)
			assert.Equal(t, tc.wantLevel, record["level"])
			assert.Equal(t, tc.wantMsg, record["msg"])
			assert.Equal(t, "GET", record["method"])
			assert.Equal(t, "/sekck-check", record["path"], "path must exclude the query string")
			assert.Equal(t, float64(tc.status), record["status"])
			assert.Equal(t, "SECCHK20261009REQID1", record["request_id"])
			assert.Equal(t, testClientIP, record["client_ip"])
			assert.Contains(t, record, "latency")

			// Field contract: the documented allowlist only.
			allowed := map[string]bool{
				"time": true, "level": true, "msg": true,
				"method": true, "path": true, "status": true, "latency": true,
				"request_id": true, "client_ip": true, "error": true,
			}
			for k := range record {
				assert.True(t, allowed[k], "unexpected log field %q", k)
			}
			if tc.handlerErr {
				assert.Contains(t, record, "error")
			}

			line := buf.String()
			for _, sentinel := range []string{
				"SECCHK20261009BEARER", // Authorization header value
				"SECCHK20261009COOKIE", // Cookie header value
				"SECCHK20261009QRY",    // query parameter value
				"SECCHK20261009PII",    // owner-name-shaped query parameter
			} {
				assert.NotContains(t, line, sentinel, "log record must not carry %q", sentinel)
			}
		})
	}
}

// ── C18: marker-tagged refusal events are captured ──────────────────────────

func TestC18_MarkerTaggedRefusalEventCaptured(t *testing.T) {
	buf := captureRequestLogs(t)
	store := newStore(t)
	r := newEngine()
	// logging outermost -> sees final 429 status; RequestID stamps the marker.
	r.Use(middleware.RequestLoggingMiddleware(), middleware.RequestID())
	r.GET("/guarded", middleware.RateLimit(store, rate.Every(time.Hour), 1), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	const marker = "SECCHK20261009REFUSAL"
	do := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/guarded", nil)
		req.RemoteAddr = testClientIP + ":443"
		req.Header.Set("X-Request-ID", marker)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		return rec
	}

	assert.Equal(t, http.StatusOK, do().Code)
	buf.Reset()
	assert.Equal(t, http.StatusTooManyRequests, do().Code)

	record := lastLogRecord(t, buf)
	assert.Equal(t, "WARN", record["level"], "429 refusal must be logged as a client error")
	assert.Equal(t, "client error", record["msg"])
	assert.Equal(t, marker, record["request_id"], "marker must survive as the refusal event's request_id")
	assert.Equal(t, float64(http.StatusTooManyRequests), record["status"])
}

// ── C18: request body content is not logged ─────────────────────────────────

func TestC18_RequestBodyNotLogged(t *testing.T) {
	buf := captureRequestLogs(t)
	r := newEngine()
	r.Use(middleware.RequestLoggingMiddleware())
	r.POST("/submit", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/submit",
		strings.NewReader(`{"secret":"secchk-u06","note":"x"}`))
	req.RemoteAddr = testClientIP + ":443"
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotContains(t, buf.String(), "secchk-u06",
		"request body must never reach the request log")
}
