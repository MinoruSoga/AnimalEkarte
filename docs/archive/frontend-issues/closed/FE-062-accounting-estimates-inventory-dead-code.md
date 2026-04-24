# FE-062: 会計・見積・在庫 — 未使用フィルタ API + 空ファイル除去

**Status**: Open
**Priority**: Medium
**Affects**: `features/accounting/api/`, `features/inventory/types/`
**Date Created**: 2026-03-18
**Related**: TASK-014

## Summary

accounting から未使用フィルタ API 関数8件を削除し、inventory の空 types ファイルを削除する。
estimates はデッドコードなし。

## 現状のコード

### 1. accounting — 8関数が未使用

```typescript
// frontend/src/features/accounting/api/get-accounting.ts
// 以下8関数が一度もインポートされていない:
// :14-25  getAccounting()（getAccountingDetail と重複）
// :14-25  useGetAccounting()（useGetAccountingDetail と重複）
// :42-57  getAccountingsByPetId()
// :42-57  useGetAccountingsByPetId()
// :60-75  getAccountingsByOwnerId()
// :60-75  useGetAccountingsByOwnerId()
// :78-93  getAccountingsByStatus()
// :78-93  useGetAccountingsByStatus()

// frontend/src/features/accounting/api/index.ts:8-15
// 上記8関数の barrel 再エクスポート（未使用）
```

**注**: `getAccountingDetail()` と `useGetAccountingDetail()` は使用されているため残す。

### 2. inventory — 空 types ファイル

```typescript
// frontend/src/features/inventory/types/index.ts（全4行）
// Inventory feature types
// Feature-specific types are defined here
// InventoryItem is imported from @/types for shared type consistency
```

コメントのみ。型は `@/types/generated/models` から直接インポートされている。

### 3. estimates — デッドコードなし ✅

全 API 関数が使用されていることを確認済み。

## 必要な変更

### 1. accounting

- `get-accounting.ts` から未使用8関数を削除
- `index.ts` から対応する再エクスポート行を削除

### 2. inventory

```bash
rm frontend/src/features/inventory/types/index.ts
# types/ ディレクトリも空になる場合は削除
rmdir frontend/src/features/inventory/types/ 2>/dev/null
```

## 完了条件

- [ ] accounting: 8関数が `get-accounting.ts` から削除されている
- [ ] accounting: `index.ts` の対応再エクスポートが削除されている
- [ ] inventory: `types/index.ts` が削除されている
- [ ] `pnpm build` パス
- [ ] `pnpm lint` パス
