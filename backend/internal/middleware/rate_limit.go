package middleware

import (
	"context"
	"fmt"
	"log/slog"
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
			s.safeEvict(10 * time.Minute)
		case <-ctx.Done():
			return
		}
	}
}

// safeEvict は evict を panic recovery 付きで実行し、cleanupLoop が継続できるようにする。
func (s *RateLimitStore) safeEvict(ttl time.Duration) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("rate limit cleanup goroutine panic", slog.Any("panic", r))
		}
	}()
	s.evict(ttl)
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
// c.ClientIP() を使用することで TRUSTED_PROXY_CIDR 設定を尊重し、
// X-Forwarded-For ヘッダーの偽装による IP スプーフィングを防止する
func RateLimit(store *RateLimitStore, rps rate.Limit, burst int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := store.getLimiter(ip, rps, burst)

		if !limiter.Allow() {
			respondError(c, http.StatusTooManyRequests, fmt.Sprintf("rate limit exceeded: %d requests per second max", int(rps)))
			return
		}

		c.Next()
	}
}

// LiffRateLimit は LIFF エンドポイント向けのレートリミッター
// 指定された requests/minute を requests/second に変換して RateLimit に委譲する
func LiffRateLimit(store *RateLimitStore, requestsPerMinute int) gin.HandlerFunc {
	return RateLimit(store, rate.Limit(requestsPerMinute)/60.0, requestsPerMinute)
}
