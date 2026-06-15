import { useMemo, useCallback } from "react";
import { createPortal } from "react-dom";
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
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { C, ICON } from "@/lib/design-tokens";
import { todayJSTISO } from "@/lib/jst-date";
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
function formatReceiptNo(billingId: string): string {
  return billingId.padStart(8, "0");
}

// ── カテゴリ集計 ─────────────────────────────────────────────────
interface CategoryBreakdown {
  medical: number;
  surgery: number;
  rv: number;       // vaccine
  food: number;
  trimming: number;
  hotel: number;
  goods: number;    // goods + training + other
}

function getCategoryBreakdown(items: AccountingItem[]): CategoryBreakdown {
  let medical = 0;
  let surgery = 0;
  let rv = 0;
  let food = 0;
  let trimming = 0;
  let hotel = 0;
  let goods = 0;
  for (const item of items) {
    const sub = item.subtotal;
    if (MEDICAL_CATEGORIES.has(item.category)) {
      medical += sub;
    } else if (item.category === "surgery") {
      surgery += sub;
    } else if (item.category === "vaccine") {
      rv += sub;
    } else if (item.category === "food") {
      food += sub;
    } else if (item.category === "trimming") {
      trimming += sub;
    } else if (item.category === "hotel") {
      hotel += sub;
    } else {
      goods += sub;
    }
  }
  return { medical, surgery, rv, food, trimming, hotel, goods };
}

// ── 支払按分（印刷用）: RV優先→診療優先で現金を割り当て ──────────
interface CashCardAmount {
  cash: number;
  card: number;
}

interface DetailedBreakdown {
  medical: CashCardAmount;
  surgery: CashCardAmount;
  rv: CashCardAmount;
  food: CashCardAmount;
  trimming: CashCardAmount;
  hotel: CashCardAmount;
  goods: CashCardAmount;
}

function apportionPayment(breakdown: CategoryBreakdown, cashTotal: number): DetailedBreakdown {
  // RV優先、次に診療、次に外科…と現金を割り当てる
  const PRIORITY: Array<keyof CategoryBreakdown> = ["rv", "medical", "surgery", "food", "trimming", "hotel", "goods"];
  let cashLeft = cashTotal;
  const result: DetailedBreakdown = {
    medical: { cash: 0, card: 0 },
    surgery: { cash: 0, card: 0 },
    rv: { cash: 0, card: 0 },
    food: { cash: 0, card: 0 },
    trimming: { cash: 0, card: 0 },
    hotel: { cash: 0, card: 0 },
    goods: { cash: 0, card: 0 },
  };
  for (const cat of PRIORITY) {
    const catTotal = breakdown[cat];
    const catCash = Math.min(catTotal, cashLeft);
    result[cat] = { cash: catCash, card: catTotal - catCash };
    cashLeft -= catCash;
  }
  return result;
}

function getRowCashTotal(accounting: Accounting): number {
  if (accounting.paymentSplits && accounting.paymentSplits.length > 1) {
    return accounting.paymentSplits
      .filter((s) => s.method === "cash")
      .reduce((sum, s) => sum + s.amount, 0);
  }
  const total = accounting.payment?.totalAmount ?? 0;
  return accounting.payment?.method === "cash" ? total : 0;
}

// ── 行・合計データ ───────────────────────────────────────────────
interface RowData {
  accounting: Accounting;
  breakdown: CategoryBreakdown;
  detailedBreakdown: DetailedBreakdown;
  subtotal: number;
  tax: number;
  discount: number;
  total: number;
}

interface TotalsData {
  medical: number;
  surgery: number;
  rv: number;
  food: number;
  trimming: number;
  hotel: number;
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

// ── 印刷エリア ──────────────────────────────────────────────────
interface CatCellProps {
  detail: CashCardAmount;
  isMixed: boolean;
}

function CatCell({ detail, isMixed }: CatCellProps) {
  const total = detail.cash + detail.card;
  if (total === 0) {
    return <td className="border border-gray-300 text-right px-1 py-0.5 text-[9pt]">-</td>;
  }
  if (!isMixed) {
    return (
      <td className="border border-gray-300 text-right px-1 py-0.5 text-[9pt]">
        ¥{total.toLocaleString()}
      </td>
    );
  }
  return (
    <td className="border border-gray-300 text-right px-1 py-0.5 text-[9pt] leading-tight">
      {detail.cash > 0 ? <div>現{detail.cash.toLocaleString()}</div> : null}
      {detail.card > 0 ? <div className="text-blue-700">カ{detail.card.toLocaleString()}</div> : null}
    </td>
  );
}

interface DailyPrintAreaProps {
  date: string;
  rows: RowData[];
  totals: TotalsData;
}

function DailyPrintArea({ date, rows, totals }: DailyPrintAreaProps) {
  const hospitalTotal = totals.medical + totals.surgery + totals.rv + totals.food + totals.goods;
  const trimmingTotal = totals.trimming + totals.hotel;

  return createPortal(
    <div
      hidden
      className="bg-white"
      data-testid="daily-print-area"
      style={{ position: "fixed", inset: 0, zIndex: 9999, overflow: "auto", padding: "8mm" }}
    >
      <style type="text/css">
        {`
          @media print {
            [data-testid="daily-print-area"] {
              display: block !important;
            }
            body > :not([data-testid="daily-print-area"]) {
              display: none !important;
            }
          }
          @page { size: A4 landscape; margin: 8mm; }
          @media print { body { margin: 0; -webkit-print-color-adjust: exact; print-color-adjust: exact; } }
        `}
      </style>

      {/* ヘッダー */}
      <div className="mb-3 text-center">
        <h1 className="text-[14pt] font-bold">日次集計一覧表</h1>
        <p className="text-[10pt]">対象日: {date} 件数: {rows.length}件</p>
      </div>

      {/* メインテーブル */}
      <table className="w-full text-[8pt] border-collapse">
        <thead>
          <tr className="bg-gray-100">
            <th className="border border-gray-400 px-1 py-0.5 text-left whitespace-nowrap">領収No</th>
            <th className="border border-gray-400 px-1 py-0.5 text-left whitespace-nowrap">飼主名</th>
            <th className="border border-gray-400 px-1 py-0.5 text-left whitespace-nowrap">ペット名</th>
            <th className="border border-gray-400 px-1 py-0.5 text-right whitespace-nowrap">診療</th>
            <th className="border border-gray-400 px-1 py-0.5 text-right whitespace-nowrap">外科</th>
            <th className="border border-gray-400 px-1 py-0.5 text-right whitespace-nowrap">RV</th>
            <th className="border border-gray-400 px-1 py-0.5 text-right whitespace-nowrap">フード</th>
            <th className="border border-gray-400 px-1 py-0.5 text-right whitespace-nowrap">トリミング</th>
            <th className="border border-gray-400 px-1 py-0.5 text-right whitespace-nowrap">ホテル</th>
            <th className="border border-gray-400 px-1 py-0.5 text-right whitespace-nowrap">用品他</th>
            <th className="border border-gray-400 px-1 py-0.5 text-center whitespace-nowrap">支払方法</th>
            <th className="border border-gray-400 px-1 py-0.5 text-right whitespace-nowrap">合計</th>
          </tr>
        </thead>
        <tbody>
          {rows.map(({ accounting: a, detailedBreakdown, total }) => {
            const isMixed = Boolean(a.paymentSplits && a.paymentSplits.length > 1);
            const paymentLabel = isMixed
              ? a.paymentSplits!.map((s) => PAYMENT_METHOD_LABELS[s.method] ?? s.method).join("/")
              : a.payment
                ? (PAYMENT_METHOD_LABELS[a.payment.method] ?? a.payment.method)
                : "-";

            return (
              <tr key={a.id}>
                <td className="border border-gray-300 px-1 py-0.5 text-[9pt] font-mono">
                  {formatReceiptNo(a.id)}
                </td>
                <td className="border border-gray-300 px-1 py-0.5 text-[9pt]">{a.ownerName}</td>
                <td className="border border-gray-300 px-1 py-0.5 text-[9pt]">{a.petName}</td>
                <CatCell detail={detailedBreakdown.medical} isMixed={isMixed} />
                <CatCell detail={detailedBreakdown.surgery} isMixed={isMixed} />
                <CatCell detail={detailedBreakdown.rv} isMixed={isMixed} />
                <CatCell detail={detailedBreakdown.food} isMixed={isMixed} />
                <CatCell detail={detailedBreakdown.trimming} isMixed={isMixed} />
                <CatCell detail={detailedBreakdown.hotel} isMixed={isMixed} />
                <CatCell detail={detailedBreakdown.goods} isMixed={isMixed} />
                <td className="border border-gray-300 px-1 py-0.5 text-[9pt] text-center whitespace-nowrap">
                  {paymentLabel}
                </td>
                <td className="border border-gray-300 px-1 py-0.5 text-right text-[9pt] font-semibold">
                  ¥{total.toLocaleString()}
                </td>
              </tr>
            );
          })}
        </tbody>
        <tfoot>
          {/* 病院合計行 */}
          <tr className="bg-gray-50 font-semibold">
            <td colSpan={3} className="border border-gray-400 px-1 py-0.5 text-[9pt]">病院合計</td>
            <td className="border border-gray-400 px-1 py-0.5 text-right text-[9pt]">
              {totals.medical > 0 ? `¥${totals.medical.toLocaleString()}` : "-"}
            </td>
            <td className="border border-gray-400 px-1 py-0.5 text-right text-[9pt]">
              {totals.surgery > 0 ? `¥${totals.surgery.toLocaleString()}` : "-"}
            </td>
            <td className="border border-gray-400 px-1 py-0.5 text-right text-[9pt]">
              {totals.rv > 0 ? `¥${totals.rv.toLocaleString()}` : "-"}
            </td>
            <td className="border border-gray-400 px-1 py-0.5 text-right text-[9pt]">
              {totals.food > 0 ? `¥${totals.food.toLocaleString()}` : "-"}
            </td>
            <td className="border border-gray-400 px-1 py-0.5 text-center text-[9pt]">-</td>
            <td className="border border-gray-400 px-1 py-0.5 text-center text-[9pt]">-</td>
            <td className="border border-gray-400 px-1 py-0.5 text-right text-[9pt]">
              {totals.goods > 0 ? `¥${totals.goods.toLocaleString()}` : "-"}
            </td>
            <td className="border border-gray-400 px-1 py-0.5 text-[9pt]" />
            <td className="border border-gray-400 px-1 py-0.5 text-right text-[9pt] font-bold">
              ¥{hospitalTotal.toLocaleString()}
            </td>
          </tr>
          {/* トリミング合計行 */}
          <tr className="bg-gray-50 font-semibold">
            <td colSpan={3} className="border border-gray-400 px-1 py-0.5 text-[9pt]">トリミング合計</td>
            <td className="border border-gray-400 px-1 py-0.5 text-center text-[9pt]">-</td>
            <td className="border border-gray-400 px-1 py-0.5 text-center text-[9pt]">-</td>
            <td className="border border-gray-400 px-1 py-0.5 text-center text-[9pt]">-</td>
            <td className="border border-gray-400 px-1 py-0.5 text-center text-[9pt]">-</td>
            <td className="border border-gray-400 px-1 py-0.5 text-right text-[9pt]">
              {totals.trimming > 0 ? `¥${totals.trimming.toLocaleString()}` : "-"}
            </td>
            <td className="border border-gray-400 px-1 py-0.5 text-right text-[9pt]">
              {totals.hotel > 0 ? `¥${totals.hotel.toLocaleString()}` : "-"}
            </td>
            <td className="border border-gray-400 px-1 py-0.5 text-center text-[9pt]">-</td>
            <td className="border border-gray-400 px-1 py-0.5 text-[9pt]" />
            <td className="border border-gray-400 px-1 py-0.5 text-right text-[9pt] font-bold">
              ¥{trimmingTotal.toLocaleString()}
            </td>
          </tr>
          {/* 全体合計行 */}
          <tr className="bg-gray-200 font-bold">
            <td colSpan={3} className="border border-gray-400 px-1 py-1 text-[9pt]">全体合計</td>
            <td className="border border-gray-400 px-1 py-1 text-right text-[9pt]">
              {totals.medical > 0 ? `¥${totals.medical.toLocaleString()}` : "-"}
            </td>
            <td className="border border-gray-400 px-1 py-1 text-right text-[9pt]">
              {totals.surgery > 0 ? `¥${totals.surgery.toLocaleString()}` : "-"}
            </td>
            <td className="border border-gray-400 px-1 py-1 text-right text-[9pt]">
              {totals.rv > 0 ? `¥${totals.rv.toLocaleString()}` : "-"}
            </td>
            <td className="border border-gray-400 px-1 py-1 text-right text-[9pt]">
              {totals.food > 0 ? `¥${totals.food.toLocaleString()}` : "-"}
            </td>
            <td className="border border-gray-400 px-1 py-1 text-right text-[9pt]">
              {totals.trimming > 0 ? `¥${totals.trimming.toLocaleString()}` : "-"}
            </td>
            <td className="border border-gray-400 px-1 py-1 text-right text-[9pt]">
              {totals.hotel > 0 ? `¥${totals.hotel.toLocaleString()}` : "-"}
            </td>
            <td className="border border-gray-400 px-1 py-1 text-right text-[9pt]">
              {totals.goods > 0 ? `¥${totals.goods.toLocaleString()}` : "-"}
            </td>
            <td className="border border-gray-400 px-1 py-1 text-[9pt]" />
            <td className="border border-gray-400 px-1 py-1 text-right text-[10pt] font-bold">
              ¥{totals.total.toLocaleString()}
            </td>
          </tr>
        </tfoot>
      </table>
    </div>,
    document.body,
  );
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
                    <TableHead className={`text-xs ${C.text60} whitespace-nowrap w-[90px]`}>領収No</TableHead>
                    {isMultiClinic ? (
                      <TableHead className={`text-xs ${C.text60} whitespace-nowrap w-[100px]`}>拠点</TableHead>
                    ) : null}
                    <TableHead className={`text-xs ${C.text60} whitespace-nowrap`}>飼主名</TableHead>
                    <TableHead className={`text-xs ${C.text60} whitespace-nowrap`}>ペット名</TableHead>
                    <TableHead className={`text-right text-xs ${C.text60} whitespace-nowrap`}>診療</TableHead>
                    <TableHead className={`text-right text-xs ${C.text60} whitespace-nowrap`}>外科</TableHead>
                    <TableHead className={`text-right text-xs ${C.text60} whitespace-nowrap`}>RV</TableHead>
                    <TableHead className={`text-right text-xs ${C.text60} whitespace-nowrap`}>フード</TableHead>
                    <TableHead className={`text-right text-xs ${C.text60} whitespace-nowrap`}>トリミング</TableHead>
                    <TableHead className={`text-right text-xs ${C.text60} whitespace-nowrap`}>ホテル</TableHead>
                    <TableHead className={`text-right text-xs ${C.text60} whitespace-nowrap`}>用品他</TableHead>
                    <TableHead className={`text-right text-xs ${C.text60} whitespace-nowrap`}>売上合計</TableHead>
                    <TableHead className={`text-center text-xs ${C.text60} whitespace-nowrap`}>支払方法</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {rows.map(({ accounting: a, breakdown, total }) => (
                    <TableRow key={a.id} className={`border-b ${C.borderLight}`}>
                      <TableCell className={`text-xs font-mono py-2 ${C.text60}`}>
                        {formatReceiptNo(a.id)}
                      </TableCell>
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
                        {breakdown.rv > 0 ? formatCurrency(breakdown.rv) : "-"}
                      </TableCell>
                      <TableCell className={`text-right text-sm font-mono py-2 ${C.text60}`}>
                        {breakdown.food > 0 ? formatCurrency(breakdown.food) : "-"}
                      </TableCell>
                      <TableCell className={`text-right text-sm font-mono py-2 ${C.text60}`}>
                        {breakdown.trimming > 0 ? formatCurrency(breakdown.trimming) : "-"}
                      </TableCell>
                      <TableCell className={`text-right text-sm font-mono py-2 ${C.text60}`}>
                        {breakdown.hotel > 0 ? formatCurrency(breakdown.hotel) : "-"}
                      </TableCell>
                      <TableCell className={`text-right text-sm font-mono py-2 ${C.text60}`}>
                        {breakdown.goods > 0 ? formatCurrency(breakdown.goods) : "-"}
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
                    <TableCell colSpan={labelColSpan} className="py-2 text-sm">合計（{rows.length}件）</TableCell>
                    <TableCell className="text-right text-sm font-mono py-2">
                      {totals.medical > 0 ? formatCurrency(totals.medical) : "-"}
                    </TableCell>
                    <TableCell className="text-right text-sm font-mono py-2">
                      {totals.surgery > 0 ? formatCurrency(totals.surgery) : "-"}
                    </TableCell>
                    <TableCell className="text-right text-sm font-mono py-2">
                      {totals.rv > 0 ? formatCurrency(totals.rv) : "-"}
                    </TableCell>
                    <TableCell className="text-right text-sm font-mono py-2">
                      {totals.food > 0 ? formatCurrency(totals.food) : "-"}
                    </TableCell>
                    <TableCell className="text-right text-sm font-mono py-2">
                      {totals.trimming > 0 ? formatCurrency(totals.trimming) : "-"}
                    </TableCell>
                    <TableCell className="text-right text-sm font-mono py-2">
                      {totals.hotel > 0 ? formatCurrency(totals.hotel) : "-"}
                    </TableCell>
                    <TableCell className="text-right text-sm font-mono py-2">
                      {totals.goods > 0 ? formatCurrency(totals.goods) : "-"}
                    </TableCell>
                    <TableCell className="text-right text-sm font-mono font-bold py-2">
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

      {/* 印刷エリア: print時のみ表示 */}
      {rows.length > 0 ? (
        <DailyPrintArea date={selectedDate} rows={rows} totals={totals} />
      ) : null}
    </>
  );
}
