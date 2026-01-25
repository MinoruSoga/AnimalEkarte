package middleware

import (
	"bytes"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// responseBodyWriter is a wrapper around gin.ResponseWriter to capture response body
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// RequestLogger returns a structured logging middleware
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get client IP
		clientIP := c.ClientIP()

		// Get status code
		statusCode := c.Writer.Status()

		// Get request method
		method := c.Request.Method

		// Build full path
		if raw != "" {
			path = path + "?" + raw
		}

		// Log the request
		logger := slog.With(
			slog.String("method", method),
			slog.String("path", path),
			slog.String("client_ip", clientIP),
			slog.Int("status_code", statusCode),
			slog.Duration("latency", latency),
			slog.Int("body_size", c.Writer.Size()),
		)

		if statusCode >= 500 {
			logger.Error("server error")
		} else if statusCode >= 400 {
			logger.Warn("client error")
		} else {
			logger.Info("request completed")
		}
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
