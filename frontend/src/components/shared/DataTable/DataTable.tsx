import { memo, ReactNode } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { STYLE } from "@/lib/design-tokens";

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
}

export const DataTable = memo(function DataTable<T>({
  columns,
  data,
  renderRow,
  emptyMessage = "データが見つかりません",
  className = "",
}: DataTableProps<T>) {
  return (
    <div className={`${STYLE.tableContainer} ${className}`}>
      <div className="flex-1 overflow-auto relative">
        <Table className="min-w-[640px]">
          <TableHeader className="sticky top-0 z-10">
            <TableRow className={STYLE.tableHeaderRow}>
              {columns.map((col, index) => (
                <TableHead
                  key={index}
                  className={`${STYLE.tableHeaderCell} ${
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
