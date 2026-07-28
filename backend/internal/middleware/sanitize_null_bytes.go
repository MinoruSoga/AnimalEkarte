package middleware

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

// DefaultJSONBodyMaxBytes is the global raw-body ceiling for non-binary
// POST/PATCH/PUT traffic (INF-02 / POC-12 / X-07). Counts bytes before
// control-character sanitization so MaxBytesReader cannot be bypassed.
const DefaultJSONBodyMaxBytes int64 = 1 << 20 // 1 MiB

// SanitizeNullBytes は POST/PATCH/PUT リクエストのボディから NULL バイトおよび
// 制御文字を除去するミドルウェア。
// PostgreSQL は NULL バイト（\u0000）を含む文字列を拒否するため、
// DB エラー（500）を防ぐためにハンドラに渡す前に除去する。
//
// 除去対象:
//   - NULL バイト: \x00
//   - 制御文字: \x01-\x08, \x0B（垂直タブ）, \x0C（フォームフィード）, \x0E-\x1F
//
// 保持するもの:
//   - \x09（水平タブ）: JSON で使用される可能性があるため保持
//   - \x0A（LF）, \x0D（CR）: 改行文字として正常テキストで使用されるため保持
//
// BUG-067: フロントエンド（axios インターセプター）での除去と同様の処理を API レベルで実施。
//
// X-1: multipart/form-data・application/octet-stream 等のバイナリボディは対象外。
// テキスト用の制御文字除去をバイナリに適用すると、0x0E-0x1F 範囲のバイトが除去され、
// アップロードされた画像等のバイナリデータが破損する（例: PNG シグネチャの 7 バイト目 0x1A）。
// 該当する Content-Type のリクエストはボディを読み取り・書き換えせずそのまま後続ハンドラに渡す。
//
// 対象判定は「application/json のみ許可」という allowlist にしない: c.ShouldBindJSON
// は Content-Type を見ずにボディを常に JSON としてパースするため、allowlist にすると
// Content-Type を省略・誤指定・偽装された JSON リクエストがサニタイズを素通りし、
// BUG-067（NULL バイトが PostgreSQL に到達する不具合）が再発する。そのため、既知の
// バイナリ系 Content-Type のみを除外する blocklist とし、それ以外は全て対象に含める。
var binaryMediaTypes = []string{
	"multipart/form-data",
	"application/octet-stream",
}

func SanitizeNullBytes() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		if method != http.MethodPost && method != http.MethodPatch && method != http.MethodPut {
			c.Next()
			return
		}

		contentType := c.Request.Header.Get("Content-Type")
		if isBinaryContentType(contentType) {
			c.Next()
			return
		}

		if c.Request.Body == nil || c.Request.ContentLength == 0 {
			c.Next()
			return
		}

		// INF-02: reject oversized declared length before wrapping. Raw body
		// bytes (including later-discarded control bytes) are bounded by
		// MaxBytesReader under the sanitizer so filtered n cannot bypass the cap.
		if c.Request.ContentLength > DefaultJSONBodyMaxBytes {
			httpapi.RespondError(c, apperrors.WrapPayloadTooLarge("request body exceeds size limit"))
			c.Abort()
			return
		}

		// 認証・route固有のbody limitより前に全量をheapへ載せないよう、
		// downstreamが読む分だけin-placeで制御文字を除去する。
		// MaxBytesReader を source 側に置き、除去前の生バイト消費を計上する。
		limited := http.MaxBytesReader(c.Writer, c.Request.Body, DefaultJSONBodyMaxBytes)
		c.Request.Body = &sanitizedBodyReader{source: limited}
		c.Request.ContentLength = -1
		if getBody := c.Request.GetBody; getBody != nil {
			c.Request.GetBody = func() (io.ReadCloser, error) {
				body, err := getBody()
				if err != nil {
					return nil, err
				}
				return &sanitizedBodyReader{
					source: http.MaxBytesReader(c.Writer, body, DefaultJSONBodyMaxBytes),
				}, nil
			}
		}

		c.Next()
	}
}

// LimitRequestBody bounds raw request body size for non-binary write methods.
// Apply on authenticated API groups (POC-12) as defense-in-depth after global
// SanitizeNullBytes; MaxBytesReader still counts raw bytes when stacked under
// sanitizedBodyReader (see SanitizeNullBytes).
func LimitRequestBody(maxBytes int64) gin.HandlerFunc {
	if maxBytes <= 0 {
		maxBytes = DefaultJSONBodyMaxBytes
	}
	return func(c *gin.Context) {
		method := c.Request.Method
		if method != http.MethodPost && method != http.MethodPatch && method != http.MethodPut {
			c.Next()
			return
		}
		contentType := c.Request.Header.Get("Content-Type")
		if isBinaryContentType(contentType) {
			c.Next()
			return
		}
		if c.Request.Body == nil || c.Request.ContentLength == 0 {
			c.Next()
			return
		}
		if c.Request.ContentLength > maxBytes {
			httpapi.RespondError(c, apperrors.WrapPayloadTooLarge("request body exceeds size limit"))
			c.Abort()
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}

type sanitizedBodyReader struct {
	source io.ReadCloser
}

func (r *sanitizedBodyReader) Read(p []byte) (int, error) {
	for {
		n, err := r.source.Read(p)
		writeIndex := 0
		for _, b := range p[:n] {
			if isSanitizedControlByte(b) {
				continue
			}
			p[writeIndex] = b
			writeIndex++
		}
		if writeIndex > 0 || err != nil || n == 0 {
			return writeIndex, err //nolint:wrapcheck // transparent Reader must preserve io.EOF identity
		}
	}
}

func (r *sanitizedBodyReader) Close() error {
	if err := r.source.Close(); err != nil {
		return fmt.Errorf("close sanitized request body: %w", err)
	}
	return nil
}

func isSanitizedControlByte(b byte) bool {
	return b == 0x00 ||
		(b >= 0x01 && b <= 0x08) ||
		b == 0x0B ||
		b == 0x0C ||
		(b >= 0x0E && b <= 0x1F)
}

// isBinaryContentType は Content-Type がサニタイズ対象外のバイナリ系かどうかを判定する。
func isBinaryContentType(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		mediaType = strings.TrimSpace(strings.SplitN(contentType, ";", 2)[0])
	}
	for _, binaryMediaType := range binaryMediaTypes {
		if strings.EqualFold(mediaType, binaryMediaType) {
			return true
		}
	}
	return false
}
