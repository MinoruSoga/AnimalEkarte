package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeNullBytes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		method      string
		body        []byte
		contentType string
		wantBody    string
	}{
		{
			name:        "NULL バイトを除去する",
			method:      http.MethodPost,
			body:        []byte("{\x00\"name\":\"田中\x00太郎\"}"),
			contentType: "application/json",
			wantBody:    `{"name":"田中太郎"}`,
		},
		{
			name:        "複数の NULL バイトを除去する",
			method:      http.MethodPost,
			body:        []byte("{\x00\"name\x00\":\"test\x00\"}"),
			contentType: "application/json",
			wantBody:    `{"name":"test"}`,
		},
		{
			name:        "制御文字 \\x01-\\x08 を除去する",
			method:      http.MethodPatch,
			body:        []byte("name\x01\x02\x03\x04\x05\x06\x07\x08value"),
			contentType: "application/json",
			wantBody:    "namevalue",
		},
		{
			name:        "垂直タブ（\\x0B）を除去する",
			method:      http.MethodPatch,
			body:        []byte("name\x0Bvalue"),
			contentType: "application/json",
			wantBody:    "namevalue",
		},
		{
			name:        "フォームフィード（\\x0C）を除去する",
			method:      http.MethodPut,
			body:        []byte("name\x0Cvalue"),
			contentType: "application/json",
			wantBody:    "namevalue",
		},
		{
			name:        "制御文字 \\x0E-\\x1F を除去する",
			method:      http.MethodPost,
			body:        []byte("name\x0E\x0F\x10\x1Fvalue"),
			contentType: "application/json",
			wantBody:    "namevalue",
		},
		{
			name:        "通常の日本語テキストは変更されない",
			method:      http.MethodPost,
			body:        []byte(`{"name":"佐藤花子","address":"東京都渋谷区"}`),
			contentType: "application/json",
			wantBody:    `{"name":"佐藤花子","address":"東京都渋谷区"}`,
		},
		{
			name:        "タブ（\\x09）は保持される",
			method:      http.MethodPost,
			body:        []byte("name\tvalue"),
			contentType: "application/json",
			wantBody:    "name\tvalue",
		},
		{
			name:        "改行（\\x0A）は保持される",
			method:      http.MethodPost,
			body:        []byte("name\nvalue"),
			contentType: "application/json",
			wantBody:    "name\nvalue",
		},
		{
			name:        "キャリッジリターン（\\x0D）は保持される",
			method:      http.MethodPost,
			body:        []byte("name\rvalue"),
			contentType: "application/json",
			wantBody:    "name\rvalue",
		},
		{
			name:        "GET リクエストはボディを変更しない",
			method:      http.MethodGet,
			body:        []byte("name\x00value"),
			contentType: "application/json",
			wantBody:    "name\x00value",
		},
		{
			name:        "DELETE リクエストはボディを変更しない",
			method:      http.MethodDelete,
			body:        []byte("name\x00value"),
			contentType: "application/json",
			wantBody:    "name\x00value",
		},
		{
			name:        "空ボディはエラーなしで通過する",
			method:      http.MethodPost,
			body:        nil,
			contentType: "application/json",
			wantBody:    "",
		},
		{
			name:        "ASCII 英数字は変更されない",
			method:      http.MethodPost,
			body:        []byte(`{"name":"John Doe 123","email":"john@example.com"}`),
			contentType: "application/json",
			wantBody:    `{"name":"John Doe 123","email":"john@example.com"}`,
		},
		{
			name:        "絵文字は変更されない",
			method:      http.MethodPost,
			body:        []byte(`{"note":"🐕🐈"}`),
			contentType: "application/json",
			wantBody:    `{"note":"🐕🐈"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			var bodyReader io.Reader
			if tt.body != nil {
				bodyReader = bytes.NewReader(tt.body)
			} else {
				bodyReader = http.NoBody
			}

			req, err := http.NewRequest(tt.method, "/test", bodyReader)
			require.NoError(t, err)
			req.Header.Set("Content-Type", tt.contentType)
			if tt.body != nil {
				req.ContentLength = int64(len(tt.body))
			}
			c.Request = req

			// Act
			handler := SanitizeNullBytes()
			handler(c)

			// Assert: ボディを読み取って検証
			var gotBody string
			if c.Request.Body != nil {
				gotBytes, readErr := io.ReadAll(c.Request.Body)
				require.NoError(t, readErr)
				gotBody = string(gotBytes)
			}

			assert.Equal(t, tt.wantBody, gotBody)
		})
	}
}

func TestSanitizeNullBytes_ContentLength(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// NULL バイト除去後に ContentLength が更新されることを確認
	t.Run("ContentLength が除去後のバイト数に更新される", func(t *testing.T) {
		originalBody := []byte("abc\x00def") // 8 バイト
		expectedBody := "abcdef"             // 6 バイト

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req, err := http.NewRequest(http.MethodPost, "/test", bytes.NewReader(originalBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.ContentLength = int64(len(originalBody))
		c.Request = req

		handler := SanitizeNullBytes()
		handler(c)

		gotBytes, err := io.ReadAll(c.Request.Body)
		require.NoError(t, err)

		assert.Equal(t, expectedBody, string(gotBytes))
		assert.Equal(t, int64(len(expectedBody)), c.Request.ContentLength)
	})
}
