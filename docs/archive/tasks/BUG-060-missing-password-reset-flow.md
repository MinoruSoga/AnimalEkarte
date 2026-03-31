# BUG-060: パスワードリセット機能が未実装

**作成日**: 2026-03-29
**ステータス**: Open
**優先度**: Medium
**領域**: Backend + Frontend
**関連**: BUG-055（認証セキュリティ）

---

## 背景・問題

パスワードリセット機能が存在しない。

- ログインフォームに「パスワードをお忘れですか？」リンクがない
- `POST /v1/auth/forgot-password` エンドポイントが存在しない
- スタッフがパスワードを忘れると、`clinic_admin` が手動でパスワードを変更するしか手段がない
- `clinic_admin` 自身がパスワードを忘れると `system_admin` の手動対応が必要

---

## リスク

| リスク | 内容 |
|--------|------|
| **セキュリティ** | `clinic_admin` が手動リセットする場合、仮パスワードを別経路（電話・チャット）で伝えることになり、パスワードが平文で流出するリスクがある |
| **UX** | 自己解決できないため運用コストが高い |
| **医療情報システム** | ユーザーが自分でアカウントを回復できることはシステムの基本要件 |

---

## 実装方針

### フロー

```
1. ユーザーが /forgot-password にアクセス
   → メールアドレスを入力して送信

2. POST /v1/auth/forgot-password { email: "..." }
   → DB で email を検索
   → 見つかった場合: password_reset_tokens テーブルにワンタイムトークンを保存
   → メールでリセットリンクを送信（https://app.example.com/reset-password?token=xxx）
   → 見つからない場合: 同じ成功レスポンスを返す（ユーザー列挙攻撃対策）

3. ユーザーがリセットリンクをクリック
   → /reset-password?token=xxx ページが開く
   → 新パスワード × 2 を入力して送信

4. POST /v1/auth/reset-password { token: "xxx", password: "new_pass" }
   → token の有効期限・未使用を確認
   → パスワードを bcrypt でハッシュして更新
   → token を使用済みにする（ワンタイム）
   → 全セッション（refresh_tokens）を revoke
```

### DB スキーマ（`migrations/001_init.sql`）

```sql
CREATE TABLE password_reset_tokens (
    id         BIGSERIAL   PRIMARY KEY,
    user_id    bigint      NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
    token_hash varchar(64) NOT NULL UNIQUE,  -- SHA-256（hex 64文字）
    expires_at timestamp   NOT NULL,         -- 発行から 30 分
    used_at    timestamp   NULL,             -- NULL = 未使用
    created_at timestamp   DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_password_reset_tokens_hash ON password_reset_tokens(token_hash) WHERE used_at IS NULL;
```

### バックエンドエンドポイント

| エンドポイント | 処理 |
|--------------|------|
| `POST /v1/auth/forgot-password` | email でユーザー検索 → token 生成 → メール送信 |
| `POST /v1/auth/reset-password` | token 検証 → パスワード更新 → 全 session revoke |

### フロントエンドページ

| ページ | パス |
|--------|------|
| パスワードリセット申請 | `/forgot-password` |
| 新パスワード設定 | `/reset-password?token=xxx` |

---

## メール送信

`smtp` または SES を使用したメール送信実装が必要。
インフラ未整備の場合は、開発初期は `slog` でトークンをログに出力する簡易実装でも可。

```go
// 開発環境での簡易実装（本番前に差し替え必須）
if cfg.Env == "development" {
    slog.Info("password reset link",
        "email", email,
        "token", rawToken,  // ← 本番では絶対にログに出さない
    )
    return nil
}
// 本番: メール送信
return s.mailer.Send(email, buildResetEmail(rawToken))
```

---

## 派生イシュー

| イシュー | 領域 | 内容 |
|---------|------|------|
| BE-081 | Backend | `password_reset_tokens` テーブル + forgot/reset エンドポイント実装 |
| FE-138 | Frontend | `/forgot-password` + `/reset-password` ページ実装 |

---

## 完了条件

- [ ] ログインフォームに「パスワードをお忘れですか？」リンクがある
- [ ] `POST /v1/auth/forgot-password` にメールアドレスを送るとリセットメールが届く（dev は log 出力）
- [ ] リセットリンクのトークンは 30 分で期限切れになる
- [ ] トークンはワンタイム（一度使用すると無効）
- [ ] `POST /v1/auth/reset-password` 成功後、全既存セッションが無効化される
- [ ] 存在しないメールアドレスを `forgot-password` に送っても同じ成功レスポンスが返る（ユーザー列挙対策）
