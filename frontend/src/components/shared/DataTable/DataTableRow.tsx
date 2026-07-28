import { TableRow } from "@/components/ui/table";
import { cn } from "@/components/ui/utils";
import { TABLE_STYLES } from "@/lib/design-tokens";

type DataTableRowProps = Omit<React.ComponentProps<typeof TableRow>, "onClick"> & {
  children: React.ReactNode;
};

export function DataTableRow({ className, children, ...props }: DataTableRowProps) {
  const { onClick: ignoredLegacyOnClick, ...nonInteractiveProps } = props as typeof props & {
    onClick?: unknown;
  };
  void ignoredLegacyOnClick;

  return (
    <TableRow
      className={cn(TABLE_STYLES.row, "cursor-default", className)}
      {...nonInteractiveProps}
    >
      {children}
    </TableRow>
  );
}
