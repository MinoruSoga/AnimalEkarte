package middleware

import (
	"bytes"
	"io"
	"mime/multipart"
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

func TestSanitizeNullBytes_StreamingBodyLengthBecomesUnknown(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// streamingでは除去後の長さを事前計算せず、未知長としてdownstreamへ渡す。
	t.Run("ContentLength を未知長へ変更する", func(t *testing.T) {
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
		assert.Equal(t, int64(-1), c.Request.ContentLength)
	})
}

// X-1: multipart/form-data のバイナリボディはサニタイズ対象外とし、
// バイト単位で無変更のまま後続ハンドラに渡す必要がある。
// PNG シグネチャ（0x89 0x50 0x4E 0x47 0x0D 0x0A 0x1A 0x0A）の 0x1A は
// 制御文字除去範囲（0x0E-0x1F）に含まれるため、対象外にしないと画像が破損する。
func TestSanitizeNullBytes_MultipartBinaryUntouched(t *testing.T) {
	gin.SetMode(gin.TestMode)

	pngSignature := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("file", "signature.png")
	require.NoError(t, err)
	_, err = part.Write(pngSignature)
	require.NoError(t, err)
	require.NoError(t, mw.Close())

	originalBody := buf.Bytes()
	// 送信前にコピーを保持しておく（ミドルウェアが originalBody を書き換えないことの確認用）
	wantBody := append([]byte(nil), originalBody...)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, err := http.NewRequest(http.MethodPost, "/medical-records/1/images/upload", bytes.NewReader(originalBody))
	require.NoError(t, err)
	contentType := mw.FormDataContentType()
	req.Header.Set("Content-Type", "Multipart/Form-Data"+contentType[len("multipart/form-data"):])
	req.ContentLength = int64(len(originalBody))
	c.Request = req

	handler := SanitizeNullBytes()
	handler(c)

	gotBody, readErr := io.ReadAll(c.Request.Body)
	require.NoError(t, readErr)

	assert.Equal(t, wantBody, gotBody, "multipart body must pass through byte-exact (PNG signature must not be corrupted)")
}

func TestSanitizeNullBytes_DoesNotPreReadNonBinaryBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := &countingSanitizerReader{reader: bytes.NewReader(bytes.Repeat([]byte("x"), 1024*1024))}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, err := http.NewRequest(http.MethodPost, "/test", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "text/plain")
	req.ContentLength = -1
	c.Request = req

	SanitizeNullBytes()(c)

	assert.Zero(t, body.bytesRead, "global middleware must not pre-read request bodies")
}

func TestSanitizeNullBytes_ControlOnlyBodyRespectsRawByteLimit(t *testing.T) {
	// INF-02: filtered-byte-only bodies must still exhaust MaxBytesReader on raw reads.
	gin.SetMode(gin.TestMode)
	raw := bytes.Repeat([]byte{0x00}, int(DefaultJSONBodyMaxBytes)+1)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, err := http.NewRequest(http.MethodPost, "/test", bytes.NewReader(raw))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(raw))
	c.Request = req

	SanitizeNullBytes()(c)
	// Declared ContentLength above the cap aborts before wrapping.
	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
}

func TestSanitizeNullBytes_ChunkedControlOnlyBodyHitsMaxBytesReader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	raw := bytes.Repeat([]byte{0x01}, int(DefaultJSONBodyMaxBytes)+8)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, err := http.NewRequest(http.MethodPost, "/test", bytes.NewReader(raw))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = -1 // force stream path
	c.Request = req

	SanitizeNullBytes()(c)
	require.Equal(t, http.StatusOK, w.Code) // middleware itself does not abort on stream
	_, readErr := io.ReadAll(c.Request.Body)
	require.Error(t, readErr)
	var maxBytesError *http.MaxBytesError
	assert.ErrorAs(t, readErr, &maxBytesError)
}

func TestLimitRequestBody_RejectsOversizedDeclaredLength(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, err := http.NewRequest(http.MethodPost, "/api/v1/pets", bytes.NewReader([]byte(`{}`)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = DefaultJSONBodyMaxBytes + 1
	c.Request = req

	LimitRequestBody(DefaultJSONBodyMaxBytes)(c)
	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	assert.True(t, c.IsAborted())
}

type countingSanitizerReader struct {
	reader    io.Reader
	bytesRead int
}

func (r *countingSanitizerReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	r.bytesRead += n
	return n, err
}

// X-1 (security review follow-up): allowlist 方式（application/json のみサニタイズ）だと、
// c.ShouldBindJSON が Content-Type を無視して常に JSON としてボディを解釈するため、
// Content-Type を省略・誤指定したリクエストが NULL バイトを含んだまま素通りし、
// BUG-067 が再発する。blocklist 方式（バイナリ系のみ除外）でこれを防いでいることを検証する。
func TestSanitizeNullBytes_NonJSONContentTypeStillSanitized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		contentType string
	}{
		{name: "Content-Type: text/plain でも除去される", contentType: "text/plain"},
		{name: "Content-Type 未指定でも除去される", contentType: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			body := []byte("{\x00\"name\":\"田中\x00太郎\"}")
			wantBody := `{"name":"田中太郎"}`

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, err := http.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
			require.NoError(t, err)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			req.ContentLength = int64(len(body))
			c.Request = req

			// Act
			handler := SanitizeNullBytes()
			handler(c)

			// Assert
			gotBytes, readErr := io.ReadAll(c.Request.Body)
			require.NoError(t, readErr)
			assert.Equal(t, wantBody, string(gotBytes))
		})
	}
}
