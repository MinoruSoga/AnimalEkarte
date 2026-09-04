import { DataTableRowButton } from "@/components/shared/DataTable/DataTableRowButton";
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

export function DiagnosisTypeRow({ item, canEdit, onEdit }: DiagnosisTypeRowProps) {
  return (
    <SortableDataTableRow
      key={item.id}
      id={item.id}
      dragLabel={`並べ替え: 診断カテゴリ ${item.name} (ID ${item.id})`}
      dragDisabled={!canEdit}
    >
      <TableCell className={`font-medium ${C.text}`}>
        {canEdit ? (
          <DataTableRowButton
            aria-label={`詳細: 診断カテゴリ ${item.name} (ID ${item.id})`}
            onClick={() => onEdit(item)}
          >
            {item.name}
          </DataTableRowButton>
        ) : (
          item.name
        )}
      </TableCell>
      <TableCell className={`${C.text70} truncate max-w-[240px]`}>
        {item.description || "-"}
      </TableCell>
      <TableCell className="text-center">
        <StatusPill isActive={item.isActive} />
      </TableCell>
      <TableCell className="text-right">
        {canEdit ? (
          <RowActionButton
            aria-label={`編集: 診断カテゴリ ${item.name} (ID ${item.id})`}
            onClick={() => onEdit(item)}
          />
        ) : null}
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

export function DiagnosisNameRow({ item, categoryMap, canEdit, onEdit }: DiagnosisNameRowProps) {
  return (
    <SortableDataTableRow
      key={item.id}
      id={item.id}
      dragLabel={`並べ替え: 診断病名 ${item.name} (ID ${item.id})`}
      dragDisabled={!canEdit}
    >
      <TableCell className={C.text70}>{categoryMap.get(item.diagnosisTypeId) ?? "-"}</TableCell>
      <TableCell className={`font-medium ${C.text}`}>
        {canEdit ? (
          <DataTableRowButton
            aria-label={`詳細: 診断病名 ${item.name} (ID ${item.id})`}
            onClick={() => onEdit(item)}
          >
            {item.name}
          </DataTableRowButton>
        ) : (
          item.name
        )}
      </TableCell>
      <TableCell className="text-center">
        <StatusPill isActive={item.isActive} />
      </TableCell>
      <TableCell className="text-right">
        {canEdit ? (
          <RowActionButton
            aria-label={`編集: 診断病名 ${item.name} (ID ${item.id})`}
            onClick={() => onEdit(item)}
          />
        ) : null}
      </TableCell>
    </SortableDataTableRow>
  );
}
