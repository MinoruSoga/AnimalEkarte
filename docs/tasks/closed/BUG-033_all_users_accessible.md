# BUG-033: staffユーザーが全ユーザー情報を閲覧できる（バックエンドアクセス制御未実装）

## 概要
一般グループ（staff）のユーザーが `GET /api/v1/users` を呼び出すと、HTTP 200 で全10ユーザーの情報が取得できる。
バックエンドのアクセス制御が未実装。

## 再現手順
1. 山田花子（vet@example.com）でログイン
2. `GET /api/v1/users` を呼び出す
3. → HTTP 200 で全ユーザー情報（メール・氏名・権限グループ等）が返る

## 期待する動作
- `admin:can_view=false` のユーザーは他ユーザー情報を参照できない（403 Forbidden）
- 管理者のみが全ユーザー一覧を参照可能

## 実装場所
- `backend/internal/handler/user_handler.go` の ListUsers ハンドラ
- 権限チェックミドルウェアを追加

## 優先度
High（個人情報漏洩リスク）

## 関連
- FUNCTIONAL_TEST_REPORT.md BUG-033
- テスト確認日: 2026-03-30
