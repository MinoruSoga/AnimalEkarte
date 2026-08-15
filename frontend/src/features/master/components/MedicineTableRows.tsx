import { memo } from "react";
import ChevronRight from "lucide-react/dist/esm/icons/chevron-right";
import GripVertical from "lucide-react/dist/esm/icons/grip-vertical";
import Plus from "lucide-react/dist/esm/icons/plus";
import { DataTableRowButton } from "@/components/shared/DataTable/DataTableRowButton";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { TableCell, TableRow } from "@/components/ui/table";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { formatCurrencyOrDash } from "@/lib/format/number";
import type { Medicine } from "@/types";

const DOSAGE_FORM_LABELS: Record<string, string> = {
  tablet: "錠剤",
  liquid: "液剤",
  injection: "注射剤",
  topical: "外用剤",
  powder: "散剤",
};

interface SortableMedicineRowProps {
  medicine: Medicine;
  onEdit: (medicine: Medicine) => void;
  grouped: boolean;
  canEdit: boolean;
}

interface MedicineCategoryHeaderRowProps {
  parentId: string;
  header: Medicine;
  itemCount: number;
  isCollapsed: boolean;
  canCreate: boolean;
  canEdit: boolean;
  onToggleGroup: (key: string) => void;
  onEdit: (medicine: Medicine) => void;
  onCreate: (parentId?: string) => void;
}

export const MedicineCategoryHeaderRow = memo(function MedicineCategoryHeaderRow({
  parentId,
  header,
  itemCount,
  isCollapsed,
  canCreate,
  canEdit,
  onToggleGroup,
  onEdit,
  onCreate,
}: MedicineCategoryHeaderRowProps) {
  return (
    <TableRow className={`${STYLE.tableRow} border-b ${C.borderLight} ${C.bgPage30} group/header ${C.hoverBgPage60}`}>
      <TableCell className="w-11 px-0">
        <span
          aria-hidden="true"
          className={`${STYLE.iconBtn32} ${C.text20} ${C.hoverBgMedium} ${C.hoverText60} cursor-grab`}
        >
          <GripVertical className={ICON.action} />
        </span>
      </TableCell>

      <TableCell className="pl-0 pr-2">
        <div className="flex items-center">
          <button
            type="button"
            onClick={(event) => {
              event.stopPropagation();
              onToggleGroup(parentId);
            }}
            className={`flex min-h-11 min-w-11 items-center gap-1.5 py-1.5 px-1 ${C.hoverBgLight} rounded-xxs transition-colors`}
          >
            <ChevronRight
              className={`${ICON.xs} ${C.text50} transition-transform duration-150 ${
                isCollapsed ? "" : "rotate-90"
              }`}
            />
            <span className={`text-base font-medium ${C.text65}`}>
              {header.name}
            </span>
            <span className={`text-base ${C.text40} ml-0.5`}>{itemCount}</span>
          </button>
          <div className="flex-1" />
          {canCreate ? (
            <button
              type="button"
              aria-label={`追加: 薬剤カテゴリ ${header.name} (ID ${header.id})`}
              onClick={() => {
                onCreate(parentId);
              }}
              className={`${STYLE.iconBtn32} ${C.text40} ${C.hoverBgMedium} ${C.hoverText} opacity-0 group-hover/header:opacity-100`}
            >
              <Plus className={ICON.xs} />
            </button>
          ) : null}
        </div>
      </TableCell>

      <TableCell className="w-[100px]" />
      <TableCell className="w-[130px]" />
      <TableCell className="w-[110px] text-center">
        <StatusPill isActive={true} />
      </TableCell>
      <TableCell className="w-[80px] text-center">
        {canEdit ? (
          <RowActionButton
            aria-label={`詳細: 薬剤カテゴリ ${header.name} (ID ${header.id})`}
            onClick={() => onEdit(header)}
            className="opacity-0 group-hover/header:opacity-100 focus-visible:opacity-100"
          />
        ) : null}
      </TableCell>
    </TableRow>
  );
});

export const SortableMedicineRow = memo(function SortableMedicineRow({
  medicine,
  onEdit,
  grouped,
  canEdit,
}: SortableMedicineRowProps) {
  return (
    <SortableDataTableRow
      id={medicine.id}
      dragLabel={`並べ替え: 薬剤 ${medicine.name} (ID ${medicine.id})`}
      dragDisabled={!canEdit}
    >
      <TableCell className={`${STYLE.tableCell} font-medium ${grouped ? "pl-12!" : "pl-2"}`}>
        {canEdit ? (
          <DataTableRowButton
            aria-label={`詳細: 薬剤 ${medicine.name} (ID ${medicine.id})`}
            onClick={() => onEdit(medicine)}
          >
            {medicine.name}
          </DataTableRowButton>
        ) : medicine.name}
      </TableCell>
      <TableCell className={`${STYLE.tableCell} w-[100px] text-center`}>
        {medicine.dosageForm ? (DOSAGE_FORM_LABELS[medicine.dosageForm] ?? medicine.dosageForm) : "-"}
      </TableCell>
      <TableCell className={`${STYLE.tableCell} w-[130px] text-right pr-4 font-mono`}>
        {formatCurrencyOrDash(medicine.price)}
      </TableCell>
      <TableCell className="w-[110px] text-center">
        <StatusPill isActive={medicine.isActive} />
      </TableCell>
      <TableCell className="w-[80px] text-center">
        {canEdit ? (
          <RowActionButton
            aria-label={`編集: 薬剤 ${medicine.name} (ID ${medicine.id})`}
            onClick={() => onEdit(medicine)}
          />
        ) : null}
      </TableCell>
    </SortableDataTableRow>
  );
});

interface MedicineRowOverlayProps {
  medicine: Medicine;
  grouped: boolean;
}

export function MedicineRowOverlay({ medicine, grouped }: MedicineRowOverlayProps) {
  return (
    <div
      className={`flex items-center h-12 ${C.bgWhite} border ${C.borderLight} rounded-xs ${STYLE.dragOverlayShadow} cursor-grabbing`}
      style={{ width: "100%" }}
    >
      <div className={`w-11 shrink-0 flex items-center justify-center ${C.text50}`}>
        <GripVertical className={ICON.action} />
      </div>
      <div className={`flex-1 min-w-0 text-base font-medium ${C.text} ${grouped ? "pl-10" : "pl-0"}`}>
        {medicine.name}
      </div>
      <div className={`w-[100px] shrink-0 text-center text-base ${C.text65}`}>
        {medicine.dosageForm ? (DOSAGE_FORM_LABELS[medicine.dosageForm] ?? medicine.dosageForm) : "-"}
      </div>
      <div className={`w-[130px] shrink-0 text-right pr-4 font-mono text-base ${C.text}`}>
        {formatCurrencyOrDash(medicine.price)}
      </div>
      <div className="w-[110px] shrink-0 flex justify-center">
        <StatusPill isActive={medicine.isActive} />
      </div>
      <div className="w-[80px] shrink-0" />
    </div>
  );
}
