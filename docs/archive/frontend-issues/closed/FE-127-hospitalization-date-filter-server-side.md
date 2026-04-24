# FE-127: 入院・ホテル一覧 日付フィルタのサーバーサイド移行

**Status**: Open
**Priority**: Medium
**Affects**: `features/hospitalization/`
**Date Created**: 2026-03-26
**Related**: TASK-031, BE-064

## Summary

入院・ホテル一覧の入院日フィルタは現状クライアントサイドのみ（全件ロード後に JS で絞り込み）。BE-064 で追加される `start_date` / `end_date` クエリパラメータを使い、サーバーサイドフィルタへ移行する。

## 現状のコード

```typescript
// frontend/src/features/hospitalization/api/get-hospitalizations.ts:12
export const getHospitalizations = async (): Promise<Hospitalization[]> => {
  const { data } = await axios.get<HospitalizationPaginatedResponse>("/v1/hospitalizations");
  // ↑ クエリパラメータなし
  return data.data.map(transformHospitalization);
};

// frontend/src/features/hospitalization/routes/HospitalizationList.tsx:120-129
// 入院日フィルタ（date-range）
const dateFilter = activeFilters.find((f) => f.key === "startDate")?.value as
  | { from?: string; to?: string } | undefined;
if (dateFilter?.from || dateFilter?.to) {
  result = result.filter((h) => {
    const d = h.startDate?.slice(0, 10) ?? "";
    return (!dateFilter.from || d >= dateFilter.from) && (!dateFilter.to || d <= dateFilter.to);
  });
}
// ↑ 全件ロード後にクライアントサイドで日付フィルタ
```

## 必要な変更

### 1. `get-hospitalizations.ts` — フィルタ型追加・クエリパラメータ送信

```typescript
// Before
export const getHospitalizations = async (): Promise<Hospitalization[]> => {
  const { data } = await axios.get<HospitalizationPaginatedResponse>("/v1/hospitalizations");
  return data.data.map(transformHospitalization);
};

export const useGetHospitalizations = () => {
  return useQuery({
    queryKey: ["hospitalizations"],
    queryFn: getHospitalizations,
  });
};

// After
export interface HospitalizationFilters {
  startDate?: string; // YYYY-MM-DD（入院開始日の範囲）
  endDate?: string;   // YYYY-MM-DD
}

export const getHospitalizations = async (filters?: HospitalizationFilters): Promise<Hospitalization[]> => {
  const params: Record<string, string> = {};
  if (filters?.startDate) params.start_date = filters.startDate;
  if (filters?.endDate) params.end_date = filters.endDate;
  const { data } = await axios.get<HospitalizationPaginatedResponse>("/v1/hospitalizations", { params });
  return data.data.map(transformHospitalization);
};

export const useGetHospitalizations = (filters?: HospitalizationFilters) => {
  return useQuery({
    queryKey: ["hospitalizations", filters],
    queryFn: () => getHospitalizations(filters),
  });
};
```

### 2. `HospitalizationList.tsx` — クライアントサイド日付フィルタ削除・API に渡す

`typeFilteredHospitalizations` の `useMemo` から入院日フィルタブロックを削除し、`useGetHospitalizations` に日付フィルタを渡す。

```typescript
// Before（HospitalizationList.tsx）
const {
  filteredHospitalizations, // useHospitalizationList() から取得（フィルタなし）
  ...
} = useHospitalizationList();

const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);

const typeFilteredHospitalizations = useMemo(() => {
  let result = filteredHospitalizations;
  // ... 入院区分フィルタ ...

  // 入院日フィルタ（date-range）← 削除対象
  const dateFilter = activeFilters.find((f) => f.key === "startDate")?.value as ...;
  if (dateFilter?.from || dateFilter?.to) {
    result = result.filter(...);
  }

  // ... species フィルタ ...
}, [filteredHospitalizations, activeFilters]);

// After — 日付フィルタをサーバーサイドに移行
import type { HospitalizationFilters } from "../api/get-hospitalizations";
import { useGetHospitalizations } from "../api/get-hospitalizations";

// useMemo で activeFilters から日付フィルタを抽出
const dateFilters = useMemo<HospitalizationFilters>(() => {
  const dateFilter = activeFilters.find((f) => f.key === "startDate")?.value as
    | { from?: string; to?: string } | undefined;
  return {
    startDate: dateFilter?.from,
    endDate: dateFilter?.to,
  };
}, [activeFilters]);

// useGetHospitalizations に dateFilters を渡す（useHospitalizationList の内部呼び出しを置き換え or 並立）
// ※ useHospitalizationList は searchTerm/statusFilter を管理するので維持し、
//   日付フィルタのある場合は useGetHospitalizations を直接呼び出してデータを取得する
const { data: allHospitalizations = [] } = useGetHospitalizations(dateFilters);

// typeFilteredHospitalizations の入院日フィルタブロックを削除
const typeFilteredHospitalizations = useMemo(() => {
  let result = allHospitalizations;
  // 入院区分フィルタ（クライアントサイドのまま維持）
  // ...
  // 入院日フィルタ ← 削除
  // species フィルタ（クライアントサイドのまま維持）
  // ...
}, [allHospitalizations, activeFilters]);
```

> **注意**: `useHospitalizationList` が内部で `useHospitalizations` を呼んでいるため、リファクタリング時は `searchTerm` / `statusFilter` の適用と `dateFilters` の適用を両立させること。最もシンプルな実装は `HospitalizationList.tsx` で `useGetHospitalizations(dateFilters)` を直接呼び出し、既存の `filteredHospitalizations`（searchTerm/status フィルタ済み）の代わりに使う方法。

## プロジェクトルール遵守チェック

- [x] `any` 型なし
- [x] `FC` / `forwardRef` なし
- [x] barrel index 経由 import なし
- [x] 条件レンダー `? ... : null`
- [x] `useMemo` deps は primitive

## 依存関係

- BE-064 が先に完了している必要がある（`start_date`/`end_date` クエリパラメータが必要）

## 完了条件

- [ ] 日付フィルタを設定すると `GET /v1/hospitalizations?start_date=...&end_date=...` がネットワークタブに表示される
- [ ] `HospitalizationList.tsx` のクライアントサイド日付フィルタコードが削除されている
- [ ] 入院区分・種フィルタ・テキスト検索・ボード/リスト切り替えは引き続き正常動作する
- [ ] `pnpm build` パス
- [ ] `pnpm lint` パス
