import { type ComponentProps, useMemo } from "react";
import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";

import { DataTable } from "@/components/shared/DataTable/DataTable";
import { DataTableRowButton } from "@/components/shared/DataTable/DataTableRowButton";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { TableCell } from "@/components/ui/table";
import { C, STYLE } from "@/lib/design-tokens";

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

/** 動物種の詳細導線を保ち、編集権限がある場合だけ並べ替えと操作列を出す。 */
export function AnimalSpeciesSortableTable({
  items,
  sensors,
  onDragEnd,
  canEdit,
  onEdit,
}: AnimalSpeciesSortableTableProps) {
  // B-02 (EMR-409): 非編集者には空セルだけの「操作」列を出さず、ヘッダとセルを
  // まとめて消す。名称ボタン（詳細）とドラッグハンドルは読取専用でも維持する。
  const columns = useMemo(
    () => (canEdit ? COLUMNS : COLUMNS.filter((column) => column.header !== "操作")),
    [canEdit],
  );
  return (
    <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={onDragEnd}>
      <SortableContext items={items.map((item) => item.id)} strategy={verticalListSortingStrategy}>
        <DataTable
          headerRowClassName={STYLE.tableHeaderRow}
          headerCellClassName={STYLE.sectionLabel}
          columns={columns}
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
              {canEdit ? (
                <TableCell className="text-right">
                  <RowActionButton
                    aria-label={`編集: 動物種類 ${item.name} (ID ${item.id})`}
                    onClick={() => onEdit(item)}
                  />
                </TableCell>
              ) : null}
            </SortableDataTableRow>
          )}
        />
      </SortableContext>
    </DndContext>
  );
}

// eslint-disable-next-line react-refresh/only-export-components -- Settings ページが列定義を再利用する既存の共有 export（分離は別途検討）
export { COLUMNS as ANIMAL_SPECIES_COLUMNS };
