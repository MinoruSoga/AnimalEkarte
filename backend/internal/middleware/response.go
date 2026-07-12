package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// respondError は auth.go/liff_auth.go 系 middleware 共通のエラーレスポンスを返す
// （code/message/timestamp）。handler 層の RespondError（`{"error": msg}`）とは
// スキーマが異なる — 統一判断は BE-refactor.md §4 参照（X-17）。
func respondError(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, gin.H{
		"code":      status,
		"message":   msg,
		"timestamp": time.Now(),
	})
}
