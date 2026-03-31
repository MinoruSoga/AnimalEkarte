# TASK-043: JWT を httpOnly Cookie に移行（FE + BE）

**作成日**: 2026-03-27
**ステータス**: Closed
**優先度**: High
**領域**: Frontend + Backend / Security

---

## 概要

JWT トークンが `sessionStorage` に保存されており、XSS が発生した瞬間にトークンが盗まれる。
プロジェクトルール（CLAUDE.md）にも「httpOnly Cookie + `withCredentials: true`」が明記されているが未実装。

FE で `localStorage` に `userType`（権限情報）が保存されており、XSS で書き換えることで権限チェックのバイパスが可能になるリスクもある。

---

## 現状

```
# FE: sessionStorage に JWT を格納
sessionStorage.setItem("auth_token", data.token)  // login.ts:29
sessionStorage.getItem("auth_token")               // axios.ts:11

# FE: localStorage に権限情報を格納
localStorage.setItem("user", JSON.stringify({userType, ...}))  // use-auth.tsx:30-38
```

`auth.go` ミドルウェアは既に Cookie 読み取りをサポート（L28）。

---

## 修正方針

### Backend（`handler/auth_handler.go`）
1. Login レスポンスで `Set-Cookie: auth_token=xxx; HttpOnly; Secure; SameSite=Lax; Path=/` を設定
2. Logout で `Set-Cookie: auth_token=; Max-Age=0` で Cookie をクリア
3. `Authorization: Bearer` ヘッダ方式は削除可（または並存）

### Frontend
1. `login.ts` — `sessionStorage.setItem("auth_token", ...)` を削除
2. `logout.ts` — `sessionStorage.removeItem("auth_token")` を削除
3. `refresh-token.ts` — sessionStorage 読み取りを削除（Cookie は自動送信されるため不要）
4. `axios.ts` — `Authorization` ヘッダ手動設定を削除、`withCredentials: true` を設定
5. `use-auth.tsx` — `localStorage` への `userType`/`permissions` 書き込みを廃止
   - リロード後の状態復元は `GET /v1/me` のレスポンスのみを信頼する

---

## 注意事項

- 開発環境（http://localhost）では `Secure` フラグなしで動作させる必要がある（環境変数で切り替え）
- SameSite=Lax は POST リクエストには自動送信されないため、API は GET/PATCH/DELETE 中心の構成に問題なし
- `withCredentials: true` を設定すると CORS の `Access-Control-Allow-Origin: *` が使えなくなる → `cors.go` で `AllowOrigins` を明示する必要がある

---

## 受入条件

- [ ] Login レスポンスで httpOnly Cookie が発行されている
- [ ] `sessionStorage` に JWT が保存されなくなっている
- [ ] `localStorage` に `userType` が保存されなくなっている
- [ ] `axios.ts` で `withCredentials: true` が設定されている
- [ ] ページリロード後のセッション復元が `GET /v1/me` 経由で動作している
- [ ] `docker compose exec frontend npm run build` 成功
- [ ] `docker compose exec backend go build ./...` 成功
