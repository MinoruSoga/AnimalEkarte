package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SanitizeNullBytes は POST/PATCH/PUT リクエストのボディから NULL バイト（\u0000）を除去するミドルウェア。
// PostgreSQL は NULL バイトを含む文字列を拒否するため、DB エラー（500）を防ぐために
// ハンドラに渡す前に除去する。
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

		// NULL バイトを除去
		sanitized := bytes.ReplaceAll(body, []byte{0x00}, nil)

		// ボディを置き換え
		c.Request.Body = io.NopCloser(bytes.NewReader(sanitized))
		c.Request.ContentLength = int64(len(sanitized))

		c.Next()
	}
}
