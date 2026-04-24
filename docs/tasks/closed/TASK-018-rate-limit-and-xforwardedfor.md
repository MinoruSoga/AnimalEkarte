# TASK-018: レートリミット欠落（forgot/reset-password）と X-Forwarded-For 無検証信頼

## 概要

パスワードリセット系エンドポイントにレートリミットが未適用、かつレートリミット実装が `X-Forwarded-For` を無条件信頼しているため攻撃者にバイパスされる。

## 優先度

HIGH

---

## 問題 1: forgot-password / reset-password にレートリミットなし

### ファイル
`backend/internal/handler/handler.go:64-65`

### 問題
```go
// 現状: レートリミットなし
api.POST("/auth/forgot-password", h.ForgotPassword)
api.POST("/auth/reset-password",  h.ResetPassword)
```

- `forgot-password`: メールアドレスを総当たりしてアカウント存在確認（ユーザー列挙）に悪用可能
- `reset-password`: リセットトークンのブルートフォース攻撃に利用可能

### 修正案
```go
// handler.go
pwResetRateStore := middleware.NewRateLimitStore(ctx)
api.POST("/auth/forgot-password",
    middleware.RateLimit(pwResetRateStore, 3.0/60, 3), h.ForgotPassword)
api.POST("/auth/reset-password",
    middleware.RateLimit(pwResetRateStore, 3.0/60, 3), h.ResetPassword)
```

---

## 問題 2: X-Forwarded-For を無条件信頼しレートリミットをバイパス可能

### ファイル
`backend/internal/middleware/rate_limit.go:104-112`

### 問題
```go
func getClientIP(c *gin.Context) string {
    if xForwardedFor := c.GetHeader("X-Forwarded-For"); xForwardedFor != "" {
        // 無条件で信頼 → 攻撃者は X-Forwarded-For: 1.2.3.4 を送るだけでレートリミットを回避
        return strings.Split(xForwardedFor, ",")[0]
    }
    return c.ClientIP()
}
```

### 修正案
Gin の `TrustedProxies` を設定し、`c.ClientIP()` に統一する。`c.ClientIP()` は設定済みの TrustedProxies からのリクエストのみ `X-Forwarded-For` を使用する。

```go
// main.go（DI 配線）
r := gin.New()
// リバースプロキシの CIDR を環境変数から取得
if err := r.SetTrustedProxies(cfg.TrustedProxies); err != nil {
    log.Fatalf("failed to set trusted proxies: %v", err)
}

// rate_limit.go
func getClientIP(c *gin.Context) string {
    return c.ClientIP() // Gin が TrustedProxies に基づいて解決
}
```

```go
// config.go
TrustedProxies []string // TRUSTED_PROXIES="10.0.0.0/8,172.16.0.0/12"
```

---

## 問題 3: Logout が protected グループ外に登録されている

### ファイル
`backend/internal/handler/handler.go:62`

### 優先度
MEDIUM

### 問題
```go
// 現状: Auth ミドルウェアが適用されない
api.POST("/logout", h.Logout)
```

Auth ミドルウェアを通らないため、`auth_handler.go` 内でユーザー情報を手動チェックする複雑な分岐が必要になっている。監査ログの記録漏れリスクもある。

### 修正案
```go
// protected グループ（Auth ミドルウェア適用済み）に移動
protected.POST("/logout", h.Logout)
// handler 側の手動 user_id チェック分岐を削除
```
