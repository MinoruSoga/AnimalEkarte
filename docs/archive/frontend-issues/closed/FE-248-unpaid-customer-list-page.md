# FE-248: 未納者一覧画面（タブ切替・売掛金総額表示）

**Status**: Closed (2026-04-14, commit 8fcd1382)
**Priority**: High
**Affects**: `features/accounting`, `app/router.tsx`, `config/paths.ts`, `components/shared/Layout`
**Date Created**: 2026-04-14
**Related**: BUG-370, BE-110

## Summary

基準日時点で未納な billings を一覧表示する画面を新設。「飼主単位 / 会計単位」タブ切替、売掛金総額サマリー表示、URL クエリ同期、経過日数降順ソートを実装する。

## 現状のコード

`frontend/src/features/accounting/routes/AccountingList.tsx:60-79` のステータスフィルタには `waiting` / `completed` / `cancelled` のみで、「未納者一覧」専用ビューは存在しない。

```typescript
// frontend/src/features/accounting/routes/AccountingList.tsx:75-79
options: [
  { value: "waiting", label: "会計待ち" },
  { value: "completed", label: "会計済" },
  { value: "cancelled", label: "キャンセル" },
],
```

## 必要な変更

### 1. 型定義

```typescript
// frontend/src/features/accounting/api/types.ts に追加

export interface UnpaidByOwner {
  ownerId: number;
  ownerName: string;
  billingCount: number;
  totalAmount: number;
  oldestDate: string;  // ISO date
  latestDate: string;  // ISO date
  elapsedDays: number;
}

export interface UnpaidSummary {
  totalAmount: number;
  billingCount: number;
  ownerCount: number;
}

export type UnpaidGroupBy = "owner" | "billing";

export interface GetUnpaidParams {
  asOf: string;        // YYYY-MM-DD
  groupBy: UnpaidGroupBy;
  page: number;
  limit: number;
}

export interface GetUnpaidByOwnerResponse {
  data: UnpaidByOwner[];
  summary: UnpaidSummary;
  total: number;
  page: number;
  limit: number;
}

export interface GetUnpaidByBillingResponse {
  data: Accounting[]; // 既存 Accounting 型を再利用
  summary: UnpaidSummary;
  total: number;
  page: number;
  limit: number;
}
```

### 2. API hook

```typescript
// frontend/src/features/accounting/api/get-unpaid-billings.ts (新規)

import { useQuery } from "@tanstack/react-query";
import axios from "@/lib/axios";
import { QUERY_STALE_TIMES } from "@/lib/react-query";
import type {
  GetUnpaidParams,
  GetUnpaidByOwnerResponse,
  GetUnpaidByBillingResponse,
} from "./types";
import { transformBackendAccountingToFrontend } from "./transforms";

export async function getUnpaidByOwner(
  clinicID: number,
  params: GetUnpaidParams,
): Promise<GetUnpaidByOwnerResponse> {
  const { data } = await axios.get(`/api/clinics/${clinicID}/unpaid-billings`, {
    params: { as_of: params.asOf, group_by: "owner", page: params.page, limit: params.limit },
  });
  // snake_case → camelCase 変換は axios interceptor で実施想定
  return data;
}

export async function getUnpaidByBilling(
  clinicID: number,
  params: GetUnpaidParams,
): Promise<GetUnpaidByBillingResponse> {
  const { data } = await axios.get(`/api/clinics/${clinicID}/unpaid-billings`, {
    params: { as_of: params.asOf, group_by: "billing", page: params.page, limit: params.limit },
  });
  return {
    ...data,
    data: data.data.map(transformBackendAccountingToFrontend),
  };
}

export function useGetUnpaidByOwner(clinicID: number, params: GetUnpaidParams) {
  return useQuery({
    queryKey: ["unpaid-billings", clinicID, "owner", params.asOf, params.page, params.limit],
    queryFn: () => getUnpaidByOwner(clinicID, params),
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    enabled: params.groupBy === "owner",
  });
}

export function useGetUnpaidByBilling(clinicID: number, params: GetUnpaidParams) {
  return useQuery({
    queryKey: ["unpaid-billings", clinicID, "billing", params.asOf, params.page, params.limit],
    queryFn: () => getUnpaidByBilling(clinicID, params),
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    enabled: params.groupBy === "billing",
  });
}
```

### 3. コンポーネント変更

```typescript
// frontend/src/features/accounting/routes/UnpaidCustomerList.tsx (新規)

import { useCallback, useMemo, useState } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { format } from "date-fns";

import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { Pagination } from "@/components/shared/Pagination/Pagination";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { NotionDatePicker } from "@/components/shared/NotionDatePicker";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { TableCell } from "@/components/ui/table";
import { C, STYLE } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import { formatCurrency } from "@/utils/format/number";
import { usePagination } from "@/hooks/use-pagination";
import { useClinicID } from "@/features/auth";

import { useGetUnpaidByOwner, useGetUnpaidByBilling } from "../api/get-unpaid-billings";
import type { UnpaidGroupBy } from "../api/types";

export function UnpaidCustomerList() {
  const navigate = useNavigate();
  const clinicID = useClinicID();
  const [searchParams, setSearchParams] = useSearchParams();

  const groupBy = (searchParams.get("group_by") as UnpaidGroupBy) || "owner";
  const asOfParam = searchParams.get("as_of") ?? format(new Date(), "yyyy-MM-dd");

  const { page, limit, setPage } = usePagination();

  const params = useMemo(
    () => ({ asOf: asOfParam, groupBy, page, limit }),
    [asOfParam, groupBy, page, limit],
  );

  const ownerQuery = useGetUnpaidByOwner(clinicID, params);
  const billingQuery = useGetUnpaidByBilling(clinicID, params);
  const active = groupBy === "owner" ? ownerQuery : billingQuery;

  const handleAsOfChange = useCallback(
    (date: string) => {
      const next = new URLSearchParams(searchParams);
      next.set("as_of", date);
      setSearchParams(next);
      setPage(1);
    },
    [searchParams, setSearchParams, setPage],
  );

  const handleTabChange = useCallback(
    (value: string) => {
      const next = new URLSearchParams(searchParams);
      next.set("group_by", value);
      setSearchParams(next);
      setPage(1);
    },
    [searchParams, setSearchParams, setPage],
  );

  if (active.isLoading) return <LoadingFallback />;
  if (active.isError) return <ErrorFallback />;

  const summary = active.data?.summary;

  return (
    <PageLayout title="未納者一覧">
      {/* 基準日選択 + サマリーカード */}
      <div className={`${STYLE.FLEX_BETWEEN} mb-4`}>
        <div>
          <label className={`text-sm ${C.text60}`}>基準日</label>
          <NotionDatePicker value={asOfParam} onChange={handleAsOfChange} />
        </div>
        {summary ? (
          <div className={`p-4 rounded ${C.bgPage}`}>
            <div className={`text-xs ${C.text60}`}>売掛金総額</div>
            <div className="text-2xl font-bold">{formatCurrency(summary.totalAmount)}</div>
            <div className={`text-xs ${C.text60}`}>
              {summary.billingCount}件 / {summary.ownerCount}名
            </div>
          </div>
        ) : null}
      </div>

      {/* タブ切替 */}
      <Tabs value={groupBy} onValueChange={handleTabChange}>
        <TabsList>
          <TabsTrigger value="owner">飼主単位</TabsTrigger>
          <TabsTrigger value="billing">会計単位</TabsTrigger>
        </TabsList>
      </Tabs>

      {/* テーブル */}
      {groupBy === "owner" ? (
        <DataTable headers={["飼主名", "件数", "未納額合計", "最古未納日", "最新未納日", "経過日数"]}>
          {ownerQuery.data?.data.length === 0 ? (
            <DataTableRow><TableCell colSpan={6}>未納者はいません</TableCell></DataTableRow>
          ) : (
            ownerQuery.data?.data.map((row) => (
              <DataTableRow
                key={row.ownerId}
                onClick={() => navigate(paths.owners.detail.getHref(row.ownerId))}
              >
                <TableCell>{row.ownerName}</TableCell>
                <TableCell>{row.billingCount}件</TableCell>
                <TableCell>{formatCurrency(row.totalAmount)}</TableCell>
                <TableCell>{format(new Date(row.oldestDate), "yyyy/MM/dd")}</TableCell>
                <TableCell>{format(new Date(row.latestDate), "yyyy/MM/dd")}</TableCell>
                <TableCell>{row.elapsedDays}日</TableCell>
              </DataTableRow>
            ))
          )}
        </DataTable>
      ) : (
        <DataTable headers={["飼主名", "ペット名", "診療日", "未納額", "経過日数"]}>
          {billingQuery.data?.data.length === 0 ? (
            <DataTableRow><TableCell colSpan={5}>未納者はいません</TableCell></DataTableRow>
          ) : (
            billingQuery.data?.data.map((b) => (
              <DataTableRow
                key={b.id}
                onClick={() => navigate(paths.accounting.detail.getHref(b.id))}
              >
                <TableCell>{b.ownerName}</TableCell>
                <TableCell>{b.petName}</TableCell>
                <TableCell>{format(new Date(b.scheduledDate), "yyyy/MM/dd")}</TableCell>
                <TableCell>{formatCurrency(b.totalAmount)}</TableCell>
                <TableCell>
                  {Math.floor((Date.now() - new Date(b.scheduledDate).getTime()) / 86_400_000)}日
                </TableCell>
              </DataTableRow>
            ))
          )}
        </DataTable>
      )}

      <Pagination page={page} total={active.data?.total ?? 0} limit={limit} onChange={setPage} />
    </PageLayout>
  );
}
```

### 4. ルート登録

```typescript
// frontend/src/app/router.tsx の accounting セクションに追加
{
  path: "unpaid",
  lazy: async () => {
    const { UnpaidCustomerList } = await import("@/features/accounting/routes/UnpaidCustomerList");
    return { Component: UnpaidCustomerList };
  },
},
```

### 5. paths 定義

```typescript
// frontend/src/config/paths.ts の accounting に追加
unpaid: {
  path: "/accounting/unpaid",
  getHref: () => "/accounting/unpaid",
},
```

### 6. メニュー追加

```typescript
// frontend/src/components/shared/Layout/Sidebar の accounting セクションに「未納者一覧」追加
{ label: "未納者一覧", href: paths.accounting.unpaid.getHref() }
```

### 7. Feature index export

```typescript
// frontend/src/features/accounting/index.ts に追加
export { UnpaidCustomerList } from "./routes/UnpaidCustomerList";
export { useGetUnpaidByOwner, useGetUnpaidByBilling } from "./api/get-unpaid-billings";
```

## UI 操作フロー

1. ユーザーがサイドバーから「未納者一覧」をクリック
2. デフォルト: 基準日=今日 / タブ=飼主単位 で表示
3. 売掛金総額カードに「¥1,234,567（45件 / 23名）」表示
4. 飼主単位タブ: 飼主名・件数・未納額・最古日・最新日・経過日数
5. タブを「会計単位」に切替 → URL `?group_by=billing` 同期
6. 基準日を変更 → URL `?as_of=YYYY-MM-DD` 同期、データ再取得
7. 飼主行クリック → `OwnerDetail` 画面遷移
8. 会計行クリック → `AccountingDetail` 画面遷移
9. 未納 0 件: 「未納者はいません」表示

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし（feature 内部は相対パス）
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useTransition` で pending 管理（API 書き込みはなし、Query なので不要）
- [ ] 型は `models.ts` から導出（`Accounting` は既存型を再利用）
- [ ] デザイントークン `C`, `STYLE` 使用（Hex 直指定禁止）
- [ ] `useCallback` でハンドラ安定化
- [ ] 静的配列は外部定数化

## 依存関係

- BE-110 が先に完了している必要がある（API エンドポイント `GET /api/clinics/:id/unpaid-billings`）

## 完了条件

- [ ] 型エラーなし（`docker compose exec frontend pnpm build` パス）
- [ ] ESLint エラーなし（`docker compose exec frontend pnpm lint` パス）
- [ ] AC-1〜AC-13（BUG-370 参照）すべて達成
- [ ] 既存 `AccountingList` 画面に影響なし
- [ ] URL クエリ同期動作確認（リロード・ブックマーク）
