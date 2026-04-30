// React/Framework
import { memo, useMemo, useCallback, useRef, useEffect } from "react";
import { Link, useSearchParams } from "react-router";

// External
import { Receipt, FileText, ChevronLeft, ChevronRight, AlertTriangle, ArrowDown, ArrowUp } from "lucide-react";

// Internal
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Label } from "@/components/ui/label";
import { LoadingFallback } from "@/components/shared/DataStates";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { paths } from "@/config/paths";

// Relative
import { useGetAccountings } from "../api/get-accountings";
import type { Accounting } from "../api/transforms";
import type { AccountingStatus } from "../types";

const PAGE_SIZE = 10;

/** ソート可能なフィールド */
type SortField = "date" | "amount" | "status";
type SortOrder = "asc" | "desc";

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

function parseSortField(raw: string | null): SortField {
  return raw !== null && SORT_FIELD_SET.has(raw) ? (raw as SortField) : "date";
}
function parseSortOrder(raw: string | null): SortOrder {
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
      let primary = 0;
      if (sortField === "date") {
        primary = a.scheduledDate.localeCompare(b.scheduledDate) * dir;
      } else if (sortField === "amount") {
        primary = (getAccountingAmount(a) - getAccountingAmount(b)) * dir;
      } else {
        primary = (STATUS_ORDER[a.status] - STATUS_ORDER[b.status]) * dir;
      }
      if (primary !== 0) return primary;
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
          const newPage =
            typeof updater === "function" ? updater(currentPage) : updater;
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

  const totalPages = Math.max(1, Math.ceil(sortedAccountings.length / PAGE_SIZE));
  const pagedAccountings = sortedAccountings.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE);

  const firstUnpaidIndex = useMemo(
    () =>
      sortedAccountings.findIndex(
        (acc) => acc.status === "waiting" || acc.status === "pending",
      ),
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
    return (
      <p className={`text-sm ${C.text50} px-4 py-6 text-center`}>
        会計履歴はありません。
      </p>
    );
  }

  return (
    <div className="space-y-3">
      {/* 累計支払い金額 — 完了済 1 件以上のときだけ表示 */}
      {summary.completedCount > 0 ? (
        <div
          className={`flex items-center justify-between rounded-lg ${C.bgPage} px-4 py-3 border ${C.borderMedium}`}
          data-testid="accounting-history-summary"
        >
          <span className={`text-xs ${C.text60}`}>
            累計支払い金額（精算済 {summary.completedCount} 件）
          </span>
          <span className={`text-base font-bold ${C.text} font-mono`}>
            ¥{summary.completedTotal.toLocaleString()}
          </span>
        </div>
      ) : null}

      {/* 未払い警告 — waiting / pending が 1 件でもあれば表示。クリックで先頭未払い行へスクロール */}
      {summary.unpaidCount > 0 ? (
        <div
          role="alert"
          className={`flex items-center gap-2 rounded-lg ${C.bgWarning50} ${C.textWarning} px-4 py-3 border ${C.borderWarning20} text-sm`}
        >
          <AlertTriangle className={`${ICON.action} ${C.textWarningIcon} shrink-0`} aria-hidden />
          <button
            type="button"
            onClick={handleScrollToFirstUnpaid}
            className="flex-1 text-left bg-transparent cursor-pointer hover:underline underline-offset-2"
          >
            未払いの会計が <strong>{summary.unpaidCount}</strong> 件あります。
            <span className="ml-1 font-medium">先頭の未払い行を確認する</span>
          </button>
        </div>
      ) : null}

      {/* ソート切替 — 履歴テーブルのすぐ上に配置。
          フィールドは Select、方向は Toggle ボタン。*/}
      <div className="flex items-center justify-end gap-2">
        <Label
          htmlFor="accounting-history-sort-field"
          className={`text-xs ${C.text50}`}
        >
          並び替え
        </Label>
        <Select
          value={sortField}
          onValueChange={handleSortFieldChange}
        >
          <SelectTrigger
            id="accounting-history-sort-field"
            className="h-8 w-[120px] text-xs"
            aria-label="ソート項目"
          >
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="date">受付日</SelectItem>
            <SelectItem value="amount">金額</SelectItem>
            <SelectItem value="status">ステータス</SelectItem>
          </SelectContent>
        </Select>
        <Button
          type="button"
          variant="outline"
          size="sm"
          className="h-8 px-2"
          onClick={toggleSortOrder}
          aria-label={sortOrder === "asc" ? "昇順 — クリックで降順に切替" : "降順 — クリックで昇順に切替"}
          aria-pressed={sortOrder === "asc"}
        >
          {sortOrder === "asc" ? (
            <ArrowUp className={ICON.action} />
          ) : (
            <ArrowDown className={ICON.action} />
          )}
        </Button>
      </div>

      <div className={`rounded-lg ${C.bgWhite} overflow-hidden border ${C.borderMedium}`}>
        <Table>
          <TableHeader>
            <TableRow className={`hover:bg-transparent ${C.bgPage} border-b ${C.borderMedium} h-12`}>
              <TableHead className={STYLE.tableCellMuted}>受付日</TableHead>
              <TableHead className={STYLE.tableCellMuted}>受付No</TableHead>
              <TableHead className={STYLE.tableCellMuted}>ペット</TableHead>
              <TableHead className={STYLE.tableCellMuted}>ステータス</TableHead>
              <TableHead className={`${STYLE.tableCellMuted} text-right`}>金額</TableHead>
              <TableHead className={STYLE.tableCellMuted}>操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {pagedAccountings.map((acc) => (
              <AccountingHistoryRow
                key={acc.id}
                accounting={acc}
                ref={acc.id === firstUnpaidId ? firstUnpaidRowRef : undefined}
                isScrollTarget={acc.id === firstUnpaidId}
              />
            ))}
          </TableBody>
        </Table>
      </div>

      {totalPages > 1 ? (
        <nav
          className="flex items-center justify-center gap-3 pt-1"
          role="navigation"
          aria-label="ページネーション"
        >
          <Button
            type="button"
            variant="outline"
            size="sm"
            className="h-8 px-2"
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={page <= 1}
            aria-label="前のページ"
          >
            <ChevronLeft className={ICON.action} />
            前へ
          </Button>
          <span className={`text-xs ${C.text50}`} aria-live="polite">
            {page} / {totalPages} ページ
          </span>
          <Button
            type="button"
            variant="outline"
            size="sm"
            className="h-8 px-2"
            onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
            disabled={page >= totalPages}
            aria-label="次のページ"
          >
            次へ
            <ChevronRight className={ICON.action} />
          </Button>
        </nav>
      ) : null}
    </div>
  );
});

const STATUS_LABEL: Record<AccountingStatus, string> = {
  waiting: "未精算",
  pending: "保留",
  completed: "精算済",
  cancelled: "取消",
};

const AccountingHistoryRow = memo(function AccountingHistoryRow({
  accounting,
  ref,
  isScrollTarget = false,
}: {
  accounting: Accounting;
  ref?: React.Ref<HTMLTableRowElement>;
  isScrollTarget?: boolean;
}) {
  const detailHref = paths.accounting.detail.getHref(accounting.id);
  const totalAmount = accounting.payment?.billingAmount ?? accounting.payment?.totalAmount ?? 0;
  const isCompleted = accounting.status === "completed";

  return (
    <TableRow
      ref={ref}
      tabIndex={isScrollTarget ? -1 : undefined}
      data-testid={isScrollTarget ? "first-unpaid-row" : undefined}
      className={`transition-colors ${C.borderDivider} ${C.hoverBgPage} h-12`}
    >
      <TableCell className={STYLE.tableCell}>
        {accounting.scheduledDate || "-"}
      </TableCell>
      <TableCell className={STYLE.tableCell}>{accounting.id}</TableCell>
      <TableCell className={STYLE.tableCell}>{accounting.petName || "-"}</TableCell>
      <TableCell className={STYLE.tableCell}>
        <Badge
          variant={isCompleted ? "default" : "outline"}
          className="font-normal text-xs"
        >
          {STATUS_LABEL[accounting.status] ?? accounting.status}
        </Badge>
      </TableCell>
      <TableCell className={`${STYLE.tableCell} text-right font-mono`}>
        {isCompleted ? `¥${totalAmount.toLocaleString()}` : "-"}
      </TableCell>
      <TableCell className="py-2">
        <div className="flex items-center justify-end gap-3">
          {isCompleted ? (
            <Link
              to={detailHref}
              className={`inline-flex items-center gap-1 text-xs ${C.textBrand} hover:underline`}
              aria-label={`受付No ${accounting.id} の明細兼領収書を表示`}
            >
              <Receipt className={ICON.action} />
              明細兼領収書
            </Link>
          ) : null}
          <Link
            to={detailHref}
            className={`inline-flex items-center gap-1 text-xs ${C.text65} hover:underline`}
            aria-label={`受付No ${accounting.id} の会計詳細を開く`}
          >
            <FileText className={ICON.action} />
            詳細
            <ChevronRight className={ICON.action} />
          </Link>
        </div>
      </TableCell>
    </TableRow>
  );
});
