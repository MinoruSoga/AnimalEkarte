# BUG-058: 権限グループの取得・更新・削除で clinic_id チェックなし（クロステナントアクセス可能）

**Status**: Closed（Superseded by TASK-049 / BE-082）
> 権限グループを company 単位に変更する（TASK-049）ことで、クロスクリニックアクセスの問題が構造的に解消される。本バグは TASK-049 完了後に無効となる。
**Priority**: Critical
**Affects**: internal/handler/permission_group_handler.go, internal/service/permission_group_service.go, internal/repository/permission_group_repository.go
**Date Created**: 2026-03-29
**Related**: BUG-056, TASK-048

---

## Summary

`GET/PUT/DELETE /v1/permission-groups/:id` および `POST /v1/permission-groups/:id/rules`
において、ハンドラーが `clinic_id` を検証していない。

**クリニック A のユーザーが、ID を推測してクリニック B の権限グループを
取得・変更・削除できる。**

マルチテナントの根幹を脅かす CRITICAL バグ。

---

## 再現手順

```bash
# クリニック A のユーザー JWT でクリニック B のグループを取得
curl http://localhost:8080/api/v1/permission-groups/999 \
  -H "Cookie: access_token=<clinic_A_jwt>"
# → 200 OK（本来は 404 または 403 であるべき）

# クリニック A のユーザーがクリニック B のグループルールを書き換え
curl -X POST http://localhost:8080/api/v1/permission-groups/999/rules \
  -H "Cookie: access_token=<clinic_A_jwt>" \
  -d '{"rules": [{"resource": "medical-records", "canDelete": true}]}'
# → 200 OK（本来は 404 または 403 であるべき）
```

---

## 原因

```go
// permission_group_handler.go（現状）
func (h *Handler) GetPermissionGroup(c *gin.Context) {
    id, ok := extractGroupID(c)
    if !ok {
        return
    }
    // ❌ clinic_id を取得・検証していない
    group, err := h.svc.PermissionGroup.GetByID(c.Request.Context(), id)
    // ...
}

// permission_group_service.go（現状）
func (s *permissionGroupService) GetByID(ctx context.Context, id uint64) (*model.PermissionGroup, error) {
    // ❌ clinic_id を受け取らず、id のみで検索
    return s.repo.FindByID(ctx, id)
}

// permission_group_repository.go（現状）
func (r *permissionGroupRepository) FindByID(ctx context.Context, id uint64) (*model.PermissionGroup, error) {
    var group model.PermissionGroup
    // ❌ clinic_id で WHERE していない
    err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&group).Error
    return &group, err
}
```

---

## 影響を受けるエンドポイント

| エンドポイント | 操作 | リスク |
|--------------|------|--------|
| `GET /v1/permission-groups/:id` | 他クリニックのグループ詳細・ルールを取得 | 機密情報漏洩 |
| `PUT /v1/permission-groups/:id` | 他クリニックのグループ名・説明を変更 | データ改ざん |
| `DELETE /v1/permission-groups/:id` | 他クリニックのグループを削除 | サービス破壊 |
| `POST /v1/permission-groups/:id/rules` | 他クリニックの権限ルールを書き換え | 権限昇格 |

---

## 修正方針

全エンドポイントで `clinic_id` を取得し、service/repository まで伝達する。

### ハンドラー

```go
// permission_group_handler.go（修正後）
func (h *Handler) GetPermissionGroup(c *gin.Context) {
    id, ok := extractGroupID(c)
    if !ok {
        return
    }
    clinicID, ok := extractClinicID(c)  // ← 追加
    if !ok {
        return
    }
    group, err := h.svc.PermissionGroup.GetByID(c.Request.Context(), id, clinicID)  // ← clinicID 追加
    if err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusOK, toPermissionGroupResponse(group))
}
```

同様に `UpdatePermissionGroup`, `DeletePermissionGroup`, `SetPermissionGroupRules` も修正する。

### サービス

```go
// permission_group_service.go（修正後）
func (s *permissionGroupService) GetByID(
    ctx context.Context,
    id, clinicID uint64,  // ← clinicID 追加
) (*model.PermissionGroup, error) {
    return s.repo.FindByID(ctx, id, clinicID)
}
```

### リポジトリ

```go
// permission_group_repository.go（修正後）
func (r *permissionGroupRepository) FindByID(
    ctx context.Context,
    id, clinicID uint64,  // ← clinicID 追加
) (*model.PermissionGroup, error) {
    var group model.PermissionGroup
    err := r.db.WithContext(ctx).
        Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", id, clinicID).  // ← clinic_id 追加
        First(&group).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, apperrors.WrapNotFound(fmt.Errorf("permission group %d not found", id))
    }
    return &group, err
}
```

`clinic_id` が一致しないグループは `ErrNotFound` を返す
（`403 Forbidden` ではなく `404 Not Found` にすることでグループ存在有無の情報漏洩を防ぐ）。

---

## 確認コマンド

```bash
docker compose exec backend go build ./...
docker compose exec backend golangci-lint run ./...
docker compose exec backend go test ./... -v
```

---

## 受入条件

- [ ] 他クリニックの権限グループ ID を指定した `GET /v1/permission-groups/:id` が `404` を返す
- [ ] 他クリニックの権限グループ ID を指定した `PUT /v1/permission-groups/:id` が `404` を返す
- [ ] 他クリニックの権限グループ ID を指定した `DELETE /v1/permission-groups/:id` が `404` を返す
- [ ] 他クリニックの権限グループ ID を指定した `POST /v1/permission-groups/:id/rules` が `404` を返す
- [ ] 同一クリニックの権限グループは従来通り操作できる
- [ ] `docker compose exec backend go build ./...` 成功
- [ ] `docker compose exec backend golangci-lint run ./...` エラー 0 件
