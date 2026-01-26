package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLoggingMiddleware リクエストロギングミドルウェア
func RequestLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// リクエストボディを読み取り（POST/PUTリクエストの場合）
		var requestBody []byte
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			if c.Request.Body != nil {
				requestBody, _ = io.ReadAll(c.Request.Body)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
			}
		}

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

		// パスとクエリ
		if raw != "" {
			path = path + "?" + raw
		}

		// 構造化ログ出力
		logAttrs := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status_code", statusCode),
			slog.Duration("latency", latency),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", c.Request.UserAgent()),
		}

		// リクエストボディをログに追加（サイズ制限付き）
		if len(requestBody) > 0 && len(requestBody) < 1024 {
			logAttrs = append(logAttrs, slog.String("request_body", string(requestBody)))
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

		// slog.Logに渡すために変換
		args := make([]any, len(logAttrs)*2)
		for i, attr := range logAttrs {
			args[i*2] = attr.Key
			args[i*2+1] = attr.Value
		}

		slog.Log(c.Request.Context(), logLevel, message, args...)
	}
}

// RequestID adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	// Simple implementation - in production, consider using UUID
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
