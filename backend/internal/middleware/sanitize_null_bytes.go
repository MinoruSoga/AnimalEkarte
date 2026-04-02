package middleware

import (
	"bytes"
	"io"
	"net/http"

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
func SanitizeNullBytes() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		if method != http.MethodPost && method != http.MethodPatch && method != http.MethodPut {
			c.Next()
			return
		}

		if c.Request.Body == nil || c.Request.ContentLength == 0 {
			c.Next()
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Next()
			return
		}
		_ = c.Request.Body.Close()

		// NULL バイトおよび制御文字を除去。
		// bytes.Map は戻り値 -1 でその rune をドロップする。
		sanitized := bytes.Map(func(r rune) rune {
			switch {
			case r == 0x00: // NULL バイト
				return -1
			case r >= 0x01 && r <= 0x08: // SOH〜BS
				return -1
			case r == 0x0B: // 垂直タブ（VT）
				return -1
			case r == 0x0C: // フォームフィード（FF）
				return -1
			case r >= 0x0E && r <= 0x1F: // SO〜US
				return -1
			default:
				return r
			}
		}, body)

		// ボディを置き換え
		c.Request.Body = io.NopCloser(bytes.NewReader(sanitized))
		c.Request.ContentLength = int64(len(sanitized))

		c.Next()
	}
}
