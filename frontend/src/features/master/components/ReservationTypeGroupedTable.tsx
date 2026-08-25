import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useSortableList } from "@/hooks/use-sortable-list";
import { C, STYLE } from "@/lib/design-tokens";
import { useReorderReservationTypes } from "../api/reservation-types";
import type { ReservationType } from "../api/reservation-types";
import type { ReservationTypeGroup } from "../api/reservation-type-groups";
import { MASTER_TABLE_COL } from "../constants/styles";
import { ReservationTypeGroupedTableBody } from "./ReservationTypeGroupedTableBody";
import { groupReservationTypesByGroupId } from "./reservation-type-grouped-table-model";

interface ReservationTypeGroupedTableProps {
  groups: ReservationTypeGroup[];
  categories: ReservationType[];
  onCategoryEdit: (category: ReservationType) => void;
  onGroupEdit: (group: ReservationTypeGroup) => void;
  onCategoryAddInGroup: (groupId: string | undefined) => void;
  canEdit: boolean;
}

export function ReservationTypeGroupedTable({
  groups,
  categories,
  onCategoryEdit,
  onGroupEdit,
  onCategoryAddInGroup,
  canEdit,
}: ReservationTypeGroupedTableProps) {
  const [collapsed, setCollapsed] = useState<Set<string>>(new Set());
  const reorderMutation = useReorderReservationTypes();
  const resetOrderRef = useRef<() => void>(() => {});

  const handleReorder = useCallback(
    (newIds: string[]) => {
      if (!canEdit) return;
      reorderMutation.mutate(
        { ids: newIds.map(Number) },
        { onError: () => resetOrderRef.current() },
      );
    },
    [canEdit, reorderMutation],
  );

  const toggleCollapse = useCallback((id: string) => {
    setCollapsed((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  }, []);

  const { orderedItems, sensors, handleDragStart, handleDragEnd, handleDragCancel, resetOrder } =
    useSortableList({ items: categories, onReorder: handleReorder });

  useEffect(() => {
    resetOrderRef.current = resetOrder;
  }, [resetOrder]);

  const groupedCategories = useMemo(
    () => groupReservationTypesByGroupId(orderedItems),
    [orderedItems],
  );

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCenter}
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
      onDragCancel={handleDragCancel}
    >
      <SortableContext items={orderedItems.map((item) => item.id)} strategy={verticalListSortingStrategy}>
        <div className={`rounded-xs border ${C.borderLight} overflow-x-auto overflow-y-hidden ${C.bgWhite}`}>
          <table className="w-full border-collapse">
            <thead>
              <tr className={STYLE.tableHeaderRow}>
                <th data-c18-structural-cell className="w-11 px-0" />
                <th className={`text-left ${STYLE.tableHeaderCell}`}>名称</th>
                <th className={`text-left ${STYLE.tableHeaderCell} w-56`}>備考</th>
                <th className={`text-center ${STYLE.tableHeaderCell} ${MASTER_TABLE_COL.w100} whitespace-nowrap`}>ステータス</th>
                <th data-c18-structural-cell className="w-20" />
              </tr>
            </thead>
            <ReservationTypeGroupedTableBody
              groups={groups}
              groupedCategories={groupedCategories}
              collapsed={collapsed}
              canEdit={canEdit}
              onToggleCollapse={toggleCollapse}
              onCategoryEdit={onCategoryEdit}
              onGroupEdit={onGroupEdit}
              onCategoryAddInGroup={onCategoryAddInGroup}
            />
          </table>
        </div>
      </SortableContext>
    </DndContext>
  );
}
