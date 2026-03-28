# BUG-031: 権限グループ割当で存在しない group_id を指定すると HTTP 500 が返る

## 種類
バグ（バックエンド — バリデーション未実装）

## 重要度
中

## 発見日
2026-03-28

## 再現手順
1. `PUT /api/v1/users/:id/permission-groups` に存在しない group_id を指定して送信
   ```json
   { "group_ids": [9999] }
   ```

## 期待動作
- HTTP 400 Bad Request または HTTP 404 Not Found が返る
- エラーメッセージ: 「指定した権限グループが存在しません」等

## 実際の動作
- HTTP 500 Internal Server Error が返る（`{"error":"internal server error"}`）

## 影響
- 存在しない group_id を指定した場合に情報が漏れない（ただし 500 はセキュリティ上望ましくない）
- バリデーションエラーとして適切に処理されない

## 修正方針
### バックエンド
- `backend/internal/service/` の SetUserPermissionGroups（または同等処理）で、group_ids の各 ID が存在するか検証
- 存在しない ID がある場合は `errors.ErrNotFound` を wrap して HTTP 404 or `errors.ErrInvalidInput` で HTTP 400 を返す

## 対象ファイル（推定）
- `backend/internal/service/user_service.go`（または permission_group_service.go）
