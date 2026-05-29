import { type ComponentProps } from "react";
import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";

import { DataTable } from "@/components/shared/DataTable/DataTable";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { TableCell } from "@/components/ui/table";
import { C } from "@/lib/design-tokens";

import type { AnimalSpecies } from "../api/animal-species";

const COLUMNS = [
  { header: "", className: "w-[32px]" },
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
          columns={COLUMNS}
          data={items}
          emptyMessage="動物種類が登録されていません"
          renderRow={(item) => (
            <SortableDataTableRow key={item.id} id={item.id} onClick={() => onEdit(item)}>
              <TableCell className={`font-medium text-base ${C.text}`}>{item.name}</TableCell>
              <TableCell className="text-center">
                <NotionStatusPill isActive={item.isActive} />
              </TableCell>
              <TableCell className="p-0 text-right">
                {canEdit ? <RowActionButton onClick={() => onEdit(item)} /> : null}
              </TableCell>
            </SortableDataTableRow>
          )}
        />
      </SortableContext>
    </DndContext>
  );
}

export { COLUMNS as ANIMAL_SPECIES_COLUMNS };
