# BE-078: リフレッシュトークン実装（デュアルトークン移行）

**Status**: Open
**Priority**: High
**Affects**: internal/model/auth.go（新規）, internal/handler/auth_handler.go, internal/service/auth_service.go, internal/repository/auth_repository.go, migrations/001_init.sql
**Date Created**: 2026-03-29
**Related**: BUG-055, FE-135（Axios インターセプター）

---

## Summary

現在の認証は **24時間有効な JWT 1枚** のみ。サーバー側での無効化手段がなく、
ログアウト後もトークンが有効であり、漏洩時のリスクが高い。

デュアルトークン方式に移行し、サーバー側でセッションを管理可能にする。

| 項目 | 現在 | 移行後 |
|------|------|--------|
| アクセストークン有効期限 | 24時間 | 15分 |
| リフレッシュトークン | なし | 不透明トークン（7日間・DB管理） |
| ログアウト時の無効化 | 不可能 | リフレッシュトークンを DB で revoke |
| 漏洩時の最大被害時間 | 24時間 | 15分 |

---

## 実装手順

### 1. `refresh_tokens` テーブル追加（`migrations/001_init.sql`）

```sql
CREATE TABLE refresh_tokens (
    id          BIGSERIAL   PRIMARY KEY,
    user_id     bigint      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    clinic_id   bigint      NOT NULL REFERENCES clinics(id),
    token_hash  varchar(64) NOT NULL UNIQUE,  -- SHA-256 ハッシュ（hex 64文字）
    expires_at  timestamp   NOT NULL,
    revoked_at  timestamp   NULL,
    created_at  timestamp   DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_refresh_tokens_user   FOREIGN KEY (user_id)   REFERENCES users(id)   ON DELETE CASCADE,
    CONSTRAINT fk_refresh_tokens_clinic FOREIGN KEY (clinic_id) REFERENCES clinics(id)
);

-- 検索インデックス（token_hash で引く）
CREATE INDEX idx_refresh_tokens_hash ON refresh_tokens(token_hash) WHERE revoked_at IS NULL;

-- ユーザー単位のセッション一覧取得用
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id, expires_at DESC);
```

**設計理由**:
- `token_hash` に SHA-256 ハッシュを保存（生トークンはDBに置かない。DBが漏洩しても生トークンを得られない）
- `revoked_at` NULL = 有効、NOT NULL = 無効化済み
- `ON DELETE CASCADE` でユーザー削除時に自動削除

### 2. `RefreshToken` モデル追加（`backend/internal/model/auth.go`）

```go
package model

import "time"

// RefreshToken は refresh_tokens テーブルのモデル。
type RefreshToken struct {
    ID        uint64     `gorm:"primaryKey"           json:"id"`
    UserID    uint64     `gorm:"not null"             json:"user_id"`
    ClinicID  uint64     `gorm:"not null"             json:"clinic_id"`
    TokenHash string     `gorm:"not null;uniqueIndex" json:"-"`  // 外部に返さない
    ExpiresAt time.Time  `gorm:"not null"             json:"expires_at"`
    RevokedAt *time.Time `gorm:"default:null"         json:"revoked_at"`
    CreatedAt time.Time                               `json:"created_at"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }

// IsExpired はトークンが期限切れかどうか返す。
func (r *RefreshToken) IsExpired() bool {
    return time.Now().After(r.ExpiresAt)
}

// IsRevoked はトークンが無効化済みかどうか返す。
func (r *RefreshToken) IsRevoked() bool {
    return r.RevokedAt != nil
}

// IsValid はトークンが有効（期限内かつ revoke されていない）かどうか返す。
func (r *RefreshToken) IsValid() bool {
    return !r.IsExpired() && !r.IsRevoked()
}
```

### 3. リポジトリ追加（`backend/internal/repository/auth_repository.go`）

```go
type AuthRepository interface {
    CreateRefreshToken(ctx context.Context, token *model.RefreshToken) error
    FindRefreshToken(ctx context.Context, tokenHash string) (*model.RefreshToken, error)
    RevokeRefreshToken(ctx context.Context, id uint64) error
    RevokeAllUserTokens(ctx context.Context, userID uint64) error  // 全端末ログアウト用
    DeleteExpiredTokens(ctx context.Context) error                 // 定期クリーンアップ
}
```

### 4. サービス層の変更（`backend/internal/service/auth_service.go`）

#### ログインレスポンス変更

```go
type LoginResult struct {
    AccessToken  string    `json:"access_token"`   // JWT 15分
    RefreshToken string    `json:"refresh_token"`  // 不透明トークン 7日間
    ExpiresAt    time.Time `json:"expires_at"`
    User         *UserInfo `json:"user"`
}

func (s *authService) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
    // ... 既存のユーザー認証処理 ...

    // アクセストークン: JWT 15分
    accessToken, err := s.generateJWT(user, clinicID, 15*time.Minute)
    if err != nil {
        return nil, fmt.Errorf("failed to generate access token: %w", err)
    }

    // リフレッシュトークン: 不透明トークン
    rawToken := generateSecureToken()  // crypto/rand で 32バイト生成
    tokenHash := sha256Hex(rawToken)   // SHA-256 ハッシュ化

    refreshToken := &model.RefreshToken{
        UserID:    user.ID,
        ClinicID:  clinicID,
        TokenHash: tokenHash,
        ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
    }
    if err := s.authRepo.CreateRefreshToken(ctx, refreshToken); err != nil {
        return nil, fmt.Errorf("failed to create refresh token: %w", err)
    }

    return &LoginResult{
        AccessToken:  accessToken,
        RefreshToken: rawToken,  // 生トークンをレスポンスに含める（DBにはハッシュのみ）
        ExpiresAt:    time.Now().Add(15 * time.Minute),
        User:         toUserInfo(user),
    }, nil
}
```

#### リフレッシュエンドポイント（新規）

```go
func (s *authService) RefreshToken(ctx context.Context, rawToken string) (*LoginResult, error) {
    tokenHash := sha256Hex(rawToken)

    existing, err := s.authRepo.FindRefreshToken(ctx, tokenHash)
    if err != nil || !existing.IsValid() {
        return nil, fmt.Errorf("invalid or expired refresh token: %w", apperrors.ErrUnauthorized)
    }

    // トークンローテーション: 古いトークンを revoke してから新しいトークンを発行
    if err := s.authRepo.RevokeRefreshToken(ctx, existing.ID); err != nil {
        return nil, fmt.Errorf("failed to revoke old token: %w", err)
    }

    // 新しいアクセストークン + リフレッシュトークンを発行
    return s.issueTokenPair(ctx, existing.UserID, existing.ClinicID)
}
```

#### ログアウト変更

```go
func (s *authService) Logout(ctx context.Context, rawRefreshToken string) error {
    tokenHash := sha256Hex(rawRefreshToken)
    token, err := s.authRepo.FindRefreshToken(ctx, tokenHash)
    if err != nil {
        return nil  // すでに無効化されていてもエラーにしない
    }
    return s.authRepo.RevokeRefreshToken(ctx, token.ID)
}
```

### 5. ハンドラー変更（`backend/internal/handler/auth_handler.go`）

#### Cookie 設定

```go
func setTokenCookies(c *gin.Context, accessToken, refreshToken string) {
    // アクセストークン: 15分 + httpOnly + Secure + SameSite=Strict
    c.SetCookie("access_token", accessToken,
        int(15*time.Minute/time.Second),
        "/", "", true, true)
    // SameSite は Gin の SetSameSite で設定
    c.SetSameSite(http.SameSiteStrictMode)

    // リフレッシュトークン: 7日 + httpOnly + Secure + SameSite=Strict
    // パスを /v1/auth/refresh に限定してアクセストークンより安全に
    c.SetCookie("refresh_token", refreshToken,
        int(7*24*time.Hour/time.Second),
        "/v1/auth/refresh", "", true, true)
}

// POST /v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
    // ... binding ...
    result, err := h.service.Login(c.Request.Context(), input)
    if err != nil {
        RespondError(c, err)
        return
    }
    setTokenCookies(c, result.AccessToken, result.RefreshToken)
    c.JSON(http.StatusOK, toMeResponse(result.User))
}

// POST /v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
    rawToken, err := c.Cookie("refresh_token")
    if err != nil {
        RespondError(c, fmt.Errorf("missing refresh token: %w", apperrors.ErrUnauthorized))
        return
    }
    result, err := h.service.RefreshToken(c.Request.Context(), rawToken)
    if err != nil {
        // リフレッシュ失敗時: Cookie を削除してフロントに再ログインを促す
        c.SetCookie("access_token", "", -1, "/", "", true, true)
        c.SetCookie("refresh_token", "", -1, "/v1/auth/refresh", "", true, true)
        RespondError(c, err)
        return
    }
    setTokenCookies(c, result.AccessToken, result.RefreshToken)
    c.JSON(http.StatusOK, toMeResponse(result.User))
}

// POST /v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
    rawToken, _ := c.Cookie("refresh_token")
    if rawToken != "" {
        _ = h.service.Logout(c.Request.Context(), rawToken)
    }
    c.SetCookie("access_token", "", -1, "/", "", true, true)
    c.SetCookie("refresh_token", "", -1, "/v1/auth/refresh", "", true, true)
    c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}
```

### 6. JWT 有効期限変更

```go
// 既存の JWT 生成の有効期限を 24h → 15min に変更
const accessTokenTTL = 15 * time.Minute
```

### 7. ルーター登録（`cmd/api/main.go` または `handler/router.go`）

```go
auth := v1.Group("/auth")
{
    auth.POST("/login",   authHandler.Login)
    auth.POST("/refresh", authHandler.RefreshToken)  // 新規追加
    auth.POST("/logout",  authHandler.Logout)        // 変更: refresh_token を revoke するように
}
```

---

## ユーティリティ関数

```go
// backend/internal/service/token_utils.go

import (
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
)

// generateSecureToken は crypto/rand で 32バイトの安全なランダムトークンを生成する。
func generateSecureToken() string {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        panic(fmt.Sprintf("failed to generate secure token: %v", err))
    }
    return hex.EncodeToString(b)
}

// sha256Hex は文字列の SHA-256 ハッシュを hex 文字列で返す。
func sha256Hex(s string) string {
    h := sha256.Sum256([]byte(s))
    return hex.EncodeToString(h[:])
}
```

---

## 確認コマンド

```bash
# Go ビルド確認
docker compose exec backend go build ./...

# DB リセット（refresh_tokens テーブルが作成されることを確認）
make reset

# テスト
docker compose exec backend go test ./... -v

# Lint
docker compose exec backend golangci-lint run ./...
```

---

## 受入条件

- [ ] `refresh_tokens` テーブルが DB に存在する
- [ ] `POST /v1/auth/login` が `access_token`（15分）と `refresh_token`（7日）の2つの httpOnly Cookie をセットする
- [ ] `POST /v1/auth/refresh` が `refresh_token` Cookie を検証し、新しいトークンペアを発行する（トークンローテーション）
- [ ] `POST /v1/auth/logout` が `refresh_token` を DB で revoke し、両 Cookie を削除する
- [ ] revoke 済みの `refresh_token` での `/v1/auth/refresh` が 401 を返す
- [ ] 期限切れの `refresh_token` での `/v1/auth/refresh` が 401 を返す
- [ ] `docker compose exec backend go build ./...` 成功
- [ ] `docker compose exec backend golangci-lint run ./...` エラー 0 件
- [ ] `docker compose exec backend go test ./... -v` 成功
