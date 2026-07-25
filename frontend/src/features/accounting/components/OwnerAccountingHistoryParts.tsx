import { memo } from "react";
import type React from "react";
import { Link } from "react-router";
import { AlertTriangle, ArrowDown, ArrowUp, ChevronRight, FileText, Receipt } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { paths } from "@/config/paths";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { ACCOUNTING_STATUS_LABELS } from "@/constants/accounting-status";
import { formatCurrency } from "@/lib/format/number";
import type { Accounting } from "../api/transforms";

export type AccountingHistorySortField = "date" | "amount" | "status";
export type AccountingHistorySortOrder = "asc" | "desc";

interface AccountingHistorySummaryProps {
  completedCount: number;
  completedTotal: number;
}

export function AccountingHistorySummary({
  completedCount,
  completedTotal,
}: AccountingHistorySummaryProps) {
  if (completedCount === 0) return null;

  return (
    <div
      className={`flex items-center justify-between rounded-lg ${C.bgPage} px-4 py-3 border ${C.borderMedium}`}
      data-testid="accounting-history-summary"
    >
      <span className={`text-xs ${C.text60}`}>
        累計支払い金額（精算済 {completedCount} 件）
      </span>
      <span className={`text-base font-bold ${C.text} font-mono`}>
        {formatCurrency(completedTotal)}
      </span>
    </div>
  );
}

interface AccountingHistoryUnpaidAlertProps {
  unpaidCount: number;
  onScrollToFirstUnpaid: () => void;
}

export function AccountingHistoryUnpaidAlert({
  unpaidCount,
  onScrollToFirstUnpaid,
}: AccountingHistoryUnpaidAlertProps) {
  if (unpaidCount === 0) return null;

  return (
    <div
      role="alert"
      className={`flex items-center gap-2 rounded-lg ${C.bgWarning50} ${C.textWarning} px-4 border ${C.borderWarning20} text-sm`}
    >
      <AlertTriangle className={`${ICON.action} ${C.textWarningIcon} shrink-0`} aria-hidden />
      <button
        type="button"
        onClick={onScrollToFirstUnpaid}
        className="flex min-h-11 flex-1 items-center text-left bg-transparent cursor-pointer hover:underline underline-offset-2"
      >
        未払いの会計が <strong>{unpaidCount}</strong> 件あります。
        <span className="ml-1 font-medium">先頭の未払い行を確認する</span>
      </button>
    </div>
  );
}

interface AccountingHistorySortControlsProps {
  sortField: AccountingHistorySortField;
  sortOrder: AccountingHistorySortOrder;
  onSortFieldChange: (value: string) => void;
  onToggleSortOrder: () => void;
}

export function AccountingHistorySortControls({
  sortField,
  sortOrder,
  onSortFieldChange,
  onToggleSortOrder,
}: AccountingHistorySortControlsProps) {
  return (
    <div className="flex items-center justify-end gap-2">
      <Label htmlFor="accounting-history-sort-field" className={`text-xs ${C.text50}`}>
        並び替え
      </Label>
      <Select value={sortField} onValueChange={onSortFieldChange}>
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
        onClick={onToggleSortOrder}
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
  );
}

interface AccountingHistoryTableProps {
  accountings: Accounting[];
  firstUnpaidId: string | null;
  firstUnpaidRowRef: React.RefObject<HTMLTableRowElement | null>;
}

export function AccountingHistoryTable({
  accountings,
  firstUnpaidId,
  firstUnpaidRowRef,
}: AccountingHistoryTableProps) {
  return (
    <div className={`rounded-lg ${C.bgWhite} overflow-hidden border ${C.borderMedium}`}>
      <Table>
        <TableHeader>
          <TableRow className={`hover:bg-transparent ${C.bgPage} border-b ${C.borderMedium} h-12`}>
            <TableHead>受付日</TableHead>
            <TableHead>受付No</TableHead>
            <TableHead>ペット</TableHead>
            <TableHead>ステータス</TableHead>
            <TableHead className="text-right">金額</TableHead>
            <TableHead>操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {accountings.map((accounting) => (
            <AccountingHistoryRow
              key={accounting.id}
              accounting={accounting}
              ref={accounting.id === firstUnpaidId ? firstUnpaidRowRef : undefined}
              isScrollTarget={accounting.id === firstUnpaidId}
            />
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

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
          {ACCOUNTING_STATUS_LABELS[accounting.status] ?? accounting.status}
        </Badge>
      </TableCell>
      <TableCell className={`${STYLE.tableCell} text-right font-mono`}>
        {isCompleted ? formatCurrency(totalAmount) : "-"}
      </TableCell>
                    <TableCell>
        <div className="flex items-center justify-end gap-3">
          {isCompleted ? (
            <Link
              to={detailHref}
              className={`inline-flex min-h-11 min-w-11 items-center gap-1 text-xs ${C.textBrand} hover:underline`}
              aria-label={`受付No ${accounting.id} の明細兼領収書を表示`}
            >
              <Receipt className={ICON.action} />
              明細兼領収書
            </Link>
          ) : null}
          <Link
            to={detailHref}
            className={`inline-flex min-h-11 min-w-11 items-center gap-1 text-xs ${C.text65} hover:underline`}
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
