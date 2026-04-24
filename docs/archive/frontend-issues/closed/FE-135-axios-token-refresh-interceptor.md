# FE-135: Axios インターセプター — 401 時のトークン自動リフレッシュ

**Status**: Open
**Priority**: High
**Affects**: lib/axios.ts
**Date Created**: 2026-03-29
**Related**: BUG-055, BE-078（先に完了必要）
**Blocked by**: BE-078（`POST /v1/auth/refresh` エンドポイントが必要）

---

## Summary

BE-078 でデュアルトークン方式が実装されると、アクセストークンの有効期限が **24時間 → 15分** に短縮される。
このとき、フロントエンドが 401 を受け取るたびに手動でリフレッシュするのは現実的でない。

Axios インターセプターで 401 を検知し、`POST /v1/auth/refresh` を自動呼び出しして
元のリクエストをリトライする仕組みを実装する。

---

## 現状

```typescript
// lib/axios.ts（現状）
import axios from "axios";

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  withCredentials: true,  // httpOnly Cookie 送信
});

export default api;
```

401 を受け取った場合のリトライ処理がない。
アクセストークンが 15分で切れると、全 API リクエストが 401 で失敗しログイン画面に飛ばされる。

---

## 実装

### `lib/axios.ts` の変更

```typescript
import axios, { AxiosError, AxiosRequestConfig } from "axios";

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  withCredentials: true,  // httpOnly Cookie を常に送信
});

// リフレッシュ中の並行リクエストを待機させるキュー
let isRefreshing = false;
let pendingRequests: Array<{
  resolve: (value: unknown) => void;
  reject: (reason?: unknown) => void;
}> = [];

// リフレッシュ完了後に待機中リクエストを再実行
function processQueue(error: AxiosError | null) {
  for (const pending of pendingRequests) {
    if (error) {
      pending.reject(error);
    } else {
      pending.resolve(undefined);
    }
  }
  pendingRequests = [];
}

// レスポンスインターセプター
api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as AxiosRequestConfig & { _retry?: boolean };

    // 401 以外、またはリフレッシュ自体が失敗した場合はそのままエラーを返す
    if (
      error.response?.status !== 401 ||
      originalRequest._retry ||
      originalRequest.url === "/v1/auth/refresh"
    ) {
      return Promise.reject(error);
    }

    if (isRefreshing) {
      // すでにリフレッシュ中なら、完了を待ってリトライ
      return new Promise((resolve, reject) => {
        pendingRequests.push({
          resolve: () => resolve(api(originalRequest)),
          reject,
        });
      });
    }

    originalRequest._retry = true;
    isRefreshing = true;

    try {
      // refresh_token Cookie を使ってアクセストークンを更新
      await api.post("/v1/auth/refresh");
      processQueue(null);
      return api(originalRequest);
    } catch (refreshError) {
      processQueue(refreshError as AxiosError);
      // リフレッシュ失敗 = セッション切れ → ログインページへ
      window.location.href = "/login";
      return Promise.reject(refreshError);
    } finally {
      isRefreshing = false;
    }
  },
);

export default api;
```

---

## 設計の説明

### 並行リクエストのキューイング

アクセストークンが切れた瞬間に複数の API リクエストが同時に走ることがある（例: ページロード時）。
`isRefreshing` フラグでリフレッシュが1回だけ実行されるよう制御し、
並行リクエストは `pendingRequests` キューで待機させてリフレッシュ後に一括リトライする。

### リフレッシュ自体の 401 は再帰しない

`originalRequest.url === "/v1/auth/refresh"` チェックにより、
リフレッシュ自体が 401 を返した場合（リフレッシュトークン期限切れ・revoke 済み）は
インターセプターが再帰しない。そのままログインページへリダイレクトする。

### `_retry` フラグ

1回リフレッシュしてリトライしたリクエストに `_retry: true` を付けることで、
リトライ後も 401 が返った場合（権限なし等）に再度リフレッシュしない。

---

## BE-078 との連携

| 項目 | BE-078 | FE-135 |
|------|--------|--------|
| リフレッシュエンドポイント | `POST /v1/auth/refresh` を実装 | `await api.post("/v1/auth/refresh")` で呼び出し |
| Cookie 管理 | `Set-Cookie` でトークンを更新 | `withCredentials: true` で自動送受信 |
| ログアウト時の revoke | `POST /v1/auth/logout` で DB revoke | 既存のログアウト処理を呼ぶだけ（変更不要） |

---

## テスト

```typescript
// lib/axios.test.ts（新規作成）

import { describe, it, expect, vi, beforeEach } from "vitest";
import MockAdapter from "axios-mock-adapter";
import api from "./axios";

describe("Axios interceptor", () => {
  const mock = new MockAdapter(api);

  beforeEach(() => {
    mock.reset();
  });

  it("401 を受け取ったとき /v1/auth/refresh を呼ぶ", async () => {
    mock.onGet("/v1/owners").replyOnce(401).onGet("/v1/owners").reply(200, []);
    mock.onPost("/v1/auth/refresh").reply(200);

    await api.get("/v1/owners");

    expect(mock.history.post.some((r) => r.url === "/v1/auth/refresh")).toBe(true);
  });

  it("リフレッシュが失敗したとき window.location.href が /login になる", async () => {
    const hrefSpy = vi.spyOn(window, "location", "get").mockReturnValue({
      ...window.location,
      href: "",
    } as Location);

    mock.onGet("/v1/owners").reply(401);
    mock.onPost("/v1/auth/refresh").reply(401);

    await expect(api.get("/v1/owners")).rejects.toThrow();

    hrefSpy.mockRestore();
  });

  it("リフレッシュ中の並行リクエストは完了後にリトライされる", async () => {
    let refreshCount = 0;
    mock
      .onGet("/v1/owners").reply(401)
      .onGet("/v1/owners").reply(200, [])
      .onGet("/v1/pets").reply(401)
      .onGet("/v1/pets").reply(200, []);
    mock.onPost("/v1/auth/refresh").reply(() => {
      refreshCount++;
      return [200];
    });

    await Promise.all([api.get("/v1/owners"), api.get("/v1/pets")]);

    // リフレッシュは1回だけ
    expect(refreshCount).toBe(1);
  });
});
```

---

## 変更ファイル一覧

| ファイル | 変更内容 |
|---------|---------|
| `lib/axios.ts` | レスポンスインターセプター追加（401 → refresh → retry） |
| `lib/axios.test.ts` | インターセプターの単体テスト追加 |

---

## 受入条件

- [ ] 401 レスポンスを受け取ったとき `POST /v1/auth/refresh` が自動的に呼ばれる
- [ ] リフレッシュ成功後、元のリクエストが自動的にリトライされる
- [ ] リフレッシュ失敗（401）時に `/login` にリダイレクトされる
- [ ] 複数の並行リクエストが 401 を受け取ったとき、リフレッシュは1回だけ実行される
- [ ] リフレッシュ中の並行リクエストはリフレッシュ完了後に一括リトライされる
- [ ] `docker compose exec frontend pnpm build` 成功
- [ ] `docker compose exec frontend pnpm lint` エラー 0 件
- [ ] `docker compose exec frontend pnpm test:run` 成功
