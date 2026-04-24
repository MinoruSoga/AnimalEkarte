# FE-142: ログイン失敗エラーメッセージが英語の生 HTTP エラーになる

**Status**: Open
**Priority**: Low
**Affects**: features/auth/
**Date Created**: 2026-03-29
**Related**: BUG-047

---

## Summary

ログイン失敗時に「Request failed with status code 401」という英語の生 HTTP エラーが表示される。
「メールアドレスまたはパスワードが違います」等の日本語メッセージに変換すべき。

---

## 実装手順

### 1. 原因調査

`features/auth/` のログインフォームコンポーネントおよびエラーハンドリングを確認：

```bash
grep -rn "error\|catch\|message" frontend/src/features/auth/
```

現在の実装（推測）:
```typescript
// ❌ axios の error.message をそのまま表示
setError(error.message); // "Request failed with status code 401"
```

### 2. 修正: 401 を日本語メッセージに変換

```typescript
import type { AxiosError } from "axios";

// features/auth/routes/LoginPage.tsx（または useLoginForm.ts）
const handleLoginError = (error: unknown) => {
  if (error instanceof AxiosError) {
    if (error.response?.status === 401) {
      return "メールアドレスまたはパスワードが違います";
    }
    if (error.response?.status === 429) {
      return "ログイン試行回数が多すぎます。しばらくしてからお試しください";
    }
  }
  return "ログインに失敗しました。しばらく経ってから再度お試しください";
};
```

`lib/error-messages.ts` にある `ERROR_MESSAGES` と統合するか、
`showErrorToast` ヘルパーを通じて表示することも可。

### 3. 表示方式の確認

現在フォーム内赤字表示 or トースト表示かを確認し、プロジェクト標準（Toaster）に統一する。

---

## 受入条件

- [ ] 誤認証時に「メールアドレスまたはパスワードが違います」が日本語で表示される
- [ ] axios の生エラーメッセージ（英語）が UI に露出しない
- [ ] `docker compose exec frontend pnpm lint` エラー 0 件
