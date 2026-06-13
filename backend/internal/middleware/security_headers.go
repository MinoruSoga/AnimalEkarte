package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders は OWASP 推奨のセキュリティヘッダを全レスポンスに付与するミドルウェア。
// isProduction が true の場合は HSTS ヘッダも付与する（HTTPS 前提）。
func SecurityHeaders(isProduction bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'none'")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		// BUG-137: API レスポンスのキャッシュを禁止（個人データ保護）
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
		c.Header("Pragma", "no-cache")
		if isProduction {
			// HSTS は HTTPS 環境のみ有効にする
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}
		c.Next()
	}
}
