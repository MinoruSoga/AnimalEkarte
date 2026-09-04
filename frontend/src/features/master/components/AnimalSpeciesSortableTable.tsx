import { type ComponentProps } from "react";
import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";

import {
  DataTable,
  DESIGN_TABLE_HEADER_ROW,
  DESIGN_TABLE_HEADER_CELL,
} from "@/components/shared/DataTable/DataTable";
import { DataTableRowButton } from "@/components/shared/DataTable/DataTableRowButton";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { TableCell } from "@/components/ui/table";
import { C } from "@/lib/design-tokens";

import type { AnimalSpecies } from "../api/animal-species";

const COLUMNS = [
  { header: "", className: "w-11 px-0" },
  { header: "動物種類名" },
  { header: "ステータス", className: "w-[100px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

interface AnimalSpeciesSortableTableProps {
  items: AnimalSpecies[];
  sensors: ComponentProps<typeof DndContext>["sensors"];
  onDragEnd: NonNullable<ComponentProps<typeof DndContext>["onDragEnd"]>;
  canEdit: boolean;
  onEdit: (item: AnimalSpecies) => void;
}

export function AnimalSpeciesSortableTable({
  items,
  sensors,
  onDragEnd,
  canEdit,
  onEdit,
}: AnimalSpeciesSortableTableProps) {
  return (
    <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={onDragEnd}>
      <SortableContext items={items.map((item) => item.id)} strategy={verticalListSortingStrategy}>
        <DataTable
          headerRowClassName={DESIGN_TABLE_HEADER_ROW}
          headerCellClassName={DESIGN_TABLE_HEADER_CELL}
          columns={COLUMNS}
          data={items}
          emptyMessage="動物種類が登録されていません"
          renderRow={(item) => (
            <SortableDataTableRow
              key={item.id}
              id={item.id}
              dragLabel={`並べ替え: 動物種類 ${item.name} (ID ${item.id})`}
              dragDisabled={!canEdit}
            >
              <TableCell className={`font-medium ${C.text}`}>
                <DataTableRowButton
                  aria-label={`詳細: 動物種類 ${item.name} (ID ${item.id})`}
                  onClick={() => onEdit(item)}
                >
                  {item.name}
                </DataTableRowButton>
              </TableCell>
              <TableCell className="text-center">
                <StatusPill isActive={item.isActive} />
              </TableCell>
              <TableCell className="text-right">
                {canEdit ? (
                  <RowActionButton
                    aria-label={`編集: 動物種類 ${item.name} (ID ${item.id})`}
                    onClick={() => onEdit(item)}
                  />
                ) : null}
              </TableCell>
            </SortableDataTableRow>
          )}
        />
      </SortableContext>
    </DndContext>
  );
}

export { COLUMNS as ANIMAL_SPECIES_COLUMNS };
