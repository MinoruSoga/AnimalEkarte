# BUG-031: 存在しない group_id を指定すると 500 Internal Server Error

## 概要
権限グループ割当 API に存在しない group_id を指定すると 500 エラーが返る。
適切なバリデーション（404 Not Found）がない。

## 再現手順
1. `PUT /api/v1/users/:id/groups` に `{"group_ids":[9999]}` を送信
2. → HTTP 500 Internal Server Error

## 期待する動作
- 存在しない group_id の場合: 400 Bad Request または 404 Not Found
- エラーメッセージ: `"group_id 9999 not found"`

## 実装場所
- `backend/internal/service/` または `handler/` の権限グループ割当処理
- group_id の存在チェックを追加

## 優先度
Medium

## 関連
- FUNCTIONAL_TEST_REPORT.md BUG-031
- テスト確認日: 2026-03-30
