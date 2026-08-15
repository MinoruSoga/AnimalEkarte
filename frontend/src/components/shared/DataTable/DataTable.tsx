import { memo, ReactNode } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { C, STYLE } from "@/lib/design-tokens";

/**
 * DESIGN.md `ex-data-table-cell` 準拠の header chrome（eyebrow/sectionLabel + canvas-soft）。
 * headerRowClassName / headerCellClassName への opt-in 用共有定数。
 * MedicalRecords.tsx / OwnersListTable.tsx のローカル定数と同一値（コミット済み e93b53a2 パターン）。
 */
export const DESIGN_TABLE_HEADER_ROW = `border-b ${C.borderLight} ${C.bgPage} h-11`;
export const DESIGN_TABLE_HEADER_CELL = STYLE.sectionLabel;

interface DataTableProps<T> {
  columns: {
    header: ReactNode;
    className?: string;
    align?: "left" | "center" | "right";
  }[];
  data: T[];
  renderRow: (item: T) => ReactNode;
  emptyMessage?: string;
  className?: string;
  /**
   * DESIGN.md `ex-data-table-cell` の header 行/セル用クラスを丸ごと置き換える opt-in オーバーライド。
   * 既定（未指定）は STYLE.tableHeaderRow/tableHeaderCell を維持するため、既存の他画面の見た目は変わらない。
   * 併記ではなく置換であることに注意（Tailwind の同一 specificity クラス競合を避けるため）。
   */
  headerRowClassName?: string;
  headerCellClassName?: string;
}

export const DataTable = memo(function DataTable<T>({
  columns,
  data,
  renderRow,
  emptyMessage = "データが見つかりません",
  className = "",
  headerRowClassName,
  headerCellClassName,
}: DataTableProps<T>) {
  return (
    <div className={`${STYLE.tableContainer} ${className}`}>
      {/* min-w-0 + overflow-auto: narrow viewports scroll instead of clipping (BUG-458).
          Avoid fixed min-w-[640px] which forced off-screen status/action columns. */}
      <div className="relative min-w-0 flex-1 overflow-auto">
        <Table className="w-full min-w-0">
          <TableHeader className="sticky top-0 z-10">
            <TableRow className={headerRowClassName ?? STYLE.tableHeaderRow}>
              {columns.map((col, index) => (
                <TableHead
                  key={index}
                  className={`${headerCellClassName ?? STYLE.tableHeaderCell} ${
                    col.align === "right" ? "text-right" :
                    col.align === "center" ? "text-center" : ""
                  } ${col.className || ""}`}
                >
                  {col.header}
                </TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {data.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={columns.length}
                  className={STYLE.tableEmpty}
                >
                  {emptyMessage}
                </TableCell>
              </TableRow>
            ) : (
              data.map((item) => renderRow(item))
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}) as <T>(props: DataTableProps<T>) => React.ReactElement;
