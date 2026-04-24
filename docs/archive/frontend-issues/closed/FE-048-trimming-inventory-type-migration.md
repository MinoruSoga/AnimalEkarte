# FE-048: trimming・inventory ドメイン — models.ts 型移行（Request型導出化）

**親タスク**: [TASK-012](../../docs/tasks/open/TASK-012-models-ts-type-migration-by-domain.md)
**Status**: Open
**Priority**: Medium
**Date Created**: 2026-03-18

## Summary

両ドメインともResponse型は models.ts を使用済み。Request型が `src/types/trimming.ts` 等に手書きされているため、models.ts から導出する。

## 現状

### trimming
```typescript
// src/types/trimming.ts
// CreateTrimmingRequest, UpdateTrimmingRequest 手書き
// TrimmingRecord as BackendTrimmingRecord は models.ts import 済み
```

### inventory
```typescript
// features/inventory/api/types.ts
// CreateInventoryItemRequest, UpdateInventoryItemRequest 手書き
// InventoryItem は models.ts import 済み
```

## src/types/index.ts から使用している手書き型

### trimming
- `TrimmingRecord`（src/types/index.ts 手書き）— transforms, routes, api で使用
- `TrimmingCourse`, `TrimmingOption`, `TargetSize`（src/types/index.ts 手書き）— models.ts に対応型あり

### inventory
- `InventoryItem`（src/types/index.ts 手書き）— routes, hooks で使用。models.ts に `InventoryItem` 存在

## 必要な変更

### trimming
1. `src/types/trimming.ts` の Request 型を `Omit<TrimmingRecord, ...>` で導出
2. `TrimmingFormData`（UI固有型）は `features/trimming/types/` に移動
3. transforms/routes が src/types/index.ts の手書き `TrimmingRecord` を使用 → models.ts 由来に統一

### inventory
1. `features/inventory/api/types.ts` の Request 型を `Omit<InventoryItem, ...>` で導出
2. routes/hooks が src/types/index.ts の手書き `InventoryItem` を使用 → models.ts 由来に統一

## 完了条件

- [ ] trimming Request 型が models.ts から導出されている
- [ ] inventory Request 型が models.ts から導出されている
- [ ] src/types/index.ts の手書き TrimmingRecord/InventoryItem への依存が解消
- [ ] `pnpm build` 成功・型エラーなし
