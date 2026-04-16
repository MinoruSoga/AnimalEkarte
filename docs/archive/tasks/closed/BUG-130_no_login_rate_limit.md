# BUG-130: ログインエンドポイントにレートリミットがない

## 概要
`POST /api/v1/login` にレートリミットが実装されていない。
攻撃者がブルートフォース攻撃で無制限にパスワード試行可能。

## 脆弱性分類
- **CWE-307**: Improper Restriction of Excessive Authentication Attempts
- **OWASP A07:2021**: Identification and Authentication Failures
- **影響**: パスワードブルートフォース攻撃に対する防御なし

## 再現手順
```bash
# 20回連続でパスワード試行 — すべて受け付けられる
for i in $(seq 1 20); do
  curl -s -o /dev/null -w "%{http_code}\n" \
    -X POST http://localhost:8080/api/v1/login \
    -H 'Content-Type: application/json' \
    -d '{"email":"admin@example.com","password":"wrong'$i'"}'
done
# → 全て 401, レートリミットなし
```

## ブラウザテスト結果
20回連続ログイン失敗 → 全て 400/401、制限なし。その後正常パスワードで即座にログイン成功。

## 期待する動作
- IP ベースのレートリミット: 5回/分（失敗時）
- アカウントベースのロック: 10回連続失敗で一時ロック（5-15分）
- レートリミット超過時は `429 Too Many Requests` を返す

## 修正方針

### オプション1: Gin ミドルウェアでレートリミット

```go
import "golang.org/x/time/rate"

// IP ベースのレートリミッター
var loginLimiters = sync.Map{}

func LoginRateLimit() gin.HandlerFunc {
    return func(c *gin.Context) {
        ip := c.ClientIP()
        limiter, _ := loginLimiters.LoadOrStore(ip, rate.NewLimiter(rate.Every(time.Minute/5), 5))
        if !limiter.(*rate.Limiter).Allow() {
            c.JSON(http.StatusTooManyRequests, gin.H{"error": "リクエストが多すぎます。しばらく待ってから再試行してください"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

### オプション2: Redis ベース（推奨、本番環境向け）
- `login_attempts:<email>` キーでカウント
- 5回失敗で 15分ロック
- 成功時にカウンタリセット

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/api.md` — Security
> "Implement rate limiting"

明確にレートリミットの実装が求められている。

### `.claude/rules/security.md` — Authentication
> "Use secure session management"

ブルートフォース対策はセッション管理の基本要件。

## 優先度
**High** — 本番環境ではブルートフォース攻撃に対する防御が必須。
デモ環境の `password` のような弱いパスワードが突破される。

## 関連ファイル
- `backend/cmd/api/main.go` — ルート登録
- `backend/internal/handler/auth_handler.go` — Login ハンドラ
