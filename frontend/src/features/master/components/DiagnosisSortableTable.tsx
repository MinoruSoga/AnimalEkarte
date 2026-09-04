import { type ComponentProps, type ReactNode } from "react";
import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";

import {
  DataTable,
  DESIGN_TABLE_HEADER_ROW,
  DESIGN_TABLE_HEADER_CELL,
} from "@/components/shared/DataTable/DataTable";

interface DiagnosisSortableTableProps<T extends { id: string }> {
  items: T[];
  columns: {
    header: ReactNode;
    className?: string;
    align?: "left" | "center" | "right";
  }[];
  emptyMessage: string;
  sensors: ComponentProps<typeof DndContext>["sensors"];
  onDragEnd: NonNullable<ComponentProps<typeof DndContext>["onDragEnd"]>;
  renderRow: (item: T) => ReactNode;
}

export function DiagnosisSortableTable<T extends { id: string }>({
  items,
  columns,
  emptyMessage,
  sensors,
  onDragEnd,
  renderRow,
}: DiagnosisSortableTableProps<T>) {
  return (
    <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={onDragEnd}>
      <SortableContext items={items.map((item) => item.id)} strategy={verticalListSortingStrategy}>
        <DataTable
          headerRowClassName={DESIGN_TABLE_HEADER_ROW}
          headerCellClassName={DESIGN_TABLE_HEADER_CELL}
          columns={columns}
          data={items}
          emptyMessage={emptyMessage}
          renderRow={renderRow}
        />
      </SortableContext>
    </DndContext>
  );
}
