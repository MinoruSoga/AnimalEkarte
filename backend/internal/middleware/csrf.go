package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireXRequestedWith は POST/PATCH/DELETE に対して X-Requested-With ヘッダを強制する CSRF 対策 middleware。
// SEC-601: クロスサイト偽装リクエストの防止。
func RequireXRequestedWith() gin.HandlerFunc {
	return func(c *gin.Context) {
		// GET, HEAD, OPTIONS は除外（冪等メソッド）
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// X-Requested-With チェック（XMLHttpRequest が標準値）
		if c.Request.Header.Get("X-Requested-With") == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "X-Requested-With header required for state-changing requests"})
			return
		}

		c.Next()
	}
}
