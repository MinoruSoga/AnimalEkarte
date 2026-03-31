# BUG-063: 論理削除ユーザーがログイン可能

**作成日**: 2026-03-29
**ステータス**: Open
**優先度**: Medium
**領域**: Backend
**関連**: BUG-061（アカウント停止の即時反映）

---

## 背景・問題

`user_accounts.deleted_at IS NOT NULL`（論理削除済み）のユーザーがログインエンドポイントで認証を通過できる。

`account_status = 'active'` はチェックしているが、`deleted_at IS NULL` を確認していない。

---

## 修正方針

`FindByEmail` に `AND deleted_at IS NULL` を追加する（またはサービス層でチェックを追加）。
エラーメッセージは「メールアドレスまたはパスワードが正しくありません」で統一し、削除済みであることを外部に漏らさない。

**BUG-061 との関係**: BUG-061 の `FindActiveByID` 実装時に `deleted_at IS NULL` を複合チェックするため、同時修正を推奨。

---

## 派生イシュー

| イシュー | 領域 | 内容 |
|---------|------|------|
| BUG-063（BE） | Backend | `FindByEmail` / `auth_service.Login` に `deleted_at IS NULL` チェック追加 |

詳細実装は `backend/issues/open/BUG-063-deleted-user-can-login.md` を参照。

---

## 完了条件

- [ ] `deleted_at IS NOT NULL` のユーザーでログインすると 401 が返る
- [ ] `account_status = 'inactive'` のユーザーでログインすると 401 が返る
- [ ] エラーメッセージが「メールアドレスまたはパスワードが正しくありません」で統一されている
- [ ] 正常なアクティブユーザーのログインに影響しない
- [ ] `docker compose exec backend go build ./...` 成功

---

## 依存・関連

- **BUG-061** と同時修正を推奨（`FindActiveByID` で両方解消）
