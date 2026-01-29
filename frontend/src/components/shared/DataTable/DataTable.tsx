import { ReactNode } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

interface DataTableProps<T> {
  columns: {
    header: string;
    className?: string;
    align?: "left" | "center" | "right";
  }[];
  data: T[];
  renderRow: (item: T) => ReactNode;
  emptyMessage?: string;
  className?: string;
}

export function DataTable<T>({
  columns,
  data,
  renderRow,
  emptyMessage = "データが見つかりません",
  className = "",
}: DataTableProps<T>) {
  return (
    <div className={`bg-[#fafafa] border border-[#e5e5e5] rounded-lg shadow-sm flex flex-col h-[calc(100vh-220px)] ${className}`}>
      <div className="flex-1 overflow-auto relative">
        <Table className="min-w-[800px]">
          <TableHeader className="sticky top-0 z-10 shadow-sm">
            <TableRow className="border-b-[#e5e5e5] bg-[#fafafa] hover:bg-[#fafafa] h-11">
              {columns.map((col, index) => (
                <TableHead
                  key={index}
                  className={`text-sm font-medium text-[#737373] h-11 ${
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
                  className="text-center py-12 text-[#37352F]/60 text-sm"
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
}
