import { useState, useCallback } from "react";
import { useSearchParams } from "react-router";
import { History } from "lucide-react";
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { toJSTWallDate } from "@/lib/jst-date";
import { Pagination } from "@/components/shared/Pagination";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { useGetCashRegisterCloses } from "../api/get-cash-register-closes";
import type { CashRegisterClose } from "../api/get-cash-register-closes";
import { monthDateRange } from "../lib/month-date-range";
import { ResourceCashRegisterClose } from "@/types/generated/models";
import {
  CashRegisterHistoryFilters,
  type HistoryPeriodFilter,
} from "../components/CashRegisterHistoryFilters";
import { CashRegisterHistoryTable } from "../components/CashRegisterHistoryTable";
import { CashRegisterHistoryDetailDialog } from "../components/CashRegisterHistoryDetailDialog";
import { parseHighlightDate } from "../lib/cash-register-history-model";

const HISTORY_PAGE_SIZE = 20;

export function CashRegisterHistoryPage() {
  const now = toJSTWallDate(new Date());
  const [searchParams] = useSearchParams();
  const drillDownTarget = parseHighlightDate(searchParams.get("date"));
  const highlightDate = drillDownTarget ? searchParams.get("date") : null;

  const [year, setYear] = useState<number>(drillDownTarget?.year ?? now.getFullYear());
  const [month, setMonth] = useState<number>(drillDownTarget?.month ?? now.getMonth() + 1);
  const [periodFilter, setPeriodFilter] = useState<HistoryPeriodFilter>("all");
  const [selectedClose, setSelectedClose] = useState<CashRegisterClose | null>(null);
  const [singleDate, setSingleDate] = useState<string | null>(highlightDate);
  const [page, setPage] = useState<number>(1);

  const { startDate, endDate } = singleDate
    ? { startDate: singleDate, endDate: singleDate }
    : monthDateRange(year, month);

  const { data, isLoading, isError } = useGetCashRegisterCloses({
    start_date: startDate,
    end_date: endDate,
    page,
    limit: HISTORY_PAGE_SIZE,
  });

  const handleYearChange = useCallback(
    (e: React.ChangeEvent<HTMLSelectElement>) => {
      setYear(Number(e.target.value));
      setSingleDate(null);
      setPage(1);
    },
    [setPage],
  );

  const handleMonthChange = useCallback(
    (e: React.ChangeEvent<HTMLSelectElement>) => {
      setMonth(Number(e.target.value));
      setSingleDate(null);
      setPage(1);
    },
    [setPage],
  );

  const handlePeriodFilterChange = useCallback((e: React.ChangeEvent<HTMLSelectElement>) => {
    setPeriodFilter(e.target.value as HistoryPeriodFilter);
  }, []);

  const yearOptions = Array.from({ length: 5 }, (_, i) => now.getFullYear() - i);

  const rows = (data?.data ?? []).filter(
    (close) => periodFilter === "all" || close.period === periodFilter,
  );

  const total = data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / HISTORY_PAGE_SIZE));
  const startIndex = total === 0 ? 0 : (page - 1) * HISTORY_PAGE_SIZE + 1;
  const endIndex = Math.min(page * HISTORY_PAGE_SIZE, total);

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

        <CashRegisterHistoryFilters
          year={year}
          month={month}
          periodFilter={periodFilter}
          yearOptions={yearOptions}
          onYearChange={handleYearChange}
          onMonthChange={handleMonthChange}
          onPeriodFilterChange={handlePeriodFilterChange}
        />

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
            <CashRegisterHistoryTable
              rows={rows}
              highlightDate={highlightDate}
              onSelect={setSelectedClose}
            />
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

        <CashRegisterHistoryDetailDialog
          selectedClose={selectedClose}
          onOpenChange={(open) => {
            if (!open) setSelectedClose(null);
          }}
        />
      </div>
    </PageLayout>
  );
}
