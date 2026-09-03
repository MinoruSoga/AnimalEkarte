import ChevronDown from "lucide-react/dist/esm/icons/chevron-down";
import ChevronRight from "lucide-react/dist/esm/icons/chevron-right";

import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { DataTableRowButton } from "@/components/shared/DataTable/DataTableRowButton";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { TableCell } from "@/components/ui/table";
import { C, ICON } from "@/lib/design-tokens";
import { formatCurrencyOrDash } from "@/lib/format/number";
import type { TreatmentItem } from "@/lib/transforms/treatment";

import type {
  TreatmentTreeItem,
  TreatmentVirtualRow,
} from "../lib/treatment-plan-tab-content-model";

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

  return <ChildTreatmentRow item={row.item} canEdit={canEdit} onEdit={() => onEdit(row.item)} />;
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
  const expandLabel = `治療プラン ${item.name} (ID ${item.id}) の子項目を${isExpanded ? "折りたたむ" : "展開"}`;

  return (
    <SortableDataTableRow
      key={item.id}
      id={item.id}
      dragLabel={`並べ替え: 治療プラン ${item.name} (ID ${item.id})`}
      dragDisabled={!canEdit}
    >
      <TableCell>
        <div className="flex items-center gap-1">
          <TreatmentExpandButton
            hasChildren={hasChildren}
            isExpanded={isExpanded}
            ariaLabel={expandLabel}
            onToggle={onToggleExpanded}
          />
          <DataTableRowButton
            aria-label={`詳細: 治療プラン ${item.name} (ID ${item.id})`}
            className="text-base"
            onClick={onEdit}
          >
            {item.name}
          </DataTableRowButton>
          {hasChildren ? (
            <span className={`text-base ${C.text25} ml-0.5`}>{item.children.length}</span>
          ) : null}
        </div>
      </TableCell>
      <TreatmentPriceCell price={item.price} />
      <TableCell className="text-center">
        <StatusPill isActive={item.isActive} />
      </TableCell>
      <TableCell className="text-right">
        {canEdit ? (
          <RowActionButton
            aria-label={`編集: 治療プラン ${item.name} (ID ${item.id})`}
            onClick={onEdit}
          />
        ) : null}
      </TableCell>
    </SortableDataTableRow>
  );
}

interface TreatmentExpandButtonProps {
  hasChildren: boolean;
  isExpanded: boolean;
  ariaLabel: string;
  onToggle: () => void;
}

function TreatmentExpandButton({
  hasChildren,
  isExpanded,
  ariaLabel,
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
      aria-label={ariaLabel}
      className={`size-[22px] min-h-11 min-w-11 flex items-center justify-center rounded-xxs ${C.text40} ${C.hoverBgMedium} transition-colors shrink-0`}
    >
      {isExpanded ? <ChevronDown className={ICON.xs} /> : <ChevronRight className={ICON.xs} />}
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
    <DataTableRow>
      <TableCell className="w-11 px-0" />
      <TableCell>
        <div className="flex items-center gap-1 pl-[22px]">
          <span className="size-[22px] shrink-0" />
          <DataTableRowButton
            aria-label={`詳細: 治療プラン ${item.name} (ID ${item.id})`}
            className="text-base font-normal"
            onClick={onEdit}
          >
            {item.name}
          </DataTableRowButton>
        </div>
      </TableCell>
      <TreatmentPriceCell price={item.price} />
      <TableCell className="text-center">
        <StatusPill isActive={item.isActive} />
      </TableCell>
      <TableCell className="text-right">
        {canEdit ? (
          <RowActionButton
            aria-label={`編集: 治療プラン ${item.name} (ID ${item.id})`}
            onClick={onEdit}
          />
        ) : null}
      </TableCell>
    </DataTableRow>
  );
}

function TreatmentPriceCell({ price }: { price: number }) {
  return (
    <TableCell className="text-right">
      <span className={`text-base ${C.text70} font-mono`}>{formatCurrencyOrDash(price)}</span>
    </TableCell>
  );
}
