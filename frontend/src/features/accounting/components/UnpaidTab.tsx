import { useMemo, useState, useCallback } from "react";
import { useSearchParams } from "react-router";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Pagination } from "@/components/shared/Pagination/Pagination";
import { DataTableRowLink } from "@/components/shared/DataTable/DataTableRowLink";
import { LoadingFallback, ErrorFallback, EmptyState } from "@/components/shared/DataStates";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { C, STYLE } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import { formatCurrency } from "@/lib/format/number";
import { daysSince, currentJSTYearMonth } from "@/lib/jst-date";

import {
  useGetUnpaidByOwner,
  useGetUnpaidByBilling,
  useGetUnpaidMonthly,
  type UnpaidOwner,
} from "../api/get-unpaid-billings";

type GroupBy = "owner" | "billing" | "monthly";

export function UnpaidTab() {
  const [searchParams, setSearchParams] = useSearchParams();

  const rawGroupBy = searchParams.get("group_by");
  const groupBy: GroupBy = rawGroupBy === "billing" ? "billing" : rawGroupBy === "monthly" ? "monthly" : "owner";

  // #120: start_date/end_date 必須 — 両方揃うまでクエリは発火しない
  const startDate = searchParams.get("start_date") ?? "";
  const endDate = searchParams.get("end_date") ?? "";

  // #114: month param (YYYY-MM), default は JST 当月
  const monthParam = searchParams.get("month") ?? currentJSTYearMonth();
  const monthParts = monthParam.split("-").map(Number);
  const yearNum = monthParts[0] ?? 0;
  const monthNum = monthParts[1] ?? 0;

  const [page, setPage] = useState(1);
  const limit = 20;

  const handleStartDateChange = useCallback((next: string) => {
    setSearchParams((prev) => {
      const p = new URLSearchParams(prev);
      if (next) p.set("start_date", next);
      else p.delete("start_date");
      return p;
    }, { replace: true });
    setPage(1);
  }, [setSearchParams]);

  const handleEndDateChange = useCallback((next: string) => {
    setSearchParams((prev) => {
      const p = new URLSearchParams(prev);
      if (next) p.set("end_date", next);
      else p.delete("end_date");
      return p;
    }, { replace: true });
    setPage(1);
  }, [setSearchParams]);

  const handleGroupByChange = useCallback((next: GroupBy) => {
    setSearchParams((prev) => {
      const p = new URLSearchParams(prev);
      p.set("group_by", next);
      return p;
    }, { replace: true });
    setPage(1);
  }, [setSearchParams]);

  const handleMonthChange = useCallback((next: string) => {
    setSearchParams((prev) => {
      const p = new URLSearchParams(prev);
      if (next) p.set("month", next);
      else p.delete("month");
      return p;
    }, { replace: true });
    setPage(1);
  }, [setSearchParams]);

  // groupBy: "monthly" のとき enabled=false になるよう型を統一
  const ownerQuery = useGetUnpaidByOwner({ startDate, endDate, groupBy, page, limit });
  const billingQuery = useGetUnpaidByBilling({ startDate, endDate, groupBy, page, limit });
  const monthlyQuery = useGetUnpaidMonthly({ year: yearNum, month: monthNum, page, limit });

  const summary = ownerQuery.data?.summary;
  const monthlySummary = monthlyQuery.data?.summary;

  const isLoading = groupBy === "owner" ? ownerQuery.isLoading
    : groupBy === "billing" ? billingQuery.isLoading
    : monthlyQuery.isLoading;
  const isError = groupBy === "owner" ? ownerQuery.isError
    : groupBy === "billing" ? billingQuery.isError
    : monthlyQuery.isError;

  const ownerRows = useMemo<UnpaidOwner[]>(() => ownerQuery.data?.data ?? [], [ownerQuery.data]);
  const monthlyRows = useMemo(() => monthlyQuery.data?.data ?? [], [monthlyQuery.data]);

  return (
    <div className="flex flex-col gap-4">
      {/* 期間絞り込み + 表示単位切り替え */}
      <div className="flex items-end gap-4 flex-wrap">
        {groupBy !== "monthly" ? (
          <>
            <div className="space-y-1.5">
              <Label htmlFor="startDate" className={`text-sm ${C.text60}`}>開始日</Label>
              <Input
                id="startDate"
                type="date"
                value={startDate}
                onChange={(e) => handleStartDateChange(e.target.value)}
                className="h-9 text-sm"
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="endDate" className={`text-sm ${C.text60}`}>終了日</Label>
              <Input
                id="endDate"
                type="date"
                value={endDate}
                onChange={(e) => handleEndDateChange(e.target.value)}
                className="h-9 text-sm"
              />
            </div>
          </>
        ) : (
          <div className="space-y-1.5">
            <Label htmlFor="monthPicker" className={`text-sm ${C.text60}`}>対象月</Label>
            <Input
              id="monthPicker"
              type="month"
              value={monthParam}
              onChange={(e) => handleMonthChange(e.target.value)}
              className="h-9 text-sm"
            />
          </div>
        )}
        <div className="flex gap-2">
          <Button
            type="button"
            variant={groupBy === "owner" ? "default" : "outline"}
            size="sm"
            onClick={() => handleGroupByChange("owner")}
          >
            飼主単位
          </Button>
          <Button
            type="button"
            variant={groupBy === "billing" ? "default" : "outline"}
            size="sm"
            onClick={() => handleGroupByChange("billing")}
          >
            会計単位
          </Button>
          <Button
            type="button"
            variant={groupBy === "monthly" ? "default" : "outline"}
            size="sm"
            onClick={() => handleGroupByChange("monthly")}
          >
            月次繰越
          </Button>
        </div>
      </div>

      {/* 売掛金サマリーカード (owner/billing モード) */}
      {groupBy !== "monthly" && summary ? (
        <div className={`rounded-lg border ${C.borderLight} p-4 ${C.bgWhite}`}>
          <p className={`text-xs ${C.text50} mb-1`}>売掛金総額</p>
          <p className="text-heading-3 font-bold">{formatCurrency(summary.total_amount)}</p>
          <p className={`text-xs ${C.text60} mt-1`}>
            {summary.billing_count}件 / {summary.owner_count}名
          </p>
        </div>
      ) : null}

      {/* 月次繰越サマリーカード */}
      {groupBy === "monthly" && monthlySummary ? (
        <div className={`rounded-lg border ${C.borderLight} p-4 ${C.bgWhite}`}>
          <div className="grid grid-cols-3 gap-4">
            <div>
              <p className={`text-xs ${C.text50} mb-1`}>前月繰越</p>
              <p className="text-xl font-bold">{formatCurrency(monthlySummary.prev_month_carryover)}</p>
            </div>
            <div>
              <p className={`text-xs ${C.text50} mb-1`}>当月未払い</p>
              <p className="text-xl font-bold">{formatCurrency(monthlySummary.current_month_unpaid)}</p>
            </div>
            <div>
              <p className={`text-xs ${C.text50} mb-1`}>次月繰越</p>
              <p className="text-xl font-bold">{formatCurrency(monthlySummary.next_month_carryover)}</p>
            </div>
          </div>
        </div>
      ) : null}

      {isLoading ? <LoadingFallback /> : null}
      {isError ? <ErrorFallback message="データの取得に失敗しました" /> : null}

      {/* 飼主単位テーブル */}
      {!isLoading && !isError && groupBy === "owner" ? (
        ownerRows.length === 0 ? (
          <p className={`text-sm ${C.text50} py-8 text-center`}>未納者はいません</p>
        ) : (
          <div className={`rounded-lg border ${C.borderLight} ${C.bgWhite} overflow-hidden`}>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>飼主名</TableHead>
                  <TableHead className="text-right">件数</TableHead>
                  <TableHead className="text-right">未納額合計</TableHead>
                  <TableHead>最古未納日</TableHead>
                  <TableHead>最新未納日</TableHead>
                  <TableHead className="text-right">経過日数</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {ownerRows.map((row) => (
                  <TableRow key={row.owner_id} className={STYLE.tableRowHover}>
                    <TableCell className="font-medium">
                      <DataTableRowLink
                        to={paths.owners.detail.getHref(String(row.owner_id))}
                        aria-label={`飼主詳細: ${row.owner_name} (ID ${row.owner_id})`}
                      >
                        {row.owner_name}
                      </DataTableRowLink>
                    </TableCell>
                    <TableCell className="text-right">{row.count}</TableCell>
                    <TableCell className="text-right font-mono">
                      {formatCurrency(row.total_amount)}
                    </TableCell>
                    <TableCell>{row.oldest_scheduled}</TableCell>
                    <TableCell>{row.latest_scheduled}</TableCell>
                    <TableCell className="text-right">
                      {daysSince(row.oldest_scheduled, endDate)}日
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )
      ) : null}

      {/* 会計単位テーブル */}
      {!isLoading && !isError && groupBy === "billing" ? (
        (billingQuery.data?.data ?? []).length === 0 ? (
          <EmptyState message="未納会計はありません" />
        ) : (
          <div className={`rounded-lg border ${C.borderLight} ${C.bgWhite} overflow-hidden`}>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>飼主名</TableHead>
                  <TableHead>ペット名</TableHead>
                  <TableHead>診療日</TableHead>
                  <TableHead className="text-right">未納額</TableHead>
                  <TableHead className="text-right">経過日数</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {(billingQuery.data?.data ?? []).map((b) => {
                  // BUG-007: outstanding_amount を優先（クレジット訂正差額）。未設定時は明細合計へフォールバック。
                  const unpaidAmount =
                    (b.outstandingAmount ?? 0) > 0
                      ? (b.outstandingAmount as number)
                      : b.items.reduce((s: number, i: { subtotal: number; taxAmount: number }) => s + i.subtotal + i.taxAmount, 0);
                  return (
                    <TableRow key={b.id} className={STYLE.tableRowHover}>
                      <TableCell className="font-medium">
                        <DataTableRowLink
                          to={paths.accounting.detail.getHref(b.id)}
                          aria-label={`会計詳細: ${b.ownerName} / ${b.petName} (ID ${b.id})`}
                        >
                          {b.ownerName}
                        </DataTableRowLink>
                      </TableCell>
                      <TableCell>{b.petName}</TableCell>
                      <TableCell>{b.scheduledDate || "-"}</TableCell>
                      <TableCell className="text-right font-mono">
                        {formatCurrency(unpaidAmount)}
                      </TableCell>
                      <TableCell className="text-right">
                        {b.scheduledDate ? `${daysSince(b.scheduledDate, endDate)}日` : "-"}
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </div>
        )
      ) : null}

      {/* 月次繰越テーブル */}
      {!isLoading && !isError && groupBy === "monthly" ? (
        monthlyRows.length === 0 ? (
          <EmptyState message="対象月の未納データがありません" />
        ) : (
          <div className={`rounded-lg border ${C.borderLight} ${C.bgWhite} overflow-hidden`}>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>飼主名</TableHead>
                  <TableHead>ペット名</TableHead>
                  <TableHead className="text-right">前月繰越</TableHead>
                  <TableHead className="text-right">当月未払い</TableHead>
                  <TableHead className="text-right">次月繰越</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {monthlyRows.map((row) => (
                  <TableRow key={`${row.owner_id}-${row.pet_id ?? "none"}`} className={STYLE.tableRowHover}>
                    <TableCell className="font-medium">
                      <DataTableRowLink
                        to={paths.owners.detail.getHref(String(row.owner_id))}
                        aria-label={`飼主詳細: ${row.owner_name} (ID ${row.owner_id})`}
                      >
                        {row.owner_name}
                      </DataTableRowLink>
                    </TableCell>
                    <TableCell>{row.pet_name || "-"}</TableCell>
                    <TableCell className="text-right font-mono">
                      {formatCurrency(row.prev_month_carryover)}
                    </TableCell>
                    <TableCell className="text-right font-mono">
                      {formatCurrency(row.current_month_unpaid)}
                    </TableCell>
                    <TableCell className="text-right font-mono">
                      {formatCurrency(row.next_month_carryover)}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )
      ) : null}

      {/* ページネーション */}
      {(() => {
        const data = groupBy === "owner" ? ownerQuery.data
          : groupBy === "billing" ? billingQuery.data
          : monthlyQuery.data;
        if (!data || data.total <= limit) return null;
        const total = data.total;
        const totalPages = Math.max(1, Math.ceil(total / limit));
        const startIndex = (page - 1) * limit + 1;
        const endIndex = Math.min(total, page * limit);
        return (
          <Pagination
            currentPage={page}
            totalPages={totalPages}
            totalCount={total}
            startIndex={startIndex}
            endIndex={endIndex}
            onPageChange={setPage}
            onPrev={() => setPage((currentPage) => Math.max(1, currentPage - 1))}
            onNext={() => setPage((currentPage) => Math.min(totalPages, currentPage + 1))}
          />
        );
      })()}
    </div>
  );
}
