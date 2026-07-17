import { TableCell } from "@/components/ui/table";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { C } from "@/lib/design-tokens";

import {
  TARGET_SIZE_LABELS,
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
    <DataTableRow key={item.id} onClick={canEdit ? () => onEdit(item) : undefined}>
      <TableCell className={`font-medium text-base ${C.text}`}>
        {item.name}
      </TableCell>
      <TableCell className={`text-base ${C.text70}`}>
        {item.targetSize ? TARGET_SIZE_LABELS[item.targetSize] : "-"}
      </TableCell>
      <TableCell className={`text-base ${C.text70}`}>
        {formatTrimmingDuration(item.duration)}
      </TableCell>
      <TrimmingPriceCell price={item.price} />
      <TableCell className="text-center">
        <StatusPill isActive={item.isActive} />
      </TableCell>
      <TableCell className="p-0 text-right">
        {canEdit ? <RowActionButton onClick={() => onEdit(item)} /> : null}
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
    <DataTableRow key={item.id} onClick={canEdit ? () => onEdit(item) : undefined}>
      <TableCell className={`font-medium text-base ${C.text}`}>
        {item.name}
      </TableCell>
      <TableCell className={`text-base ${C.text70}`}>
        {formatTrimmingDuration(item.duration)}
      </TableCell>
      <TableCell className="text-center">
        <CombinablePill combinable={item.combinable} />
      </TableCell>
      <TrimmingPriceCell price={item.price} />
      <TableCell className="text-center">
        <StatusPill isActive={item.isActive} />
      </TableCell>
      <TableCell className="p-0 text-right">
        {canEdit ? <RowActionButton onClick={() => onEdit(item)} /> : null}
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
    <TableCell className={`text-right font-mono text-base ${C.text}`}>
      {price != null ? `¥${price.toLocaleString()}` : "-"}
    </TableCell>
  );
}

function formatTrimmingDuration(duration: number | null) {
  return duration != null ? `${duration}分` : "-";
}
