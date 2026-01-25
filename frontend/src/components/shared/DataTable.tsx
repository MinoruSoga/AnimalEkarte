import { ReactNode } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../ui/table";

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
    <div className={`bg-white border border-[rgba(55,53,47,0.16)] rounded-md shadow-sm flex flex-col h-[calc(100vh-220px)] ${className}`}>
      <div className="flex-1 overflow-auto relative">
        <div className="min-w-[800px]">
        <Table>
          <TableHeader className="sticky top-0 z-10 shadow-sm">
            <TableRow className="border-b-[rgba(55,53,47,0.09)] bg-[#F7F6F3] hover:bg-[#F7F6F3] h-12">
              {columns.map((col, index) => (
                <TableHead
                  key={index}
                  className={`text-sm font-medium text-[#37352F]/80 h-12 ${
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
    </div>
  );
}
