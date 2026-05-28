import GripVertical from "lucide-react/dist/esm/icons/grip-vertical";
import Maximize2 from "lucide-react/dist/esm/icons/maximize-2";
import MoreHorizontal from "lucide-react/dist/esm/icons/more-horizontal";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { TableCell } from "@/components/ui/table";
import { C, ICON, STYLE } from "@/lib/design-tokens";
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

export function SortableMedicineRow({
  medicine,
  onEdit,
  grouped,
  canEdit,
}: SortableMedicineRowProps) {
  return (
    <SortableDataTableRow id={medicine.id} onClick={canEdit ? () => onEdit(medicine) : undefined}>
      <TableCell className={`${STYLE.tableCell} font-medium ${grouped ? "pl-12!" : "pl-2"}`}>
        {medicine.name}
      </TableCell>
      <TableCell className={`${STYLE.tableCell} w-[100px] text-center text-base`}>
        {medicine.dosageForm ? (DOSAGE_FORM_LABELS[medicine.dosageForm] ?? medicine.dosageForm) : "-"}
      </TableCell>
      <TableCell className={`${STYLE.tableCell} w-[130px] text-right pr-4 font-mono`}>
        {medicine.price > 0 ? `¥${medicine.price.toLocaleString()}` : "-"}
      </TableCell>
      <TableCell className="w-[110px] py-2.5 text-center">
        <NotionStatusPill isActive={medicine.isActive} />
      </TableCell>
      <TableCell className="w-[80px] py-2.5 text-center" onClick={(e) => e.stopPropagation()}>
        {canEdit ? (
          <div className="flex items-center justify-center gap-1">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button
                  type="button"
                  className={`${STYLE.iconBtn28} ${C.text40} ${C.hoverBgMedium} ${C.hoverText}`}
                >
                  <MoreHorizontal className={ICON.action} />
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onClick={() => onEdit(medicine)}>編集</DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
            <button
              type="button"
              onClick={() => onEdit(medicine)}
              className={`${STYLE.iconBtn28} ${C.text40} ${C.hoverBgMedium} ${C.hoverText}`}
            >
              <Maximize2 className={ICON.xs} />
            </button>
          </div>
        ) : null}
      </TableCell>
    </SortableDataTableRow>
  );
}

interface MedicineRowOverlayProps {
  medicine: Medicine;
  grouped: boolean;
}

export function MedicineRowOverlay({ medicine, grouped }: MedicineRowOverlayProps) {
  return (
    <div
      className={`flex items-center h-12 ${C.bgWhite} border ${C.borderLight} rounded-[4px] ${STYLE.dragOverlayShadow} cursor-grabbing`}
      style={{ width: "100%" }}
    >
      <div className={`w-8 shrink-0 flex items-center justify-center ${C.text50}`}>
        <GripVertical className={ICON.action} />
      </div>
      <div className={`flex-1 min-w-0 text-base font-medium ${C.text} ${grouped ? "pl-10" : "pl-0"}`}>
        {medicine.name}
      </div>
      <div className={`w-[100px] shrink-0 text-center text-base ${C.text65}`}>
        {medicine.dosageForm ? (DOSAGE_FORM_LABELS[medicine.dosageForm] ?? medicine.dosageForm) : "-"}
      </div>
      <div className={`w-[130px] shrink-0 text-right pr-4 font-mono text-base ${C.text}`}>
        {medicine.price > 0 ? `¥${medicine.price.toLocaleString()}` : "-"}
      </div>
      <div className="w-[110px] shrink-0 flex justify-center">
        <NotionStatusPill isActive={medicine.isActive} />
      </div>
      <div className="w-[80px] shrink-0" />
    </div>
  );
}
