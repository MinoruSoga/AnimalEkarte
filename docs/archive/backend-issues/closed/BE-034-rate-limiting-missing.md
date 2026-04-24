# BE-034: レートリミット未実装

## 問題
全APIエンドポイントにレートリミットが設定されていない。
特に `/api/v1/login` へのブルートフォース攻撃が無制限に可能。

## リスク
- ログインエンドポイントへのブルートフォース攻撃
- DDoS攻撃による全エンドポイントのリソース枯渇

## 修正方針
```go
// middleware/rate_limit.go
import "golang.org/x/time/rate"

// ログイン: 5リクエスト/分/IP
// 一般: 100リクエスト/分/IP
func RateLimitMiddleware(r rate.Limit, b int) gin.HandlerFunc { ... }
```

`go.mod` に `golang.org/x/time` を追加。
`handler.go` の `/api/v1/login` グループに適用。

## 優先度
HIGH（セキュリティ）
