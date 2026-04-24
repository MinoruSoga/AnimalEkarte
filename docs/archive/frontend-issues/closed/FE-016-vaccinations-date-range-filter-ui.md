# FE-016: 予防接種一覧 — 日付範囲フィルタUI追加

**Status**: Open
**Priority**: Medium
**Affects**: vaccinations feature — 一覧画面
**Date Created**: 2026-03-17
**Related**: TASK-002, BE-042

## Summary

予防接種一覧画面に日付範囲フィルタ（DateRangePicker）を追加し、API に `start_date` / `end_date` パラメータを送信する。FE-015（健康診断）と同一パターン。

## 現状のコード

### API hook — フィルタパラメータなし

```typescript
// frontend/src/features/vaccinations/api/get-vaccinations.ts:7-17
export const getVaccinations = async (): Promise<VaccinationRecord[]> => {
  const { data } = await axios.get<{ data: BackendVaccination[] }>(
    "/v1/vaccinations"  // ← パラメータなし
  );
  return (data.data ?? []).map(transformVaccination);
};

export const useGetVaccinations = () => {
  return useQuery({
    queryKey: ["vaccinations"],
    queryFn: getVaccinations,
  });
};
```

### 一覧画面 — テキスト検索のみ

```typescript
// frontend/src/features/vaccinations/routes/VaccinationList.tsx:31-98
const [searchTerm, setSearchTerm] = useState("");
const deferredSearchTerm = useDeferredValue(searchTerm);
const { data: filteredRecords } = useVaccinations(deferredSearchTerm);

// 行66-72: SearchFilterBar のみ
<SearchFilterBar
  searchTerm={searchTerm}
  onSearchChange={handleSearchChange}
  placeholder="飼主名、ペット名、予防接種名..."
  count={filteredRecords.length}
/>
```

### useVaccinations — クライアント側フィルタ

```typescript
// frontend/src/features/vaccinations/hooks/useVaccinations.ts:1-19
export function useVaccinations(searchTerm: string) {
  const { data = [], isLoading, error } = useGetVaccinations();
  const filteredRecords = useMemo(() => {
    if (!searchTerm) return data;
    const lowerTerm = searchTerm.toLowerCase();
    return data.filter(
      (r) =>
        r.ownerName.toLowerCase().includes(lowerTerm) ||
        r.petName.toLowerCase().includes(lowerTerm) ||
        r.vaccineName.toLowerCase().includes(lowerTerm)
    );
  }, [data, searchTerm]);
  return { data: filteredRecords, isLoading, error };
}
```

## 必要な変更

### 1. API hook — 日付パラメータ追加

```typescript
// frontend/src/features/vaccinations/api/get-vaccinations.ts

interface VaccinationFilters {
  startDate?: string;  // YYYY-MM-DD
  endDate?: string;    // YYYY-MM-DD
}

export const getVaccinations = async (
  filters?: VaccinationFilters,
): Promise<VaccinationRecord[]> => {
  const params: Record<string, string> = {};
  if (filters?.startDate) params.start_date = filters.startDate;
  if (filters?.endDate) params.end_date = filters.endDate;
  const { data } = await axios.get<{ data: BackendVaccination[] }>(
    "/v1/vaccinations", { params },
  );
  return (data.data ?? []).map(transformVaccination);
};

export const useGetVaccinations = (filters?: VaccinationFilters) => {
  return useQuery({
    queryKey: ["vaccinations", filters],
    queryFn: () => getVaccinations(filters),
  });
};
```

### 2. useVaccinations — フィルタパラメータ受け取り

```typescript
// frontend/src/features/vaccinations/hooks/useVaccinations.ts

export function useVaccinations(
  searchTerm: string,
  filters?: VaccinationFilters,
) {
  const { data = [], isLoading, error } = useGetVaccinations(filters);
  // ... テキスト検索フィルタはクライアント側で維持
}
```

### 3. 一覧画面 — DateRangePicker 追加

```typescript
// frontend/src/features/vaccinations/routes/VaccinationList.tsx

const [dateRange, setDateRange] = useState<{ from?: Date; to?: Date }>({});

const filters = useMemo(() => ({
  startDate: dateRange.from ? format(dateRange.from, "yyyy-MM-dd") : undefined,
  endDate: dateRange.to ? format(dateRange.to, "yyyy-MM-dd") : undefined,
}), [dateRange]);

const { data: filteredRecords } = useVaccinations(deferredSearchTerm, filters);

// UI: SearchFilterBar の横に DateRangePicker を配置
<div className="flex items-center gap-4">
  <SearchFilterBar ... />
  <DateRangePicker
    value={dateRange}
    onChange={setDateRange}
  />
</div>
```

## UI 操作フロー

1. ユーザーが予防接種一覧画面を開く
2. デフォルトは全期間表示
3. DateRangePicker で開始日・終了日を選択
4. 指定期間内の予防接種のみ表示される
5. テキスト検索との併用が可能

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useMemo` でフィルタオブジェクト安定化
- [ ] 型は `models.ts` から導出

## 依存関係

- **BE-042** が先に完了している必要がある（`start_date` / `end_date` パラメータが必要）
- 共有コンポーネント `DateRangePicker` は既存

## 完了条件

- [ ] `get-vaccinations.ts` に startDate/endDate パラメータ追加
- [ ] `useGetVaccinations` の queryKey にフィルタを含める
- [ ] DateRangePicker UI を SearchFilterBar の横に配置
- [ ] 日付範囲選択で期間内のみ表示
- [ ] テキスト検索との併用が動作
- [ ] 型エラーなし（`docker compose exec frontend pnpm build` パス）
- [ ] ESLint エラーなし（`docker compose exec frontend pnpm lint` パス）
