# BE: BUG-033 staff ユーザーが全ユーザー情報を閲覧できる

## 概要

一般グループ（staff）のユーザーが `GET /api/v1/users` を呼び出すと
HTTP 200 で全ユーザー情報（メール・氏名・権限グループ）が取得できる。
バックエンドのアクセス制御が未実装。

## 再現手順

1. staff ユーザーの JWT で `GET /api/v1/users` を呼び出す
2. → HTTP 200 で全10ユーザー情報が返る

## 期待する動作

- `admin:can_view=false` のユーザーは 403 Forbidden を返す
- 管理者のみが全ユーザー一覧を参照可能

## 実装場所

- `backend/internal/handler/user_handler.go` の ListUsers ハンドラ
- 権限チェック: JWTのロール or 権限グループの `admin:can_view` を確認

## 優先度

High（個人情報漏洩リスク）

## 関連

- `docs/tasks/open/security/BUG-033_all_users_accessible.md`
- FUNCTIONAL_TEST_REPORT.md BUG-033
