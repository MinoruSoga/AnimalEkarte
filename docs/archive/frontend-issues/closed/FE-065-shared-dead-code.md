# FE-065: 共通基盤 — 未使用 hooks・deprecated 再エクスポート除去

**Status**: Open
**Priority**: Medium
**Affects**: `hooks/`, `main.tsx`
**Date Created**: 2026-03-18
**Related**: TASK-014

## Summary

共通 hooks ディレクトリから未使用の hook ファイル・deprecated 再エクスポートを削除し、main.tsx のコメントアウトされたテストコードを除去する。

## 現状のコード

### 1. 未使用 hook — useTableSort

```typescript
// frontend/src/hooks/useTableSort.ts — 全ファイルが未使用
// エクスポートされているが、frontend/src/ 内のどのファイルからもインポートされていない
```

### 2. 未使用 hook — useDevice, useIsMobile

```typescript
// frontend/src/hooks/use-mobile.ts
// :22-56  useDevice() — 一度もインポートされていない
// :61-77  useIsMobile() — 一度もインポートされていない
```

**注**: このファイルが `components/ui/sidebar.tsx` から参照されているか確認が必要。参照されている場合は、使用されている関数のみ残す。

### 3. deprecated 再エクスポート — use-master-items

```typescript
// frontend/src/hooks/use-master-items.ts — 全ファイルが deprecated 再エクスポート
/**
 * @deprecated Import directly from @/features/master/hooks/use-master-items
 * This re-export exists for backward compatibility.
 */
export { useMasterItems } from "@/features/master/hooks/use-master-items";
// ❌ @/hooks/use-master-items からインポートしているファイルは存在しない
```

### 4. deprecated 再エクスポート — use-service-type-color-map

```typescript
// frontend/src/hooks/use-service-type-color-map.ts — 全ファイルが再エクスポート
export { useServiceTypeColorMap } from "@/features/master/hooks/useServiceTypeColorMap";
// ❌ @/hooks/use-service-type-color-map からインポートしているファイルは存在しない
```

### 5. コメントアウトされたテストコード

```typescript
// frontend/src/main.tsx:29-32
// Test for workflow
// Retry test
// Final test
// Final workflow test
```

## 必要な変更

### 1. ファイル削除

```bash
rm frontend/src/hooks/useTableSort.ts
rm frontend/src/hooks/use-master-items.ts
rm frontend/src/hooks/use-service-type-color-map.ts
```

### 2. use-mobile.ts の処理

削除前に確認:
```bash
grep -rn "use-mobile" frontend/src/
grep -rn "useDevice\|useIsMobile" frontend/src/
```

- **どこからも参照されていない場合**: ファイルごと削除
- **sidebar.tsx から参照されている場合**: 使用されている export のみ残す

### 3. main.tsx のコメント除去

```typescript
// Before: lines 29-32
// Test for workflow
// Retry test
// Final test
// Final workflow test

// After: これら4行を削除
```

## 完了条件

- [ ] `hooks/useTableSort.ts` が削除されている
- [ ] `hooks/use-master-items.ts` が削除されている
- [ ] `hooks/use-service-type-color-map.ts` が削除されている
- [ ] `hooks/use-mobile.ts` の未使用関数が削除されている（または参照がある場合は残す判断を記録）
- [ ] `main.tsx` のテストコメント4行が削除されている
- [ ] `pnpm build` パス
- [ ] `pnpm lint` パス
