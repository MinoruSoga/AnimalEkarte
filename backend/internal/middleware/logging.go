package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLoggingMiddleware リクエストロギングミドルウェア
func RequestLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		// 処理
		c.Next()

		// レイテンシ計算
		latency := time.Since(start)

		// クライアントIP
		clientIP := c.ClientIP()

		// ステータスコード
		statusCode := c.Writer.Status()

		// ログレベル判定
		logLevel := slog.LevelInfo
		if statusCode >= 400 && statusCode < 500 {
			logLevel = slog.LevelWarn
		} else if statusCode >= 500 {
			logLevel = slog.LevelError
		}

		// 構造化ログ出力（method, path, status, latency, request_id, client_ip のみ）
		requestIDVal, _ := c.Get("request_id")
		requestID, _ := requestIDVal.(string)
		logAttrs := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", statusCode),
			slog.Duration("latency", latency),
			slog.String("request_id", requestID),
			slog.String("client_ip", clientIP),
		}

		// エラーがある場合は追加
		if len(c.Errors) > 0 {
			logAttrs = append(logAttrs, slog.String("error", c.Errors.String()))
		}

		message := "request completed"
		if statusCode >= 500 {
			message = "server error"
		} else if statusCode >= 400 {
			message = "client error"
		}

		// slog.Log に渡すために変換
		args := make([]any, len(logAttrs)*2)
		for i, attr := range logAttrs {
			args[i*2] = attr.Key
			args[i*2+1] = attr.Value
		}

		slog.Log(c.Request.Context(), logLevel, message, args...)
	}
}

// RequestID はリクエストごとに一意のIDを付与するミドルウェア
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		// BUG-147: ユーザー入力をサニタイズ（英数字・ハイフンのみ、64文字以内）
		if requestID == "" || len(requestID) > 64 || !isAlphanumericHyphen(requestID) {
			b := make([]byte, 4)
			_, _ = rand.Read(b)
			requestID = hex.EncodeToString(b)
		}

		c.Set("request_id", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)

		c.Next()
	}
}

// isAlphanumericHyphen は文字列が英数字・ハイフン・アンダースコアのみで構成されているか判定する
func isAlphanumericHyphen(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '-' && r != '_' {
			return false
		}
	}
	return true
}
