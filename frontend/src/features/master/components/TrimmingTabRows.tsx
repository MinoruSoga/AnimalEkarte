import { TableCell } from "@/components/ui/table";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { DataTableRowButton } from "@/components/shared/DataTable/DataTableRowButton";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { C } from "@/lib/design-tokens";
import { formatCurrency } from "@/lib/format/number";

import {
  TARGET_SIZE_LABELS,
  resolveTrimmingActiveFlag,
  type TrimmingCourse,
  type TrimmingOption,
} from "../api/trimming";

interface TrimmingCourseRowProps {
  item: TrimmingCourse;
  canEdit: boolean;
  onEdit: (item: TrimmingCourse) => void;
}

export function TrimmingCourseRow({
  item,
  canEdit,
  onEdit,
}: TrimmingCourseRowProps) {
  return (
    <DataTableRow key={item.id}>
      <TableCell className={`font-medium ${C.text}`}>
        <DataTableRowButton
          aria-label={`詳細: トリミングコース ${item.name} (ID ${item.id})`}
          onClick={() => onEdit(item)}
        >
          {item.name}
        </DataTableRowButton>
      </TableCell>
      <TableCell className={C.text70}>
        {item.targetSize ? TARGET_SIZE_LABELS[item.targetSize] : "-"}
      </TableCell>
      <TableCell className={C.text70}>
        {formatTrimmingDuration(item.duration)}
      </TableCell>
      <TrimmingPriceCell price={item.price} />
      <TableCell className="text-center">
        <StatusPill isActive={resolveTrimmingActiveFlag(item)} />
      </TableCell>
      <TableCell className="text-right">
        {canEdit ? (
          <RowActionButton
            onClick={() => onEdit(item)}
            aria-label={`トリミングコース「${item.name}」(ID: ${item.id}) を編集`}
          />
        ) : null}
      </TableCell>
    </DataTableRow>
  );
}

interface TrimmingOptionRowProps {
  item: TrimmingOption;
  canEdit: boolean;
  onEdit: (item: TrimmingOption) => void;
}

export function TrimmingOptionRow({
  item,
  canEdit,
  onEdit,
}: TrimmingOptionRowProps) {
  return (
    <DataTableRow key={item.id}>
      <TableCell className={`font-medium ${C.text}`}>
        <DataTableRowButton
          aria-label={`詳細: トリミングオプション ${item.name} (ID ${item.id})`}
          onClick={() => onEdit(item)}
        >
          {item.name}
        </DataTableRowButton>
      </TableCell>
      <TableCell className={C.text70}>
        {formatTrimmingDuration(item.duration)}
      </TableCell>
      <TableCell className="text-center">
        <CombinablePill combinable={item.combinable} />
      </TableCell>
      <TrimmingPriceCell price={item.price} />
      <TableCell className="text-center">
        <StatusPill isActive={resolveTrimmingActiveFlag(item)} />
      </TableCell>
      <TableCell className="text-right">
        {canEdit ? (
          <RowActionButton
            onClick={() => onEdit(item)}
            aria-label={`トリミングオプション「${item.name}」(ID: ${item.id}) を編集`}
          />
        ) : null}
      </TableCell>
    </DataTableRow>
  );
}

export function CombinablePill({ combinable }: { combinable: boolean }) {
  if (combinable) {
    return (
      <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-sm text-base ${C.bgStatusGreen} ${C.textStatusGreen}`}>
        可
      </span>
    );
  }

  return (
    <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-sm text-base ${C.bgInactive} ${C.text60}`}>
      不可
    </span>
  );
}

function TrimmingPriceCell({ price }: { price: number | null }) {
  return (
    <TableCell className={`text-right font-mono ${C.text}`}>
      {formatCurrency(price)}
    </TableCell>
  );
}

function formatTrimmingDuration(duration: number | null) {
  return duration != null ? `${duration}分` : "-";
}
