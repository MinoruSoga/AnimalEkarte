---
status: closed
---

# BE: BUG-104 権限グループ削除（FK依存チェックなし）→ 204 で権限即時剥奪

## 概要

`DELETE /api/v1/permission-groups/:id` がスタッフの存在チェックをせず 204 を返す。
割り当て済みスタッフの権限が即時剥奪される（セキュリティ影響あり）。

## 再現手順

```
DELETE /api/v1/permission-groups/3 → 204 No Content（スタッフ割り当て存在時も削除成功）
```

## 期待する動作

- `staffs.permission_group_id = :id` が存在する場合 → **409 Conflict**
- エラー: 「このグループは使用中のため削除できません。先にユーザーの権限グループを変更してください」

## 実装場所

- `backend/internal/service/permission_group_service.go` — Delete() に依存チェック追加
- `backend/internal/repository/permission_group_repository.go` — `CountStaffsByPermissionGroupID(ctx, id)` 追加

```go
count, err := s.repo.CountStaffsByPermissionGroupID(ctx, id)
if err != nil {
    return apperrors.Wrap(err, "failed to check permission group dependencies")
}
if count > 0 {
    return apperrors.ErrConflict
}
```

## 優先度

High（セキュリティ影響：権限の意図せぬ削除）

## 関連

- BUG-030（同種パターン）
- `docs/tasks/open/crash/BUG-104_permission-group-delete-no-fk-check.md`
