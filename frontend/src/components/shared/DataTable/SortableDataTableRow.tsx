import { memo, type ReactNode } from "react";
import type { UniqueIdentifier } from "@dnd-kit/core";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import GripVertical from "lucide-react/dist/esm/icons/grip-vertical";

import { TableCell } from "@/components/ui/table";
import { C, ICON } from "@/lib/design-tokens";
import { DataTableRow } from "./DataTableRow";

interface SortableDataTableRowProps {
  id: UniqueIdentifier;
  dragLabel: string;
  dragDisabled: boolean;
  children: ReactNode;
  /** Additional classes passed through to DataTableRow (e.g. `"group/row"` for CSS group hover). */
  className?: string;
  /** Opacity applied to the row while dragging. Default: `0.5`. */
  isDraggingOpacity?: number;
}

/**
 * Sortable row for DataTable with DnD support.
 * Wraps DataTableRow with useSortable and prepends a GripVertical handle cell.
 * Parent must render inside DndContext + SortableContext from @dnd-kit.
 *
 * Column definition must include a leading column: `{ header: "", className: "w-11 px-0" }`
 *
 * Optional customization props:
 * - `className` — forwarded to DataTableRow for additional classes (e.g. `"group/row"`).
 * - `isDraggingOpacity` — controls row opacity while dragging (default: `0.5`).
 */
export const SortableDataTableRow = memo(function SortableDataTableRow({
  id,
  dragLabel,
  dragDisabled,
  children,
  className,
  isDraggingOpacity = 0.5,
}: SortableDataTableRowProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    setActivatorNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable(dragDisabled ? { id, disabled: true } : { id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? isDraggingOpacity : 1,
  };

  return (
    <DataTableRow ref={setNodeRef} style={style} className={className}>
      <TableCell className="w-11 px-0">
        <button
          ref={setActivatorNodeRef}
          type="button"
          {...attributes}
          {...listeners}
          aria-label={dragLabel}
          disabled={dragDisabled}
          className={`flex min-h-11 min-w-11 touch-none items-center justify-center rounded-xs ${C.text20} ${C.hoverBgMedium} outline-none focus-visible:ring-2 ${C.focusRingAccent40} disabled:cursor-default disabled:opacity-30 ${dragDisabled ? "cursor-default" : "cursor-grab active:cursor-grabbing"}`}
        >
          <GripVertical className={ICON.action} aria-hidden="true" />
        </button>
      </TableCell>
      {children}
    </DataTableRow>
  );
});
