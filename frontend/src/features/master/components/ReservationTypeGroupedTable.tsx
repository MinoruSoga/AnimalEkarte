import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { ChevronDown, Pencil, Plus } from "lucide-react";
import { Fragment, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { TableCell } from "@/components/ui/table";
import { useSortableList } from "@/hooks/use-sortable-list";
import { C, ICON, LAYOUT, PALETTE, STYLE } from "@/lib/design-tokens";
import { useReorderReservationTypes } from "../api/reservation-types";
import type { ReservationType } from "../api/reservation-types";
import type { ReservationTypeGroup } from "../api/reservation-type-groups";

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
                        <tr className={`border-b ${C.borderLight}`}>
                          <td colSpan={5} className={`pl-10 py-2 text-sm ${C.text35} italic`}>
                            予約区分がありません
                          </td>
                        </tr>
                      )
                    ) : null}
                  </Fragment>
                );
              })}

              {uncategorizedCategories.length > 0 || groups.length === 0 ? (
                <Fragment key={UNCATEGORIZED_ID}>
                  <tr className={`border-b ${C.borderLight} ${C.bgPage}`}>
                    <td colSpan={5} className="px-2 py-0">
                      <div className="flex items-center gap-1 h-11">
                        <button
                          type="button"
                          onClick={() => toggleCollapse(UNCATEGORIZED_ID)}
                          className={`${STYLE.iconBtn20} ${C.text35} ${C.hoverBgMedium} shrink-0`}
                        >
                          <ChevronDown
                            className={`${ICON.smXs} transition-transform duration-150`}
                            style={{ transform: uncategorizedCollapsed ? "rotate(-90deg)" : "rotate(0deg)" }}
                          />
                        </button>
                        <span className={`${ICON.dotMd} rounded-full shrink-0`} style={{ backgroundColor: PALETTE.grayMedium }} />
                        <span className={`text-sm font-medium ${C.text55}`}>未分類</span>
                        <span className={`text-xs ${C.text35} tabular-nums`}>{uncategorizedCategories.length}</span>
                        {canEdit ? (
                          <button
                            type="button"
                            onClick={() => onCategoryAddInGroup(undefined)}
                            className={`ml-auto flex items-center gap-1 text-xs ${C.text45}
                              ${LAYOUT.inputCompact} ${C.hoverBgMedium} transition-colors`}
                          >
                            <Plus className={ICON.action} />追加
                          </button>
                        ) : null}
                      </div>
                    </td>
                  </tr>
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

interface ReservationTypeGroupHeaderProps {
  group: ReservationTypeGroup;
  count: number;
  isCollapsed: boolean;
  canEdit: boolean;
  onToggle: () => void;
  onGroupEdit: () => void;
  onCategoryAdd: () => void;
}

function ReservationTypeGroupHeader({
  group,
  count,
  isCollapsed,
  canEdit,
  onToggle,
  onGroupEdit,
  onCategoryAdd,
}: ReservationTypeGroupHeaderProps) {
  return (
    <tr className={`border-b ${C.borderLight} ${C.bgPage}`}>
      <td colSpan={5} className="px-2 py-0">
        <div className="flex items-center gap-1 h-11">
          <button
            type="button"
            onClick={onToggle}
            className={`${STYLE.iconBtn20} ${C.text35} ${C.hoverBgMedium} shrink-0`}
          >
            <ChevronDown
              className={`${ICON.smXs} transition-transform duration-150`}
              style={{ transform: isCollapsed ? "rotate(-90deg)" : "rotate(0deg)" }}
            />
          </button>
          <span className={`${ICON.dotMd} rounded-full shrink-0`} style={{ backgroundColor: group.color }} />
          <button
            type="button"
            onClick={onGroupEdit}
            className={`text-sm font-medium ${C.text} ${C.hoverBgLight} px-1 rounded-[3px] transition-colors`}
          >
            {group.name}
          </button>
          <span className={`text-xs ${C.text35} tabular-nums`}>{count}</span>
          {canEdit ? (
            <div className="ml-auto flex items-center gap-1">
              <button
                type="button"
                onClick={onGroupEdit}
                className={`flex items-center gap-1 text-xs ${C.text45}
                  ${LAYOUT.inputCompact} ${C.hoverBgMedium} transition-colors`}
              >
                <Pencil className={ICON.action} />編集
              </button>
              <button
                type="button"
                onClick={onCategoryAdd}
                className={`flex items-center gap-1 text-xs ${C.text45}
                  ${LAYOUT.inputCompact} ${C.hoverBgMedium} transition-colors`}
              >
                <Plus className={ICON.action} />追加
              </button>
            </div>
          ) : null}
        </div>
      </td>
    </tr>
  );
}

interface ReservationTypeRowProps {
  category: ReservationType;
  canEdit: boolean;
  onEdit: () => void;
}

function ReservationTypeRow({ category, canEdit, onEdit }: ReservationTypeRowProps) {
  return (
    <SortableDataTableRow id={category.id} onClick={onEdit}>
      <TableCell className={`font-medium text-sm ${C.text} pl-7`}>{category.name}</TableCell>
      <TableCell className={`text-sm ${C.text60} max-w-[220px] truncate`}>
        {category.description || "-"}
      </TableCell>
      <TableCell className="text-center">
        <NotionStatusPill isActive={category.isActive} />
      </TableCell>
      <TableCell className="p-0 text-right">
        {canEdit ? <RowActionButton onClick={onEdit} /> : null}
      </TableCell>
    </SortableDataTableRow>
  );
}
