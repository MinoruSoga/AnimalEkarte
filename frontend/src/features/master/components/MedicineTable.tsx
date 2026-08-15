import { Fragment, type ComponentProps } from "react";
import { DndContext, DragOverlay, closestCenter, type DragStartEvent, type DragEndEvent } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import Plus from "lucide-react/dist/esm/icons/plus";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ICON, STYLE } from "@/lib/design-tokens";
import type { Medicine } from "@/types";
import { MedicineCategoryHeaderRow, MedicineRowOverlay, SortableMedicineRow } from "./MedicineTableRows";

interface MedicineTableProps {
  sensors: ComponentProps<typeof DndContext>["sensors"];
  activeId: string | null;
  groupedMedicines: Map<string, { header: Medicine; items: Medicine[] }>;
  ungroupedMedicines: Medicine[];
  collapsedGroups: Set<string>;
  orderedMedicinesById: Map<string, Medicine>;
  canCreate: boolean;
  canEdit: boolean;
  onDragStart: (event: DragStartEvent) => void;
  onDragEnd: (event: DragEndEvent) => void;
  onDragCancel: () => void;
  onToggleGroup: (key: string) => void;
  onEdit: (medicine: Medicine) => void;
  onCreate: (parentId?: string) => void;
}

export function MedicineTable({
  sensors,
  activeId,
  groupedMedicines,
  ungroupedMedicines,
  collapsedGroups,
  orderedMedicinesById,
  canCreate,
  canEdit,
  onDragStart,
  onDragEnd,
  onDragCancel,
  onToggleGroup,
  onEdit,
  onCreate,
}: MedicineTableProps) {
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
                <TableHead className="w-11 px-0" />
                <TableHead className={`${STYLE.tableHeaderCell} pl-3`}>薬品名</TableHead>
                <TableHead className={`${STYLE.tableHeaderCell} w-[100px] text-center`}>剤形</TableHead>
                <TableHead className={`${STYLE.tableHeaderCell} w-[130px] text-right pr-4`}>
                  単価(税込)
                </TableHead>
                <TableHead className={`${STYLE.tableHeaderCell} w-[110px] text-center`}>
                  ステータス
                </TableHead>
                <TableHead className={`${STYLE.tableHeaderCell} w-[80px] text-center`}>操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {groupedMedicines.size === 0 && ungroupedMedicines.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} className={STYLE.tableEmpty}>
                    データが見つかりません
                  </TableCell>
                </TableRow>
              ) : null}

              {Array.from(groupedMedicines.entries()).map(([parentId, { header, items }]) => {
                const isCollapsed = collapsedGroups.has(parentId);

                return (
                  <Fragment key={parentId}>
                    <MedicineCategoryHeaderRow
                      parentId={parentId}
                      header={header}
                      itemCount={items.length}
                      isCollapsed={isCollapsed}
                      canCreate={canCreate}
                      canEdit={canEdit}
                      onToggleGroup={onToggleGroup}
                      onEdit={onEdit}
                      onCreate={onCreate}
                    />

                    {isCollapsed ? null : (
                      <SortableContext
                        items={items.map((medicine) => medicine.id)}
                        strategy={verticalListSortingStrategy}
                      >
                        {items.map((medicine) => (
                          <SortableMedicineRow
                            key={medicine.id}
                            medicine={medicine}
                            onEdit={onEdit}
                            grouped
                            canEdit={canEdit}
                          />
                        ))}
                      </SortableContext>
                    )}
                  </Fragment>
                );
              })}

              <SortableContext
                items={ungroupedMedicines.map((medicine) => medicine.id)}
                strategy={verticalListSortingStrategy}
              >
                {ungroupedMedicines.map((medicine) => (
                  <SortableMedicineRow
                    key={medicine.id}
                    medicine={medicine}
                    onEdit={onEdit}
                    grouped={false}
                    canEdit={canEdit}
                  />
                ))}
              </SortableContext>
            </TableBody>
          </Table>

          <DragOverlay dropAnimation={null}>
            {activeId ? (
              (() => {
                const medicine = orderedMedicinesById.get(activeId);
                if (!medicine) return null;
                return <MedicineRowOverlay medicine={medicine} grouped={Boolean(medicine.parentId)} />;
              })()
            ) : null}
          </DragOverlay>
        </DndContext>
      </div>

      {canCreate ? (
        <button
          type="button"
          onClick={() => onCreate()}
          className={STYLE.inlineAddBtn}
        >
          <Plus className={ICON.xs} />
          新しい薬剤を追加...
        </button>
      ) : null}
    </div>
  );
}
