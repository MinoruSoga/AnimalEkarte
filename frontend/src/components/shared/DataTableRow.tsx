import { TableRow } from "../ui/table";
import { cn } from "../ui/utils";
import { TABLE_STYLES } from "../../lib/design-tokens";

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
