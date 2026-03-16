package middleware

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimitStore IP別のレートリミッター管理
type RateLimitStore struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
}

// NewRateLimitStore はRateLimitStoreを初期化して返す
func NewRateLimitStore() *RateLimitStore {
	return &RateLimitStore{
		limiters: make(map[string]*rate.Limiter),
	}
}

// getLimiter はIPアドレスに対応するlimiterを取得・作成する
func (s *RateLimitStore) getLimiter(ip string, rps rate.Limit, burst int) *rate.Limiter {
	s.mu.RLock()
	limiter, exists := s.limiters[ip]
	s.mu.RUnlock()

	if exists {
		return limiter
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 二重チェック
	if limiter, exists := s.limiters[ip]; exists {
		return limiter
	}

	limiter = rate.NewLimiter(rps, burst)
	s.limiters[ip] = limiter
	return limiter
}

// RateLimit は指定されたレートでIPアドレスごとにレート制限を行う
// rps: requests per second, burst: バースト許容量
func RateLimit(store *RateLimitStore, rps rate.Limit, burst int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := getClientIP(c)
		limiter := store.getLimiter(ip, rps, burst)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": fmt.Sprintf("rate limit exceeded: %d requests per second max", int(rps)),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getClientIP はクライアントのIPアドレスを取得する
// X-Forwarded-For (プロキシ経由)、X-Real-IP (Nginx)、接続元IPを順に確認
func getClientIP(c *gin.Context) string {
	if xForwardedFor := c.GetHeader("X-Forwarded-For"); xForwardedFor != "" {
		// X-Forwarded-Forはカンマ区切りの複数IPが含まれることがある
		// 最初のIP（元のクライアント）を使用する
		for i := 0; i < len(xForwardedFor); i++ {
			if xForwardedFor[i] == ',' {
				return xForwardedFor[:i]
			}
		}
		return xForwardedFor
	}

	if xRealIP := c.GetHeader("X-Real-IP"); xRealIP != "" {
		return xRealIP
	}

	return c.ClientIP()
}
