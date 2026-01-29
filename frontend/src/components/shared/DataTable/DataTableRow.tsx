import { TableRow } from "@/components/ui/table";
import { cn } from "@/components/ui/utils";
import { TABLE_STYLES } from "@/config/design-tokens";

interface DataTableRowProps extends React.ComponentProps<typeof TableRow> {
  children: React.ReactNode;
}

export function DataTableRow({ className, children, ...props }: DataTableRowProps) {
  return (
    <TableRow
      className={cn(TABLE_STYLES.row, className)}
      {...props}
    >
      {children}
    </TableRow>
  );
}
