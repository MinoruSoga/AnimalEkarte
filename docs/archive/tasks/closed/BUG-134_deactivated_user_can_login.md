# BUG-134: 無効化 (is_active=false) されたスタッフがログインできる

## 概要
`is_active=false` に設定されたスタッフアカウントで `POST /api/v1/login` が成功する（200 OK）。
退職者や一時停止されたスタッフがシステムにアクセスし続けられる。

## 脆弱性分類
- **CWE-863**: Incorrect Authorization
- **OWASP A07:2021**: Identification and Authentication Failures
- **影響**: 無効化されたユーザーがシステムにフルアクセス可能。退職者のアカウント管理が機能しない。

## 再現手順
1. `admin@example.com` でログイン
2. `PATCH /api/v1/masters/staffs/9` で `{"is_active": false}` を送信 → 200
3. `POST /api/v1/login` で `{"email": "vet@example.com", "password": "password"}` → **200 OK**
4. 無効化されたユーザーでログイン成功

## ブラウザテスト結果
2回のテストで再現:
- 1回目: `is_active=false` → ログイン **200** ❌
- 2回目（再確認）: `is_active=false` → ログイン **200** ❌

## 期待する動作
- `is_active=false` のスタッフ → ログイン時に **401** `このアカウントは無効です`
- 既存セッション（JWT）も無効化すべき

## 現状コード（推定）

### `backend/internal/handler/auth_handler.go` — Login

ログイン処理で `accounts` テーブルからユーザーを取得した後、`staffs.is_active` のチェックが行われていない。

```go
func (h *Handler) Login(c *gin.Context) {
    // ... email/password 検証
    account, err := h.repos.Account.FindByEmail(ctx, req.Email)
    // ✅ パスワード検証
    if err := bcrypt.CompareHashAndPassword(...); err != nil { ... }
    
    // ❌ is_active チェックがない
    // staff := h.repos.Staff.FindByAccountID(ctx, account.ID)
    // if !staff.IsActive { return 401 }
    
    // JWT 発行
}
```

## 修正方針

### Login ハンドラに is_active チェック追加

```go
func (h *Handler) Login(c *gin.Context) {
    // ... email/password 検証後
    
    // スタッフの有効性チェック
    staff, err := h.repos.Staff.FindByAccountID(ctx, account.ID)
    if err != nil {
        RespondError(c, err)
        return
    }
    if !staff.IsActive {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "このアカウントは無効です"})
        return
    }
    
    // JWT 発行
}
```

### JWT ミドルウェアでも is_active チェック追加（推奨）

パスワード変更やアカウント無効化後に既存 JWT を無効化するため、
auth ミドルウェアでもリクエスト時に is_active を確認:

```go
func Auth(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // ... JWT 検証後
        // staff の is_active をチェック（DB 問い合わせまたはキャッシュ）
    }
}
```

**注意**: 毎リクエスト DB 問い合わせはパフォーマンスに影響。
Redis キャッシュまたは JWT の短い有効期限 + refresh token パターンで対応。

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/security.md` — Authentication
> "Use secure session management"

無効化ユーザーのセッション管理は認証の基本要件。

### `.claude/rules/api.md` — Security
> "Validate all user input"

ログインは最も重要な入力検証ポイント。ユーザーの状態（active/inactive）も検証すべき。

## 優先度
**High** — 退職者・停止アカウントがシステムにアクセス可能。本番環境で即時対応が必要。

## 関連チケット
- BUG-130: レートリミット（ログイン関連）
- BUG-131: パスワード更新（ログイン関連）

## 関連ファイル
- `backend/internal/handler/auth_handler.go` — Login ハンドラ
- `backend/internal/middleware/auth.go` — JWT ミドルウェア
- `backend/internal/repository/staff_repository.go` — FindByAccountID
