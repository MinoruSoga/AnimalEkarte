# FE-118: 定期健診一覧 - NotionFilter移行＋日付範囲フィルタ追加

**Status**: Open
**Priority**: Medium
**Affects**: 定期健診 (`features/checkups/routes/CheckupsList.tsx`)
**Date Created**: 2026-03-25
**Related**: TASK-028（バックエンド変更なし・フロントエンドのみ）

## Summary

`CheckupsList.tsx`は古いシンプルなテキスト入力UIを使用している。他の全一覧ページと同様に`NotionFilter`に移行し、日付範囲フィルタを追加する。

**バックエンドAPIはすでに`start_date`/`end_date`をサポート済み**（`useGetCheckups`も`CheckupFilters`型を受け取る）。FEのみの変更。

## 現状のコード

```tsx
// frontend/src/features/checkups/routes/CheckupsList.tsx
export function CheckupsList() {
  const [searchTerm, setSearchTerm] = useState("");
  const deferredSearch = useDeferredValue(searchTerm);

  const { data: checkups = [], isLoading, error } = useGetCheckups();
  // ↑ filters を渡していない（全件取得）

  const filtered = useMemo(() => {
    if (!deferredSearch) return checkups;
    // ↑ クライアントサイドフィルタのみ
    ...
  }, [checkups, deferredSearch]);

  return (
    <PageLayout title="定期健診">
      <div className="space-y-4">
        {/* 古いシンプルな検索バー */}
        <div className="flex items-center gap-2">
          <input type="text" placeholder="ペット名・飼主名・種別で検索..." ... />
          <span ...>{filtered.length}件</span>
        </div>
        {/* ← NotionFilter を使っていない */}
        {/* ← SortableHeader を使っていない */}
        {/* ← ページネーションなし */}
        ...
```

## 必要な変更

### 1. 型・定数定義の追加

```tsx
// frontend/src/features/checkups/routes/CheckupsList.tsx

// rendering-hoist-jsx: 静的定数をモジュールスコープに
const FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "date",
    label: "日付",
    type: "date-range",
    icon: Calendar,
  },
];

const CHECKUPS_SORT_PROPERTIES: SortProperty[] = [
  { key: "date", label: "実施日" },
  { key: "ownerName", label: "飼主名" },
  { key: "petName", label: "ペット名" },
  { key: "checkupTypeName", label: "健診種別" },
  { key: "nextDate", label: "次回予定" },
];
```

### 2. コンポーネント全体の書き換え

```tsx
// frontend/src/features/checkups/routes/CheckupsList.tsx

export function CheckupsList() {
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const deferredSearch = useDeferredValue(searchTerm);
  const isFiltering = searchTerm !== deferredSearch;

  // activeFilters から日付フィルタを抽出（vaccinations と同パターン）
  const filters = useMemo(() => {
    const dateFilter = activeFilters.find((f) => f.key === "date")?.value as
      | { from?: string; to?: string }
      | undefined;
    return {
      startDate: dateFilter?.from,
      endDate: dateFilter?.to,
    };
  }, [activeFilters]);

  const { data: checkups = [], isLoading, error } = useGetCheckups(filters);

  // テキスト検索はクライアントサイドで行う（vaccinations と同パターン）
  const filteredRecords = useMemo(() => {
    if (!deferredSearch) return checkups;
    const q = deferredSearch.toLowerCase();
    return checkups.filter(
      (c) =>
        c.petName.toLowerCase().includes(q) ||
        c.ownerName.toLowerCase().includes(q) ||
        c.checkupTypeName.toLowerCase().includes(q) ||
        c.result.toLowerCase().includes(q),
    );
  }, [checkups, deferredSearch]);

  const { activeSorts, setActiveSorts, toggleSort, directionFor, sortedData } =
    useSortableData(filteredRecords);

  const pagination = usePagination(sortedData, {
    pageSize: 20,
    resetKey: deferredSearch,
  });

  // ... return に NotionFilter + DataTable + Pagination を使う
}
```

### 3. JSX：旧シンプルUIをNotionFilter+DataTableに置換

旧UIの`<input type="text" ...>`と手書き`<Table>`ブロックを削除し、`VaccinationList.tsx`に準拠した形式に置換：

```tsx
return (
  <PageLayout
    title="定期健診"
    icon={<ClipboardCheck className="size-5 text-[#37352F]" />}
    maxWidth="max-w-full"
  >
    <div className="flex flex-col gap-4">
      <NotionFilter
        properties={FILTER_PROPERTIES}
        activeFilters={activeFilters}
        onFilterChange={setActiveFilters}
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        searchPlaceholder="ペット名・飼主名・種別で検索..."
        count={isLoading ? undefined : filteredRecords.length}
        sortProperties={CHECKUPS_SORT_PROPERTIES}
        activeSorts={activeSorts}
        onSortChange={setActiveSorts}
      />

      <FilteringIndicator isFiltering={isFiltering}>
        <DataTable
          columns={columns}
          data={pagination.paginatedData}
          emptyMessage="定期健診の記録がありません"
          renderRow={(c) => (
            <DataTableRow key={c.id}>
              <TableCell className="font-mono text-base text-[#37352F] py-2">{c.date ? formatDate(c.date) : "-"}</TableCell>
              <TableCell className="text-base text-[#37352F] py-2">{c.ownerName || "-"}</TableCell>
              <TableCell className="text-base text-[#37352F] py-2">{c.petName || "-"}</TableCell>
              <TableCell className="text-base text-[#37352F] py-2">{c.checkupTypeName || "-"}</TableCell>
              <TableCell className="font-mono text-base text-[#37352F] py-2">{c.nextDate ? formatDate(c.nextDate) : "-"}</TableCell>
              <TableCell className="text-base text-[#37352F] py-2 max-w-xs truncate">{c.result || "-"}</TableCell>
              <TableCell className="text-base text-[#37352F] py-2">{c.doctorName || "-"}</TableCell>
            </DataTableRow>
          )}
        />
      </FilteringIndicator>

      {pagination.totalPages > 1 ? (
        <Pagination ... />
      ) : null}
    </div>
  </PageLayout>
);
```

### 4. 追加 import

```tsx
// 追加
import { useCallback, useMemo } from "react";  // useDeferredValue は既存
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { SortableHeader } from "@/components/shared/SortableHeader/SortableHeader";
import { FilteringIndicator } from "@/components/shared/FilteringIndicator/FilteringIndicator";
import { Pagination } from "@/components/shared/Pagination/Pagination";
import { useSortableData } from "@/hooks/use-sortable-data";
import { usePagination } from "@/hooks/use-pagination";
import { Calendar } from "lucide-react";
import type { FilterProperty, ActiveFilter, SortProperty } from "@/components/shared/NotionFilter/types";

// columns は useMemo で定義（SortableHeader を使用）
```

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] 静的定数はモジュールスコープに巻き上げ（rendering-hoist-jsx）
- [ ] `columns` は `useMemo` でキャッシュ
- [ ] `useTransition` / `useCallback` の適切な使用

## 依存関係

- バックエンドの変更は不要（`useGetCheckups`・BE APIはすでに`start_date`/`end_date`対応済み）
- `make codegen` も不要

## 完了条件

- [ ] `pnpm build` が通る（型エラーなし）
- [ ] `pnpm lint` が通る
- [ ] 定期健診一覧で NotionFilter が表示される
- [ ] 日付フィルタを追加すると絞り込まれる（APIにパラメータが送信される）
- [ ] テキスト検索・ソート・ページネーションが動作する
- [ ] 旧シンプルUI（`<input type="text" ...>`）が完全に削除されている
