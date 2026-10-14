import { type ComponentProps, useMemo } from "react";
import { DndContext, DragOverlay, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { ArrowDown, ArrowUp } from "lucide-react";
import { Button } from "@/components/ui/button";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { Pagination } from "@/components/shared/Pagination";
import { usePagination } from "@/hooks/use-pagination";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { DataTableRowButton } from "@/components/shared/DataTable/DataTableRowButton";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { C, STYLE, TABLE_STYLES, ICON } from "@/lib/design-tokens";
import { formatCurrency } from "@/lib/format/number";
import { type MerchandiseMoveDirection } from "../lib/merchandise-reorder";

import type { FrontendMerchandiseItem } from "../api/merchandise-items";
import {
  formatMerchandiseTaxRate,
  MERCHANDISE_CATEGORY_LABELS,
} from "../lib/merchandise-side-panel-model";
import { MerchandiseRowOverlay } from "./MerchandiseRowOverlay";

const TABLE_COLUMNS = [
  { key: "grip", className: "w-11 px-0" },
  { key: "name", label: "品目名", className: "pl-3" },
  { key: "category", label: "カテゴリ", className: "w-[90px] text-center" },
  { key: "price", label: "単価(税込)", className: "w-[120px] text-right pr-4" },
  { key: "taxRate", label: "税率", className: "w-[80px] text-center" },
  { key: "status", label: "ステータス", className: "w-[100px] text-center" },
  // PR665: 移動2 + 編集1 の44pxボタンが1行に収まる幅（パディング含め160px）。
  { key: "action", label: "操作", className: "w-[160px] min-w-[160px] text-right pr-2" },
];

interface MerchandiseSortableTableProps {
  items: FrontendMerchandiseItem[];
  sensors: ComponentProps<typeof DndContext>["sensors"];
  activeId: string | null;
  onDragStart: NonNullable<ComponentProps<typeof DndContext>["onDragStart"]>;
  onDragEnd: NonNullable<ComponentProps<typeof DndContext>["onDragEnd"]>;
  onDragCancel: NonNullable<ComponentProps<typeof DndContext>["onDragCancel"]>;
  canEdit: boolean;
  onEdit: (item: FrontendMerchandiseItem) => void;
  /**
   * PR665 / EMR-401 C2-5: 50 行ページング維持のため、ページをまたぐ並べ替えは
   * 明示的な「前へ / 後へ」ボタンで行う。未指定または !canEdit では表示しない。
   */
  onMoveItem?: (item: FrontendMerchandiseItem, direction: MerchandiseMoveDirection) => void;
  /** PR665: 並べ替え API 実行中は移動ボタンを無効化する。 */
  isReordering?: boolean;
}

export function MerchandiseSortableTable({
  items,
  sensors,
  activeId,
  onDragStart,
  onDragEnd,
  onDragCancel,
  canEdit,
  onEdit,
  onMoveItem,
  isReordering = false,
}: MerchandiseSortableTableProps) {
  const pagination = usePagination(items, { pageSize: 50 });
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
                    品目が登録されていません
                  </TableCell>
                </TableRow>
              ) : null}
              <SortableContext
                items={pagination.paginatedData.map((item) => item.id)}
                strategy={verticalListSortingStrategy}
              >
                {pagination.paginatedData.map((item) => (
                  <SortableDataTableRow
                    key={item.id}
                    id={item.id}
                    dragLabel={`並べ替え: 品目 ${item.name} (ID ${item.id})`}
                    dragDisabled={!canEdit}
                  >
                    <TableCell className={`font-medium ${C.text}`}>
                      <DataTableRowButton
                        aria-label={`詳細: 品目 ${item.name} (ID ${item.id})`}
                        onClick={() => onEdit(item)}
                      >
                        {item.name}
                      </DataTableRowButton>
                    </TableCell>
                    <TableCell className={`${C.text70} text-center`}>
                      {MERCHANDISE_CATEGORY_LABELS[item.category] ?? item.category}
                    </TableCell>
                    <TableCell className={`text-right font-mono ${C.text} pr-4`}>
                      {formatCurrency(item.unitPrice)}
                    </TableCell>
                    <TableCell className={`${C.text70} text-center`}>
                      {formatMerchandiseTaxRate(item.taxRate)}
                    </TableCell>
                    <TableCell className="text-center">
                      <StatusPill isActive={item.isActive} />
                    </TableCell>
                    <TableCell className="text-right pr-2 whitespace-nowrap">
                      {canEdit ? (
                        <div className="flex items-center justify-end gap-1">
                          {onMoveItem ? (
                            <>
                              {/* PR665: 移動ボタンは並べ替え中に無効化する（共有 RowActionButton は disabled を持たないため直接 Button を使う）。 */}
                              <Button
                                variant="ghost"
                                size="icon"
                                aria-label={`前へ移動: 品目 ${item.name} (ID ${item.id})`}
                                disabled={isReordering}
                                className={TABLE_STYLES.actionButton}
                                onClick={(e) => {
                                  e.stopPropagation();
                                  onMoveItem(item, "up");
                                }}
                              >
                                <ArrowUp className={ICON.page} />
                              </Button>
                              <Button
                                variant="ghost"
                                size="icon"
                                aria-label={`後へ移動: 品目 ${item.name} (ID ${item.id})`}
                                disabled={isReordering}
                                className={TABLE_STYLES.actionButton}
                                onClick={(e) => {
                                  e.stopPropagation();
                                  onMoveItem(item, "down");
                                }}
                              >
                                <ArrowDown className={ICON.page} />
                              </Button>
                            </>
                          ) : null}
                          <RowActionButton
                            aria-label={`編集: 品目 ${item.name} (ID ${item.id})`}
                            onClick={() => onEdit(item)}
                          />
                        </div>
                      ) : null}
                    </TableCell>
                  </SortableDataTableRow>
                ))}
              </SortableContext>
            </TableBody>
          </Table>
          <DragOverlay dropAnimation={null}>
            {activeItem ? <MerchandiseRowOverlay item={activeItem} /> : null}
          </DragOverlay>
        </DndContext>
      </div>
      <Pagination
        currentPage={pagination.currentPage}
        totalPages={pagination.totalPages}
        totalCount={pagination.totalCount}
        startIndex={pagination.startIndex}
        endIndex={pagination.endIndex}
        onPageChange={pagination.goToPage}
        onPrev={pagination.prevPage}
        onNext={pagination.nextPage}
      />
    </div>
  );
}
