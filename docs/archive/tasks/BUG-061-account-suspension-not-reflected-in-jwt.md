# BUG-061: アカウント停止後も JWT が有効期限まで通過する

**作成日**: 2026-03-29
**ステータス**: Open
**優先度**: High
**領域**: Backend
**関連**: BUG-055（Dual-Token 移行）, BUG-063

---

## 背景・問題

`user_accounts.account_status = 'inactive'` に変更しても、
そのユーザーが持つ JWT は有効期限（現在 24時間）が切れるまで全 API エンドポイントを通過する。

スタッフを解雇・アカウント停止しても即時アクセス遮断できない。

| シナリオ | 現在の動作 | 期待する動作 |
|---------|-----------|------------|
| スタッフを即時解雇・停止 | 最大 24時間アクセス継続 | 即時遮断 |
| パスワード漏洩によるアカウント停止 | 最大 24時間悪用可能 | 即時遮断 |
| 退職者アカウントの無効化 | 有効期限まで通過 | 即時遮断 |

---

## 原因

`middleware/auth.go` が JWT の署名・有効期限のみ検証し、DB の `account_status` / `deleted_at` を確認しない。

---

## 修正方針

認証ミドルウェアで DB から `account_status = 'active' AND deleted_at IS NULL` を確認する。
全リクエストで PK 検索が 1 回増えるが、`user_accounts` は小テーブルのため許容範囲。

---

## 派生イシュー

| イシュー | 領域 | 内容 |
|---------|------|------|
| BUG-061（BE） | Backend | `FindActiveByID` リポジトリメソッド追加 + `Auth` ミドルウェアで DB チェック |

詳細実装は `backend/issues/open/BUG-061-account-suspension-not-reflected-in-jwt.md` を参照。

---

## 完了条件

- [ ] `account_status = 'inactive'` に変更後、即座に 401 が返る
- [ ] `deleted_at IS NOT NULL` のユーザーが 401 になる
- [ ] 正常なアクティブユーザーは影響を受けない
- [ ] `docker compose exec backend go build ./...` 成功

---

## 依存・関連

- **BUG-063**: 論理削除ユーザーのログイン可能バグと同時修正を推奨（`FindActiveByID` で両方解消）
- **BUG-055**: Dual-Token 移行完了後はアクセストークン 15分化により最大遅延がさらに縮小
