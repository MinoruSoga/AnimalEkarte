import { C, STYLE } from "@/lib/design-tokens";
import { TableCell, TableHead } from "@/components/ui/table";
import { formatJSTDateTimeLocal } from "@/lib/jst-date";
import { DataTableRowButton } from "@/components/shared/DataTable/DataTableRowButton";
import { formatCurrency } from "@/lib/format/number";
import type { CashRegisterClose } from "../api/get-cash-register-closes";
import { PERIOD_LABELS } from "../lib/constants";
import { diffClass, formatDiff } from "../lib/cash-register-history-model";

interface CashRegisterHistoryTableProps {
  rows: CashRegisterClose[];
  highlightDate: string | null;
  onSelect: (close: CashRegisterClose) => void;
}

export function CashRegisterHistoryTable({
  rows,
  highlightDate,
  onSelect,
}: CashRegisterHistoryTableProps) {
  if (rows.length === 0) {
    return <p className={`text-base ${C.text50} py-8 text-center`}>締め履歴がありません</p>;
  }

  return (
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
                  onClick={() => onSelect(close)}
                >
                  {close.closeDate.slice(0, 10)}
                </DataTableRowButton>
              </TableCell>
              <TableCell className={C.text}>{PERIOD_LABELS[close.period]}</TableCell>
              <TableCell className={`text-right ${C.text}`}>
                {formatCurrency(close.theoreticalCash ?? 0)}
              </TableCell>
              <TableCell className={`text-right ${C.text}`}>
                {formatCurrency(close.actualCash ?? 0)}
              </TableCell>
              <TableCell className={`text-right font-medium ${diffClass(diff)}`}>
                {formatDiff(diff)}
              </TableCell>
              <TableCell className={C.text}>{close.closedByStaffName ?? "—"}</TableCell>
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
  );
}
