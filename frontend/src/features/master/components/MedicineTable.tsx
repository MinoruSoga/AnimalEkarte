import { Fragment, type ComponentProps } from "react";
import { DndContext, DragOverlay, closestCenter, type DragStartEvent, type DragEndEvent } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import ChevronRight from "lucide-react/dist/esm/icons/chevron-right";
import GripVertical from "lucide-react/dist/esm/icons/grip-vertical";
import MoreHorizontal from "lucide-react/dist/esm/icons/more-horizontal";
import Plus from "lucide-react/dist/esm/icons/plus";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import type { Medicine } from "@/types";
import { MedicineRowOverlay, SortableMedicineRow } from "./MedicineTableRows";

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
                <TableHead className="w-8 px-0" />
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
                    <TableRow
                      onClick={() => onEdit(header)}
                      className={`${STYLE.tableRow} border-b ${C.borderLight} ${C.bgPage30} group/header ${C.hoverBgPage60}`}
                    >
                      <TableCell className="w-8 px-0 py-0">
                        <button
                          type="button"
                          tabIndex={-1}
                          onClick={(event) => event.stopPropagation()}
                          className={`${STYLE.iconBtn32} ${C.text20} ${C.hoverBgMedium} ${C.hoverText60} cursor-grab`}
                        >
                          <GripVertical className={ICON.action} />
                        </button>
                      </TableCell>

                      <TableCell className="py-0 pl-0 pr-2">
                        <div className="flex items-center">
                          <button
                            type="button"
                            onClick={(event) => {
                              event.stopPropagation();
                              onToggleGroup(parentId);
                            }}
                            className={`flex items-center gap-1.5 py-1.5 px-1 ${C.hoverBgLight} rounded-[3px] transition-colors`}
                          >
                            <ChevronRight
                              className={`${ICON.xs} ${C.text50} transition-transform duration-150 ${
                                isCollapsed ? "" : "rotate-90"
                              }`}
                            />
                            <span className={`text-base font-medium ${C.text65}`}>
                              {header.name}
                            </span>
                            <span className={`text-base ${C.text40} ml-0.5`}>{items.length}</span>
                          </button>
                          <div className="flex-1" />
                          {canCreate ? (
                            <button
                              type="button"
                              onClick={(event) => {
                                event.stopPropagation();
                                onCreate(parentId);
                              }}
                              className={`${STYLE.iconBtn32} ${C.text40} ${C.hoverBgMedium} ${C.hoverText} opacity-0 group-hover/header:opacity-100`}
                            >
                              <Plus className={ICON.xs} />
                            </button>
                          ) : null}
                        </div>
                      </TableCell>

                      <TableCell className="w-[100px] py-0" />
                      <TableCell className="w-[130px] py-0" />
                      <TableCell className="w-[110px] py-0 text-center">
                        <NotionStatusPill isActive={true} />
                      </TableCell>
                      <TableCell className="w-[80px] py-0 text-center" onClick={(event) => event.stopPropagation()}>
                        {canEdit ? (
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <button
                                type="button"
                                className={`${STYLE.iconBtn28} ${C.text40} ${C.hoverBgMedium} ${C.hoverText} opacity-0 group-hover/header:opacity-100`}
                              >
                                <MoreHorizontal className={ICON.action} />
                              </button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              <DropdownMenuItem onClick={() => onEdit(header)}>編集</DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        ) : null}
                      </TableCell>
                    </TableRow>

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
