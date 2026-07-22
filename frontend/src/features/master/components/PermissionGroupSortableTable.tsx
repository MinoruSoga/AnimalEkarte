import { type ComponentProps } from "react";
import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";

import { DataTable, DESIGN_TABLE_HEADER_ROW, DESIGN_TABLE_HEADER_CELL } from "@/components/shared/DataTable/DataTable";
import { DataTableRowButton } from "@/components/shared/DataTable/DataTableRowButton";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { TableCell } from "@/components/ui/table";
import { C } from "@/lib/design-tokens";

import type { PermissionGroup } from "../api/permission-groups";

const COLUMNS = [
  { header: "", className: "w-11 px-0" },
  { header: "グループ名", className: "flex-1" },
  { header: "ステータス", className: "w-[100px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

interface PermissionGroupSortableTableProps {
  items: PermissionGroup[];
  sensors: ComponentProps<typeof DndContext>["sensors"];
  onDragEnd: NonNullable<ComponentProps<typeof DndContext>["onDragEnd"]>;
  canEdit: boolean;
  onEdit: (item: PermissionGroup) => void;
}

export function PermissionGroupSortableTable({
  items,
  sensors,
  onDragEnd,
  canEdit,
  onEdit,
}: PermissionGroupSortableTableProps) {
  return (
    <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={onDragEnd}>
      <SortableContext items={items.map((item) => item.id)} strategy={verticalListSortingStrategy}>
        <DataTable
          headerRowClassName={DESIGN_TABLE_HEADER_ROW}
          headerCellClassName={DESIGN_TABLE_HEADER_CELL}
          columns={COLUMNS}
          data={items}
          emptyMessage="権限グループが登録されていません"
          renderRow={(item) => (
            <SortableDataTableRow
              key={item.id}
              id={item.id}
              dragLabel={`並べ替え: 権限グループ ${item.name} (ID ${item.id})`}
              dragDisabled={!canEdit}
            >
              <TableCell className={`font-medium ${C.text}`}>
                <DataTableRowButton
                  aria-label={`詳細: 権限グループ ${item.name} (ID ${item.id})`}
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
                    aria-label={`編集: 権限グループ ${item.name} (ID ${item.id})`}
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

export { COLUMNS as PERMISSION_GROUP_COLUMNS };
