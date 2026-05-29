import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { Fragment, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useSortableList } from "@/hooks/use-sortable-list";
import { C, STYLE } from "@/lib/design-tokens";
import { useReorderReservationTypes } from "../api/reservation-types";
import type { ReservationType } from "../api/reservation-types";
import type { ReservationTypeGroup } from "../api/reservation-type-groups";
import {
  ReservationTypeEmptyGroupRow,
  ReservationTypeGroupHeader,
  ReservationTypeRow,
  ReservationTypeUncategorizedHeader,
} from "./ReservationTypeGroupedTableRows";

const UNCATEGORIZED_ID = "__uncategorized__";

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
      reorderMutation.mutate(
        { ids: newIds.map(Number) },
        { onError: () => resetOrderRef.current() },
      );
    },
    [reorderMutation],
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

  const categoriesByGroupId = useMemo(() => {
    const map = new Map<string, ReservationType[]>();
    const uncategorized: ReservationType[] = [];
    for (const category of orderedItems) {
      if (category.groupId) {
        const groupCategories = map.get(category.groupId) ?? [];
        groupCategories.push(category);
        map.set(category.groupId, groupCategories);
      } else {
        uncategorized.push(category);
      }
    }
    return { map, uncategorized };
  }, [orderedItems]);

  const uncategorizedCategories = categoriesByGroupId.uncategorized;
  const uncategorizedCollapsed = collapsed.has(UNCATEGORIZED_ID);

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCenter}
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
      onDragCancel={handleDragCancel}
    >
      <SortableContext items={orderedItems.map((item) => item.id)} strategy={verticalListSortingStrategy}>
        <div className={`rounded-[4px] border ${C.borderLight} overflow-hidden ${C.bgWhite}`}>
          <table className="w-full border-collapse">
            <thead>
              <tr className={STYLE.tableHeaderRow}>
                <th className="w-8" />
                <th className={`text-left ${STYLE.tableHeaderCell} px-3`}>名称</th>
                <th className={`text-left ${STYLE.tableHeaderCell} px-3 w-56`}>備考</th>
                <th className={`text-center ${STYLE.tableHeaderCell} px-3 w-24 whitespace-nowrap`}>ステータス</th>
                <th className="w-20" />
              </tr>
            </thead>
            <tbody>
              {groups.map((group) => {
                const groupCategories = categoriesByGroupId.map.get(group.id) ?? [];
                const isCollapsed = collapsed.has(group.id);
                return (
                  <Fragment key={group.id}>
                    <ReservationTypeGroupHeader
                      group={group}
                      count={groupCategories.length}
                      isCollapsed={isCollapsed}
                      canEdit={canEdit}
                      onToggle={() => toggleCollapse(group.id)}
                      onGroupEdit={() => onGroupEdit(group)}
                      onCategoryAdd={() => onCategoryAddInGroup(group.id)}
                    />
                    {!isCollapsed ? (
                      groupCategories.length > 0 ? (
                        groupCategories.map((category) => (
                          <ReservationTypeRow
                            key={category.id}
                            category={category}
                            canEdit={canEdit}
                            onEdit={() => onCategoryEdit(category)}
                          />
                        ))
                      ) : (
                        <ReservationTypeEmptyGroupRow />
                      )
                    ) : null}
                  </Fragment>
                );
              })}

              {uncategorizedCategories.length > 0 || groups.length === 0 ? (
                <Fragment key={UNCATEGORIZED_ID}>
                  <ReservationTypeUncategorizedHeader
                    count={uncategorizedCategories.length}
                    isCollapsed={uncategorizedCollapsed}
                    canEdit={canEdit}
                    onToggle={() => toggleCollapse(UNCATEGORIZED_ID)}
                    onCategoryAdd={() => onCategoryAddInGroup(undefined)}
                  />
                  {!uncategorizedCollapsed
                    ? uncategorizedCategories.map((category) => (
                        <ReservationTypeRow
                          key={category.id}
                          category={category}
                          canEdit={canEdit}
                          onEdit={() => onCategoryEdit(category)}
                        />
                      ))
                    : null}
                </Fragment>
              ) : null}
            </tbody>
          </table>
        </div>
      </SortableContext>
    </DndContext>
  );
}
