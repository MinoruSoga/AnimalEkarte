package middleware

import (
	"context"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// limiterEntry は IP ごとのレートリミッターと最終アクセス時刻を保持する
type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type rateLimitKey struct {
	bucketID uint64
	ip       string
}

// RateLimitStore middleware bucket・IP別のレートリミッター管理（TTL eviction 付き）
type RateLimitStore struct {
	limiters     map[rateLimitKey]*limiterEntry
	mu           sync.RWMutex
	nextBucketID atomic.Uint64
}

// NewRateLimitStore はRateLimitStoreを初期化してバックグラウンドクリーンアップを開始する。
// ctx がキャンセルされると cleanupLoop ゴルーチンも終了する。
func NewRateLimitStore(ctx context.Context) *RateLimitStore {
	s := &RateLimitStore{
		limiters: make(map[rateLimitKey]*limiterEntry),
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
	for key, entry := range s.limiters {
		if entry.lastSeen.Before(threshold) {
			delete(s.limiters, key)
		}
	}
}

// getLimiter は middleware bucket と IP に対応する limiter を取得・作成し、lastSeen を更新する。
// RLock → RUnlock → Lock の TOCTOU を避けるため始めから Write Lock を取得する。
func (s *RateLimitStore) getLimiter(bucketID uint64, ip string, rps rate.Limit, burst int) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := rateLimitKey{bucketID: bucketID, ip: ip}
	if entry, exists := s.limiters[key]; exists {
		entry.lastSeen = time.Now()
		return entry.limiter
	}

	newEntry := &limiterEntry{
		limiter:  rate.NewLimiter(rps, burst),
		lastSeen: time.Now(),
	}
	s.limiters[key] = newEntry
	return newEntry.limiter
}

// retryAfterDelay はトークンを消費せずに、次の1トークンが回復するまでの
// 待ち時間を返す。Reserve は将来の予約を登録するだけなので、Cancel で
// 取り消せばバケット容量は変化しない。
func retryAfterDelay(limiter *rate.Limiter) time.Duration {
	reservation := limiter.Reserve()
	if !reservation.OK() {
		return time.Second
	}
	defer reservation.Cancel()
	return reservation.Delay()
}

func retryAfterSeconds(limiter *rate.Limiter) int {
	return max(1, int(math.Ceil(retryAfterDelay(limiter).Seconds())))
}

// rejectRateLimited は Retry-After 付きの 429 応答を返して request を中断する。
func rejectRateLimited(c *gin.Context, limiter *rate.Limiter) {
	c.Header("Retry-After", strconv.Itoa(retryAfterSeconds(limiter)))
	respondError(c, http.StatusTooManyRequests, "rate limit exceeded")
}

// RateLimit は指定されたレートでIPアドレスごとにレート制限を行う
// rps: requests per second, burst: バースト許容量
// c.ClientIP() を使用することで TRUSTED_PROXY_CIDR 設定を尊重し、
// X-Forwarded-For ヘッダーの偽装による IP スプーフィングを防止する
func RateLimit(store *RateLimitStore, rps rate.Limit, burst int) gin.HandlerFunc {
	bucketID := store.nextBucketID.Add(1)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := store.getLimiter(bucketID, ip, rps, burst)

		if !limiter.Allow() {
			rejectRateLimited(c, limiter)
			return
		}

		c.Next()
	}
}

// FailureRateLimit は IP ごとのバケットで request をゲートするが、トークンを
// 消費するのは handler が status >= 400 で応答した場合のみ。成功 request は
// 予算を消費しないため、共有 NAT 配下の正当利用でバケットが枯渇しない。
// ゲートは非消費の Tokens() チェックであり、空バケット時点で同時に到着した
// request は最大バースト分だけ over-admit され得る。これは abuse 抑止として
// 許容する（継続的な失敗は rps へ収束する）。
func FailureRateLimit(store *RateLimitStore, rps rate.Limit, burst int) gin.HandlerFunc {
	bucketID := store.nextBucketID.Add(1)
	return func(c *gin.Context) {
		limiter := store.getLimiter(bucketID, c.ClientIP(), rps, burst)

		if limiter.Tokens() < 1 {
			rejectRateLimited(c, limiter)
			return
		}

		c.Next()

		if c.Writer.Status() >= http.StatusBadRequest {
			limiter.Allow()
		}
	}
}

// FailureLimiter は key 単位の失敗カウンター（例: アカウント email ごとの
// ログイン失敗回数）。Admit が試行可否を判定し、失敗時だけ RecordFailure で
// 予算を消費する。成功した試行はバケットを消費しない。
type FailureLimiter struct {
	store    *RateLimitStore
	bucketID uint64
	rps      rate.Limit
	burst    int
}

// NewFailureLimiter は store 内に独立した key 空間を持つ FailureLimiter を返す。
func (s *RateLimitStore) NewFailureLimiter(rps rate.Limit, burst int) *FailureLimiter {
	return &FailureLimiter{
		store:    s,
		bucketID: s.nextBucketID.Add(1),
		rps:      rps,
		burst:    burst,
	}
}

// Admit は key の失敗予算に空きがあるかを返す。枯渇時は ok=false と
// 回復までの秒数を返す。ゲートは非消費チェックであり、Admit 自体は予算を
// 消費しない（失敗時に RecordFailure を呼ぶ契約）。
func (l *FailureLimiter) Admit(key string) (int, bool) {
	limiter := l.store.getLimiter(l.bucketID, key, l.rps, l.burst)
	if limiter.Tokens() >= 1 {
		return 0, true
	}
	return retryAfterSeconds(limiter), false
}

// RecordFailure は key の予算を1トークン消費する。バケットが空なら no-op
// （ゲートは Admit の責務）。
func (l *FailureLimiter) RecordFailure(key string) {
	l.store.getLimiter(l.bucketID, key, l.rps, l.burst).Allow()
}

// LiffRateLimit は LIFF エンドポイント向けのレートリミッター
// 指定された requests/minute を requests/second に変換して RateLimit に委譲する
func LiffRateLimit(store *RateLimitStore, requestsPerMinute int) gin.HandlerFunc {
	return RateLimit(store, rate.Limit(requestsPerMinute)/60.0, requestsPerMinute)
}
