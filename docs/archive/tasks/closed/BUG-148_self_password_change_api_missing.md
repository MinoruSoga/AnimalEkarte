# BUG-148: 自分のパスワード変更 API エンドポイントが未実装

## 概要
サイドバーの「パスワード変更」ボタンからダイアログが開き、フロントエンドの `ChangePasswordDialog`
コンポーネントと `changeMyPassword` API 関数は実装済みだが、バックエンドの
`PUT /api/v1/users/me/password` エンドポイントが存在しない（404）。

ユーザーが自分のパスワードを変更する機能が完全に動作しない。

## 脆弱性分類
- **機能未実装**（セキュリティ関連）
- **影響**: ユーザーが自分でパスワードを変更できない。管理者（master-staff edit 権限）を経由するしかない。

## 再現手順
1. 任意のユーザーでログイン
2. サイドバー左下の鍵アイコン（パスワード変更）をクリック
3. ダイアログが開く
4. 現在のパスワード、新しいパスワード、確認を入力して「変更する」をクリック
5. **結果**: エラー（API 404）

## API テスト結果
```bash
curl -X PUT /api/v1/users/me/password \
  -H 'Content-Type: application/json' \
  -d '{"current_password":"password","new_password":"newpass123"}'
# → 404 Not Found
```

## フロントエンド実装

### `frontend/src/features/auth/api/change-password.ts`
```typescript
export const changeMyPassword = async (input: ChangePasswordInput): Promise<void> => {
  await axios.put("/v1/users/me/password", input);  // ← 404
};
```

### `frontend/src/features/auth/components/ChangePasswordDialog.tsx`
- ダイアログ UI 実装済み
- 現在のパスワード、新しいパスワード、確認用パスワードの3フィールド
- フロントエンドバリデーション: 8文字以上、確認一致チェック

## バックエンド実装状況

### `backend/internal/handler/handler.go`
```go
// /api/v1/users/me/password のルートが登録されていない
protected.GET("/me", h.GetMe)
// ← PUT /users/me/password がない
```

### `backend/internal/handler/`
`ChangePassword` や `UpdateMyPassword` に該当するハンドラファイルが存在しない。

## 修正方針

### 1. ハンドラ実装 (`auth_handler.go`)

```go
func (h *Handler) ChangeMyPassword(c *gin.Context) {
    var req struct {
        CurrentPassword string `json:"current_password" binding:"required"`
        NewPassword     string `json:"new_password"     binding:"required,min=8"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
        return
    }

    staffID, _ := extractStaffID(c)
    staff, _ := h.repos.Staff.FindByID(ctx, staffID)
    account, _ := h.repos.Account.GetByID(ctx, *staff.AccountID)

    // 現在のパスワード検証
    if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(req.CurrentPassword)); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "現在のパスワードが正しくありません"})
        return
    }

    // 新しいパスワードをハッシュ化して更新
    hashed, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
    account.PasswordHash = string(hashed)
    h.repos.Account.Update(ctx, account)

    c.JSON(http.StatusOK, gin.H{"message": "パスワードを変更しました"})
}
```

### 2. ルート登録 (`handler.go`)

```go
protected.PUT("/users/me/password", h.ChangeMyPassword)
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/security.md` — Authentication
> "Implement proper password hashing (bcrypt/argon2)"

パスワード変更はセキュリティの基本機能。現在のパスワード検証 + bcrypt ハッシュ化が必要。

### `.claude/CLAUDE.md` — エラー処理の統一 (MANDATORY)
> Handler: `RespondError(c, err)` で統一レスポンス

## 優先度
**High** — ユーザーが自分でパスワードを変更できない。セキュリティ上の基本機能。
管理者のみが PATCH /masters/staffs/:id で変更可能だが、一般ユーザーは自分のパスワードすら変更不能。

## 関連チケット
- BUG-131（修正済み）: 管理者によるスタッフパスワード更新
- BUG-139: パスワード複雑性チェックなし

## 関連ファイル
- `frontend/src/features/auth/api/change-password.ts` — フロントエンド API（実装済み）
- `frontend/src/features/auth/components/ChangePasswordDialog.tsx` — ダイアログ UI（実装済み）
- `backend/internal/handler/auth_handler.go` — ハンドラ追加対象
- `backend/internal/handler/handler.go` — ルート登録追加
