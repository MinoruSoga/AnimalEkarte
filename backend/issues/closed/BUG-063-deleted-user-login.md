# BE: BUG-063 削除済みユーザーがログインできる

## 概要

`deleted_at` が設定された論理削除済みユーザーでもログインが成功する。
認証処理で `deleted_at IS NULL` チェックが行われていない。

## 再現手順

1. ユーザーAを論理削除（`deleted_at` をセット）
2. ユーザーAの認証情報でログイン
3. → ログイン成功（期待: 401 または「アカウントが無効です」）

## 期待する動作

- 論理削除済みユーザーはログインできない
- エラーメッセージ: 「このアカウントは無効です」（401）

## 実装場所

- `backend/internal/service/auth_service.go` の Login メソッド
- ユーザー取得クエリに `deleted_at IS NULL` 条件を追加（GORMのSoftDelete が設定されていれば自動処理されるはず）
- `is_active = true` チェックも同時に追加すること（BUG-061 と合わせて）

## 優先度

High（セキュリティ）

## 関連

- `docs/tasks/open/security/BUG-063_deleted_user_can_login.md`
- FUNCTIONAL_TEST_REPORT.md BUG-063
