import { C, ICON } from "@/lib/design-tokens";
import type { ReactNode } from "react";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import GripVertical from "lucide-react/dist/esm/icons/grip-vertical";
import { TableCell } from "@/components/ui/table";
import { DataTableRow } from "./DataTableRow";

interface SortableDataTableRowProps {
  id: string | number;
  onClick?: () => void;
  children: ReactNode;
  /** Additional classes passed through to DataTableRow (e.g. `"group/row"` for CSS group hover). */
  className?: string;
  /** Override grip cell className. Default: `"w-[32px] ${C.text20} cursor-grab"`. */
  gripClassName?: string;
  /** Opacity applied to the row while dragging. Default: `0.5`. */
  isDraggingOpacity?: number;
}

/**
 * Sortable row for DataTable with DnD support.
 * Wraps DataTableRow with useSortable and prepends a GripVertical handle cell.
 * Parent must render inside DndContext + SortableContext from @dnd-kit.
 *
 * Column definition must include a leading column: `{ header: "", className: "w-[32px]" }`
 *
 * Optional customization props:
 * - `className` — forwarded to DataTableRow for additional classes (e.g. `"group/row"`).
 * - `gripClassName` — overrides the grip cell className.
 * - `isDraggingOpacity` — controls row opacity while dragging (default: `0.5`).
 */
export function SortableDataTableRow({
  id,
  onClick,
  children,
  className,
  gripClassName = `w-[32px] ${C.text20} cursor-grab`,
  isDraggingOpacity = 0.5,
}: SortableDataTableRowProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } =
    useSortable({ id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? isDraggingOpacity : 1,
  };

  return (
    <DataTableRow ref={setNodeRef} style={style} {...attributes} onClick={onClick} className={className}>
      <TableCell className={gripClassName} {...listeners}>
        <GripVertical className={ICON.action} />
      </TableCell>
      {children}
    </DataTableRow>
  );
}
