# TASK-017: 認証系セキュリティ修正 4件（goroutine context / SMTP TLS / bcrypt コスト / RefreshToken rotation）

## 概要

auth/password_reset 系で発見された CRITICAL/HIGH セキュリティ問題を一括修正する。

## 優先度

CRITICAL / HIGH

---

## 問題 1: password_reset_service.go — goroutine 内でリクエストコンテキストを使用

### ファイル
`backend/internal/service/password_reset_service.go:95-101`

### 優先度
CRITICAL（goroutine リーク + ログ消失）

### 問題
メール送信 goroutine がリクエストスコープの `ctx` をキャプチャしている。HTTP レスポンス返却後にコンテキストがキャンセルされると `slog.ErrorContext(ctx, ...)` が canceled context に書き込まれ、ログが失われる。また goroutine のライフサイクルが不明確でリークの原因になる。

### 修正案
```go
go func() {
    bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    if sendErr := s.sendResetEmail(email, resetURL); sendErr != nil {
        slog.ErrorContext(bgCtx, "failed to send password reset email",
            slog.String("email", email),
            slog.String("error", sendErr.Error()))
    }
}()
```

---

## 問題 2: password_reset_service.go — SMTP が PlainAuth で TLS 未保証

### ファイル
`backend/internal/service/password_reset_service.go:191, 194`

### 優先度
CRITICAL（SMTP 認証情報の平文送信リスク）

### 問題
`smtp.PlainAuth` は TLS なし接続時に認証情報を平文で送信する。本番 SMTP サーバが STARTTLS を強制しない場合、ユーザー名・パスワードが盗聴される。

### 修正案
```go
// config.go の Validate() に追加
if c.GinMode == "release" && c.SMTPPort != 465 && c.SMTPPort != 587 {
    return fmt.Errorf("SMTP_PORT must be 465 (SMTPS) or 587 (STARTTLS) in release mode")
}
```

または `crypto/tls` を使った明示的 TLS 接続に変更:
```go
// TLS port 465 の場合
tlsConfig := &tls.Config{ServerName: s.cfg.SMTPHost}
conn, err := tls.Dial("tcp", addr, tlsConfig)
// ... net/smtp.NewClient(conn, host) で送信
```

---

## 問題 3: bcrypt コストが DefaultCost (10) で不足

### ファイル
- `backend/internal/handler/auth_handler.go:507`（ChangeMyPassword）
- `backend/internal/service/password_reset_service.go:124`（ResetPassword）
- `backend/internal/service/staff_service.go:204, 252`（Create/Update）

### 優先度
HIGH

### 問題
`bcrypt.DefaultCost = 10` は 2013 年当時の基準値。現代ハードウェアでは不十分。

### 修正案
```go
// 共通定数を errors パッケージまたは config に定義
const BcryptCost = 12

// 各所で使用
hashed, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
```

---

## 問題 4: RefreshToken Rotation が不完全（旧トークン無効化なし）

### ファイル
`backend/internal/handler/auth_handler.go:339-458`

### 優先度
HIGH

### 問題
`rotation` のコメントがあるが、旧 refresh_token の DB 無効化処理が存在しない。攻撃者が旧トークンを盗んでいた場合、新トークン発行後も旧トークンで再 refresh が可能（署名・有効期限が有効なため）。

### 修正案（最小対応）
```sql
-- refresh_token_blocklist テーブルを追加（または password_reset_tokens を流用）
CREATE TABLE refresh_token_blocklist (
    jti UUID PRIMARY KEY,
    expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_rtb_expires ON refresh_token_blocklist(expires_at);
```

```go
// JWT に jti (JWT ID) クレームを追加
claims := &JWTClaims{
    // ...
    RegisteredClaims: jwt.RegisteredClaims{
        ID: uuid.NewString(), // jti
        ExpiresAt: jwt.NewNumericDate(expiry),
    },
}

// Refresh 時に旧 jti をブロックリストに追加
err = s.blockJTI(ctx, oldClaims.ID, oldClaims.ExpiresAt.Time)

// 検証時にブロックリストをチェック
if s.isJTIBlocked(ctx, claims.ID) {
    return apperrors.WrapUnauthorized("token has been revoked")
}
```
