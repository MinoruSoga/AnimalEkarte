# TASK-044: BE RateLimitStore に TTL eviction 実装

**作成日**: 2026-03-27
**ステータス**: Open
**優先度**: High
**領域**: Backend / Security

---

## 概要

`middleware/rate_limit.go` の `RateLimitStore` は IP ごとに `rate.Limiter` を `sync.Map` に追加するが、削除メカニズムがない。
大量の異なるソース IP（分散プロキシ等）からリクエストを送ると、エントリが無限に蓄積してメモリが枯渇する DoS ベクトルになる。

---

## 現状コード

```go
// middleware/rate_limit.go
type RateLimitStore struct {
    limiters sync.Map  // IP → *rate.Limiter（削除なし）
}

func (s *RateLimitStore) GetLimiter(ip string) *rate.Limiter {
    limiter, ok := s.limiters.Load(ip)
    if !ok {
        newLimiter := rate.NewLimiter(rate.Limit(5), 10)
        s.limiters.Store(ip, newLimiter)  // 追加するが消えない
        return newLimiter
    }
    return limiter.(*rate.Limiter)
}
```

---

## 修正方針

最終アクセス時刻を記録し、一定時間（10分）アクセスがなければエントリを削除する。

```go
type limiterEntry struct {
    limiter  *rate.Limiter
    lastSeen time.Time
}

// バックグラウンドで定期クリーンアップ
func (s *RateLimitStore) cleanupLoop() {
    ticker := time.NewTicker(5 * time.Minute)
    for range ticker.C {
        now := time.Now()
        s.limiters.Range(func(key, value any) bool {
            entry := value.(*limiterEntry)
            if now.Sub(entry.lastSeen) > 10*time.Minute {
                s.limiters.Delete(key)
            }
            return true
        })
    }
}
```

または `github.com/patrickmn/go-cache` 等の TTL キャッシュライブラリを採用する。

---

## 受入条件

- [ ] `RateLimitStore` に TTL eviction が実装されている
- [ ] 長時間アクセスのない IP のエントリが自動削除される
- [ ] `docker compose exec backend go build ./...` 成功
- [ ] `docker compose exec backend go test ./...` 全テストパス
