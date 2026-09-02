import { TableCell } from "@/components/ui/table";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { DataTableRowButton } from "@/components/shared/DataTable/DataTableRowButton";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { C, LAYOUT, ICON, PALETTE } from "@/lib/design-tokens";
import type { Staff } from "../api/staffs";
import type { PermissionGroup } from "../api/permission-groups";

interface StaffSettingsRowProps {
  item: Staff;
  groups: PermissionGroup[];
  onEdit: (item: Staff) => void;
  canEdit: boolean;
}

export function StaffSettingsRow({
  item,
  groups,
  onEdit,
  canEdit,
}: StaffSettingsRowProps) {
  const visibleGroups = groups.slice(0, 2);
  const extraCount = groups.length - visibleGroups.length;
  return (
    <DataTableRow key={item.id}>
      <TableCell className={`font-medium ${C.text}`}>
        <DataTableRowButton
          aria-label={`詳細: スタッフ ${item.name} (ID ${item.id})`}
          onClick={() => onEdit(item)}
        >
          {item.name}
        </DataTableRowButton>
      </TableCell>
      <TableCell className={C.text}>{item.occupationName ?? "—"}</TableCell>
      <TableCell>
        <div className="flex flex-wrap items-center gap-1">
          {visibleGroups.length === 0 ? (
            <span className={`text-sm ${C.text40}`}>—</span>
          ) : (
            <>
              {visibleGroups.map((g) => (
                <span
                  key={g.id}
                  className={`inline-flex items-center gap-1 ${LAYOUT.inputCompact} text-xs`}
                  style={{
                    backgroundColor: g.color ? `${g.color}18` : `${PALETTE.primary}0f`,
                    color: g.color ?? PALETTE.primary,
                  }}
                >
                  <span
                    className={`${ICON.dotSm} rounded-full shrink-0`}
                    style={{ backgroundColor: g.color ?? PALETTE.defaultGray }}
                  />
                  {g.name}
                </span>
              ))}
              {extraCount > 0 ? (
                <span className={`text-xs ${C.text40}`}>+{extraCount}</span>
              ) : null}
            </>
          )}
        </div>
      </TableCell>
      <TableCell className="text-center">
        <StatusPill isActive={item.isActive} />
      </TableCell>
      <TableCell className="text-right">
        {canEdit ? (
          <RowActionButton
            onClick={() => onEdit(item)}
            aria-label={`スタッフ「${item.name}」(ID: ${item.id}) を編集`}
          />
        ) : null}
      </TableCell>
    </DataTableRow>
  );
}
