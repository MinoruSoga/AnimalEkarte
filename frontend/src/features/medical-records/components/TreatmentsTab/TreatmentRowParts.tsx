import { ChevronDown, ChevronUp, Shield } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { TableCell } from "@/components/ui/table";
import { DeleteIconButton } from "@/components/shared/DeleteIconButton/DeleteIconButton";
import { BADGE, C, ICON, STYLE } from "@/lib/design-tokens";
import { formatCurrency } from "@/lib/format/number";
import type { TreatmentItemType } from "../../types";

const ITEM_TYPE_LABELS: Record<TreatmentItemType, string> = {
  consultation: "診療",
  procedure: "処置",
  medicine: "薬品",
  other: "その他",
};

const ITEM_TYPE_BADGE: Record<TreatmentItemType, string> = {
  consultation: BADGE.blue,
  procedure: BADGE.purple,
  medicine: BADGE.green,
  other: BADGE.muted,
};

interface TreatmentSelectionCellProps {
  checked: boolean;
  onChange: (checked: boolean | "indeterminate") => void;
}

export function TreatmentSelectionCell({ checked, onChange }: TreatmentSelectionCellProps) {
  return (
    <TableCell className="w-10 text-center">
      <Checkbox
        checked={checked}
        onCheckedChange={onChange}
        className={`${C.dataCheckedBgAccent} ${C.dataCheckedBorderAccent}`}
      />
    </TableCell>
  );
}

interface TreatmentTypeCellProps {
  itemType: TreatmentItemType;
}

export function TreatmentTypeCell({ itemType }: TreatmentTypeCellProps) {
  const badgeClass = ITEM_TYPE_BADGE[itemType] ?? BADGE.muted;
  const typeLabel = ITEM_TYPE_LABELS[itemType] ?? itemType;

  return (
    <TableCell className="w-24">
      <span
        className={`inline-flex items-center h-[22px] px-2 text-xs font-medium rounded border ${badgeClass}`}
      >
        {typeLabel}
      </span>
    </TableCell>
  );
}

interface TreatmentInsuranceCellProps {
  checked: boolean;
  onChange: (checked: boolean | "indeterminate") => void;
}

export function TreatmentInsuranceCell({ checked, onChange }: TreatmentInsuranceCellProps) {
  return (
    <TableCell className="w-16 text-center">
      <Checkbox
        checked={checked}
        onCheckedChange={onChange}
        className={`${C.dataCheckedBgBrand} ${C.dataCheckedBorderBrand}`}
      />
      {checked ? <Shield className={`${ICON.xs} mt-0.5 mx-auto ${C.textStatusGreen}`} /> : null}
    </TableCell>
  );
}

interface TreatmentSubtotalCellProps {
  subtotal: number;
}

export function TreatmentSubtotalCell({ subtotal }: TreatmentSubtotalCellProps) {
  return (
    <TableCell className="w-28 text-right">
      <span className={`text-sm font-medium ${C.text} font-mono`}>{formatCurrency(subtotal)}</span>
    </TableCell>
  );
}

interface TreatmentRowActionsProps {
  isFirst: boolean;
  isLast: boolean;
  canDelete: boolean;
  onMoveUp: () => void;
  onMoveDown: () => void;
  onDelete: () => void;
}

export function TreatmentRowActions({
  isFirst,
  isLast,
  canDelete,
  onMoveUp,
  onMoveDown,
  onDelete,
}: TreatmentRowActionsProps) {
  return (
    <TableCell className="w-36">
      <div className="flex items-center gap-0.5 justify-end">
        <Button
          variant="ghost"
          size="icon"
          className={`${STYLE.iconBtn28} ${C.text40} ${C.hoverText} disabled:opacity-20`}
          onClick={onMoveUp}
          disabled={isFirst}
          title="上に移動"
        >
          <ChevronUp className={ICON.xs} />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          className={`${STYLE.iconBtn28} ${C.text40} ${C.hoverText} disabled:opacity-20`}
          onClick={onMoveDown}
          disabled={isLast}
          title="下に移動"
        >
          <ChevronDown className={ICON.xs} />
        </Button>
        {canDelete ? <DeleteIconButton onClick={onDelete} className={STYLE.iconBtn28} /> : null}
      </div>
    </TableCell>
  );
}
