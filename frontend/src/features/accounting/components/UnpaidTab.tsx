import { useMemo, useState, useCallback } from "react";
import { useSearchParams } from "react-router";

import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { currentJSTYearMonth, currentJSTMonthDateRange } from "@/lib/jst-date";

import {
  useGetUnpaidByOwner,
  useGetUnpaidByBilling,
  useGetUnpaidMonthly,
  type UnpaidOwner,
} from "../api/get-unpaid-billings";
import { UnpaidTabFilters } from "./UnpaidTabFilters";
import { parseUnpaidGroupBy, type UnpaidGroupBy } from "../lib/unpaid-tab-model";
import { UnpaidTabSummaries } from "./UnpaidTabSummaries";
import {
  UnpaidBillingTable,
  UnpaidMonthlyTable,
  UnpaidOwnerTable,
  UnpaidTabPagination,
} from "./UnpaidTabTables";

export function UnpaidTab() {
  const [searchParams, setSearchParams] = useSearchParams();

  const groupBy: UnpaidGroupBy = parseUnpaidGroupBy(searchParams.get("group_by"));

  // 未指定時は当月を既定にする。空のまま API を発火しないと「未納者はいません」と誤表示する (BUG-002)。
  const monthRange = currentJSTMonthDateRange();
  const startDate = searchParams.get("start_date") ?? monthRange.start;
  const endDate = searchParams.get("end_date") ?? monthRange.end;

  // #114: month param (YYYY-MM), default は JST 当月
  const monthParam = searchParams.get("month") ?? currentJSTYearMonth();
  const monthParts = monthParam.split("-").map(Number);
  const yearNum = monthParts[0] ?? 0;
  const monthNum = monthParts[1] ?? 0;

  const [page, setPage] = useState(1);
  const limit = 20;

  const handleStartDateChange = useCallback(
    (next: string) => {
      setSearchParams(
        (prev) => {
          const p = new URLSearchParams(prev);
          if (next) p.set("start_date", next);
          else p.delete("start_date");
          return p;
        },
        { replace: true },
      );
      setPage(1);
    },
    [setSearchParams],
  );

  const handleEndDateChange = useCallback(
    (next: string) => {
      setSearchParams(
        (prev) => {
          const p = new URLSearchParams(prev);
          if (next) p.set("end_date", next);
          else p.delete("end_date");
          return p;
        },
        { replace: true },
      );
      setPage(1);
    },
    [setSearchParams],
  );

  const handleGroupByChange = useCallback(
    (next: UnpaidGroupBy) => {
      setSearchParams(
        (prev) => {
          const p = new URLSearchParams(prev);
          p.set("group_by", next);
          return p;
        },
        { replace: true },
      );
      setPage(1);
    },
    [setSearchParams],
  );

  const handleMonthChange = useCallback(
    (next: string) => {
      setSearchParams(
        (prev) => {
          const p = new URLSearchParams(prev);
          if (next) p.set("month", next);
          else p.delete("month");
          return p;
        },
        { replace: true },
      );
      setPage(1);
    },
    [setSearchParams],
  );

  // groupBy: "monthly" のとき enabled=false になるよう型を統一
  const ownerQuery = useGetUnpaidByOwner({ startDate, endDate, groupBy, page, limit });
  const billingQuery = useGetUnpaidByBilling({ startDate, endDate, groupBy, page, limit });
  const monthlyQuery = useGetUnpaidMonthly({ year: yearNum, month: monthNum, page, limit });

  const summary = ownerQuery.data?.summary;
  const monthlySummary = monthlyQuery.data?.summary;

  const isLoading =
    groupBy === "owner"
      ? ownerQuery.isLoading
      : groupBy === "billing"
        ? billingQuery.isLoading
        : monthlyQuery.isLoading;
  const isError =
    groupBy === "owner"
      ? ownerQuery.isError
      : groupBy === "billing"
        ? billingQuery.isError
        : monthlyQuery.isError;

  const ownerRows = useMemo<UnpaidOwner[]>(() => ownerQuery.data?.data ?? [], [ownerQuery.data]);
  const monthlyRows = useMemo(() => monthlyQuery.data?.data ?? [], [monthlyQuery.data]);
  const billingRows = billingQuery.data?.data ?? [];
  const listData =
    groupBy === "owner"
      ? ownerQuery.data
      : groupBy === "billing"
        ? billingQuery.data
        : monthlyQuery.data;

  return (
    <div className="flex flex-col gap-4">
      <UnpaidTabFilters
        groupBy={groupBy}
        startDate={startDate}
        endDate={endDate}
        monthParam={monthParam}
        onStartDateChange={handleStartDateChange}
        onEndDateChange={handleEndDateChange}
        onMonthChange={handleMonthChange}
        onGroupByChange={handleGroupByChange}
      />

      <UnpaidTabSummaries groupBy={groupBy} summary={summary} monthlySummary={monthlySummary} />

      {isLoading ? <LoadingFallback /> : null}
      {isError ? <ErrorFallback message="データの取得に失敗しました" /> : null}

      {!isLoading && !isError && groupBy === "owner" ? (
        <UnpaidOwnerTable rows={ownerRows} endDate={endDate} />
      ) : null}

      {!isLoading && !isError && groupBy === "billing" ? (
        <UnpaidBillingTable billings={billingRows} endDate={endDate} />
      ) : null}

      {!isLoading && !isError && groupBy === "monthly" ? (
        <UnpaidMonthlyTable rows={monthlyRows} />
      ) : null}

      <UnpaidTabPagination
        total={listData?.total ?? 0}
        page={page}
        limit={limit}
        onPageChange={setPage}
      />
    </div>
  );
}
