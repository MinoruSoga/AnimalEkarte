# TASK-042: BE Clinic/UserAccount の認可チェック実装

**作成日**: 2026-03-27
**ステータス**: Open
**優先度**: Critical
**領域**: Backend / Security

---

## 概要

認証済みであれば誰でも以下の操作が可能な状態になっており、権限昇格・他テナント破壊のリスクがある。

1. **Clinic**: staff/clinic_admin が任意のクリニックを作成・削除できる
2. **UserAccount**: 任意ユーザーが自分の `user_type` を `system_admin` に変更できる
3. **UserAccount**: `ListUsers` のクエリパラメータ `clinic_id` を上書きすることで別テナントのユーザー一覧を取得できる

---

## 対象箇所

### SEC-C02: Clinic CRUD 認可不足
- `handler/clinic_handler.go:76-100` — `CreateClinic`
- `handler/clinic_handler.go:103-114` — `DeleteClinic`
- `handler/clinic_handler.go:15-22` — `ListClinics`（全クリニック返却）

**修正**: `CreateClinic`/`DeleteClinic` は `system_admin` のみ許可。`ListClinics` はユーザーの所属クリニックのみ返す。

### SEC-C03: UserAccount 権限昇格
- `handler/user_account_handler.go:131-159` — `UpdateUser` で `user_type` を任意変更可能
- `handler/user_account_handler.go:107-127` — `SetUserPermissionGroups` で自身に全権限付与可能
- `handler/user_account_handler.go:22-29` — `ListUsers` の clinic_id クエリパラメータ上書き

**修正**:
- `user_type` の変更は `system_admin` のみ許可（RBAC ミドルウェアまたは handler 内チェック）
- `SetUserPermissionGroups` は `clinic_admin` 以上のみ許可
- `ListUsers` の clinic_id クエリパラメータ上書きを削除（JWT の clinic_id のみ使用）

---

## 修正パターン

```go
// handler 内でロールチェック
func (h *Handler) UpdateUser(c *gin.Context) {
    currentUser, ok := extractAuthUser(c)
    if !ok { return }

    if req.UserType != nil && currentUser.UserType != model.UserTypeSystemAdmin {
        RespondError(c, apperrors.WrapForbidden("user_type change requires system_admin"))
        return
    }
    // ...
}
```

---

## 受入条件

- [ ] `CreateClinic`/`DeleteClinic` は `system_admin` 以外から 403 を返す
- [ ] `ListClinics` は自分の所属クリニックのみ返す
- [ ] `user_type` 変更は `system_admin` 以外から 403 を返す
- [ ] `SetUserPermissionGroups` は `clinic_admin` 未満から 403 を返す
- [ ] `ListUsers` の clinic_id クエリパラメータ上書きが削除されている
- [ ] `docker compose exec backend go test ./...` 全テストパス
