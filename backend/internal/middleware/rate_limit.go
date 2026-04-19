package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// limiterEntry は IP ごとのレートリミッターと最終アクセス時刻を保持する
type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimitStore IP別のレートリミッター管理（TTL eviction 付き）
type RateLimitStore struct {
	limiters map[string]*limiterEntry
	mu       sync.RWMutex
}

// NewRateLimitStore はRateLimitStoreを初期化してバックグラウンドクリーンアップを開始する。
// ctx がキャンセルされると cleanupLoop ゴルーチンも終了する。
func NewRateLimitStore(ctx context.Context) *RateLimitStore {
	s := &RateLimitStore{
		limiters: make(map[string]*limiterEntry),
	}
	go s.cleanupLoop(ctx)
	return s
}

// cleanupLoop は 5 分ごとに 10 分以上アクセスのない IP エントリを削除する。
// ctx がキャンセルされるとループを終了する。
func (s *RateLimitStore) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.evict(10 * time.Minute)
		case <-ctx.Done():
			return
		}
	}
}

// evict は ttl より古いエントリを削除する（テスト用にも公開）
func (s *RateLimitStore) evict(ttl time.Duration) {
	threshold := time.Now().Add(-ttl)
	s.mu.Lock()
	defer s.mu.Unlock()
	for ip, entry := range s.limiters {
		if entry.lastSeen.Before(threshold) {
			delete(s.limiters, ip)
		}
	}
}

// getLimiter はIPアドレスに対応するlimiterを取得・作成し、lastSeen を更新する。
// RLock → RUnlock → Lock の TOCTOU を避けるため始めから Write Lock を取得する。
func (s *RateLimitStore) getLimiter(ip string, rps rate.Limit, burst int) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entry, exists := s.limiters[ip]; exists {
		entry.lastSeen = time.Now()
		return entry.limiter
	}

	newEntry := &limiterEntry{
		limiter:  rate.NewLimiter(rps, burst),
		lastSeen: time.Now(),
	}
	s.limiters[ip] = newEntry
	return newEntry.limiter
}

// RateLimit は指定されたレートでIPアドレスごとにレート制限を行う
// rps: requests per second, burst: バースト許容量
func RateLimit(store *RateLimitStore, rps rate.Limit, burst int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := getClientIP(c)
		limiter := store.getLimiter(ip, rps, burst)

		if !limiter.Allow() {
			respondError(c, http.StatusTooManyRequests, fmt.Sprintf("rate limit exceeded: %d requests per second max", int(rps)))
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
