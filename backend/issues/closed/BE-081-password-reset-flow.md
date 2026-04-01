# BE-081: パスワードリセット機能実装（forgot-password / reset-password エンドポイント）

**Status**: Open
**Priority**: Medium
**Affects**: migrations/001_init.sql, internal/handler/auth_handler.go, internal/service/auth_service.go, internal/repository/（新規）
**Date Created**: 2026-03-29
**Related**: BUG-060, FE-138

---

## Summary

`POST /v1/auth/forgot-password` と `POST /v1/auth/reset-password` エンドポイントが未実装。
スタッフがパスワードを忘れた場合の自己回復フローがない。

---

## 実装手順

### 1. `migrations/001_init.sql` に `password_reset_tokens` テーブル追加

```sql
CREATE TABLE password_reset_tokens (
    id         BIGSERIAL   PRIMARY KEY,
    user_id    bigint      NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
    token_hash varchar(64) NOT NULL UNIQUE,  -- SHA-256（hex 64文字）
    expires_at timestamptz NOT NULL,         -- 発行から 30 分
    used_at    timestamptz NULL,             -- NULL = 未使用
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_password_reset_tokens_hash
    ON password_reset_tokens(token_hash)
    WHERE used_at IS NULL;
```

### 2. Go モデル（`internal/model/auth.go` または新規ファイル）

```go
type PasswordResetToken struct {
    ID        uint64     `gorm:"primaryKey;autoIncrement"`
    UserID    uint64     `gorm:"not null"`
    TokenHash string     `gorm:"not null;uniqueIndex;size:64"`
    ExpiresAt time.Time  `gorm:"not null"`
    UsedAt    *time.Time
    CreatedAt time.Time  `gorm:"not null;default:now()"`
}

func (PasswordResetToken) TableName() string {
    return "password_reset_tokens"
}
```

### 3. リポジトリ（`internal/repository/password_reset_repository.go`）

```go
type PasswordResetRepository interface {
    Create(ctx context.Context, token *model.PasswordResetToken) error
    FindValidByHash(ctx context.Context, tokenHash string) (*model.PasswordResetToken, error)
    MarkUsed(ctx context.Context, id uint64) error
}
```

### 4. サービスメソッド（`internal/service/auth_service.go`）

#### `ForgotPassword(ctx, email string) error`

```go
func (s *authService) ForgotPassword(ctx context.Context, email string) error {
    // ユーザー列挙対策: ユーザー不在でも成功レスポンスを返す
    user, err := s.userRepo.FindByEmail(ctx, email)
    if err != nil {
        if errors.Is(err, apperrors.ErrNotFound) {
            slog.InfoContext(ctx, "forgot-password: email not found (silent)")
            return nil
        }
        return fmt.Errorf("forgot password: %w", err)
    }

    // ワンタイムトークン生成（32バイトランダム）
    rawToken, err := generateSecureToken(32)
    if err != nil {
        return fmt.Errorf("generate token: %w", err)
    }
    tokenHash := sha256Hex(rawToken)

    token := &model.PasswordResetToken{
        UserID:    user.ID,
        TokenHash: tokenHash,
        ExpiresAt: time.Now().Add(30 * time.Minute),
    }
    if err := s.resetRepo.Create(ctx, token); err != nil {
        return fmt.Errorf("create reset token: %w", err)
    }

    // メール送信（dev 環境はログ出力）
    if s.cfg.Env == "development" {
        slog.InfoContext(ctx, "password reset token", "email", email, "token", rawToken)
        return nil
    }
    return s.mailer.SendPasswordReset(email, rawToken)
}
```

#### `ResetPassword(ctx, token, newPassword string) error`

```go
func (s *authService) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
    tokenHash := sha256Hex(rawToken)
    resetToken, err := s.resetRepo.FindValidByHash(ctx, tokenHash)
    if err != nil {
        return apperrors.WrapNotFound("token", "invalid or expired")
    }
    if time.Now().After(resetToken.ExpiresAt) {
        return apperrors.WrapInvalidInput("token expired")
    }

    hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
    if err != nil {
        return fmt.Errorf("hash password: %w", err)
    }

    if err := s.userRepo.UpdatePassword(ctx, resetToken.UserID, string(hash)); err != nil {
        return fmt.Errorf("update password: %w", err)
    }
    if err := s.resetRepo.MarkUsed(ctx, resetToken.ID); err != nil {
        return fmt.Errorf("mark token used: %w", err)
    }
    // 全 refresh_tokens を revoke
    _ = s.tokenRepo.RevokeAllByUserID(ctx, resetToken.UserID)

    slog.InfoContext(ctx, "password reset completed", slog.Uint64("user_id", resetToken.UserID))
    return nil
}
```

### 5. ハンドラー（`internal/handler/auth_handler.go`）

```go
// POST /v1/auth/forgot-password
func (h *Handler) ForgotPassword(c *gin.Context) {
    var req struct {
        Email string `json:"email" binding:"required,email"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        RespondError(c, apperrors.WrapInvalidInput("email is required"))
        return
    }
    if err := h.svc.Auth.ForgotPassword(c.Request.Context(), req.Email); err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "If this email is registered, a reset link has been sent."})
}

// POST /v1/auth/reset-password
func (h *Handler) ResetPassword(c *gin.Context) {
    var req struct {
        Token    string `json:"token"    binding:"required"`
        Password string `json:"password" binding:"required,min=8"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
        return
    }
    if err := h.svc.Auth.ResetPassword(c.Request.Context(), req.Token, req.Password); err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Password reset successful."})
}
```

### 6. ルート登録（`cmd/api/handler.go`）

```go
auth := v1.Group("/auth")
{
    // 既存 ...
    auth.POST("/forgot-password", h.ForgotPassword)
    auth.POST("/reset-password", h.ResetPassword)
}
```

---

## セキュリティ要件

- トークンは 32 バイトランダム（`crypto/rand`）
- DB にはハッシュ（SHA-256）のみ保存、生トークンはメールのみ
- 有効期限 30 分
- ワンタイム（使用済みは再利用不可）
- ユーザー不在でも同じレスポンス（ユーザー列挙攻撃対策）
- パスワードリセット成功後に全 refresh_token を revoke

---

## 確認コマンド

```bash
make reset
docker compose exec backend go build ./...
docker compose exec backend go test ./... -v
```

---

## 受入条件

- [ ] `password_reset_tokens` テーブルが migration に存在する
- [ ] `POST /v1/auth/forgot-password` → 200（メール/ログ出力）
- [ ] 存在しないメールで `forgot-password` → 200（ユーザー列挙対策）
- [ ] `POST /v1/auth/reset-password` → 200・パスワード更新
- [ ] 期限切れ・使用済みトークン → 400
- [ ] パスワードリセット後に refresh_tokens が revoke される
- [ ] `docker compose exec backend go build ./...` 成功
- [ ] `docker compose exec backend go test ./... -v` 成功
