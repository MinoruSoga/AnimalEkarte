# FE-123: 会計精算画面 - 課税区分・税率・税額 確認編集 UI

**Status**: Closed
**Priority**: High
**Affects**: features/accounting
**Date Created**: 2026-03-25
**Related**: TASK-029, BE-061（先行必須）

## Summary

会計精算画面の明細行（billing_items）に課税区分・税率・税額を表示し、インライン編集できる UI を実装する。
明細の課税区分/税率変更時に税額を即時更新し、Billing 合計も再計算する。

## 現状のコード

```typescript
// frontend/src/features/accounting/routes/AccountingDetail.tsx（現状）
// billing_items を一覧表示している
// 現在: name, unit_price, quantity, total 程度の列が存在
// tax_type, tax_rate, tax_amount の列は未実装

// frontend/src/features/accounting/api/types.ts（現状）
// BillingItem に tax_rate は存在するが tax_type は未実装

// frontend/src/features/accounting/api/transforms.ts（現状）
// transformBillingItem(): tax_type, tax_amount なし
```

## 必要な変更

### 1. 型定義の更新（api/types.ts）

```typescript
// frontend/src/features/accounting/api/types.ts

// BE-061 + make codegen 後に BillingItem に tax_type が追加される
import type { BillingItem, TaxType } from "@/types/generated/models";

// transforms で使う変換後の型
export interface AccountingItem {
  id: string;
  billingId: string;
  name: string;
  unitPrice: number;
  quantity: number;
  taxType: TaxType;    // 新規追加
  taxRate: number;
  taxAmount: number;   // 新規追加（BE が計算して返す）
  subtotal: number;    // 新規追加（単価×数量）
  // ...
}

// Update リクエスト型
export interface UpdateBillingItemRequest {
  unit_price?: number;
  quantity?: number;
  tax_type?: TaxType;  // 新規追加
  tax_rate?: number;   // 新規追加
  is_insurance_applicable?: boolean;
}
```

### 2. transforms.ts の更新

```typescript
// frontend/src/features/accounting/api/transforms.ts

export function transformBillingItem(item: BillingItem): AccountingItem {
  return {
    id: String(item.id),
    billingId: String(item.billing_id),
    name: item.name,
    unitPrice: item.unit_price,
    quantity: item.quantity,
    taxType: item.tax_type,     // 新規追加
    taxRate: item.tax_rate,
    taxAmount: item.tax_amount, // 新規追加（BE から受け取る計算済み値）
    subtotal: item.subtotal,    // 新規追加
    // ...
  };
}
```

### 3. 会計明細テーブルへの課税区分列追加

```typescript
// frontend/src/features/accounting/routes/AccountingDetail.tsx
// または billing item を表示するコンポーネント

// 明細行の表示列（Figmaデザインなし - 既存の明細テーブルスタイルに合わせる）
// | 品目名 | 単価 | 数量 | 課税区分 | 税率 | 税額 | 小計 |
// |--------|------|------|---------|------|------|------|
// | 診察料  | 5,000| 1.0 | 外税    | 10% | 500  | 5,500|
// | 薬剤費  | 3,000| 1.0 | 非課税   | 10% | 0    | 3,000|

// 課税区分セル: TaxTypeSelector コンポーネント（FE-122 で作成）を再利用
// 税率セル: TaxRateSelector コンポーネント（FE-122 で作成）を再利用
// 税額セル: 表示のみ（BE が計算した値を表示）
```

### 4. インライン編集の実装

```typescript
// 課税区分または税率を変更した場合、PATCH /v1/billing-items/:id を呼ぶ
// → BE-061 が Billing の合計を自動再計算する
// → React Query の invalidateQueries で画面を再取得

// useTransition でサブミット管理
const [isPending, startTransition] = useTransition();

const handleTaxTypeChange = useCallback((itemId: string, taxType: TaxType) => {
  startTransition(async () => {
    await updateBillingItem(itemId, { tax_type: taxType });
    // invalidateQueries(['billing', billingId])
  });
}, []);
```

### 5. API hook の更新（api/update-billing-item.ts）

```typescript
// frontend/src/features/accounting/api/update-billing-item.ts（新規または更新）
// PATCH /v1/billing-items/:id を呼ぶ mutation hook

export async function updateBillingItem(
  itemId: string,
  request: UpdateBillingItemRequest
): Promise<BillingItem> {
  const { data } = await axios.patch(`/v1/billing-items/${itemId}`, request);
  return data;
}

export function useUpdateBillingItem() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ itemId, request }: { itemId: string; request: UpdateBillingItemRequest }) =>
      updateBillingItem(itemId, request),
    onSuccess: (_, { itemId }) => {
      // 親 billing も invalidate（合計が再計算されている）
      queryClient.invalidateQueries({ queryKey: ['billing'] });
    },
  });
}
```

## UI 操作フロー

1. ユーザーが会計精算画面を開く
2. 明細行に「課税区分」「税率」「税額」列が表示される
3. 課税区分のセレクト（外税/内税/非課税）を変更
4. PATCH /v1/billing-items/:id が呼ばれる
5. 画面が更新され、税額・合計が再計算された値で表示される

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useTransition` で pending 管理（`useState(false)` + `setIsPending` 禁止）
- [ ] 型は `models.ts` から導出（`TaxType`）
- [ ] `TaxTypeSelector`, `TaxRateSelector` は `@/components/shared/` から直接 import

## 依存関係

- BE-061 が先に完了している必要がある
- FE-122 で作成した `TaxTypeSelector`, `TaxRateSelector` コンポーネントが必要
- `make codegen` で `models.ts` に `BillingItem.tax_type`, `BillingItem.tax_amount` が含まれていること

## 完了条件

- [ ] 会計精算画面の明細テーブルに「課税区分」「税率」「税額」列が表示される
- [ ] 課税区分・税率をセレクトで変更でき、PATCH が呼ばれる
- [ ] 変更後に税額（行）・税合計（集計行）が正しく更新される
- [ ] 外税 10%: 税額 = 単価 × 数量 × 0.10 で正しく表示される
- [ ] 内税 10%: 税額 = 単価 × 数量 × 0.10 ÷ 1.10 で正しく表示される
- [ ] 非課税: 税額 = 0 で表示される
- [ ] `pnpm build` 型エラーなし
- [ ] `pnpm lint` エラーなし
