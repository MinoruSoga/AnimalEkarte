# BUG-062: パスワード変更エンドポイントが未実装

**Status**: Open
**Priority**: High
**Affects**: internal/handler/user_account_handler.go, internal/service/user_account_service.go
**Date Created**: 2026-03-29
**Related**: BUG-060（パスワードリセット）, BUG-055

---

## Summary

ログイン中のユーザーが自分のパスワードを変更するエンドポイントが存在しない。

- `PUT /v1/users/:id` にはパスワードフィールドが含まれていない
- パスワード変更手段がなく、初期パスワードを使い続けるリスクがある
- clinic_admin が手動対応するしかなく、運用コストが高い

---

## 現状

```go
// user_account_handler.go — UpdateUser()
type updateUserRequest struct {
    DisplayName *string          `json:"display_name"`
    AvatarURL   *string          `json:"avatar_url"`
    UserType    *string          `json:"user_type"`
    Status      *string          `json:"status"`
    StaffID     *uint64          `json:"staff_id"`
    // ❌ password フィールドなし
}
```

---

## 実装方針

### エンドポイント

`PUT /v1/users/me/password` — 認証済みユーザーが自分のパスワードを変更する

```go
type changePasswordRequest struct {
    CurrentPassword string `json:"current_password" binding:"required"`
    NewPassword     string `json:"new_password"     binding:"required,min=8"`
}
```

**設計上の注意**:
- `:id` ではなく `/me/password` とし、他ユーザーのパスワードを変更できないようにする
- `current_password` を必須にすることで、セッションを奪われた場合のリスクを低減
- clinic_admin が他ユーザーのパスワードをリセットする場合は BUG-060 のリセットフローを使う

### サービス実装

```go
// user_account_service.go
func (s *userAccountService) ChangePassword(
    ctx context.Context,
    userID uint64,
    currentPassword, newPassword string,
) error {
    account, err := s.repo.FindByID(ctx, userID)
    if err != nil {
        return fmt.Errorf("user not found: %w", apperrors.ErrNotFound)
    }

    // 現在のパスワードを検証
    if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(currentPassword)); err != nil {
        return fmt.Errorf("current password is incorrect: %w", apperrors.ErrInvalidInput)
    }

    // 新パスワードをハッシュ化
    hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
    if err != nil {
        return fmt.Errorf("failed to hash password: %w", err)
    }

    return s.repo.UpdatePasswordHash(ctx, userID, string(hash))
}
```

### パスワード変更後のセッション処理

パスワード変更後、**既存セッションの扱いは BUG-055 の実装状況に依存**する：

- BUG-055 未対応の場合: Cookie の JWT はそのまま（24時間後に期限切れ）
- BUG-055 対応後（Dual-Token）: `refresh_tokens` テーブルの全レコードを revoke することで全端末ログアウト

BUG-055 対応後は以下を追加する：

```go
// パスワード変更後、全 refresh_token を revoke
_ = s.authRepo.RevokeAllUserTokens(ctx, userID)
```

---

## 受入条件

- [ ] `PUT /v1/users/me/password` で自分のパスワードを変更できる
- [ ] `current_password` が正しくない場合 400 を返す
- [ ] 新パスワードが 8 文字未満の場合 400 を返す
- [ ] 他ユーザーのパスワードを変更しようとしても自分のパスワードのみ変更される
- [ ] `docker compose exec backend go build ./...` 成功
- [ ] `docker compose exec backend golangci-lint run ./...` エラー 0 件
