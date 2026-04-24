# FE-100: LineItem 型・金額計算ロジック共有化

**Status**: Closed
**Priority**: Low
**Affects**: medical-records/components/, estimates/components/
**Date Created**: 2026-03-24
**Related**: TASK-024

## Summary

`TreatmentTable.tsx`（medical-records, 400行超）と `EstimateLineItems.tsx`（estimates, 118行）は、明細行アイテムの金額計算（単価×数量、割引額、合計）ロジックを独立して実装している。UI 構造は異なるため統合せず、型定義と計算ロジックのみを共有ユーティリティとして切り出す。

## 現状のコード

```typescript
// frontend/src/features/estimates/components/EstimateLineItems/EstimateLineItems.tsx:70-80
// 金額フォーマット・計算がインライン実装
{item.discountAmount > 0 ? `-${formatCurrency(item.discountAmount)}` : '-'}
// 合計行: 各アイテムの amount を直接参照
```

```typescript
// frontend/src/features/medical-records/components/TreatmentTable.tsx（400行超）
// 同様の金額計算ロジックが独立実装
// status 管理（"未完了", "完了", "-"）、割引計算、小計計算が各所に散在
```

## 必要な変更

### 1. LineItem 共有型定義

```typescript
// frontend/src/utils/line-item-helpers.ts（新規作成）

/**
 * 明細行アイテムの共通型
 * TreatmentTable / EstimateLineItems 両方に対応できる最小構造
 */
export interface LineItemBase {
  id: string | number;
  unitPrice: number;      // 単価
  quantity: number;       // 数量
  discountAmount?: number; // 割引額（任意）
  discountRate?: number;   // 割引率（任意）
}

/**
 * 明細行アイテムの金額計算
 */
export function calcLineItemAmount(item: LineItemBase): {
  subtotal: number;    // 単価 × 数量
  discount: number;    // 割引額
  total: number;       // 小計 - 割引
} {
  const subtotal = item.unitPrice * item.quantity;
  const discount = item.discountAmount
    ?? (item.discountRate ? Math.floor(subtotal * item.discountRate) : 0);
  return {
    subtotal,
    discount,
    total: subtotal - discount,
  };
}

/**
 * 明細行リストの合計金額計算
 */
export function calcLineItemsTotal(items: LineItemBase[]): {
  subtotalSum: number;
  discountSum: number;
  totalSum: number;
} {
  return items.reduce(
    (acc, item) => {
      const { subtotal, discount, total } = calcLineItemAmount(item);
      return {
        subtotalSum: acc.subtotalSum + subtotal,
        discountSum: acc.discountSum + discount,
        totalSum: acc.totalSum + total,
      };
    },
    { subtotalSum: 0, discountSum: 0, totalSum: 0 }
  );
}
```

### 2. EstimateLineItems.tsx の変更

```typescript
// frontend/src/features/estimates/components/EstimateLineItems/EstimateLineItems.tsx
// Before: インライン金額計算
// After: calcLineItemAmount / calcLineItemsTotal を使用

import { calcLineItemAmount, calcLineItemsTotal } from "@/utils/line-item-helpers";

// 合計行の計算
const totals = useMemo(() => calcLineItemsTotal(lineItems), [lineItems]);
```

### 3. TreatmentTable.tsx の変更

```typescript
// frontend/src/features/medical-records/components/TreatmentTable.tsx
// Before: 独立した金額計算ロジック
// After: calcLineItemAmount / calcLineItemsTotal を使用

import { calcLineItemAmount, calcLineItemsTotal } from "@/utils/line-item-helpers";
```

## 注意事項

- **UI は変更しない**: TreatmentTable（グリッドレイアウト・inline editing・status 管理）と EstimateLineItems（shadcn Table）の UI 構造はそれぞれ維持する
- 計算ロジックの切り出しのみ。表示フォーマット（`formatCurrency`）はそのまま各コンポーネントで使用
- TreatmentTable は 400行超の複雑なコンポーネントのため、計算ロジック部分のみを慎重に抽出すること

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] barrel index 経由 import なし（`utils/line-item-helpers` を直接 import）
- [ ] `useMemo` で計算結果をキャッシュ（APIリスト由来のJSX計算）
- [ ] `FC` / `forwardRef` なし

## 依存関係

- Backend 変更なし。FE-098/099 とも独立。
- ただし TreatmentTable は複雑なため、FE-098/099 完了後に着手推奨。

## 完了条件

- [ ] `frontend/src/utils/line-item-helpers.ts` が作成されている
- [ ] `EstimateLineItems.tsx` の金額計算が `calcLineItemsTotal` を使用している
- [ ] `TreatmentTable.tsx` の金額計算が `calcLineItemAmount` / `calcLineItemsTotal` を使用している
- [ ] medical-records カルテ画面の処置テーブルで金額表示が変化していない
- [ ] estimates 見積画面の明細金額・合計が変化していない
- [ ] `pnpm lint` エラーなし
- [ ] `pnpm build` エラーなし
