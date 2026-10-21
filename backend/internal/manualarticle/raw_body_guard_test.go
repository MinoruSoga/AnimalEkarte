package manualarticle

// raw_body_guard_test.go — PUT manual edit 専用 raw body guard の検証（NOT_RUN）。
// global sanitizer より前に登録される前提で、raw 0x00 / 不正 JSON は claim 前
// 400、上限超過は 413、検査通過 bytes は書換えず後段へ戻ることを確認する。

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const guardTestLimit = 1024

func guardResult(t *testing.T, target, body, contentType string, chunked bool) (int, []byte) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	var seen []byte
	r.Use(RawBodyGuard(guardTestLimit))
	r.PUT("/api/v1/manual/articles/:category/:slug", func(c *gin.Context) {
		var b bytes.Buffer
		_, _ = b.ReadFrom(c.Request.Body)
		seen = b.Bytes()
		c.Status(http.StatusOK)
	})
	r.PUT("/api/v1/other/:x", func(c *gin.Context) {
		var b bytes.Buffer
		_, _ = b.ReadFrom(c.Request.Body)
		seen = b.Bytes()
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, target, bytes.NewReader([]byte(body)))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if chunked {
		req.TransferEncoding = []string{"chunked"}
		req.ContentLength = -1
	}
	r.ServeHTTP(w, req)
	return w.Code, seen
}

func TestRawBodyGuard_RawNULIs400(t *testing.T) {
	raw := append([]byte(`{"operation_key":"k"`), 0x00)
	raw = append(raw, '}')
	code, _ := guardResult(t, "/api/v1/manual/articles/screens/intro", string(raw), "application/json", false)
	assert.Equal(t, http.StatusBadRequest, code)
}

func TestRawBodyGuard_InvalidJSONIs400(t *testing.T) {
	code, _ := guardResult(t, "/api/v1/manual/articles/screens/intro", `{not json`, "application/json", false)
	assert.Equal(t, http.StatusBadRequest, code)
}

func TestRawBodyGuard_NoContentTypeStillGuarded(t *testing.T) {
	raw := append([]byte(`{"a":`), 0x00)
	raw = append(raw, '}')
	code, _ := guardResult(t, "/api/v1/manual/articles/screens/intro", string(raw), "", false)
	assert.Equal(t, http.StatusBadRequest, code)
}

func TestRawBodyGuard_WrongContentTypeStillGuarded(t *testing.T) {
	code, _ := guardResult(t, "/api/v1/manual/articles/screens/intro", `{`, "text/plain", false)
	assert.Equal(t, http.StatusBadRequest, code)
}

func TestRawBodyGuard_ChunkedBodyStillGuarded(t *testing.T) {
	code, _ := guardResult(t, "/api/v1/manual/articles/screens/intro", `{`, "application/json", true)
	assert.Equal(t, http.StatusBadRequest, code)
}

func TestRawBodyGuard_OversizeIs413(t *testing.T) {
	body := string(bytes.Repeat([]byte("a"), guardTestLimit+1))
	code, _ := guardResult(t, "/api/v1/manual/articles/screens/intro", body, "application/json", false)
	assert.Equal(t, http.StatusRequestEntityTooLarge, code)
}

func TestRawBodyGuard_ValidBodyPassesByteIdentical(t *testing.T) {
	body := `{"operation_key":"k","article_id":"42","content":{"title":"t"}}`
	code, seen := guardResult(t, "/api/v1/manual/articles/screens/intro", body, "application/json", false)
	require.Equal(t, http.StatusOK, code)
	assert.Equal(t, body, string(seen))
}

func TestRawBodyGuard_OtherRouteUntouched(t *testing.T) {
	// 対象外 route は body を読まず通過 — raw NUL を含んでも guard は発動しない。
	raw := append([]byte{0x00, 0x01}, 'x')
	code, seen := guardResult(t, "/api/v1/other/1", string(raw), "application/octet-stream", false)
	require.Equal(t, http.StatusOK, code)
	assert.Equal(t, raw, seen)
}
