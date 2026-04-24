# FE-047: accounting ドメイン — models.ts 型移行（Request型 + types/index.ts 導出化）

**親タスク**: [TASK-012](../../docs/tasks/open/TASK-012-models-ts-type-migration-by-domain.md)
**Status**: Open
**Priority**: Medium
**Date Created**: 2026-03-18

## Summary

api/types.ts は models.ts（Billing, BillingItem, Payment）を使用済み。Request 型が手書き + types/index.ts に5個の手書き型が残存。

## 現状

### api/types.ts（models.ts import 済み ✅、Request型手書き ❌）
```typescript
import type { Billing, BillingItem, Payment } from "@/types/generated/models";
// CreateAccountingRequest — 手書き interface
// UpdateAccountingRequest — 手書き interface
```

### features/accounting/types/index.ts（5個手書き）
| 手書き型 | models.ts 対応型 |
|---------|----------------|
| `AccountingStatus` | `BillingStatus`（models.ts に存在） |
| `PaymentMethod` | `PaymentMethod`（models.ts に存在） |
| `ItemCategory` | `ItemCategory`（models.ts に存在） |
| `AccountingItem` | `BillingItem`（models.ts に存在）から導出 |
| `PaymentInfo` | `Payment`（models.ts に存在）から導出 |
| `Accounting` | `Billing`（models.ts に存在）から導出 |

## 必要な変更

1. `api/types.ts`: Request 型を models.ts から Omit/Partial で導出
2. `types/index.ts`: AccountingStatus → `BillingStatus`, PaymentMethod → `PaymentMethod` を models.ts から import
3. AccountingItem/PaymentInfo/Accounting を models.ts の対応型（BillingItem/Payment/Billing）ベースで導出

## 完了条件

- [ ] api/types.ts の Request 型が models.ts から導出されている
- [ ] types/index.ts の enum/型が models.ts から import されている
- [ ] `pnpm build` 成功・型エラーなし
