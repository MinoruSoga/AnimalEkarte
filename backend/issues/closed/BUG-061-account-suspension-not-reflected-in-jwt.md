# BUG-061: アカウント停止後も JWT が有効期限まで通過する

**Status**: Open
**Priority**: High
**Affects**: internal/middleware/auth.go, internal/service/auth_service.go
**Date Created**: 2026-03-29
**Related**: BUG-055（Dual-Token 移行）, BUG-063

---

## Summary

`user_accounts.account_status = 'inactive'` に変更しても、
そのユーザーが持つ JWT は有効期限（現在 24時間）が切れるまで全 API エンドポイントを通過する。

スタッフを解雇・アカウント停止しても即時アクセス遮断できない。

---

## 現状の問題

```go
// middleware/auth.go（現状）
func Auth(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractToken(c)
        claims, err := validateJWT(token, jwtSecret)
        if err != nil {
            // JWT 署名・有効期限のみ検証
            RespondError(c, apperrors.ErrUnauthorized)
            c.Abort()
            return
        }
        // ❌ DB の account_status を確認しない
        c.Set("user_id", claims.UserID)
        c.Set("clinic_id", claims.ClinicID)
        c.Set("user_type", claims.UserType)
        c.Next()
    }
}
```

---

## 影響シナリオ

| シナリオ | 現在の動作 | 期待する動作 |
|---------|-----------|------------|
| スタッフを即時解雇・停止 | 最大 24時間アクセス継続 | 即時遮断 |
| パスワード漏洩によるアカウント停止 | 最大 24時間悪用可能 | 即時遮断 |
| 退職者アカウントの無効化 | 有効期限まで通過 | 即時遮断 |

---

## 修正方針

### 短期（推奨）: ミドルウェアで DB チェック

リクエストごとに DB から `account_status` と `deleted_at` を確認する。

```go
// middleware/auth.go（修正後）
func Auth(jwtSecret string, userRepo repository.UserAccountRepository) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractToken(c)
        claims, err := validateJWT(token, jwtSecret)
        if err != nil {
            RespondError(c, apperrors.ErrUnauthorized)
            c.Abort()
            return
        }

        // ✅ DB でアカウント状態を確認
        userID, _ := strconv.ParseUint(claims.UserID, 10, 64)
        account, err := userRepo.FindActiveByID(c.Request.Context(), userID)
        if err != nil || account == nil {
            // deleted_at IS NOT NULL または status != 'active' の場合
            RespondError(c, apperrors.ErrUnauthorized)
            c.Abort()
            return
        }

        c.Set("user_id", claims.UserID)
        c.Set("clinic_id", claims.ClinicID)
        c.Set("user_type", claims.UserType)
        c.Next()
    }
}
```

```go
// repository/user_account_repository.go
func (r *userAccountRepository) FindActiveByID(ctx context.Context, id uint64) (*model.UserAccount, error) {
    var account model.UserAccount
    err := r.db.WithContext(ctx).
        Where("id = ? AND account_status = 'active' AND deleted_at IS NULL", id).
        First(&account).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &account, err
}
```

**パフォーマンス**: 全リクエストで DB SELECT が1回増える。
`user_accounts` は小さいテーブル（従業員数 ≒ 数十〜数百件）でかつ PK 検索なので許容範囲。

### 長期（BUG-055 完了後）: アクセストークン短命化

BUG-055 の Dual-Token 移行でアクセストークンを 15分に短縮すると、
最大 15分でアカウント停止が反映される。DB チェックと組み合わせることが理想。

---

## 受入条件

- [ ] `account_status = 'inactive'` に変更後、即座に 401 が返る
- [ ] `deleted_at IS NOT NULL` のユーザーが 401 になる
- [ ] 正常なアクティブユーザーは影響を受けない
- [ ] `docker compose exec backend go build ./...` 成功
- [ ] `docker compose exec backend golangci-lint run ./...` エラー 0 件
- [ ] `docker compose exec backend go test ./... -v` 成功
