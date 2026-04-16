# BUG-136: Refresh Token の Cookie は発行されるがエンドポイントが未実装

## 概要
ログイン時に `refresh_token` Cookie が `Path=/api/v1/auth/refresh` で発行されるが、
対応するエンドポイント `POST /api/v1/auth/refresh` が存在しない（404）。

access_token の有効期限が切れた後、ユーザーは再ログインを強制される。
refresh_token による透過的なセッション更新ができない。

## 脆弱性分類
- **機能未実装**（セキュリティ設計の不整合）
- **影響**: 
  - UX: access_token 期限切れで突然ログアウトされる
  - セキュリティ: refresh_token Cookie が使われないまま発行されている（不要な Cookie）

## 再現手順
```bash
# ログイン
curl -c cookies.txt -X POST /api/v1/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@example.com","password":"password"}'

# cookies.txt に refresh_token が設定される
# Path=/api/v1/auth/refresh

# refresh エンドポイントにアクセス
curl -b cookies.txt -X POST /api/v1/auth/refresh
# → 404 Not Found
```

## 現状コード

### `backend/internal/handler/auth_handler.go:252-265`
```go
// refresh_token Cookie は発行されるが...
http.SetCookie(c.Writer, &http.Cookie{
    Name:     refreshTokenCookieName,  // "refresh_token"
    Value:    refreshToken,
    Path:     "/api/v1/auth/refresh",  // ← このパスのエンドポイントがない
    HttpOnly: true,
    Secure:   isProduction,
    SameSite: sameSite,
})
```

### `backend/internal/handler/handler.go`
```go
// /api/v1/auth/refresh ルートが登録されていない
api.POST("/login", h.Login)
api.POST("/logout", h.Logout)
// ← /auth/refresh がない
```

## 修正方針

### 案A: Refresh エンドポイントを実装
```go
// handler.go
api.POST("/auth/refresh", h.RefreshToken)

// auth_handler.go
func (h *Handler) RefreshToken(c *gin.Context) {
    refreshToken, err := c.Cookie(refreshTokenCookieName)
    if err != nil {
        c.JSON(401, gin.H{"error": "refresh token not found"})
        return
    }
    // refresh_token を検証
    // 新しい access_token を発行
    // 新しい refresh_token を発行（rotation）
}
```

### 案B: Refresh Token を廃止（短期対応）
refresh_token Cookie の発行を停止し、access_token の有効期限を長めに設定。
ただし、セキュリティ上は短い access_token + refresh token が推奨。

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/security.md` — Authentication
> "Use secure session management"

Refresh token パターンはセキュリティベストプラクティス。
access_token を短命（15-30分）にし、refresh_token で透過更新する。

## 優先度
**Medium** — 機能未実装。現状は access_token の期限が長ければ問題にならないが、
不要な Cookie が発行されている点は修正すべき。

## 関連ファイル
- `backend/internal/handler/auth_handler.go:252-265` — refresh_token Cookie 発行
- `backend/internal/handler/handler.go` — ルート登録（refresh エンドポイント追加）
