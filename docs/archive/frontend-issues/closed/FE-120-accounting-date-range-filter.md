# FE-120: 会計管理一覧 - 日付範囲フィルタ追加（loaderData→useQuery移行含む）

**Status**: Open
**Priority**: Medium
**Affects**: 会計管理 (`features/accounting/routes/Accounting.tsx`, `features/accounting/api/get-accountings.ts`)
**Date Created**: 2026-03-25
**Related**: TASK-028, BE-057

## Summary

会計管理一覧に日付範囲フィルタを追加する。現在はloaderDataでデータを取得しているため、日付フィルタをリアクティブに渡せない。`useGetAccountings` hookに日付フィルタを追加し、コンポーネントの取得方法をloaderDataからuseQueryに移行する。

**BE-057の完了後に着手すること**（バックエンドAPIが`start_date`/`end_date`に対応していることが前提）。

## 現状のコード

```ts
// frontend/src/features/accounting/api/get-accountings.ts:14-25
export const getAccountings = async (
  status?: AccountingStatus,
): Promise<Accounting[]> => {
  const params: Record<string, string> = {};
  if (status) params.status = status;
  const { data } = await axios.get<AccountingsListResponse>("/v1/accountings", { params });
  // ↑ start_date / end_date なし
  return data.data.map(transformToAccounting);
};
```

```tsx
// frontend/src/features/accounting/routes/Accounting.tsx:73-74
export function Accounting() {
  const { accountings } = useLoaderData<AccountingsLoaderData>();
  // ↑ loaderData で取得 → フィルタ変更にリアクティブに対応できない

  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  // ← NotionFilter の FILTER_PROPERTIES には "status" のみ。日付フィルタなし
```

## 必要な変更

### 1. `get-accountings.ts` に date フィルタ追加

```ts
// frontend/src/features/accounting/api/get-accountings.ts

// ステータスはクライアントサイドフィルタに残すため、APIフィルタには日付のみ
// AccountingStatus の import も不要になるので削除する
// Before: import type { Accounting, AccountingStatus } from "../types";
// After:  import type { Accounting } from "../types";

export interface AccountingFilters {
  startDate?: string; // YYYY-MM-DD
  endDate?: string;   // YYYY-MM-DD
}

export const getAccountings = async (
  filters?: AccountingFilters,
): Promise<Accounting[]> => {
  const params: Record<string, string> = {};
  if (filters?.startDate) params.start_date = filters.startDate;
  if (filters?.endDate) params.end_date = filters.endDate;
  const { data } = await axios.get<AccountingsListResponse>("/v1/accountings", { params });
  return data.data.map(transformToAccounting);
};

export const useGetAccountings = (filters?: AccountingFilters) => {
  return useQuery({
    queryKey: ["accountings", filters],  // filters をキャッシュキーに含める
    queryFn: () => getAccountings(filters),
  });
};
```

### 2. `Accounting.tsx` の変更

```tsx
// frontend/src/features/accounting/routes/Accounting.tsx

// rendering-hoist-jsx: FILTER_PROPERTIES に date-range を追加
const FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "date",
    label: "日付",
    type: "date-range",
    icon: Calendar,  // ← import { Calendar } from "lucide-react"
  },
  {
    key: "status",
    label: "ステータス",
    type: "select",
    icon: CircleDot,
    options: [
      { value: "waiting", label: "会計待ち" },
      { value: "completed", label: "会計済" },
      { value: "cancelled", label: "キャンセル" },
    ],
  },
];

export function Accounting() {
  const navigate = useNavigate();

  // ① loaderData を廃止。useGetAccountings に移行
  // const { accountings } = useLoaderData<AccountingsLoaderData>();  ← 削除

  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const deferredSearch = useDeferredValue(searchTerm);
  const isFiltering = searchTerm !== deferredSearch;

  // ② activeFilters から日付フィルタのみを抽出してAPIに渡す
  // ステータスフィルタは is_not/is_empty 等の条件があるためクライアントサイドのまま維持
  const apiFilters = useMemo(() => {
    const dateFilter = activeFilters.find((f) => f.key === "date")?.value as
      | { from?: string; to?: string }
      | undefined;
    return {
      startDate: dateFilter?.from,
      endDate: dateFilter?.to,
    };
  }, [activeFilters]);

  // ③ useGetAccountings に apiFilters（日付のみ）を渡す
  const { data: accountings = [], isLoading, isError } = useGetAccountings(apiFilters);

  // ④ ステータスフィルタはクライアントサイドで維持（既存実装を変更しない）
  const filteredRecords = useMemo(() => {
    let result = accountings;

    // ステータスフィルタ（クライアントサイド維持: is_not/is_empty 条件に対応済みのため）
    const statusFilter = activeFilters.find((f) => f.key === "status");
    if (statusFilter && typeof statusFilter.value === "string") {
      result = result.filter((r) => {
        switch (statusFilter.condition) {
          case "is": return r.status === statusFilter.value;
          case "is_not": return r.status !== statusFilter.value;
          case "is_empty": return !r.status;
          case "is_not_empty": return !!r.status;
          default: return r.status === statusFilter.value;
        }
      });
    }

    // テキスト検索
    if (deferredSearch) {
      const lowerTerm = deferredSearch.toLowerCase();
      result = result.filter(
        (r) =>
          r.ownerName.toLowerCase().includes(lowerTerm) ||
          r.petName.toLowerCase().includes(lowerTerm),
      );
    }

    return result;
  }, [accountings, activeFilters, deferredSearch]);

  // ⑤ ローディング・エラー状態を追加（loaderData廃止後は useQuery が状態を管理）
  if (isLoading) return <LoadingFallback />;
  if (isError) return <ErrorFallback />;

  // ... 以下は既存と同様（sortedData, pagination, renderRow など）
```

### 3. `loaders.ts` の削除 + `router.tsx` の修正

`accountingsLoader` は useQuery 移行後は不要。`loaders.ts` を削除し、`router.tsx` を以下の通り変更する。

```ts
// frontend/src/app/router.tsx:278-285
// Before（Promise.all で loaders.ts を import している）
{
  index: true,
  lazy: async () => {
    const [{ Accounting }, { accountingsLoader }] = await Promise.all([
      import("@/features/accounting/routes/Accounting"),
      import("@/features/accounting/loaders"),
    ]);
    return { Component: Accounting, loader: accountingsLoader };
  },
},

// After（単一 import に縮小、loader 削除）
{
  index: true,
  lazy: async () => {
    const { Accounting } = await import("@/features/accounting/routes/Accounting");
    return { Component: Accounting };
  },
},
```

その後 `frontend/src/features/accounting/loaders.ts` を削除する。

## 追加 import

```tsx
// Accounting.tsx に追加
import { Calendar } from "lucide-react";  // 既存の CircleDot と同行に追加
import { useGetAccountings } from "../api/get-accountings";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates/DataStates";

// 削除
// import { useLoaderData } from "react-router";
// import type { AccountingsLoaderData } from "../loaders";
```

## プロジェクトルール遵守チェック

- [ ] `any` 型なし（`AccountingFilters` 型を定義）
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `FILTER_PROPERTIES` はモジュールスコープに巻き上げ（rendering-hoist-jsx）
- [ ] `apiFilters` は `useMemo` でキャッシュ（日付フィルタのみ）
- [ ] `queryKey` に `filters` を含める（キャッシュ正常動作）
- [ ] ステータスフィルタは `is_not`/`is_empty`/`is_not_empty` を含むためクライアントサイドのまま

## 依存関係

- **BE-057 が先に完了している必要がある**（バックエンドが`start_date`/`end_date`を受け付けること）

## 完了条件

- [ ] `pnpm build` が通る（型エラーなし）
- [ ] `pnpm lint` が通る
- [ ] 会計管理一覧の NotionFilter に「日付」フィルタが表示される
- [ ] 日付を指定するとAPIに`start_date`/`end_date`パラメータが送信される
- [ ] `get-accountings.ts` から `AccountingStatus` の import が削除されている（`AccountingFilters` に status なし）
- [ ] ステータスフィルタ（is/is_not/is_empty）がクライアントサイドで引き続き動作する
- [ ] 日付フィルタとステータスフィルタの複合フィルタが正しく動作する
- [ ] `frontend/src/app/router.tsx` の accounting route が Promise.all なしの単一 import に変更されている（loader も削除）
- [ ] `frontend/src/features/accounting/loaders.ts` が削除されている
- [ ] テキスト検索・ソート・ページネーションが引き続き動作する
