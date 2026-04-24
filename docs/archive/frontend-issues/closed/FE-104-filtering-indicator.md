# FE-104: FilteringIndicator コンポーネント共有化

**Status**: Closed
**Priority**: Low
**Affects**: owners, vaccinations, inventory, trimming, examinations, medical-records, accounting
**Date Created**: 2026-03-24
**Related**: TASK-024

## Summary

`useDeferredValue` による検索フィルタリング中の opacity トランジションが 8 箇所で重複実装されており、スタイル指定方法（className vs inline style）も不統一。`FilteringIndicator` コンポーネントに共通化する。

## 現状のコード

```typescript
// frontend/src/features/owners/routes/OwnersList.tsx:256, 485
const isFiltering = searchTerm !== deferredSearchTerm;  // 全ページで同一
...
<div className={
  isFiltering
    ? "opacity-60 transition-opacity duration-150"
    : "transition-opacity duration-150"
}>
```

```typescript
// frontend/src/features/vaccinations/routes/VaccinationList.tsx:59, 213（同じ）
const isFiltering = searchTerm !== deferredSearchTerm;
...
<div className={isFiltering ? "opacity-60 transition-opacity duration-150" : "transition-opacity duration-150"}>
```

```typescript
// frontend/src/features/accounting/routes/Accounting.tsx:98, 329（inline style バリアント）
const isFiltering = searchTerm !== deferredSearch;
...
<div style={{ opacity: isFiltering ? 0.7 : 1, transition: "opacity 150ms" }}>
// ↑ className でなく style プロパティを使用（不統一）
```

同様のパターン: `InventoryList.tsx`, `TrimmingList.tsx`, `Examinations.tsx`

## 必要な変更

### 1. FilteringIndicator コンポーネント作成

```typescript
// frontend/src/components/shared/FilteringIndicator/FilteringIndicator.tsx（新規作成）

interface FilteringIndicatorProps {
  isFiltering: boolean;
  children: React.ReactNode;
  className?: string;
}

export function FilteringIndicator({
  isFiltering,
  children,
  className,
}: FilteringIndicatorProps) {
  return (
    <div
      className={`transition-opacity duration-150 ${isFiltering ? "opacity-60" : ""} ${className ?? ""}`}
    >
      {children}
    </div>
  );
}
```

### 2. 各リストページの置き換え（例: OwnersList.tsx）

```typescript
// Before:
const isFiltering = searchTerm !== deferredSearchTerm;
...
<div className={isFiltering ? "opacity-60 transition-opacity duration-150" : "transition-opacity duration-150"}>
  {/* テーブル内容 */}
</div>

// After:
import { FilteringIndicator } from "@/components/shared/FilteringIndicator/FilteringIndicator";

<FilteringIndicator isFiltering={isFiltering}>
  {/* テーブル内容 */}
</FilteringIndicator>
```

Accounting.tsx の inline style バリアントも className ベースに統一。

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] barrel index 経由 import なし（`FilteringIndicator/FilteringIndicator` を直接 import）
- [ ] `FC` / `forwardRef` なし
- [ ] `isFiltering` の `const` 宣言は各ページで残す（`useDeferredValue` との比較ロジックはページ内で完結）

## 依存関係

- Backend 変更なし。他の FE イシューとも独立。FE-101（useSortableData）と同時進行可。

## 完了条件

- [ ] `frontend/src/components/shared/FilteringIndicator/FilteringIndicator.tsx` が作成されている
- [ ] 対象 7 ファイルで `FilteringIndicator` コンポーネントを使用している
- [ ] Accounting.tsx の inline style が削除されている
- [ ] 各リストページの検索中 opacity トランジションが変化なし
- [ ] `pnpm lint` エラーなし
- [ ] `pnpm build` エラーなし

## クローズ情報

- **Closed At**: 2026-03-24
- **変更ファイル**:
  - `frontend/src/components/shared/FilteringIndicator/FilteringIndicator.tsx` — 新規作成
  - `frontend/src/features/owners/routes/OwnersList.tsx` — FilteringIndicator に置き換え
  - `frontend/src/features/vaccinations/routes/VaccinationList.tsx` — 同上
  - `frontend/src/features/inventory/routes/InventoryList.tsx` — 同上
  - `frontend/src/features/trimming/routes/TrimmingList.tsx` — 同上
  - `frontend/src/features/examinations/routes/Examinations.tsx` — 同上
  - `frontend/src/features/medical-records/routes/MedicalRecords.tsx` — 同上
  - `frontend/src/features/accounting/routes/Accounting.tsx` — inline style を FilteringIndicator に置き換え
