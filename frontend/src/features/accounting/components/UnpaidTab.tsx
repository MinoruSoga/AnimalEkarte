import { useMemo, useState, useCallback } from "react";
import { useNavigate, useSearchParams } from "react-router";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Pagination } from "@/components/shared/Pagination/Pagination";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { C } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import { formatCurrency } from "@/utils/format/number";

import {
  useGetUnpaidByOwner,
  useGetUnpaidByBilling,
  type UnpaidOwner,
} from "../api/get-unpaid-billings";

type GroupBy = "owner" | "billing";

function todayISO(): string {
  const d = new Date();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${d.getFullYear()}-${mm}-${dd}`;
}

function daysSince(isoDate: string, baseDate: string): number {
  const from = new Date(isoDate);
  const to = new Date(baseDate);
  const diff = Math.floor((to.getTime() - from.getTime()) / (1000 * 60 * 60 * 24));
  return Math.max(0, diff);
}

export function UnpaidTab() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  const groupBy: GroupBy = (searchParams.get("group_by") as GroupBy) === "billing" ? "billing" : "owner";
  const baseDate = searchParams.get("reference_date") || todayISO();
  const [page, setPage] = useState(1);
  const limit = 20;

  const handleBaseDateChange = useCallback((next: string) => {
    setSearchParams((prev) => {
      const p = new URLSearchParams(prev);
      p.set("reference_date", next);
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

  const ownerQuery = useGetUnpaidByOwner({ baseDate, groupBy, page, limit });
  const billingQuery = useGetUnpaidByBilling({ baseDate, groupBy, page, limit });

  const summary = ownerQuery.data?.summary;
  const isLoading = groupBy === "owner" ? ownerQuery.isLoading : billingQuery.isLoading;
  const isError = groupBy === "owner" ? ownerQuery.isError : billingQuery.isError;

  const ownerRows = useMemo<UnpaidOwner[]>(() => ownerQuery.data?.data ?? [], [ownerQuery.data]);

  return (
    <div className="flex flex-col gap-4">
      {/* 基準日 + 表示単位 */}
      <div className="flex items-end gap-4">
        <div className="space-y-1.5">
          <Label htmlFor="baseDate" className={`text-sm ${C.text60}`}>基準日</Label>
          <Input
            id="baseDate"
            type="date"
            value={baseDate}
            onChange={(e) => handleBaseDateChange(e.target.value)}
            className="h-9 text-sm"
          />
        </div>
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
        </div>
      </div>

      {/* サマリーカード */}
      {summary ? (
        <div className={`rounded-lg border ${C.borderLight} p-4 bg-white`}>
          <p className={`text-xs ${C.text50} mb-1`}>売掛金総額</p>
          <p className="text-2xl font-bold">{formatCurrency(summary.total_amount)}</p>
          <p className={`text-xs ${C.text60} mt-1`}>
            {summary.billing_count}件 / {summary.owner_count}名
          </p>
        </div>
      ) : null}

      {isLoading ? <LoadingFallback /> : null}
      {isError ? <ErrorFallback message="データの取得に失敗しました" /> : null}

      {!isLoading && !isError && groupBy === "owner" ? (
        ownerRows.length === 0 ? (
          <p className={`text-sm ${C.text50} py-8 text-center`}>未納者はいません</p>
        ) : (
          <div className={`rounded-lg border ${C.borderLight} bg-white overflow-hidden`}>
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
                  <TableRow
                    key={row.owner_id}
                    className="cursor-pointer hover:bg-gray-50"
                    onClick={() => navigate(paths.owners.detail.getHref(String(row.owner_id)))}
                  >
                    <TableCell className="font-medium">{row.owner_name}</TableCell>
                    <TableCell className="text-right">{row.count}</TableCell>
                    <TableCell className="text-right font-mono">
                      {formatCurrency(row.total_amount)}
                    </TableCell>
                    <TableCell>{row.oldest_scheduled}</TableCell>
                    <TableCell>{row.latest_scheduled}</TableCell>
                    <TableCell className="text-right">
                      {daysSince(row.oldest_scheduled, baseDate)}日
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )
      ) : null}

      {!isLoading && !isError && groupBy === "billing" ? (
        (billingQuery.data?.data ?? []).length === 0 ? (
          <p className={`text-sm ${C.text50} py-8 text-center`}>未納会計はありません</p>
        ) : (
          <div className={`rounded-lg border ${C.borderLight} bg-white overflow-hidden`}>
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
                  const total = b.items.reduce((s, i) => s + i.subtotal + i.taxAmount, 0);
                  return (
                    <TableRow
                      key={b.id}
                      className="cursor-pointer hover:bg-gray-50"
                      onClick={() => navigate(paths.accounting.detail.getHref(b.id))}
                    >
                      <TableCell className="font-medium">{b.ownerName}</TableCell>
                      <TableCell>{b.petName}</TableCell>
                      <TableCell>{b.scheduledDate || "-"}</TableCell>
                      <TableCell className="text-right font-mono">
                        {formatCurrency(total)}
                      </TableCell>
                      <TableCell className="text-right">
                        {b.scheduledDate ? `${daysSince(b.scheduledDate, baseDate)}日` : "-"}
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </div>
        )
      ) : null}

      {(() => {
        const data = groupBy === "owner" ? ownerQuery.data : billingQuery.data;
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
            onPrev={() => setPage(Math.max(1, page - 1))}
            onNext={() => setPage(Math.min(totalPages, page + 1))}
          />
        );
      })()}
    </div>
  );
}
