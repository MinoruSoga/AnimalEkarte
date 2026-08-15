import { type ComponentProps, useMemo } from "react";
import { DndContext, DragOverlay, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { Plus } from "lucide-react";

import { RowActionButton } from "@/components/shared/RowActionButton";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { DataTableRowButton } from "@/components/shared/DataTable/DataTableRowButton";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { C, ICON, STYLE } from "@/lib/design-tokens";

import type { Cage } from "../api/cages";
import {
  CAGE_SIZE_LABELS,
  CAGE_TYPE_LABELS,
  formatCagePrice,
} from "./cage-side-panel-model";
import { CageRowOverlay } from "./CageRowOverlay";

const TABLE_COLUMNS = [
  { key: "grip", className: "w-11 px-0" },
  { key: "name", label: "ケージ名", className: "pl-3" },
  { key: "type", label: "エリア", className: "w-[100px]" },
  { key: "size", label: "サイズ", className: "w-[90px]" },
  { key: "price", label: "単価(税込)", className: "w-[120px] text-right pr-4" },
  { key: "status", label: "ステータス", className: "w-[100px] text-center" },
  { key: "action", label: "操作", className: "w-[80px] text-right pr-2" },
];

interface CageSortableTableProps {
  items: Cage[];
  sensors: ComponentProps<typeof DndContext>["sensors"];
  activeId: string | null;
  onDragStart: NonNullable<ComponentProps<typeof DndContext>["onDragStart"]>;
  onDragEnd: NonNullable<ComponentProps<typeof DndContext>["onDragEnd"]>;
  onDragCancel: NonNullable<ComponentProps<typeof DndContext>["onDragCancel"]>;
  canCreate: boolean;
  canEdit: boolean;
  onEdit: (item: Cage) => void;
  onNew: () => void;
}

export function CageSortableTable({
  items,
  sensors,
  activeId,
  onDragStart,
  onDragEnd,
  onDragCancel,
  canCreate,
  canEdit,
  onEdit,
  onNew,
}: CageSortableTableProps) {
  const activeItem = useMemo(
    () => items.find((item) => item.id === activeId) ?? null,
    [items, activeId],
  );

  return (
    <div className={STYLE.tableContainer}>
      <div className="flex-1 overflow-auto relative">
        <DndContext
          sensors={sensors}
          collisionDetection={closestCenter}
          onDragStart={onDragStart}
          onDragEnd={onDragEnd}
          onDragCancel={onDragCancel}
        >
          <Table>
            <TableHeader className="sticky top-0 z-10">
              <TableRow className={STYLE.tableHeaderRow}>
                {TABLE_COLUMNS.map((col) => (
                  <TableHead key={col.key} className={`${STYLE.tableHeaderCell} ${col.className}`}>
                    {col.label ?? ""}
                  </TableHead>
                ))}
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} className={STYLE.tableEmpty}>
                    ケージが登録されていません
                  </TableCell>
                </TableRow>
              ) : null}
              <SortableContext items={items.map((item) => item.id)} strategy={verticalListSortingStrategy}>
                {items.map((item) => (
                  <SortableDataTableRow
                    key={item.id}
                    id={item.id}
                    dragLabel={`並べ替え: ケージ ${item.name} (ID ${item.id})`}
                    dragDisabled={!canEdit}
                  >
                    <TableCell className={`font-medium ${C.text}`}>
                      <DataTableRowButton
                        aria-label={`詳細: ケージ ${item.name} (ID ${item.id})`}
                        onClick={() => onEdit(item)}
                      >
                        {item.name}
                      </DataTableRowButton>
                    </TableCell>
                    <TableCell className={C.text70}>
                      {CAGE_TYPE_LABELS[item.cageType] || item.cageType}
                    </TableCell>
                    <TableCell className={C.text70}>
                      {CAGE_SIZE_LABELS[item.cageSize] || item.cageSize}
                    </TableCell>
                    <TableCell className={`text-right font-mono ${C.text} pr-4`}>
                      {formatCagePrice(item.price)}
                    </TableCell>
                    <TableCell className="text-center">
                      <StatusPill isActive={item.isActive} />
                    </TableCell>
                    <TableCell className="text-right pr-2">
                      {canEdit ? (
                        <RowActionButton
                          aria-label={`編集: ケージ ${item.name} (ID ${item.id})`}
                          onClick={() => onEdit(item)}
                        />
                      ) : null}
                    </TableCell>
                  </SortableDataTableRow>
                ))}
              </SortableContext>
            </TableBody>
          </Table>
          <DragOverlay dropAnimation={null}>
            {activeItem ? <CageRowOverlay cage={activeItem} /> : null}
          </DragOverlay>
        </DndContext>
      </div>
      {canCreate ? (
        <button type="button" onClick={onNew} className={STYLE.inlineAddBtn}>
          <Plus className={ICON.xs} />
          新しいケージを追加...
        </button>
      ) : null}
    </div>
  );
}
