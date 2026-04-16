# BUG-303: auth/estimates — ハードコード URL パス文字列3箇所

## 概要

`config/paths.ts` 規約違反。auth と estimates の3ファイルで `navigate('/...')` / `<Link to="/...">` にリテラル文字列が使われていた。

## 影響ファイル

| ファイル | 違反箇所 |
|---------|---------|
| `frontend/src/features/estimates/hooks/use-estimate-form.ts` | line 126, 128 |
| `frontend/src/features/auth/routes/ForgotPasswordPage.tsx` | line 68, 102 |
| `frontend/src/features/auth/components/LoginForm.tsx` | line 148, 229 |

## 違反箇所と修正

### use-estimate-form.ts
```ts
// Before
navigate(`/estimates/${estimate.id}`);
navigate('/estimates');

// After
navigate(paths.estimates.detail.getHref(estimate.id));
navigate(paths.estimates.getHref());
```

### ForgotPasswordPage.tsx
```tsx
// Before (2箇所)
<Link to="/login">

// After
<Link to={paths.auth.login.getHref()}>
```

### LoginForm.tsx
```tsx
// Before
<Navigate to="/" replace />
<Link to="/forgot-password">

// After
<Navigate to={paths.home.getHref()} replace />
<Link to={paths.auth.forgotPassword.getHref()}>
```

## 適用ルール

- `config/paths.ts` でURL管理: ハードコードされた URL パス文字列は禁止

## ステータス

✅ 修正済み
