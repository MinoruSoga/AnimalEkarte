import { useMemo, useCallback } from "react";
import { useSearchParams } from "react-router";

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
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { C } from "@/lib/design-tokens";
import { formatCurrency } from "@/utils/format/number";

import { useGetAccountings } from "../api/get-accountings";
import { useGetDailySummary } from "../api/get-daily-summary";
import type { ClinicDailySummaryItem, DailySummary, PerClinicDailySummaryResponse } from "../api/get-daily-summary";
import type { Accounting, AccountingItem } from "../types";

// ── 定数 ────────────────────────────────────────────────────────
const PAYMENT_METHOD_LABELS: Record<string, string> = {
  cash: "現金",
  credit_card: "カード",
  electronic_money: "電子マネー",
};

const MEDICAL_CATEGORIES = new Set(["examination", "test", "procedure", "medicine"]);
const EMPTY_CLINIC_NAME_MAP = new Map<string, string>();

// ── ヘルパー ─────────────────────────────────────────────────────
function todayISO(): string {
  const jst = new Date(Date.now() + 9 * 60 * 60 * 1000);
  return jst.toISOString().slice(0, 10);
}

interface CategoryBreakdown {
  medical: number;
  surgery: number;
  food: number;
  goods: number;
}

function getCategoryBreakdown(items: AccountingItem[]): CategoryBreakdown {
  let medical = 0;
  let surgery = 0;
  let food = 0;
  let goods = 0;
  for (const item of items) {
    const sub = item.subtotal;
    if (MEDICAL_CATEGORIES.has(item.category)) {
      medical += sub;
    } else if (item.category === "surgery") {
      surgery += sub;
    } else if (item.category === "food") {
      food += sub;
    } else {
      goods += sub;
    }
  }
  return { medical, surgery, food, goods };
}

interface RowData {
  accounting: Accounting;
  breakdown: CategoryBreakdown;
  subtotal: number;
  tax: number;
  discount: number;
  total: number;
}

interface TotalsData {
  medical: number;
  surgery: number;
  food: number;
  goods: number;
  subtotal: number;
  tax: number;
  discount: number;
  total: number;
}

// ── サマリーカード ────────────────────────────────────────────────
function SummaryCard({ label, value }: { label: string; value: string }) {
  return (
    <div className={`rounded-lg border ${C.borderLight} px-4 py-3 ${C.bgWhite} min-w-[100px]`}>
      <p className={`text-xs ${C.text50} mb-0.5`}>{label}</p>
      <p className={`text-base font-semibold font-mono ${C.text}`}>{value}</p>
    </div>
  );
}

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
  const selectedDate = searchParams.get("daily_date") ?? todayISO();

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
        const subtotal =
          a.payment?.subtotal ??
          a.items.reduce((s, i) => s + i.subtotal, 0);
        const tax =
          a.payment?.taxTotal ??
          a.items.reduce((s, i) => s + i.taxAmount, 0);
        const discount = Math.abs(a.payment?.discountAmount ?? 0);
        const total = a.payment?.totalAmount ?? subtotal + tax;
        return { accounting: a, breakdown, subtotal, tax, discount, total };
      });
  }, [accountings]);

  const totals = useMemo<TotalsData>(() => {
    return rows.reduce(
      (acc, r) => ({
        medical: acc.medical + r.breakdown.medical,
        surgery: acc.surgery + r.breakdown.surgery,
        food: acc.food + r.breakdown.food,
        goods: acc.goods + r.breakdown.goods,
        subtotal: acc.subtotal + r.subtotal,
        tax: acc.tax + r.tax,
        discount: acc.discount + r.discount,
        total: acc.total + r.total,
      }),
      { medical: 0, surgery: 0, food: 0, goods: 0, subtotal: 0, tax: 0, discount: 0, total: 0 },
    );
  }, [rows]);

  return (
    <div className="flex flex-col gap-4">
      {/* 日付選択 + 集計カード */}
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

        {summary ? (
          <div className="flex flex-wrap gap-2" data-testid="daily-summary-cards">
            <SummaryCard label="会計件数" value={`${summary.billing_count}件`} />
            <SummaryCard label="売上合計" value={formatCurrency(summary.grand_total)} />
            {summary.payment_totals.map((pt) => (
              <SummaryCard
                key={pt.method}
                label={PAYMENT_METHOD_LABELS[pt.method] ?? pt.method}
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
                      label={PAYMENT_METHOD_LABELS[pt.method] ?? pt.method}
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
          <p className={`text-sm ${C.text50} py-8 text-center`} data-testid="daily-empty">
            当日の会計データがありません
          </p>
        ) : (
          <div
            className={`overflow-x-auto rounded-lg border ${C.borderLight} ${C.bgWhite}`}
            data-testid="daily-accounting-table"
          >
            <Table>
              <TableHeader>
                <TableRow className={`${C.bgPage30}`}>
                  {isMultiClinic ? (
                    <TableHead className={`text-xs ${C.text60} whitespace-nowrap w-[100px]`}>拠点</TableHead>
                  ) : null}
                  <TableHead className={`text-xs ${C.text60} whitespace-nowrap`}>飼主名</TableHead>
                  <TableHead className={`text-xs ${C.text60} whitespace-nowrap`}>ペット名</TableHead>
                  <TableHead className={`text-right text-xs ${C.text60} whitespace-nowrap`}>診療</TableHead>
                  <TableHead className={`text-right text-xs ${C.text60} whitespace-nowrap`}>外科</TableHead>
                  <TableHead className={`text-right text-xs ${C.text60} whitespace-nowrap`}>フード</TableHead>
                  <TableHead className={`text-right text-xs ${C.text60} whitespace-nowrap`}>用品/他</TableHead>
                  <TableHead className={`text-right text-xs ${C.text60} whitespace-nowrap`}>純売上</TableHead>
                  <TableHead className={`text-right text-xs ${C.text60} whitespace-nowrap`}>消費税</TableHead>
                  <TableHead className={`text-right text-xs ${C.text60} whitespace-nowrap`}>値引金</TableHead>
                  <TableHead className={`text-right text-xs ${C.text60} whitespace-nowrap`}>売上合計</TableHead>
                  <TableHead className={`text-center text-xs ${C.text60} whitespace-nowrap`}>支払方法</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {rows.map(({ accounting: a, breakdown, subtotal, tax, discount, total }) => (
                  <TableRow key={a.id} className={`border-b ${C.borderLight}`}>
                    {isMultiClinic ? (
                      <TableCell className={`text-sm ${C.text60} py-2 whitespace-nowrap`}>
                        {clinicNameById.get(a.clinicId) ?? a.clinicId}
                      </TableCell>
                    ) : null}
                    <TableCell className={`text-sm ${C.text} py-2 font-medium whitespace-nowrap`}>
                      {a.ownerName}
                    </TableCell>
                    <TableCell className={`text-sm ${C.text} py-2 whitespace-nowrap`}>
                      {a.petName}
                    </TableCell>
                    <TableCell className={`text-right text-sm font-mono py-2 ${C.text60}`}>
                      {breakdown.medical > 0 ? formatCurrency(breakdown.medical) : "-"}
                    </TableCell>
                    <TableCell className={`text-right text-sm font-mono py-2 ${C.text60}`}>
                      {breakdown.surgery > 0 ? formatCurrency(breakdown.surgery) : "-"}
                    </TableCell>
                    <TableCell className={`text-right text-sm font-mono py-2 ${C.text60}`}>
                      {breakdown.food > 0 ? formatCurrency(breakdown.food) : "-"}
                    </TableCell>
                    <TableCell className={`text-right text-sm font-mono py-2 ${C.text60}`}>
                      {breakdown.goods > 0 ? formatCurrency(breakdown.goods) : "-"}
                    </TableCell>
                    <TableCell className={`text-right text-sm font-mono py-2 ${C.text}`}>
                      {formatCurrency(subtotal)}
                    </TableCell>
                    <TableCell className={`text-right text-sm font-mono py-2 ${C.text60}`}>
                      {formatCurrency(tax)}
                    </TableCell>
                    <TableCell className={`text-right text-sm font-mono py-2 ${C.text60}`}>
                      {discount > 0 ? formatCurrency(discount) : "-"}
                    </TableCell>
                    <TableCell className={`text-right text-sm font-mono font-semibold py-2 ${C.text}`}>
                      {formatCurrency(total)}
                    </TableCell>
                    <TableCell className={`text-center text-sm py-2 whitespace-nowrap ${C.text60}`}>
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
                  <TableCell colSpan={isMultiClinic ? 3 : 2} className="py-2 text-sm">合計（{rows.length}件）</TableCell>
                  <TableCell className="text-right text-sm font-mono py-2">
                    {totals.medical > 0 ? formatCurrency(totals.medical) : "-"}
                  </TableCell>
                  <TableCell className="text-right text-sm font-mono py-2">
                    {totals.surgery > 0 ? formatCurrency(totals.surgery) : "-"}
                  </TableCell>
                  <TableCell className="text-right text-sm font-mono py-2">
                    {totals.food > 0 ? formatCurrency(totals.food) : "-"}
                  </TableCell>
                  <TableCell className="text-right text-sm font-mono py-2">
                    {totals.goods > 0 ? formatCurrency(totals.goods) : "-"}
                  </TableCell>
                  <TableCell className="text-right text-sm font-mono py-2">
                    {formatCurrency(totals.subtotal)}
                  </TableCell>
                  <TableCell className="text-right text-sm font-mono py-2">
                    {formatCurrency(totals.tax)}
                  </TableCell>
                  <TableCell className="text-right text-sm font-mono py-2">
                    {totals.discount > 0 ? formatCurrency(totals.discount) : "-"}
                  </TableCell>
                  <TableCell className="text-right text-sm font-mono py-2">
                    {formatCurrency(totals.total)}
                  </TableCell>
                  <TableCell />
                </TableRow>
              </TableFooter>
            </Table>
          </div>
        )
      ) : null}
    </div>
  );
}
