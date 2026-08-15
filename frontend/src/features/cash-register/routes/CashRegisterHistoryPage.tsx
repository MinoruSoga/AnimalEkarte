import { useState, useCallback } from "react";
import { useSearchParams } from "react-router";
import { History } from "lucide-react";
import { C, ICON, LAYOUT, STYLE } from "@/lib/design-tokens";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { TableCell, TableHead } from "@/components/ui/table";
import { formatJSTDateTimeLocal, toJSTWallDate } from "@/lib/jst-date";
import { Pagination } from "@/components/shared/Pagination";
import { DataTableRowButton } from "@/components/shared/DataTable/DataTableRowButton";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { formatCurrency } from "@/lib/format/number";
import { useGetCashRegisterCloses } from "../api/get-cash-register-closes";
import type { CashRegisterClose } from "../api/get-cash-register-closes";
import { PERIOD_LABELS, PERIOD_OPTIONS, type CashRegisterPeriod } from "../constants";
import { summarizeCategoryTotals } from "../category-breakdown";
import { monthDateRange } from "../month-date-range";
import { ResourceCashRegisterClose } from "@/types/generated/models";

type PeriodFilter = "all" | CashRegisterPeriod;

// BE のデフォルト limit（query_helpers.go）と一致させる
const HISTORY_PAGE_SIZE = 20;

// 月次集計レポートからのドリルダウン用 ?date=YYYY-MM-DD を年月にパースする
function parseHighlightDate(dateParam: string | null): { year: number; month: number } | null {
  if (!dateParam) return null;
  const matched = /^(\d{4})-(\d{2})-(\d{2})$/.exec(dateParam);
  if (!matched) return null;
  return { year: Number(matched[1]), month: Number(matched[2]) };
}

function diffClass(diff: number): string {
  return diff === 0 ? C.textStatusGreen : C.danger;
}

function formatDiff(diff: number): string {
  return `${diff >= 0 ? "+" : ""}${diff.toLocaleString()}`;
}

export function CashRegisterHistoryPage() {
  const now = toJSTWallDate(new Date());
  const [searchParams] = useSearchParams();
  const drillDownTarget = parseHighlightDate(searchParams.get("date"));
  // 妥当な YYYY-MM-DD のみハイライト対象とする（不正な値は無視）
  const highlightDate = drillDownTarget ? searchParams.get("date") : null;

  const [year, setYear] = useState<number>(drillDownTarget?.year ?? now.getFullYear());
  const [month, setMonth] = useState<number>(drillDownTarget?.month ?? now.getMonth() + 1);
  const [periodFilter, setPeriodFilter] = useState<PeriodFilter>("all");
  const [selectedClose, setSelectedClose] = useState<CashRegisterClose | null>(null);
  // ドリルダウン(?date=)初期表示は該当日 1 日分に絞る。年月セレクタを操作したら
  // null に落として月レンジ(月初〜月末)表示へ切り替える。
  const [singleDate, setSingleDate] = useState<string | null>(highlightDate);
  const [page, setPage] = useState<number>(1);

  // BE 契約: start_date <= close_date <= end_date（両端含む）。年月→月初/月末を導出する。
  const { startDate, endDate } = singleDate
    ? { startDate: singleDate, endDate: singleDate }
    : monthDateRange(year, month);

  const { data, isLoading, isError } = useGetCashRegisterCloses({
    start_date: startDate,
    end_date: endDate,
    page,
    limit: HISTORY_PAGE_SIZE,
  });

  const handleYearChange = useCallback((e: React.ChangeEvent<HTMLSelectElement>) => {
    setYear(Number(e.target.value));
    setSingleDate(null);
    setPage(1);
  }, [setPage]);

  const handleMonthChange = useCallback((e: React.ChangeEvent<HTMLSelectElement>) => {
    setMonth(Number(e.target.value));
    setSingleDate(null);
    setPage(1);
  }, [setPage]);

  const handlePeriodFilterChange = useCallback((e: React.ChangeEvent<HTMLSelectElement>) => {
    setPeriodFilter(e.target.value as PeriodFilter);
  }, []);

  const yearOptions = Array.from({ length: 5 }, (_, i) => now.getFullYear() - i);

  // period(AM/PM/EMG)はクライアント側フィルタのため、現在ページ内の行だけに作用する。
  const rows = (data?.data ?? []).filter(
    (close) => periodFilter === "all" || close.period === periodFilter,
  );

  // ページネーションはサーバー側。BE の total と FE 管理の page/limit から算出する。
  const total = data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / HISTORY_PAGE_SIZE));
  const startIndex = total === 0 ? 0 : (page - 1) * HISTORY_PAGE_SIZE + 1;
  const endIndex = Math.min(page * HISTORY_PAGE_SIZE, total);

  const detailSubtotals = selectedClose
    ? summarizeCategoryTotals(selectedClose.categoryBreakdown)
    : [];
  const detailDiff = selectedClose
    ? (selectedClose.actualCash ?? 0) - (selectedClose.theoreticalCash ?? 0)
    : 0;

  return (
    <PageLayout
      title="締め履歴"
      resource={ResourceCashRegisterClose}
      icon={<History className={`${ICON.page} ${C.text}`} />}
      maxWidth={LAYOUT.pageContentMaxWidth.full}
    >
      <div className="space-y-6">
        {highlightDate ? (
          <p className={`text-base ${C.text60}`}>
            月次集計レポートから <span className={`font-medium ${C.text}`}>{highlightDate}</span>{" "}
            の締めをハイライト表示しています。
          </p>
        ) : null}

        {/* 絞り込み */}
        <div className="flex flex-wrap gap-3 items-center">
          <div>
            <label htmlFor="hist_year" className={`${STYLE.formLabel} mr-2`}>
              年
            </label>
            <select
              id="hist_year"
              value={year}
              onChange={handleYearChange}
              className={`${STYLE.formInput} rounded-xs border px-3 inline-block w-auto`}
            >
              {yearOptions.map((y) => (
                <option key={y} value={y}>
                  {y}年
                </option>
              ))}
            </select>
          </div>
          <div>
            <label htmlFor="hist_month" className={`${STYLE.formLabel} mr-2`}>
              月
            </label>
            <select
              id="hist_month"
              value={month}
              onChange={handleMonthChange}
              className={`${STYLE.formInput} rounded-xs border px-3 inline-block w-auto`}
            >
              {Array.from({ length: 12 }, (_, i) => i + 1).map((m) => (
                <option key={m} value={m}>
                  {m}月
                </option>
              ))}
            </select>
          </div>
          <div>
            <label htmlFor="hist_period" className={`${STYLE.formLabel} mr-2`}>
              区分
            </label>
            <select
              id="hist_period"
              value={periodFilter}
              onChange={handlePeriodFilterChange}
              className={`${STYLE.formInput} rounded-xs border px-3 inline-block w-auto`}
            >
              <option value="all">すべて</option>
              {PERIOD_OPTIONS.map((p) => (
                <option key={p} value={p}>
                  {PERIOD_LABELS[p]}
                </option>
              ))}
            </select>
          </div>
        </div>

        {isLoading ? (
          <div className="flex items-center justify-center py-8">
            <p className={`text-base ${C.text50}`}>読み込み中...</p>
          </div>
        ) : isError ? (
          <div className="flex items-center justify-center py-8">
            <p className={`text-base ${C.danger}`}>データの取得に失敗しました</p>
          </div>
        ) : (
          <div className={`${C.bgWhite} rounded-lg border ${C.borderLight} overflow-x-auto`}>
            {rows.length > 0 ? (
              <table className="w-full text-sm">
                <thead>
                  <tr className={STYLE.tableHeaderRow}>
                    <TableHead>日付</TableHead>
                    <TableHead>区分</TableHead>
                    <TableHead className="text-right">理論現金</TableHead>
                    <TableHead className="text-right">実際の現金</TableHead>
                    <TableHead className="text-right">差額</TableHead>
                    <TableHead>担当者</TableHead>
                    <TableHead>締め時刻</TableHead>
                  </tr>
                </thead>
                <tbody>
                  {rows.map((close) => {
                    const diff = (close.actualCash ?? 0) - (close.theoreticalCash ?? 0);
                    const isHighlighted =
                      highlightDate != null && close.closeDate.slice(0, 10) === highlightDate;
                    return (
                      <tr
                        key={close.id}
                        data-highlighted={isHighlighted ? "true" : undefined}
                        className={`border-b ${C.borderLight} ${
                          isHighlighted ? C.bgBrandLight40 : STYLE.tableRow
                        }`}
                      >
                        <TableCell className={C.text}>
                          <DataTableRowButton
                            aria-label={`締め詳細: ${close.closeDate} ${PERIOD_LABELS[close.period]} (ID ${close.id})`}
                            onClick={() => setSelectedClose(close)}
                          >
                            {close.closeDate}
                          </DataTableRowButton>
                        </TableCell>
                        <TableCell className={C.text}>
                          {PERIOD_LABELS[close.period]}
                        </TableCell>
                        <TableCell className={`text-right ${C.text}`}>
                          {formatCurrency(close.theoreticalCash ?? 0)}
                        </TableCell>
                        <TableCell className={`text-right ${C.text}`}>
                          {formatCurrency(close.actualCash ?? 0)}
                        </TableCell>
                        <TableCell className={`text-right font-medium ${diffClass(diff)}`}>
                          {formatDiff(diff)}
                        </TableCell>
                        <TableCell className={C.text}>
                          {close.closedByStaffName ?? "—"}
                        </TableCell>
                        <TableCell className={C.text60}>
                          {close.closedAt
                            ? formatJSTDateTimeLocal(close.closedAt).slice(5).replace("T", " ")
                            : "—"}
                        </TableCell>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            ) : (
              <p className={`text-base ${C.text50} py-8 text-center`}>
                締め履歴がありません
              </p>
            )}
          </div>
        )}

        {!isLoading && !isError && total > HISTORY_PAGE_SIZE ? (
          <Pagination
            currentPage={page}
            totalPages={totalPages}
            totalCount={total}
            startIndex={startIndex}
            endIndex={endIndex}
            onPageChange={(p) => setPage(p)}
            onPrev={() => setPage((p) => Math.max(1, p - 1))}
            onNext={() => setPage((p) => Math.min(totalPages, p + 1))}
          />
        ) : null}

        <Dialog
          open={selectedClose !== null}
          onOpenChange={(open) => {
            if (!open) setSelectedClose(null);
          }}
        >
          <DialogContent>
            {selectedClose ? (
              <>
                <DialogHeader>
                  <DialogTitle>
                    {selectedClose.closeDate} {PERIOD_LABELS[selectedClose.period]} の締め詳細
                  </DialogTitle>
                  <DialogDescription>
                    この締めレコードの集計内訳と差額を表示しています。
                  </DialogDescription>
                </DialogHeader>

                <div className="space-y-4">
                  <dl className="grid grid-cols-2 gap-x-4 gap-y-2 text-base">
                    <dt className={C.text60}>理論現金</dt>
                    <dd className={`text-right ${C.text}`}>
                      {formatCurrency(selectedClose.theoreticalCash ?? 0)}
                    </dd>
                    <dt className={C.text60}>実際の現金</dt>
                    <dd className={`text-right ${C.text}`}>
                      {formatCurrency(selectedClose.actualCash ?? 0)}
                    </dd>
                    <dt className={C.text60}>差額</dt>
                    <dd className={`text-right font-medium ${diffClass(detailDiff)}`}>
                      {formatDiff(detailDiff)}
                    </dd>
                    <dt className={C.text60}>担当者</dt>
                    <dd className={`text-right ${C.text}`}>
                      {selectedClose.closedByStaffName ?? "—"}
                    </dd>
                    <dt className={C.text60}>締め日時</dt>
                    <dd className={`text-right ${C.text}`}>
                      {selectedClose.closedAt
                        ? formatJSTDateTimeLocal(selectedClose.closedAt).replace("T", " ")
                        : "—"}
                    </dd>
                  </dl>

                  <div>
                    <p className={`text-sm font-medium ${C.text70} mb-1`}>メモ</p>
                    <p className={`text-base ${C.text}`}>
                      {selectedClose.memo ? selectedClose.memo : "—"}
                    </p>
                  </div>

                  <div>
                    <p className={`text-sm font-medium ${C.text70} mb-2`}>部門別内訳</p>
                    {detailSubtotals.length > 0 ? (
                      <dl className="space-y-1">
                        {detailSubtotals.map((row) => (
                          <div key={row.label} className="flex justify-between text-base">
                            <dt className={C.text60}>{row.label}</dt>
                            <dd className={C.text}>
                              {formatCurrency(row.total)}
                              {/* DEC-40: 未分類・要確認の件数は締め時点 snapshot。欠落は「記録なし」 */}
                              {row.count !== undefined ? (
                                <span className={`ml-2 ${C.text60}`}>
                                  {row.count === null ? "記録なし" : `${row.count}件`}
                                </span>
                              ) : null}
                            </dd>
                          </div>
                        ))}
                      </dl>
                    ) : (
                      <p className={`text-base ${C.text50}`}>内訳データなし</p>
                    )}
                  </div>
                </div>
              </>
            ) : null}
          </DialogContent>
        </Dialog>
      </div>
    </PageLayout>
  );
}
