# FE-126: トリミング一覧 日付フィルタのサーバーサイド移行

**Status**: Open
**Priority**: Medium
**Affects**: `features/trimming/`
**Date Created**: 2026-03-26
**Related**: TASK-031, BE-063

## Summary

トリミング一覧の日付フィルタは現状クライアントサイドのみ（全件ロード後に JS で絞り込み）。BE-063 で追加される `start_date` / `end_date` クエリパラメータを使い、他ページと同様にサーバーサイドフィルタへ移行する。

## 現状のコード

```typescript
// frontend/src/features/trimming/api/get-trimmings.ts:7
export const getTrimmings = async (): Promise<TrimmingUI[]> => {
  const { data } = await axios.get<TrimmingListResponse>("/v1/trimmings");
  // ↑ クエリパラメータなし
  return data.data.map(transformTrimming);
};

// frontend/src/features/trimming/hooks/use-trimming-records.ts:65-78
return result.filter((r) => {
  const recordDate = r.date.slice(0, 10);
  const matchesDate =
    (!from || recordDate >= from) &&
    (!to || recordDate <= to);
  return matchesKeyword && matchesDate;
});
// ↑ 全件ロード後にクライアントサイドで日付フィルタ
```

## 必要な変更

### 1. `get-trimmings.ts` — フィルタ型追加・クエリパラメータ送信

```typescript
// Before
export const getTrimmings = async (): Promise<TrimmingUI[]> => {
  const { data } = await axios.get<TrimmingListResponse>("/v1/trimmings");
  return data.data.map(transformTrimming);
};

export const useGetTrimmings = () => {
  return useQuery({
    queryKey: ["trimmings"],
    queryFn: getTrimmings,
  });
};

// After
export interface TrimmingFilters {
  startDate?: string; // YYYY-MM-DD
  endDate?: string;   // YYYY-MM-DD
}

export const getTrimmings = async (filters?: TrimmingFilters): Promise<TrimmingUI[]> => {
  const params: Record<string, string> = {};
  if (filters?.startDate) params.start_date = filters.startDate;
  if (filters?.endDate) params.end_date = filters.endDate;
  const { data } = await axios.get<TrimmingListResponse>("/v1/trimmings", { params });
  return data.data.map(transformTrimming);
};

export const useGetTrimmings = (filters?: TrimmingFilters) => {
  return useQuery({
    queryKey: ["trimmings", filters],
    queryFn: () => getTrimmings(filters),
  });
};
```

### 2. `use-trimming-records.ts` — クライアントサイド日付フィルタ削除・サーバーサイドへ移行

```typescript
// Before
export function useFilterTrimmingRecords(
  searchTerm: string,
  dateRange: DateRange,
  activeFilters?: ActiveFilter[],
) {
  const { data: trimmingRecords = [], ... } = useGetTrimmings();
  const { from, to } = dateRange;

  const filteredRecords = useMemo(() => {
    // ... status/species/staff フィルタ ...
    return result.filter((r) => {
      const recordDate = r.date.slice(0, 10);
      const matchesDate = (!from || recordDate >= from) && (!to || recordDate <= to);
      return matchesKeyword && matchesDate;
    });
  }, [trimmingRecords, searchTerm, from, to, activeFilters]);
}

// After
import type { TrimmingFilters } from "../api/get-trimmings";

export function useFilterTrimmingRecords(
  searchTerm: string,
  filters?: TrimmingFilters,        // dateRange から TrimmingFilters に変更
  activeFilters?: ActiveFilter[],
) {
  const { data: trimmingRecords = [], ... } = useGetTrimmings(filters); // フィルタをAPIに渡す

  const filteredRecords = useMemo(() => {
    let result = trimmingRecords;

    // status / species / staff フィルタ（クライアントサイドのまま維持）
    // ... 既存の filter ロジック ...

    // テキスト検索
    return result.filter((r) =>
      searchTerm === "" ||
      r.ownerName.toLowerCase().includes(searchTerm.toLowerCase()) ||
      r.petName.toLowerCase().includes(searchTerm.toLowerCase()),
    );
    // ↑ 日付フィルタの行を削除（サーバーサイドに移行）
  }, [trimmingRecords, searchTerm, activeFilters]);
}
```

### 3. `TrimmingList.tsx` — dateRange を TrimmingFilters に置き換え

```typescript
// Before
// activeFilters から日付フィルタ抽出（dateRange 形式）
const dateRange = useMemo(() => {
  const dateFilter = activeFilters.find((f) => f.key === "date")?.value as
    | { from?: string; to?: string } | undefined;
  return { from: dateFilter?.from ?? "", to: dateFilter?.to ?? "" };
}, [activeFilters]);

const { data: filteredRecords, ... } = useFilterTrimmingRecords(deferredSearch, dateRange, activeFilters);

// After
import type { TrimmingFilters } from "../api/get-trimmings";

const filters = useMemo<TrimmingFilters>(() => {
  const dateFilter = activeFilters.find((f) => f.key === "date")?.value as
    | { from?: string; to?: string } | undefined;
  return {
    startDate: dateFilter?.from,
    endDate: dateFilter?.to,
  };
}, [activeFilters]);

const { data: filteredRecords, ... } = useFilterTrimmingRecords(deferredSearch, filters, activeFilters);
```

## プロジェクトルール遵守チェック

- [x] `any` 型なし
- [x] `FC` / `forwardRef` なし
- [x] barrel index 経由 import なし
- [x] 条件レンダー `? ... : null`
- [x] `useCallback` / `useMemo` の deps は primitive

## 依存関係

- BE-063 が先に完了している必要がある（`start_date`/`end_date` クエリパラメータが必要）

## 完了条件

- [ ] 日付フィルタを設定すると `GET /v1/trimmings?start_date=...&end_date=...` がネットワークタブに表示される
- [ ] クライアントサイドの日付フィルタコード（`recordDate >= from` 比較）が `use-trimming-records.ts` から削除されている
- [ ] ステータス・種・スタッフフィルタは引き続き正常動作する
- [ ] `pnpm build` パス
- [ ] `pnpm lint` パス
