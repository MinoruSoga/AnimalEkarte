package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS はCORSヘッダーを設定するミドルウェア。
// allowedOrigin は許可オリジンのカンマ区切り文字列（呼び出し側が config.CORSAllowedOrigin を注入する）。
// 例: http://localhost:3000,http://localhost:3001,https://reserve.noah-karte.com,https://liff.line.me
func CORS(allowedOrigin string) gin.HandlerFunc {
	if allowedOrigin == "" {
		// 開発環境デフォルト: 管理画面 + LIFF App (localhost:3001) + LINE内ブラウザ
		allowedOrigin = "http://localhost:3000,http://localhost:3001,https://liff.line.me"
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		// 設定されたオリジンのみ許可
		for allowed := range strings.SplitSeq(allowedOrigin, ",") {
			if strings.TrimSpace(allowed) == origin {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				c.Writer.Header().Set("Timing-Allow-Origin", origin)
				// SEC-601: キャッシュポイズニング対策（Origin別に異なるレスポンス）
				c.Writer.Header().Set("Vary", "Origin")
				break
			}
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		// SEC-601: X-Requested-With (CSRF対策ヘッダ)を許可
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID, X-Clinic-ID, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
