# BE: BUG-031 存在しない group_id 指定で 500 エラー

## 概要

権限グループ割当 API に存在しない `group_id` を指定すると 500 エラーが返る。
404 Not Found または 400 Bad Request を返すべき。

## 再現手順

```
PUT /api/v1/users/:id/groups
{"group_ids": [9999]}
→ HTTP 500 Internal Server Error
```

## 期待する動作

- 存在しない group_id の場合: 404 Not Found
- エラーメッセージ: `"group_id 9999 not found"`

## 実装場所

- `backend/internal/service/` の権限グループ割当処理
- group_id 存在チェック → `apperrors.FromGORM` で 404 を返す

## 優先度

Medium

## 関連

- `docs/tasks/open/security/BUG-031_invalid_group_id_500.md`
- FUNCTIONAL_TEST_REPORT.md BUG-031
