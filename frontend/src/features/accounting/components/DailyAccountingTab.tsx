import { useMemo, useCallback } from "react";
import { useSearchParams } from "react-router";
import { Printer } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Table,
  TableBody,
  TableCell,
  TableFooter,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { LoadingFallback, ErrorFallback, EmptyState } from "@/components/shared/DataStates";
import { C, ICON } from "@/lib/design-tokens";
import { PAYMENT_METHOD_LABELS } from "@/constants/payment-method";
import { todayJSTISO } from "@/lib/jst-date";
import { formatCurrency } from "@/lib/format/number";

import { useGetAccountings } from "../api/get-accountings";
import { useGetDailySummary } from "../api/get-daily-summary";
import type { ClinicDailySummaryItem, DailySummary, PerClinicDailySummaryResponse } from "../api/get-daily-summary";
import type { PaymentMethod } from "../types";

import { DailyPrintArea } from "./DailyAccountingPrintArea";
import { SummaryCard } from "./DailyAccountingTabParts";
import {
  formatReceiptNo,
  getCategoryBreakdown,
  apportionPayment,
  getRowCashTotal,
} from "./daily-accounting-utils";
import type { RowData, TotalsData } from "./daily-accounting-utils";

const EMPTY_CLINIC_NAME_MAP = new Map<string, string>();

function isSingleSummary(data: DailySummary | PerClinicDailySummaryResponse): data is DailySummary {
  return "billing_count" in data;
}

// ── メインコンポーネント ──────────────────────────────────────────
interface DailyAccountingTabProps {
  /** 拠点横断表示 (#86 段階3): 2件以上の場合に拠点別集計を表示する。 */
  selectedClinicIds?: string[];
  /** clinicId → 医院名（所属医院由来） */
  clinicNameById?: Map<string, string>;
}

export function DailyAccountingTab({
  selectedClinicIds,
  clinicNameById = EMPTY_CLINIC_NAME_MAP,
}: DailyAccountingTabProps) {
  const [searchParams, setSearchParams] = useSearchParams();
  const isMultiClinic = selectedClinicIds !== undefined && selectedClinicIds.length > 1;
  const selectedDate = searchParams.get("daily_date") ?? todayJSTISO();

  const handleDateChange = useCallback((next: string) => {
    setSearchParams((prev) => {
      const p = new URLSearchParams(prev);
      p.set("daily_date", next);
      return p;
    }, { replace: true });
  }, [setSearchParams]);

  const { data: accountings = [], isLoading, isError } = useGetAccountings({
    startDate: selectedDate,
    endDate: selectedDate,
    clinicIds: selectedClinicIds,
  });

  const { data: summaryRaw } = useGetDailySummary(selectedDate, selectedClinicIds);
  const summary = summaryRaw && isSingleSummary(summaryRaw) ? summaryRaw : null;
  const perClinicSummaries: ClinicDailySummaryItem[] =
    summaryRaw && !isSingleSummary(summaryRaw) ? summaryRaw.per_clinic : [];

  const rows = useMemo<RowData[]>(() => {
    return accountings
      .filter((a) => a.status === "completed")
      .map((a) => {
        const breakdown = getCategoryBreakdown(a.items);
        const cashTotal = getRowCashTotal(a);
        const detailedBreakdown = apportionPayment(breakdown, cashTotal);
        const subtotal =
          a.payment?.subtotal ??
          a.items.reduce((s, i) => s + i.subtotal, 0);
        const tax =
          a.payment?.taxTotal ??
          a.items.reduce((s, i) => s + i.taxAmount, 0);
        const discount = Math.abs(a.payment?.discountAmount ?? 0);
        const total = a.payment?.totalAmount ?? subtotal + tax;
        return { accounting: a, breakdown, detailedBreakdown, subtotal, tax, discount, total };
      });
  }, [accountings]);

  const totals = useMemo<TotalsData>(() => {
    return rows.reduce(
      (acc, r) => ({
        medical: acc.medical + r.breakdown.medical,
        surgery: acc.surgery + r.breakdown.surgery,
        rv: acc.rv + r.breakdown.rv,
        food: acc.food + r.breakdown.food,
        trimming: acc.trimming + r.breakdown.trimming,
        hotel: acc.hotel + r.breakdown.hotel,
        goods: acc.goods + r.breakdown.goods,
        subtotal: acc.subtotal + r.subtotal,
        tax: acc.tax + r.tax,
        discount: acc.discount + r.discount,
        total: acc.total + r.total,
      }),
      { medical: 0, surgery: 0, rv: 0, food: 0, trimming: 0, hotel: 0, goods: 0, subtotal: 0, tax: 0, discount: 0, total: 0 },
    );
  }, [rows]);

  // isMultiClinic 時の colSpan: 領収No + 拠点 + 飼主 + ペット = 4
  // single clinic 時の colSpan: 領収No + 飼主 + ペット = 3
  const labelColSpan = isMultiClinic ? 4 : 3;

  return (
    <>
      <div className="flex flex-col gap-4">
        {/* 日付選択 + 印刷ボタン + 集計カード */}
        <div className="flex flex-wrap items-end gap-4">
          <div className="space-y-1.5">
            <Label htmlFor="daily-date" className={`text-sm ${C.text60}`}>対象日</Label>
            <Input
              id="daily-date"
              type="date"
              value={selectedDate}
              onChange={(e) => handleDateChange(e.target.value)}
              className="h-9 text-sm"
            />
          </div>

          {rows.length > 0 ? (
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => window.print()}
              data-testid="daily-print-button"
            >
              <Printer className={`mr-2 ${ICON.action}`} />
              印刷
            </Button>
          ) : null}

          {summary ? (
            <div className="flex flex-wrap gap-2" data-testid="daily-summary-cards">
              <SummaryCard label="会計件数" value={`${summary.billing_count}件`} />
              <SummaryCard label="売上合計" value={formatCurrency(summary.grand_total)} />
              {summary.payment_totals.map((pt) => (
                <SummaryCard
                  key={pt.method}
                  label={PAYMENT_METHOD_LABELS[pt.method as PaymentMethod] ?? pt.method}
                  value={formatCurrency(pt.total)}
                />
              ))}
            </div>
          ) : null}
          {perClinicSummaries.length > 0 ? (
            <div className="flex flex-col gap-2" data-testid="daily-summary-per-clinic">
              {perClinicSummaries.map((cs) => (
                <div key={cs.clinic_id} className={`rounded-lg border ${C.borderLight} px-3 py-2 ${C.bgWhite}`}>
                  <p className={`text-xs font-medium ${C.text60} mb-1.5`}>
                    {clinicNameById.get(String(cs.clinic_id)) ?? `拠点 ${cs.clinic_id}`}
                  </p>
                  <div className="flex flex-wrap gap-2">
                    <SummaryCard label="会計件数" value={`${cs.summary.billing_count}件`} />
                    <SummaryCard label="売上合計" value={formatCurrency(cs.summary.grand_total)} />
                    {cs.summary.payment_totals.map((pt) => (
                      <SummaryCard
                        key={pt.method}
                        label={PAYMENT_METHOD_LABELS[pt.method as PaymentMethod] ?? pt.method}
                        value={formatCurrency(pt.total)}
                      />
                    ))}
                  </div>
                </div>
              ))}
            </div>
          ) : null}
        </div>

        {isLoading ? <LoadingFallback /> : null}
        {isError ? <ErrorFallback message="データの取得に失敗しました" /> : null}

        {!isLoading && !isError ? (
          rows.length === 0 ? (
            <EmptyState data-testid="daily-empty" message="当日の会計データがありません" />
          ) : (
            <div
              className={`overflow-x-auto rounded-lg border ${C.borderLight} ${C.bgWhite}`}
              data-testid="daily-accounting-table"
            >
              <Table>
                <TableHeader>
                  <TableRow className={`${C.bgPage30}`}>
                    <TableHead className={`${C.text60} whitespace-nowrap w-[90px]`}>領収No</TableHead>
                    {isMultiClinic ? (
                      <TableHead className={`${C.text60} whitespace-nowrap w-[100px]`}>拠点</TableHead>
                    ) : null}
                    <TableHead className={`${C.text60} whitespace-nowrap`}>飼主名</TableHead>
                    <TableHead className={`${C.text60} whitespace-nowrap`}>ペット名</TableHead>
                    <TableHead className={`text-right ${C.text60} whitespace-nowrap`}>診療</TableHead>
                    <TableHead className={`text-right ${C.text60} whitespace-nowrap`}>外科</TableHead>
                    <TableHead className={`text-right ${C.text60} whitespace-nowrap`}>RV</TableHead>
                    <TableHead className={`text-right ${C.text60} whitespace-nowrap`}>フード</TableHead>
                    <TableHead className={`text-right ${C.text60} whitespace-nowrap`}>トリミング</TableHead>
                    <TableHead className={`text-right ${C.text60} whitespace-nowrap`}>ホテル</TableHead>
                    <TableHead className={`text-right ${C.text60} whitespace-nowrap`}>用品他</TableHead>
                    <TableHead className={`text-right ${C.text60} whitespace-nowrap`}>売上合計</TableHead>
                    <TableHead className={`text-center ${C.text60} whitespace-nowrap`}>支払方法</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {rows.map(({ accounting: a, breakdown, total }) => (
                    <TableRow key={a.id} className={`border-b ${C.borderLight}`}>
                      <TableCell className={`font-mono ${C.text60}`}>
                        {formatReceiptNo(a.id)}
                      </TableCell>
                      {isMultiClinic ? (
                        <TableCell className={`text-sm ${C.text60} whitespace-nowrap`}>
                          {clinicNameById.get(a.clinicId) ?? a.clinicId}
                        </TableCell>
                      ) : null}
                      <TableCell className={`text-sm ${C.text} font-medium whitespace-nowrap`}>
                        {a.ownerName}
                      </TableCell>
                      <TableCell className={`text-sm ${C.text} whitespace-nowrap`}>
                        {a.petName}
                      </TableCell>
                      <TableCell className={`text-right text-sm font-mono ${C.text60}`}>
                        {breakdown.medical > 0 ? formatCurrency(breakdown.medical) : "-"}
                      </TableCell>
                      <TableCell className={`text-right text-sm font-mono ${C.text60}`}>
                        {breakdown.surgery > 0 ? formatCurrency(breakdown.surgery) : "-"}
                      </TableCell>
                      <TableCell className={`text-right text-sm font-mono ${C.text60}`}>
                        {breakdown.rv > 0 ? formatCurrency(breakdown.rv) : "-"}
                      </TableCell>
                      <TableCell className={`text-right text-sm font-mono ${C.text60}`}>
                        {breakdown.food > 0 ? formatCurrency(breakdown.food) : "-"}
                      </TableCell>
                      <TableCell className={`text-right text-sm font-mono ${C.text60}`}>
                        {breakdown.trimming > 0 ? formatCurrency(breakdown.trimming) : "-"}
                      </TableCell>
                      <TableCell className={`text-right text-sm font-mono ${C.text60}`}>
                        {breakdown.hotel > 0 ? formatCurrency(breakdown.hotel) : "-"}
                      </TableCell>
                      <TableCell className={`text-right text-sm font-mono ${C.text60}`}>
                        {breakdown.goods > 0 ? formatCurrency(breakdown.goods) : "-"}
                      </TableCell>
                      <TableCell className={`text-right text-sm font-mono font-semibold ${C.text}`}>
                        {formatCurrency(total)}
                      </TableCell>
                      <TableCell className={`text-center text-sm whitespace-nowrap ${C.text60}`}>
                        {a.paymentSplits && a.paymentSplits.length > 1
                          ? a.paymentSplits.map((s) => PAYMENT_METHOD_LABELS[s.method] ?? s.method).join(" / ")
                          : a.payment
                            ? (PAYMENT_METHOD_LABELS[a.payment.method] ?? a.payment.method)
                            : "-"}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
                <TableFooter>
                  <TableRow className={`font-bold border-t-2 ${C.borderLight}`}>
                    <TableCell colSpan={labelColSpan} className="text-sm">合計（{rows.length}件）</TableCell>
                    <TableCell className="text-right text-sm font-mono">
                      {totals.medical > 0 ? formatCurrency(totals.medical) : "-"}
                    </TableCell>
                    <TableCell className="text-right text-sm font-mono">
                      {totals.surgery > 0 ? formatCurrency(totals.surgery) : "-"}
                    </TableCell>
                    <TableCell className="text-right text-sm font-mono">
                      {totals.rv > 0 ? formatCurrency(totals.rv) : "-"}
                    </TableCell>
                    <TableCell className="text-right text-sm font-mono">
                      {totals.food > 0 ? formatCurrency(totals.food) : "-"}
                    </TableCell>
                    <TableCell className="text-right text-sm font-mono">
                      {totals.trimming > 0 ? formatCurrency(totals.trimming) : "-"}
                    </TableCell>
                    <TableCell className="text-right text-sm font-mono">
                      {totals.hotel > 0 ? formatCurrency(totals.hotel) : "-"}
                    </TableCell>
                    <TableCell className="text-right text-sm font-mono">
                      {totals.goods > 0 ? formatCurrency(totals.goods) : "-"}
                    </TableCell>
                    <TableCell className="text-right text-sm font-mono">
                      <span className="font-bold">{formatCurrency(totals.total)}</span>
                    </TableCell>
                    <TableCell />
                  </TableRow>
                </TableFooter>
              </Table>
            </div>
          )
        ) : null}
      </div>

      {/* 印刷エリア: print時のみ表示 */}
      {rows.length > 0 ? (
        <DailyPrintArea date={selectedDate} rows={rows} totals={totals} />
      ) : null}
    </>
  );
}
