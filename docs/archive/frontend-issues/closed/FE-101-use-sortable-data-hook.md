# FE-101: useSortableData カスタムフック共有化

**Status**: Closed
**Priority**: High
**Affects**: owners, vaccinations, inventory, trimming, examinations, medical-records, accounting（全リストページ）
**Date Created**: 2026-03-24
**Related**: TASK-024

## Summary

全リストページ（7〜8画面）で `toggleSort` / `directionFor` / `sortedData` の **完全に同一なソートロジック** が独立実装されている。`useSortableData<T>()` カスタムフックとして共有化し、重複を排除する。

## 現状のコード

```typescript
// 以下のコードが 17+ 箇所で完全に重複している

// frontend/src/features/owners/routes/OwnersList.tsx:205-248
// frontend/src/features/vaccinations/routes/VaccinationList.tsx:79-115
// frontend/src/features/inventory/routes/InventoryList.tsx:115-156
// frontend/src/features/trimming/routes/TrimmingList.tsx:148-184
// frontend/src/features/examinations/routes/Examinations.tsx:93-129
// frontend/src/features/medical-records/routes/MedicalRecords.tsx:62-98
// frontend/src/features/accounting/routes/Accounting.tsx:141-186

const toggleSort = useCallback((key: SortKey) => {
  setActiveSorts((prev) => {
    const existing = prev.find((s) => s.key === key);
    if (!existing) {
      return [{ key, direction: "asc" as const }];
    }
    if (existing.direction === "asc") {
      return prev.map((s) =>
        s.key === key ? { ...s, direction: "desc" as const } : s
      );
    }
    return prev.filter((s) => s.key !== key);
  });
}, []);

const directionFor = useCallback(
  (key: SortKey): "ascending" | "descending" | "none" => {
    const sort = activeSorts.find((s) => s.key === key);
    if (!sort) return "none";
    return sort.direction === "asc" ? "ascending" : "descending";
  },
  [activeSorts],
);

const sortedData = useMemo(() => {
  if (activeSorts.length === 0) return [...filteredData];
  const sorted = [...filteredData];
  sorted.sort((a, b) => {
    for (const sort of activeSorts) {
      const aVal = String(a[sort.key as keyof typeof a] ?? "");
      const bVal = String(b[sort.key as keyof typeof b] ?? "");
      const cmp = aVal.localeCompare(bVal, "ja");
      if (cmp !== 0) return sort.direction === "asc" ? cmp : -cmp;
    }
    return 0;
  });
  return sorted;
}, [filteredData, activeSorts]);
```

## 必要な変更

### 1. useSortableData フック作成

```typescript
// frontend/src/hooks/useSortableData.ts（新規作成）

import { useState, useCallback, useMemo } from "react";

interface ActiveSort {
  key: string;
  direction: "asc" | "desc";
}

interface UseSortableDataOptions<T> {
  numericKeys?: (keyof T)[];  // 数値ソートするキー（デフォルト: 文字列ソート）
}

export function useSortableData<T extends Record<string, unknown>>(
  data: T[],
  options?: UseSortableDataOptions<T>
) {
  const [activeSorts, setActiveSorts] = useState<ActiveSort[]>([]);

  const toggleSort = useCallback((key: keyof T & string) => {
    setActiveSorts((prev) => {
      const existing = prev.find((s) => s.key === key);
      if (!existing) return [{ key, direction: "asc" as const }];
      if (existing.direction === "asc") {
        return prev.map((s) =>
          s.key === key ? { ...s, direction: "desc" as const } : s
        );
      }
      return prev.filter((s) => s.key !== key);
    });
  }, []);

  const directionFor = useCallback(
    (key: keyof T & string): "ascending" | "descending" | "none" => {
      const sort = activeSorts.find((s) => s.key === key);
      if (!sort) return "none";
      return sort.direction === "asc" ? "ascending" : "descending";
    },
    [activeSorts],
  );

  const sortedData = useMemo(() => {
    if (activeSorts.length === 0) return [...data];
    const sorted = [...data];
    sorted.sort((a, b) => {
      for (const sort of activeSorts) {
        const key = sort.key as keyof T;
        let cmp = 0;
        if (options?.numericKeys?.includes(key)) {
          cmp = Number(a[key] ?? 0) - Number(b[key] ?? 0);
        } else {
          const aVal = String(a[key] ?? "");
          const bVal = String(b[key] ?? "");
          cmp = aVal.localeCompare(bVal, "ja");
        }
        if (cmp !== 0) return sort.direction === "asc" ? cmp : -cmp;
      }
      return 0;
    });
    return sorted;
  }, [data, activeSorts, options?.numericKeys]);

  return { activeSorts, setActiveSorts, toggleSort, directionFor, sortedData };
}
```

### 2. 各リストページの置き換え（例: OwnersList.tsx）

```typescript
// frontend/src/features/owners/routes/OwnersList.tsx
// Before: activeSorts useState + toggleSort + directionFor + sortedData useMemo（40行）を削除
// After:

import { useSortableData } from "@/hooks/useSortableData";

// 呼び出し（1行に集約）
const { activeSorts, setActiveSorts, toggleSort, directionFor, sortedData } =
  useSortableData(filteredOwners);
```

同じ置き換えを以下 6 ファイルで実施:
- `VaccinationList.tsx`
- `InventoryList.tsx`
- `TrimmingList.tsx`
- `Examinations.tsx`
- `MedicalRecords.tsx`
- `Accounting.tsx`

## プロジェクトルール遵守チェック

- [ ] `any` 型なし（`Record<string, unknown>` + keyof で型安全）
- [ ] barrel index 経由 import なし（`hooks/useSortableData` を直接 import）
- [ ] `useCallback` deps に `options?.numericKeys` を含める（プリミティブ相当ではないが配列のため注意）
- [ ] `FC` / `forwardRef` なし

## 依存関係

- Backend 変更なし。FE-098/099/100 とも独立。単独で実施可能。

## 完了条件

- [ ] `frontend/src/hooks/useSortableData.ts` が作成されている
- [ ] 対象 7 ファイル（OwnersList/VaccinationList/InventoryList/TrimmingList/Examinations/MedicalRecords/Accounting）で `useSortableData()` を使用している
- [ ] 各リストページのソート（昇順↑ / 降順↓ / 解除）が変更前と同一動作
- [ ] `pnpm lint` エラーなし
- [ ] `pnpm build` エラーなし

## クローズ情報

- **Closed At**: 2026-03-24
- **変更ファイル**:
  - `frontend/src/hooks/use-sortable-data.ts`（新規作成）
  - `frontend/src/features/owners/routes/OwnersList.tsx`
  - `frontend/src/features/vaccinations/routes/VaccinationList.tsx`
  - `frontend/src/features/inventory/routes/InventoryList.tsx`
  - `frontend/src/features/trimming/routes/TrimmingList.tsx`
  - `frontend/src/features/examinations/routes/Examinations.tsx`
  - `frontend/src/features/medical-records/routes/MedicalRecords.tsx`
  - `frontend/src/features/accounting/routes/Accounting.tsx`
