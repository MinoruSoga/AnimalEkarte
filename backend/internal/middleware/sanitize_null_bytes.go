package middleware

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

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

		// 認証・route固有のbody limitより前に全量をheapへ載せないよう、
		// downstreamが読む分だけin-placeで制御文字を除去する。
		c.Request.Body = &sanitizedBodyReader{source: c.Request.Body}
		c.Request.ContentLength = -1
		if getBody := c.Request.GetBody; getBody != nil {
			c.Request.GetBody = func() (io.ReadCloser, error) {
				body, err := getBody()
				if err != nil {
					return nil, err
				}
				return &sanitizedBodyReader{source: body}, nil
			}
		}

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
