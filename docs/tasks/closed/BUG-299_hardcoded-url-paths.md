# BUG-299: ハードコードURLパス — config/paths.ts未使用

## 概要

`features/estimates/` と `features/auth/` でURLパスがハードコードされており、`config/paths.ts` を経由していない。規約: ハードコードURLパス文字列禁止、`paths.xxx.getHref()` を使用すること。

## 違反箇所

### estimates ドメイン（5件）

| ファイル | 行 | コード |
|---------|-----|-------|
| `EstimateList.tsx:244` | `navigate("/estimates/new")` | `paths.estimates.new.getHref()` で解決可 |
| `EstimateForm.tsx:266` | `` navigate(`/estimates/${estimateId}`) `` | `paths.estimates.detail.getHref(estimateId)` で解決可 |
| `EstimateForm.tsx:270` | `navigate('/estimates')` | `paths.estimates.getHref()` で解決可 |
| `EstimateDetail.tsx:30` | `navigate('/estimates')` | `paths.estimates.getHref()` で解決可 |
| `EstimateDetail.tsx:52` | `navigate('/estimates')` | `paths.estimates.getHref()` で解決可 |
| `EstimateDetail.tsx:62` | `` navigate(`/estimates/${id}/edit`) `` | `paths.estimates.edit.getHref(id)` で解決可 |

### auth ドメイン（3件）

| ファイル | 行 | コード |
|---------|-----|-------|
| `ResetPasswordPage.tsx:56` | `navigate("/login")` | `paths.auth.login.getHref()` で解決可 |
| `ResetPasswordPage.tsx:79` | `<Link to="/forgot-password">` | `paths.auth.forgotPassword.getHref()` — paths.ts に追加が必要 |
| `ResetPasswordPage.tsx:168` | `<Link to="/login">` | `paths.auth.login.getHref()` で解決可 |

**注意**: `/forgot-password` と `/reset-password` は `config/paths.ts` の `auth` に未定義。追加が必要。

## 修正

```typescript
// config/paths.ts の auth セクションに追加
auth: {
  login: { path: "/login", getHref: () => "/login" },
  forgotPassword: { path: "/forgot-password", getHref: () => "/forgot-password" },
  resetPassword: { path: "/reset-password", getHref: () => "/reset-password" },
},
```

## ステータス

- [x] ドキュメント作成
- [x] 実装完了
