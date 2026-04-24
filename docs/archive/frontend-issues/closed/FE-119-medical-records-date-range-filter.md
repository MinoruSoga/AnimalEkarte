# FE-119: カルテ一覧 - 日付範囲フィルタ追加

**Status**: Open
**Priority**: Medium
**Affects**: カルテ管理 (`features/medical-records/`)
**Date Created**: 2026-03-25
**Related**: TASK-028, BE-056

## Summary

カルテ一覧の`NotionFilter`に`date-range`プロパティを追加し、診療日での絞り込みを可能にする。現在は`NotionFilter`に`properties={[]}`が渡されており、フィルタが機能しない状態。

**BE-056の完了後に着手すること**（バックエンドAPIが`start_date`/`end_date`に対応していることが前提）。

## 現状のコード

```tsx
// frontend/src/features/medical-records/routes/MedicalRecords.tsx:103-110
<NotionFilter
  properties={[]}          // ← 空配列。フィルタ機能なし
  activeFilters={[]}       // ← ハードコード空配列
  onFilterChange={() => {}} // ← no-op
  searchTerm={searchTerm}
  onSearchChange={setSearchTerm}
  searchPlaceholder="飼主名、ペット名、カルテNo、主訴で検索..."
  count={filteredRecords.length}
  sortProperties={MEDICAL_RECORD_SORT_PROPERTIES}
  activeSorts={activeSorts}
  onSortChange={setActiveSorts}
/>
```

```ts
// frontend/src/features/medical-records/api/get-medical-records.ts
export const getMedicalRecords = async (): Promise<MedicalRecord[]> => {
  const { data } = await axios.get<MedicalRecordsListResponse>("/v1/medical-records");
  // ↑ クエリパラメータなし
  return data.data.map(transformMedicalRecord);
};
```

```ts
// frontend/src/features/medical-records/hooks/use-medical-records.ts
export function useFilterMedicalRecords(searchTerm: string) {
  const { data: records = [], isLoading, isError } = useGetMedicalRecords();
  // ↑ date フィルタを受け取らない
```

## 必要な変更

### 1. `get-medical-records.ts` に filters 追加

```ts
// frontend/src/features/medical-records/api/get-medical-records.ts

export interface MedicalRecordFilters {
  startDate?: string; // YYYY-MM-DD
  endDate?: string;   // YYYY-MM-DD
}

export const getMedicalRecords = async (
  filters?: MedicalRecordFilters,
): Promise<MedicalRecord[]> => {
  const params: Record<string, string> = {};
  if (filters?.startDate) params.start_date = filters.startDate;
  if (filters?.endDate) params.end_date = filters.endDate;
  const { data } = await axios.get<MedicalRecordsListResponse>("/v1/medical-records", { params });
  return data.data.map(transformMedicalRecord);
};

export const useGetMedicalRecords = (filters?: MedicalRecordFilters) => {
  return useQuery({
    queryKey: ["medical-records", filters],  // filters をキャッシュキーに含める
    queryFn: () => getMedicalRecords(filters),
  });
};
```

### 2. `use-medical-records.ts` に date フィルタ引数追加

```ts
// frontend/src/features/medical-records/hooks/use-medical-records.ts
import type { MedicalRecordFilters } from "../api/get-medical-records";

export function useFilterMedicalRecords(
  searchTerm: string,
  filters?: MedicalRecordFilters,
) {
  const { data: records = [], isLoading, isError } = useGetMedicalRecords(filters);

  const filteredRecords = useMemo(() => {
    if (!searchTerm) return records;
    const lowerTerm = searchTerm.toLowerCase();
    return records.filter(
      (r) =>
        r.ownerName.toLowerCase().includes(lowerTerm) ||
        r.petName.toLowerCase().includes(lowerTerm) ||
        r.recordNo.toLowerCase().includes(lowerTerm) ||
        r.chiefComplaint.toLowerCase().includes(lowerTerm)
    );
  }, [records, searchTerm]);

  return { data: filteredRecords, isLoading, isError };
}
```

### 3. `MedicalRecords.tsx` に NotionFilter 設定追加

```tsx
// frontend/src/features/medical-records/routes/MedicalRecords.tsx

// 追加 import
import { Calendar } from "lucide-react";
import { useState } from "react";  // activeFilters state のため（既存のuseStateに追加）
import type { FilterProperty, ActiveFilter } from "@/components/shared/NotionFilter/types";
import type { MedicalRecordFilters } from "../api/get-medical-records";

// rendering-hoist-jsx: 静的定数をモジュールスコープに
const FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "date",
    label: "診療日",
    type: "date-range",
    icon: Calendar,
  },
];

// コンポーネント内:
export function MedicalRecords() {
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);  // ← 追加
  const deferredSearch = useDeferredValue(searchTerm);

  // activeFilters から日付フィルタを抽出
  const filters = useMemo<MedicalRecordFilters>(() => {
    const dateFilter = activeFilters.find((f) => f.key === "date")?.value as
      | { from?: string; to?: string }
      | undefined;
    return {
      startDate: dateFilter?.from,
      endDate: dateFilter?.to,
    };
  }, [activeFilters]);

  const { data: filteredRecords, isLoading, isError } = useFilterMedicalRecords(deferredSearch, filters);

  // ... 以下は既存と同様

  // NotionFilter の props を更新:
  return (
    ...
    <NotionFilter
      properties={FILTER_PROPERTIES}    // ← [] から変更
      activeFilters={activeFilters}      // ← state に変更
      onFilterChange={setActiveFilters}  // ← no-op から変更
      searchTerm={searchTerm}
      onSearchChange={setSearchTerm}
      searchPlaceholder="飼主名、ペット名、カルテNo、主訴で検索..."
      count={filteredRecords.length}
      sortProperties={MEDICAL_RECORD_SORT_PROPERTIES}
      activeSorts={activeSorts}
      onSortChange={setActiveSorts}
    />
    ...
  );
}
```

## プロジェクトルール遵守チェック

- [ ] `any` 型なし（`MedicalRecordFilters` 型を定義）
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし（直接ファイルimport）
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `FILTER_PROPERTIES` はモジュールスコープに巻き上げ（rendering-hoist-jsx）
- [ ] `filters` は `useMemo` でキャッシュ（activeFilters 変化時のみ再計算）
- [ ] `queryKey` に `filters` を含める（キャッシュ正常動作）

## 依存関係

- **BE-056 が先に完了している必要がある**（バックエンドが`start_date`/`end_date`を受け付けること）

## 完了条件

- [ ] `pnpm build` が通る（型エラーなし）
- [ ] `pnpm lint` が通る
- [ ] カルテ一覧の NotionFilter に「診療日」フィルタが表示される
- [ ] 日付を指定するとAPIに`start_date`/`end_date`パラメータが送信される
- [ ] 日付フィルタなし（デフォルト）は全件表示される
- [ ] テキスト検索・ソートと日付フィルタが共存して動作する
