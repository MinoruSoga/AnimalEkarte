# BUG-295: app/router.tsx — Feature内部への深掘りimport

## 概要

`frontend/src/app/router.tsx` でFeatureの `index.ts` (barrel) を経由せず、Feature内部ファイルを直接インポートしている箇所が4件存在する。Feature Indexingルール違反。

## 問題箇所

```typescript
// L40: ❌ 深掘りimport
const { ForgotPasswordPage } = await import("@/features/auth/routes/ForgotPasswordPage");

// L47: ❌ 深掘りimport
const { ResetPasswordPage } = await import("@/features/auth/routes/ResetPasswordPage");

// L401: ❌ 深掘りimport
const { AccountingList } = await import("@/features/accounting/routes/AccountingList");

// L412: ❌ 深掘りimport
const { AccountingPetSelection } = await import("@/features/accounting/routes/AccountingPetSelection");
```

すべてのコンポーネントはすでに各Featureの `index.ts` からexportされている:
- `auth/index.ts:5`: `export { ForgotPasswordPage } from "./routes/ForgotPasswordPage"`
- `auth/index.ts:6`: `export { ResetPasswordPage } from "./routes/ResetPasswordPage"`
- `accounting/index.ts:1-3`: `AccountingList`, `AccountingPetSelection` ともにexport済み

## 修正

```typescript
// L40: ✅ barrel経由
const { ForgotPasswordPage } = await import("@/features/auth");

// L47: ✅ barrel経由
const { ResetPasswordPage } = await import("@/features/auth");

// L401: ✅ barrel経由
const { AccountingList } = await import("@/features/accounting");

// L412: ✅ barrel経由
const { AccountingPetSelection } = await import("@/features/accounting");
```

## ステータス

- [x] ドキュメント作成
- [x] 実装完了
