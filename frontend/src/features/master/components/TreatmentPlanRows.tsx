import ChevronDown from "lucide-react/dist/esm/icons/chevron-down";
import ChevronRight from "lucide-react/dist/esm/icons/chevron-right";

import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { TableCell } from "@/components/ui/table";
import { C, ICON } from "@/lib/design-tokens";
import type { TreatmentItem } from "@/lib/transforms/treatment";

import type { TreatmentTreeItem, TreatmentVirtualRow } from "./treatment-plan-tab-content-model";

interface TreatmentPlanRowProps {
  row: TreatmentVirtualRow;
  canEdit: boolean;
  onEdit: (item: TreatmentItem) => void;
  onToggleExpanded: (id: string) => void;
}

export function TreatmentPlanRow({
  row,
  canEdit,
  onEdit,
  onToggleExpanded,
}: TreatmentPlanRowProps) {
  if (row.type === "root") {
    return (
      <RootTreatmentRow
        item={row.item}
        isExpanded={row.isExpanded}
        canEdit={canEdit}
        onEdit={() => onEdit(row.item)}
        onToggleExpanded={() => onToggleExpanded(row.item.id)}
      />
    );
  }

  return (
    <ChildTreatmentRow
      item={row.item}
      canEdit={canEdit}
      onEdit={() => onEdit(row.item)}
    />
  );
}

interface RootTreatmentRowProps {
  item: TreatmentTreeItem;
  isExpanded: boolean;
  canEdit: boolean;
  onEdit: () => void;
  onToggleExpanded: () => void;
}

function RootTreatmentRow({
  item,
  isExpanded,
  canEdit,
  onEdit,
  onToggleExpanded,
}: RootTreatmentRowProps) {
  const hasChildren = item.children.length > 0;

  return (
    <SortableDataTableRow key={item.id} id={item.id} onClick={onEdit}>
      <TableCell>
        <div className="flex items-center gap-1">
          <TreatmentExpandButton
            hasChildren={hasChildren}
            isExpanded={isExpanded}
            onToggle={onToggleExpanded}
          />
          <span className={`text-base font-medium ${C.text}`}>{item.name}</span>
          {hasChildren ? (
            <span className={`text-base ${C.text25} ml-0.5`}>{item.children.length}</span>
          ) : null}
        </div>
      </TableCell>
      <TreatmentPriceCell price={item.price} />
      <TableCell className="text-center">
        <StatusPill isActive={item.isActive} />
      </TableCell>
      <TableCell className="p-0 text-right">
        {canEdit ? <RowActionButton onClick={onEdit} /> : null}
      </TableCell>
    </SortableDataTableRow>
  );
}

interface TreatmentExpandButtonProps {
  hasChildren: boolean;
  isExpanded: boolean;
  onToggle: () => void;
}

function TreatmentExpandButton({
  hasChildren,
  isExpanded,
  onToggle,
}: TreatmentExpandButtonProps) {
  if (!hasChildren) {
    return <span className="size-[22px] shrink-0" />;
  }

  return (
    <button
      type="button"
      onClick={(event) => {
        event.stopPropagation();
        onToggle();
      }}
      className={`size-[22px] flex items-center justify-center rounded-[3px] ${C.text40} ${C.hoverBgMedium} transition-colors shrink-0`}
    >
      {isExpanded ? (
        <ChevronDown className={ICON.xs} />
      ) : (
        <ChevronRight className={ICON.xs} />
      )}
    </button>
  );
}

interface ChildTreatmentRowProps {
  item: TreatmentItem;
  onEdit: () => void;
  canEdit: boolean;
}

function ChildTreatmentRow({ item, onEdit, canEdit }: ChildTreatmentRowProps) {
  return (
    <DataTableRow onClick={onEdit}>
      <TableCell className="w-[32px]" />
      <TableCell>
        <div className="flex items-center gap-1 pl-[22px]">
          <span className="size-[22px] shrink-0" />
          <span className={`text-base ${C.text}`}>{item.name}</span>
        </div>
      </TableCell>
      <TreatmentPriceCell price={item.price} />
      <TableCell className="text-center">
        <StatusPill isActive={item.isActive} />
      </TableCell>
      <TableCell className="p-0 text-right">
        {canEdit ? <RowActionButton onClick={onEdit} /> : null}
      </TableCell>
    </DataTableRow>
  );
}

function TreatmentPriceCell({ price }: { price: number }) {
  return (
    <TableCell className="text-right">
      <span className={`text-base ${C.text70} font-mono`}>
        {price > 0 ? `¥${price.toLocaleString()}` : "-"}
      </span>
    </TableCell>
  );
}
