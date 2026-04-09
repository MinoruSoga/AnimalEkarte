package middleware

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS はCORSヘッダーを設定するミドルウェア。
// 許可オリジンは環境変数 CORS_ALLOWED_ORIGIN にカンマ区切りで指定する。
// 例: http://localhost:3000,http://localhost:3001,https://reserve.noah-karte.com,https://liff.line.me
func CORS() gin.HandlerFunc {
	allowedOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
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
				break
			}
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID, X-Clinic-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
