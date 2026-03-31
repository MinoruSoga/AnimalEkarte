# BUG-055: 認証 — 単一JWT24時間・サーバー側無効化不可

## 概要

現在の認証は JWT 1本（有効期限 24時間）のみで、サーバー側でトークンを
無効化する手段がない。ログアウトしてもトークンが盗まれていれば最大24時間
悪用され続ける。

## 重要度

**HIGH** — トークン漏洩時の被害ウィンドウが大きすぎる。

## 現状の問題点

| 問題 | 影響 |
|------|------|
| トークン有効期限が24時間 | 漏洩時に最大24時間悪用可能 |
| サーバー側に無効化手段がない | ログアウト・パスワード変更後も旧トークンが有効 |
| アカウント停止が即時反映されない | `account_status = 'inactive'` にしても既存セッションは継続 |
| リフレッシュトークンがない | 再認証の仕組みがなく、セッション継続性の設計が単純すぎる |

## 修正方針: Dual-Token 方式への移行

### 1. `refresh_tokens` テーブルの追加

```sql
CREATE TABLE refresh_tokens (
  id          BIGSERIAL PRIMARY KEY,
  user_id     BIGINT NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
  token_hash  TEXT NOT NULL UNIQUE,   -- SHA-256(平文トークン)
  expires_at  TIMESTAMPTZ NOT NULL,
  revoked_at  TIMESTAMPTZ,            -- NULL = 有効
  user_agent  TEXT,
  ip_address  INET,
  created_at  TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_hash ON refresh_tokens(token_hash);
```

### 2. ログインエンドポイントの変更 (`POST /v1/login`)

```
変更前: JWT(24時間) を Set-Cookie: auth_token
変更後: JWT(15分)  を Set-Cookie: access_token
        opaque(7日) を Set-Cookie: refresh_token
        refresh_tokens テーブルに SHA-256(opaque) を保存
```

### 3. リフレッシュエンドポイントの追加 (`POST /v1/auth/refresh`)

1. Cookie の `refresh_token` を SHA-256 でハッシュ
2. `refresh_tokens` で検索（期限・revoked_at IS NULL）
3. 旧トークンを `revoked_at = now()` で無効化（ローテーション）
4. 新 JWT(15分) + 新 refresh_token(7日) を発行・保存
5. 両 Cookie を更新して返す

### 4. ログアウトエンドポイントの変更 (`POST /v1/logout`)

```
変更前: Cookie を MaxAge=-1 でクリアするだけ
変更後: refresh_tokens テーブルの該当レコードを revoked_at = now() に更新
        → サーバー側でセッションを即時無効化
        両 Cookie を MaxAge=-1 でクリア
```

### 5. フロントエンド Axios インターセプター

```typescript
// レスポンスインターセプターで 401 を検知してリフレッシュ
axios.interceptors.response.use(
  (res) => res,
  async (error) => {
    if (error.response?.status === 401 && !error.config._retry) {
      error.config._retry = true;
      try {
        await axios.post("/v1/auth/refresh");  // Cookie 自動送信
        return axios(error.config);            // 元リクエストをリトライ
      } catch {
        window.location.href = "/login";
      }
    }
    return Promise.reject(error);
  }
);
```

### 6. 強制ログアウト (`DELETE /v1/users/:id/sessions`)

パスワード変更・アカウント停止時にサーバー側で全セッションを無効化する。

```go
// 対象ユーザーの全 refresh_tokens を revoked_at = now() に更新
db.Model(&RefreshToken{}).
  Where("user_id = ? AND revoked_at IS NULL", userID).
  Update("revoked_at", time.Now())
```

## 影響範囲

- `backend/migrations/001_init.sql` — `refresh_tokens` テーブル追加
- `backend/internal/model/` — `RefreshToken` モデル追加
- `backend/internal/handler/auth_handler.go` — Login/Logout/Refresh エンドポイント
- `backend/internal/service/auth_service.go` — トークン生成・検証ロジック
- `backend/internal/repository/` — `RefreshTokenRepository` 追加
- `frontend/src/lib/axios.ts` — レスポンスインターセプターでリフレッシュ処理
- `frontend/src/features/auth/api/refresh-token.ts` — リフレッシュ API

## 派生イシュー

| イシュー | 領域 | 内容 |
|---------|------|------|
| BE-078 | Backend | `refresh_tokens` テーブル追加・RefreshToken モデル・`POST /v1/auth/refresh` エンドポイント・ログアウト revoke 対応 |
| FE-135 | Frontend | Axios インターセプター — 401 時の自動リフレッシュ・並行リクエストキューイング・リフレッシュ失敗時のログインリダイレクト |

## 関連

- `docs/AUTH.md` §4 セッション管理・トークンリフレッシュフロー
- `docs/AUTH.md` §9 実装状態と残課題
