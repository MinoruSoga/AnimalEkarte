# BE: BUG-061 無効化されたユーザーの JWT が有効期限まで有効

## 概要

ユーザーを無効化・停止した後も、そのユーザーが持つ JWT は有効期限まで有効のまま。
停止後もログイン状態が維持される。

## 再現手順

1. ユーザーAでログイン（JWT取得）
2. 管理者がユーザーAを無効化（`is_active = false`）
3. ユーザーAのJWTで `/api/v1/me` にリクエスト
4. → HTTP 200（期待: 401）

## 期待する動作

- ユーザー無効化後、そのユーザーのJWTは即時無効化される
- 実装: 認証ミドルウェアで `users.is_active = true` チェック

## 実装場所

- `backend/internal/middleware/auth.go`
- JWT検証後に `users.is_active` を DB で確認するロジックを追加
- パフォーマンスを考慮し Redis キャッシュ（TTL: 5分）も検討

## 優先度

High（セキュリティ）

## 関連

- `docs/tasks/open/security/BUG-061_inactive_user_jwt.md`
- FUNCTIONAL_TEST_REPORT.md BUG-061
