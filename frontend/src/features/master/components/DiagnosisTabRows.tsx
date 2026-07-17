import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { TableCell } from "@/components/ui/table";
import { C } from "@/lib/design-tokens";

import type { DiagnosisName, DiagnosisType } from "../api/diagnosis";

interface DiagnosisTypeRowProps {
  item: DiagnosisType;
  canEdit: boolean;
  onEdit: (item: DiagnosisType) => void;
}

export function DiagnosisTypeRow({
  item,
  canEdit,
  onEdit,
}: DiagnosisTypeRowProps) {
  return (
    <SortableDataTableRow
      key={item.id}
      id={item.id}
      onClick={canEdit ? () => onEdit(item) : undefined}
    >
      <TableCell className={`font-medium text-base ${C.text}`}>
        {item.name}
      </TableCell>
      <TableCell className={`text-base ${C.text70} truncate max-w-[240px]`}>
        {item.description || "-"}
      </TableCell>
      <TableCell className="text-center">
        <StatusPill isActive={item.isActive} />
      </TableCell>
      <TableCell className="p-0 text-right">
        {canEdit ? <RowActionButton onClick={() => onEdit(item)} /> : null}
      </TableCell>
    </SortableDataTableRow>
  );
}

interface DiagnosisNameRowProps {
  item: DiagnosisName;
  categoryMap: Map<string, string>;
  canEdit: boolean;
  onEdit: (item: DiagnosisName) => void;
}

export function DiagnosisNameRow({
  item,
  categoryMap,
  canEdit,
  onEdit,
}: DiagnosisNameRowProps) {
  return (
    <SortableDataTableRow
      key={item.id}
      id={item.id}
      onClick={canEdit ? () => onEdit(item) : undefined}
    >
      <TableCell className={`text-base ${C.text70}`}>
        {categoryMap.get(item.diagnosisTypeId) ?? "-"}
      </TableCell>
      <TableCell className={`font-medium text-base ${C.text}`}>
        {item.name}
      </TableCell>
      <TableCell className="text-center">
        <StatusPill isActive={item.isActive} />
      </TableCell>
      <TableCell className="p-0 text-right">
        {canEdit ? <RowActionButton onClick={() => onEdit(item)} /> : null}
      </TableCell>
    </SortableDataTableRow>
  );
}
