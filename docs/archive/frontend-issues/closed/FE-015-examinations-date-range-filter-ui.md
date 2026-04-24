# FE-015: 健康診断一覧 — 日付範囲フィルタUI追加

**Status**: Open
**Priority**: Medium
**Affects**: examinations feature — 一覧画面
**Date Created**: 2026-03-17
**Related**: TASK-002, BE-041

## Summary

健康診断一覧画面に日付範囲フィルタ（DateRangePicker）を追加し、API に `start_date` / `end_date` パラメータを送信する。既存の共有コンポーネント `DateRangePicker` を利用する。

## 現状のコード

### API hook — フィルタパラメータなし

```typescript
// frontend/src/features/examinations/api/get-examinations.ts:14-24
export const getExaminations = async (): Promise<ExaminationRecord[]> => {
  const { data } = await axios.get<ExaminationsListResponse>("/v1/examinations");
  // ← パラメータなし
  return data.data.map(transformExamination);
};

export const useGetExaminations = () => {
  return useQuery({
    queryKey: ["examinations"],
    queryFn: getExaminations,
  });
};
```

### 一覧画面 — テキスト検索のみ

```typescript
// frontend/src/features/examinations/routes/Examinations.tsx:37-39
const [searchTerm, setSearchTerm] = useState("");
const deferredSearch = useDeferredValue(searchTerm);
const { data: filteredRecords, isLoading } = useExaminationRecords(deferredSearch);

// 行69-74: SearchFilterBar のみ
<SearchFilterBar
  searchTerm={searchTerm}
  onSearchChange={setSearchTerm}
  placeholder="飼主名、ペット名、検査種別..."
  count={isLoading ? undefined : filteredRecords.length}
/>
```

### useExaminationRecords — クライアント側フィルタ

```typescript
// frontend/src/features/examinations/hooks/useExaminationRecords.ts:4-19
export function useExaminationRecords(searchTerm: string) {
  const { data = [], isLoading, error } = useGetExaminations();
  const filteredRecords = useMemo(() => {
    if (!searchTerm) return data;
    const lowerTerm = searchTerm.toLowerCase();
    return data.filter(
      (r) =>
        r.ownerName.toLowerCase().includes(lowerTerm) ||
        r.petName.toLowerCase().includes(lowerTerm) ||
        r.testType.toLowerCase().includes(lowerTerm)
    );
  }, [data, searchTerm]);
  return { data: filteredRecords, isLoading, error };
}
```

### 共有コンポーネント — DateRangePicker

```
frontend/src/components/shared/DateRangePicker/  ← 既に存在
```

## 必要な変更

### 1. API hook — 日付パラメータ追加

```typescript
// frontend/src/features/examinations/api/get-examinations.ts

interface ExaminationFilters {
  startDate?: string;  // YYYY-MM-DD
  endDate?: string;    // YYYY-MM-DD
}

export const getExaminations = async (
  filters?: ExaminationFilters,
): Promise<ExaminationRecord[]> => {
  const params: Record<string, string> = {};
  if (filters?.startDate) params.start_date = filters.startDate;
  if (filters?.endDate) params.end_date = filters.endDate;
  const { data } = await axios.get<ExaminationsListResponse>("/v1/examinations", { params });
  return data.data.map(transformExamination);
};

export const useGetExaminations = (filters?: ExaminationFilters) => {
  return useQuery({
    queryKey: ["examinations", filters],
    queryFn: () => getExaminations(filters),
  });
};
```

### 2. useExaminationRecords — フィルタパラメータ受け取り

```typescript
// frontend/src/features/examinations/hooks/useExaminationRecords.ts

export function useExaminationRecords(
  searchTerm: string,
  filters?: ExaminationFilters,
) {
  const { data = [], isLoading, error } = useGetExaminations(filters);
  // ... テキスト検索フィルタはクライアント側で維持
}
```

### 3. 一覧画面 — DateRangePicker 追加

```typescript
// frontend/src/features/examinations/routes/Examinations.tsx

const [dateRange, setDateRange] = useState<{ from?: Date; to?: Date }>({});

const filters = useMemo(() => ({
  startDate: dateRange.from ? format(dateRange.from, "yyyy-MM-dd") : undefined,
  endDate: dateRange.to ? format(dateRange.to, "yyyy-MM-dd") : undefined,
}), [dateRange]);

const { data: filteredRecords, isLoading } = useExaminationRecords(
  deferredSearch, filters,
);

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

1. ユーザーが健康診断一覧画面を開く
2. デフォルトは全期間表示
3. DateRangePicker で開始日・終了日を選択
4. 指定期間内の健康診断のみ表示される
5. テキスト検索との併用が可能

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useMemo` でフィルタオブジェクト安定化
- [ ] 型は `models.ts` から導出

## 依存関係

- **BE-041** が先に完了している必要がある（`start_date` / `end_date` パラメータが必要）
- 共有コンポーネント `DateRangePicker` は既存

## 完了条件

- [ ] `get-examinations.ts` に startDate/endDate パラメータ追加
- [ ] `useGetExaminations` の queryKey にフィルタを含める
- [ ] DateRangePicker UI を SearchFilterBar の横に配置
- [ ] 日付範囲選択で期間内のみ表示
- [ ] テキスト検索との併用が動作
- [ ] 型エラーなし（`docker compose exec frontend pnpm build` パス）
- [ ] ESLint エラーなし（`docker compose exec frontend pnpm lint` パス）
