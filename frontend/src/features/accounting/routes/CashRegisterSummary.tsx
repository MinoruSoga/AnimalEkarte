// BUG-368: レジ締め日次集計ページ（表示専用・印刷対応）
import { useMemo, useState, useCallback } from "react";
import { useNavigate } from "react-router";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { C } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import { formatCurrency } from "@/utils/format/number";

import { useGetDailySummary } from "../api/get-daily-summary";

const PAYMENT_METHOD_LABELS: Record<string, string> = {
  cash: "現金",
  credit_card: "クレジットカード",
  electronic_money: "電子マネー",
};

const CATEGORY_LABELS: Record<string, string> = {
  examination: "診察",
  test: "検査",
  procedure: "処置",
  surgery: "手術",
  medicine: "薬品",
  food: "フード",
  goods: "グッズ",
  other: "その他",
};

function todayISO(): string {
  const d = new Date();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${d.getFullYear()}-${mm}-${dd}`;
}

export function CashRegisterSummary() {
  const navigate = useNavigate();
  const [date, setDate] = useState(todayISO);

  const { data, isLoading, isError, error } = useGetDailySummary(date);

  const handleDateChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setDate(e.target.value);
  }, []);

  const handleBack = useCallback(() => {
    navigate(paths.accounting.getHref());
  }, [navigate]);

  const handlePrint = useCallback(() => {
    window.print();
  }, []);

  const paymentRows = useMemo(
    () =>
      (data?.payment_totals ?? []).map((p) => ({
        label: PAYMENT_METHOD_LABELS[p.method] ?? p.method,
        total: p.total,
      })),
    [data?.payment_totals],
  );

  const categoryRows = useMemo(
    () =>
      (data?.category_totals ?? []).map((c) => ({
        label: CATEGORY_LABELS[c.category] ?? c.category,
        total: c.total,
      })),
    [data?.category_totals],
  );

  return (
    <PageLayout
      title="レジ締め日次集計"
      onBack={handleBack}
    >
      {/* 日付選択 */}
      <div className="flex items-center gap-4 mb-6 print:hidden">
        <Label htmlFor="summary-date">集計日</Label>
        <Input
          id="summary-date"
          type="date"
          value={date}
          onChange={handleDateChange}
          className="w-48"
        />
        <Button variant="outline" onClick={handlePrint}>
          印刷
        </Button>
      </div>

      {isLoading ? (
        <LoadingFallback />
      ) : isError ? (
        <ErrorFallback message={(error as Error)?.message} />
      ) : (
        <div className="space-y-6">
          {/* サマリー */}
          <div className={`flex gap-8 p-4 rounded-lg border ${C.borderMuted}`}>
            <div>
              <p className={`text-sm ${C.textMuted}`}>件数</p>
              <p className={`text-2xl font-bold ${C.text}`}>
                {data?.billing_count ?? 0}件
              </p>
            </div>
            <div>
              <p className={`text-sm ${C.textMuted}`}>売上合計</p>
              <p className={`text-2xl font-bold ${C.text}`}>
                {formatCurrency(data?.grand_total ?? 0)}
              </p>
            </div>
          </div>

          {/* 支払方法別集計 */}
          <section>
            <h2 className={`text-base font-semibold mb-2 ${C.text}`}>
              支払方法別
            </h2>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>支払方法</TableHead>
                  <TableHead className="text-right">金額</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {paymentRows.length > 0 ? (
                  paymentRows.map((row) => (
                    <TableRow key={row.label}>
                      <TableCell>{row.label}</TableCell>
                      <TableCell className="text-right">{formatCurrency(row.total)}</TableCell>
                    </TableRow>
                  ))
                ) : (
                  <TableRow>
                    <TableCell colSpan={2} className={`text-center ${C.textMuted}`}>
                      データなし
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </section>

          {/* 診療区分別集計 */}
          <section>
            <h2 className={`text-base font-semibold mb-2 ${C.text}`}>
              診療区分別
            </h2>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>区分</TableHead>
                  <TableHead className="text-right">金額</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {categoryRows.length > 0 ? (
                  categoryRows.map((row) => (
                    <TableRow key={row.label}>
                      <TableCell>{row.label}</TableCell>
                      <TableCell className="text-right">{formatCurrency(row.total)}</TableCell>
                    </TableRow>
                  ))
                ) : (
                  <TableRow>
                    <TableCell colSpan={2} className={`text-center ${C.textMuted}`}>
                      データなし
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </section>
        </div>
      )}
    </PageLayout>
  );
}
