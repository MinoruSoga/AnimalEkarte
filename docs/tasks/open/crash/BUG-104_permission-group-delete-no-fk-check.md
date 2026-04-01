# BUG-104: 権限グループ削除（FK制約なし）→ ユーザー権限が即時剥奪

## 概要

`DELETE /api/v1/permission-groups/:id` が権限グループに割り当てられているユーザーの存在チェックをせず
204 No Content で削除成功してしまう。割り当てられていたユーザーの権限が即時剥奪され、
予期せぬ権限昇格・剥奪が発生する。

## 症状

- `/settings/permission-groups` で「一般」グループの削除を実行
- DELETE /api/v1/permission-groups/3 → **204 No Content**
- リスト 3件 → 2件（削除成功）
- 「一般」グループに割り当てられていたスタッフのアクセス権限が消滅

## 期待する動作

- ユーザー（staff）が当該権限グループに割り当てられている場合 → **409 Conflict**
- エラートースト: 「このグループは使用中のため削除できません。先にユーザーの権限グループを変更してください」
- リストは変化しない

## 根本原因

`backend/internal/service/permission_group_service.go` の Delete メソッドに依存チェックがない。
`staffs.permission_group_id` FK 参照が存在するかのチェックがない。

## 修正内容

```go
// permission_group_service.go Delete() 内
count, err := s.repo.CountStaffsByPermissionGroupID(ctx, id)
if err != nil {
    return apperrors.Wrap(err, "failed to check permission group dependencies")
}
if count > 0 {
    return apperrors.ErrConflict
}
```

## 影響ファイル

- `backend/internal/service/permission_group_service.go`
- `backend/internal/repository/permission_group_repository.go`（CountStaffsByPermissionGroupID 追加）

## 優先度

High（セキュリティ影響：権限の意図せぬ削除）

## 関連

- BUG-030（同系統バグ。権限グループは修正漏れ）
- テスト確認日: 2026-04-01（ローカル環境）
- DELETE http://localhost:8080/api/v1/permission-groups/3 [204] 確認
