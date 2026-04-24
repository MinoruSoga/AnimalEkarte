# FE-105: formatCurrency インライン実装を共有 util へ統一

**Status**: Closed
**Priority**: Low
**Affects**: accounting/routes/, estimates/routes/
**Date Created**: 2026-03-24
**Related**: TASK-024

## Summary

`frontend/src/utils/format/number.ts` に `formatCurrency()` が共有実装されているにもかかわらず、`Accounting.tsx` と `EstimateList.tsx` が `new Intl.NumberFormat("ja-JP", { style: "currency", currency: "JPY" })` をインラインで重複実装している。

## 現状のコード

```typescript
// frontend/src/utils/format/number.ts（共有実装・既存）
export function formatCurrency(amount: number): string {
  return new Intl.NumberFormat("ja-JP", { style: "currency", currency: "JPY" }).format(amount);
}
```

```typescript
// frontend/src/features/accounting/routes/Accounting.tsx:73-78（インライン重複）
new Intl.NumberFormat("ja-JP", { style: "currency", currency: "JPY" }).format(amount)
```

```typescript
// frontend/src/features/estimates/routes/EstimateList.tsx:53-54（インライン重複）
new Intl.NumberFormat("ja-JP", { style: "currency", currency: "JPY" }).format(amount)
```

## 必要な変更

### Accounting.tsx

```typescript
// Before: インライン実装削除
// After:
import { formatCurrency } from "@/utils/format/number";
// インライン Intl.NumberFormat を formatCurrency(amount) に置き換え
```

### EstimateList.tsx

```typescript
// Before: インライン実装削除
// After:
import { formatCurrency } from "@/utils/format/number";
// インライン Intl.NumberFormat を formatCurrency(amount) に置き換え
```

## 完了条件

- [ ] `Accounting.tsx` のインライン `Intl.NumberFormat` が `formatCurrency()` に置き換えられている
- [ ] `EstimateList.tsx` のインライン `Intl.NumberFormat` が `formatCurrency()` に置き換えられている
- [ ] 金額表示フォーマットが変化なし
- [ ] `pnpm lint` エラーなし
- [ ] `pnpm build` エラーなし
