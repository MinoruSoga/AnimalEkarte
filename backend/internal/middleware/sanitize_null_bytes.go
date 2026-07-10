package middleware

import (
	"bytes"
	"io"
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
// bytes.Map（UTF-8 コードポイント単位の走査）をバイナリに適用すると、不正な UTF-8
// バイト列が U+FFFD に置換されたり、0x0E-0x1F 範囲のバイトが誤って除去されたりして、
// アップロードされた画像等のバイナリデータが破損する（例: PNG シグネチャの 7 バイト目 0x1A）。
// 該当する Content-Type のリクエストはボディを読み取り・書き換えせずそのまま後続ハンドラに渡す。
//
// 対象判定は「application/json のみ許可」という allowlist にしない: c.ShouldBindJSON
// は Content-Type を見ずにボディを常に JSON としてパースするため、allowlist にすると
// Content-Type を省略・誤指定・偽装された JSON リクエストがサニタイズを素通りし、
// BUG-067（NULL バイトが PostgreSQL に到達する不具合）が再発する。そのため、既知の
// バイナリ系 Content-Type のみを除外する blocklist とし、それ以外は全て対象に含める。
var binaryContentTypePrefixes = []string{
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

// isBinaryContentType は Content-Type がサニタイズ対象外のバイナリ系かどうかを判定する。
func isBinaryContentType(contentType string) bool {
	for _, prefix := range binaryContentTypePrefixes {
		if strings.HasPrefix(contentType, prefix) {
			return true
		}
	}
	return false
}
