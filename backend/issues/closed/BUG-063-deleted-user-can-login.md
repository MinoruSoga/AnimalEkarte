# BUG-063: 論理削除ユーザーがログイン可能

**Status**: Open
**Priority**: Medium
**Affects**: internal/repository/user_account_repository.go, internal/service/auth_service.go
**Date Created**: 2026-03-29
**Related**: BUG-061（アカウント停止の即時反映）

---

## Summary

`user_accounts.deleted_at IS NOT NULL`（論理削除済み）のユーザーが
ログインエンドポイントで認証を通過できる。

ログイン処理で `account_status = 'active'` はチェックしているが、
`deleted_at IS NULL` を確認していない。

---

## 現状

```go
// auth_service.go — Login（現状推定）
func (s *authService) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
    account, err := s.userRepo.FindByEmail(ctx, input.Email)
    if err != nil {
        return nil, apperrors.ErrUnauthorized
    }

    // ✅ status チェックあり
    if account.AccountStatus != model.AccountStatusActive {
        return nil, fmt.Errorf("account is not active: %w", apperrors.ErrUnauthorized)
    }

    // ❌ deleted_at チェックなし
    // deleted_at IS NOT NULL でも status = 'active' なら通過する

    // ... bcrypt 検証・JWT 発行 ...
}
```

```go
// repository/user_account_repository.go — FindByEmail（現状推定）
func (r *userAccountRepository) FindByEmail(ctx context.Context, email string) (*model.UserAccount, error) {
    var account model.UserAccount
    // ❌ deleted_at のフィルタなし
    err := r.db.WithContext(ctx).Where("email = ?", email).First(&account).Error
    return &account, err
}
```

---

## 修正方針

### リポジトリの `FindByEmail` に `deleted_at IS NULL` を追加

```go
// repository/user_account_repository.go（修正後）
func (r *userAccountRepository) FindByEmail(ctx context.Context, email string) (*model.UserAccount, error) {
    var account model.UserAccount
    err := r.db.WithContext(ctx).
        Where("email = ? AND deleted_at IS NULL", email).  // ← 追加
        First(&account).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, apperrors.ErrNotFound
    }
    return &account, err
}
```

### または auth_service.go でチェックを追加

```go
// auth_service.go（修正後）
if account.DeletedAt != nil {
    return nil, fmt.Errorf("account has been deleted: %w", apperrors.ErrUnauthorized)
}
if account.AccountStatus != model.AccountStatusActive {
    return nil, fmt.Errorf("account is not active: %w", apperrors.ErrUnauthorized)
}
```

**エラーメッセージは「存在しない or 無効」で統一**し、削除済みであることを外部に漏らさない。

```go
// ユーザー列挙攻撃対策: 削除済み・無効・パスワード不一致をすべて同じメッセージにする
return nil, fmt.Errorf("invalid credentials: %w", apperrors.ErrUnauthorized)
```

---

## BUG-061 との関係

BUG-061（ミドルウェアでの DB チェック追加）を実装する際に、
`FindActiveByID()` が `deleted_at IS NULL AND account_status = 'active'` を
複合チェックするため、本バグも合わせて解消される。

**優先度**: BUG-061 と同時に修正することを推奨する。

---

## 受入条件

- [ ] `deleted_at IS NOT NULL` のユーザーでログインすると 401 が返る
- [ ] `account_status = 'inactive'` のユーザーでログインすると 401 が返る
- [ ] エラーメッセージが「メールアドレスまたはパスワードが正しくありません」で統一されている（削除済みであることを漏らさない）
- [ ] 正常なアクティブユーザーのログインに影響しない
- [ ] `docker compose exec backend go build ./...` 成功
