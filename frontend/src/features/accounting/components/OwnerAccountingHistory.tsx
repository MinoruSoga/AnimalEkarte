// React/Framework
import { memo, useMemo, useCallback, useRef, useEffect } from "react";
import { useSearchParams } from "react-router";

// Internal
import { LoadingFallback } from "@/components/shared/DataStates";
import { Pagination } from "@/components/shared/Pagination";
import { C } from "@/lib/design-tokens";

// Relative
import { useGetAccountings } from "../api/get-accountings";
import type { Accounting } from "../api/transforms";
import type { AccountingStatus } from "../types";
import {
  AccountingHistorySortControls,
  AccountingHistorySummary,
  AccountingHistoryTable,
  AccountingHistoryUnpaidAlert,
} from "./OwnerAccountingHistoryParts";
import type {
  AccountingHistorySortField,
  AccountingHistorySortOrder,
} from "./OwnerAccountingHistoryParts";

const PAGE_SIZE = 10;

/**
 * ステータスの業務的優先度（昇順時の並び）。
 *   waiting (未精算) を最優先で確認したい → 上に出す
 *   cancelled は確認優先度が低い → 末尾
 * これは UI の表示順固有の都合で、永続化する仕様ではない。
 */
const STATUS_ORDER: Record<AccountingStatus, number> = {
  waiting: 0,
  pending: 1,
  completed: 2,
  cancelled: 3,
};

/** Accounting → 金額（completed 以外は 0 として扱う） */
function getAccountingAmount(acc: Accounting): number {
  if (acc.status !== "completed") return 0;
  return acc.payment?.billingAmount ?? acc.payment?.totalAmount ?? 0;
}

// URL クエリパラメータ のパース関数 (無効値はデフォルト値にフォールバック)
const SORT_FIELD_SET = new Set<string>(["date", "amount", "status"]);

function parseSortField(raw: string | null): AccountingHistorySortField {
  return raw !== null && SORT_FIELD_SET.has(raw) ? (raw as AccountingHistorySortField) : "date";
}
function parseSortOrder(raw: string | null): AccountingHistorySortOrder {
  return raw === "asc" ? "asc" : "desc";
}
function parsePage(raw: string | null): number {
  const n = Number(raw);
  return Number.isInteger(n) && n >= 1 ? n : 1;
}

interface OwnerAccountingHistoryProps {
  ownerId: string;
}

/**
 * 飼主詳細画面に表示する会計履歴セクション。
 *
 * - GET /api/v1/accountings?owner_id=... を直接叩いて履歴を取得する。
 * - 各行から `/accounting/:id` の詳細ページに遷移できる。
 *   完了済 (status === "completed") の会計は、詳細ページに既存の
 *   「明細兼領収書」プレビュー＆印刷ボタンが備わっているため、
 *   その導線を再利用する形で「明細兼領収書」リンクラベルを表示する。
 * - 新しい再発行 API は導入しない（既存の AccountingDocument プレビューを再利用）。
 */
export const OwnerAccountingHistory = memo(function OwnerAccountingHistory({
  ownerId,
}: OwnerAccountingHistoryProps) {
  const { data: accountings, isLoading, isError } = useGetAccountings({ ownerId });

  // URL クエリ同期: ah_sort / ah_order / ah_page
  // デフォルト値 (date/desc/1) は URL に出力しない（URLを汚さない）
  const [searchParams, setSearchParams] = useSearchParams();
  const sortField = parseSortField(searchParams.get("ah_sort"));
  const sortOrder = parseSortOrder(searchParams.get("ah_order"));
  const page = parsePage(searchParams.get("ah_page"));

  /**
   * 並び替え。
   *   - 第一キー: sortField + sortOrder
   *   - tie-break: 第二キーは常に「受付日降順 → ID 降順」（決定的）
   *
   * 金額キーは completed 以外を 0 として扱うため、未完了が金額順で
   * 先頭に固まる可能性がある。本仕様では tie-break 経路で受付日順に
   * 整列するためその範囲では決定性は保たれる。
   */
  const sortedAccountings = useMemo<Accounting[]>(() => {
    if (!accountings) return [];
    const dir = sortOrder === "asc" ? 1 : -1;
    return [...accountings].sort((a, b) => {
      if (sortField === "date") {
        const cmp = a.scheduledDate.localeCompare(b.scheduledDate) * dir;
        if (cmp !== 0) return cmp;
      } else if (sortField === "amount") {
        const cmp = (getAccountingAmount(a) - getAccountingAmount(b)) * dir;
        if (cmp !== 0) return cmp;
      } else {
        const cmp = (STATUS_ORDER[a.status] - STATUS_ORDER[b.status]) * dir;
        if (cmp !== 0) return cmp;
      }
      // tie-break: 受付日降順 → ID 降順（決定的に同値解消）
      if (a.scheduledDate !== b.scheduledDate) {
        return b.scheduledDate.localeCompare(a.scheduledDate);
      }
      return Number(b.id) - Number(a.id);
    });
  }, [accountings, sortField, sortOrder]);

  const handleSortFieldChange = useCallback(
    (v: string) => {
      setSearchParams(
        (prev) => {
          const next = new URLSearchParams(prev);
          if (v === "date") {
            next.delete("ah_sort");
          } else {
            next.set("ah_sort", v);
          }
          next.delete("ah_page");
          return next;
        },
        { replace: true },
      );
    },
    [setSearchParams],
  );

  const toggleSortOrder = useCallback(() => {
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev);
        const current = parseSortOrder(prev.get("ah_order"));
        const newOrder = current === "asc" ? "desc" : "asc";
        if (newOrder === "desc") {
          next.delete("ah_order");
        } else {
          next.set("ah_order", newOrder);
        }
        next.delete("ah_page");
        return next;
      },
      { replace: true },
    );
  }, [setSearchParams]);

  const setPage = useCallback(
    (updater: number | ((prev: number) => number)) => {
      setSearchParams(
        (prev) => {
          const currentPage = parsePage(prev.get("ah_page"));
          const newPage = typeof updater === "function" ? updater(currentPage) : updater;
          const next = new URLSearchParams(prev);
          if (newPage === 1) {
            next.delete("ah_page");
          } else {
            next.set("ah_page", String(newPage));
          }
          return next;
        },
        { replace: true },
      );
    },
    [setSearchParams],
  );

  // FE5-15: ページ数・切り出しは src/hooks/use-pagination.ts と同一アルゴリズム。
  // ah_page は URL クエリで永続化する要件があり、usePagination フックは内部 state のみで
  // 外部制御ページを受け付けないため、フック自体は使わず同じ計算式のみ踏襲する。
  const totalCount = sortedAccountings.length;
  const totalPages = Math.max(1, Math.ceil(totalCount / PAGE_SIZE));
  const pagedAccountings = sortedAccountings.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE);
  const startIndex = totalCount === 0 ? 0 : (page - 1) * PAGE_SIZE + 1;
  const endIndex = Math.min(page * PAGE_SIZE, totalCount);

  const firstUnpaidIndex = useMemo(
    () =>
      sortedAccountings.findIndex((acc) => acc.status === "waiting" || acc.status === "pending"),
    [sortedAccountings],
  );
  const firstUnpaidId = firstUnpaidIndex >= 0 ? sortedAccountings[firstUnpaidIndex].id : null;

  const firstUnpaidRowRef = useRef<HTMLTableRowElement | null>(null);
  const shouldScrollRef = useRef(false);

  const handleScrollToFirstUnpaid = useCallback(() => {
    if (firstUnpaidIndex < 0) return;
    const targetPage = Math.ceil((firstUnpaidIndex + 1) / PAGE_SIZE);
    if (targetPage !== page) {
      shouldScrollRef.current = true;
      setPage(targetPage);
    } else {
      firstUnpaidRowRef.current?.scrollIntoView({ behavior: "smooth", block: "nearest" });
      firstUnpaidRowRef.current?.focus();
    }
  }, [firstUnpaidIndex, page, setPage]);

  useEffect(() => {
    if (shouldScrollRef.current) {
      shouldScrollRef.current = false;
      firstUnpaidRowRef.current?.scrollIntoView({ behavior: "smooth", block: "nearest" });
      firstUnpaidRowRef.current?.focus();
    }
  }, [page]);

  /**
   * サマリ集計:
   *   - completedTotal: 累計支払い金額。`completed` の会計のみ集計。
   *     金額は row 表示と同じ `billingAmount ?? totalAmount ?? 0` を採用し、
   *     表上の数字と整合する。`cancelled` は除外（取消済は支払発生していない）。
   *   - completedCount: 累計集計の母数（行数表記用）。
   *   - unpaidCount: 未払い件数。`waiting` と `pending` を未払い扱いとし、
   *     `cancelled` は警告対象外。
   *
   * 空データ時はすべて 0 となり、summary strip と未払い警告は両方非表示にする。
   */
  const summary = useMemo(() => {
    let completedTotal = 0;
    let completedCount = 0;
    let unpaidCount = 0;
    // ソート結果に依らないので未ソートの accountings から計算する
    for (const acc of accountings ?? []) {
      if (acc.status === "completed") {
        completedTotal += acc.payment?.billingAmount ?? acc.payment?.totalAmount ?? 0;
        completedCount += 1;
      } else if (acc.status === "waiting" || acc.status === "pending") {
        unpaidCount += 1;
      }
    }
    return { completedTotal, completedCount, unpaidCount };
  }, [accountings]);

  if (isLoading) return <LoadingFallback />;
  if (isError) {
    return (
      <p className={`text-sm ${C.danger}`} role="alert">
        会計履歴の取得に失敗しました
      </p>
    );
  }

  if (sortedAccountings.length === 0) {
    return <p className={`text-sm ${C.text50} px-4 py-6 text-center`}>会計履歴はありません。</p>;
  }

  return (
    <div className="space-y-3">
      <AccountingHistorySummary
        completedCount={summary.completedCount}
        completedTotal={summary.completedTotal}
      />
      <AccountingHistoryUnpaidAlert
        unpaidCount={summary.unpaidCount}
        onScrollToFirstUnpaid={handleScrollToFirstUnpaid}
      />
      <AccountingHistorySortControls
        sortField={sortField}
        sortOrder={sortOrder}
        onSortFieldChange={handleSortFieldChange}
        onToggleSortOrder={toggleSortOrder}
      />
      <AccountingHistoryTable
        accountings={pagedAccountings}
        firstUnpaidId={firstUnpaidId}
        firstUnpaidRowRef={firstUnpaidRowRef}
      />
      {totalPages > 1 ? (
        <nav aria-label="ページネーション">
          <Pagination
            currentPage={page}
            totalPages={totalPages}
            totalCount={totalCount}
            startIndex={startIndex}
            endIndex={endIndex}
            onPageChange={setPage}
            onPrev={() => setPage((p) => Math.max(1, p - 1))}
            onNext={() => setPage((p) => Math.min(totalPages, p + 1))}
          />
        </nav>
      ) : null}
    </div>
  );
});
