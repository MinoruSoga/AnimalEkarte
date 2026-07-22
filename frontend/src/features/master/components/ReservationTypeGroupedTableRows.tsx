import { ChevronDown, Pencil, Plus } from "lucide-react";
import { DataTableRowButton } from "@/components/shared/DataTable/DataTableRowButton";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { TableCell } from "@/components/ui/table";
import { C, ICON, LAYOUT, PALETTE } from "@/lib/design-tokens";
import type { ReservationType } from "../api/reservation-types";
import type { ReservationTypeGroup } from "../api/reservation-type-groups";

interface ReservationTypeGroupHeaderProps {
  group: ReservationTypeGroup;
  count: number;
  isCollapsed: boolean;
  canEdit: boolean;
  onToggle: () => void;
  onGroupEdit: () => void;
  onCategoryAdd: () => void;
}

export function ReservationTypeGroupHeader({
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
            aria-label={`${group.name}グループを${isCollapsed ? "展開" : "折りたたむ"}`}
            aria-expanded={!isCollapsed}
            className={`flex min-h-11 min-w-11 items-center justify-center rounded-xxs ${C.text35} ${C.hoverBgMedium} shrink-0`}
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
            className={`min-h-11 min-w-11 text-sm font-medium ${C.text} ${C.hoverBgLight} px-1 rounded-xxs transition-colors`}
          >
            {group.name}
          </button>
          <span className={`text-xs ${C.text35} tabular-nums`}>{count}</span>
          {canEdit ? (
            <div className="ml-auto flex items-center gap-1">
              <button
                type="button"
                onClick={onGroupEdit}
                className={`flex min-h-11 min-w-11 items-center gap-1 text-xs ${C.text45}
                  ${LAYOUT.inputCompact} ${C.hoverBgMedium} transition-colors`}
              >
                <Pencil className={ICON.action} />編集
              </button>
              <button
                type="button"
                onClick={onCategoryAdd}
                className={`flex min-h-11 min-w-11 items-center gap-1 text-xs ${C.text45}
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

interface ReservationTypeUncategorizedHeaderProps {
  count: number;
  isCollapsed: boolean;
  canEdit: boolean;
  onToggle: () => void;
  onCategoryAdd: () => void;
}

export function ReservationTypeUncategorizedHeader({
  count,
  isCollapsed,
  canEdit,
  onToggle,
  onCategoryAdd,
}: ReservationTypeUncategorizedHeaderProps) {
  return (
    <tr className={`border-b ${C.borderLight} ${C.bgPage}`}>
      <td colSpan={5} className="px-2 py-0">
        <div className="flex items-center gap-1 h-11">
          <button
            type="button"
            onClick={onToggle}
            aria-label={`未分類を${isCollapsed ? "展開" : "折りたたむ"}`}
            aria-expanded={!isCollapsed}
            className={`flex min-h-11 min-w-11 items-center justify-center rounded-xxs ${C.text35} ${C.hoverBgMedium} shrink-0`}
          >
            <ChevronDown
              className={`${ICON.smXs} transition-transform duration-150`}
              style={{ transform: isCollapsed ? "rotate(-90deg)" : "rotate(0deg)" }}
            />
          </button>
          <span className={`${ICON.dotMd} rounded-full shrink-0`} style={{ backgroundColor: PALETTE.grayMedium }} />
          <span className={`text-sm font-medium ${C.text55}`}>未分類</span>
          <span className={`text-xs ${C.text35} tabular-nums`}>{count}</span>
          {canEdit ? (
            <button
              type="button"
              onClick={onCategoryAdd}
              className={`ml-auto flex min-h-11 min-w-11 items-center gap-1 text-xs ${C.text45}
                ${LAYOUT.inputCompact} ${C.hoverBgMedium} transition-colors`}
            >
              <Plus className={ICON.action} />追加
            </button>
          ) : null}
        </div>
      </td>
    </tr>
  );
}

export function ReservationTypeEmptyGroupRow() {
  return (
    <tr className={`border-b ${C.borderLight}`}>
      <td colSpan={5} className={`pl-10 py-2 text-sm ${C.text35} italic`}>
        予約区分がありません
      </td>
    </tr>
  );
}

interface ReservationTypeRowProps {
  category: ReservationType;
  canEdit: boolean;
  onEdit: () => void;
}

export function ReservationTypeRow({ category, canEdit, onEdit }: ReservationTypeRowProps) {
  return (
    <SortableDataTableRow
      id={category.id}
      dragLabel={`並べ替え: 予約区分 ${category.name} (ID ${category.id})`}
      dragDisabled={!canEdit}
    >
      <TableCell className={`font-medium text-sm ${C.text} pl-7`}>
        <DataTableRowButton
          aria-label={`詳細: 予約区分 ${category.name} (ID ${category.id})`}
          onClick={onEdit}
        >
          {category.name}
        </DataTableRowButton>
      </TableCell>
      <TableCell className={`text-sm ${C.text60} max-w-[220px] truncate`}>
        {category.description || "-"}
      </TableCell>
      <TableCell className="text-center">
        <StatusPill isActive={category.isActive} />
      </TableCell>
      <TableCell className="text-right">
        {canEdit ? (
          <RowActionButton
            aria-label={`編集: 予約区分 ${category.name} (ID ${category.id})`}
            onClick={onEdit}
          />
        ) : null}
      </TableCell>
    </SortableDataTableRow>
  );
}
