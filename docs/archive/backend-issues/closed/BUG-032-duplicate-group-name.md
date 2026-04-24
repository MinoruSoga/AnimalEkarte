# BE: BUG-032 権限グループ名の重複作成が可能

## 概要

権限グループ作成時に既存のグループ名と同じ名前で新規作成が成功してしまう。
UNIQUE制約がないか、サービス層で重複チェックが未実装。

## 再現手順

```
POST /api/v1/permission-groups
{"name": "管理者"}（既存と同名）
→ HTTP 201 Created（期待: 409 Conflict）
```

## 期待する動作

- 同名グループが既存の場合: 409 Conflict を返す
- エラーメッセージ: `"permission group name already exists"`

## 実装場所

- `backend/internal/service/permission_group_service.go` の Create メソッド
  - 同名チェック追加 → 既存なら `apperrors.ErrConflict` を返す
- または DB マイグレーションで `permission_groups.name` に UNIQUE制約を追加

## 優先度

Medium

## 関連

- `docs/tasks/open/security/BUG-032_duplicate_group_name.md`
- FUNCTIONAL_TEST_REPORT.md BUG-032
