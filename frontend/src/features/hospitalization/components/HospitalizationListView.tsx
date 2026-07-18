// React/Framework
import { memo } from "react";

// Internal
import { TableCell } from "@/components/ui/table";
import { DataTable, DESIGN_TABLE_HEADER_ROW, DESIGN_TABLE_HEADER_CELL } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { StatusBadge } from "@/components/shared/StatusBadge/StatusBadge";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { getHospitalizationStatusColor, getHospitalizationTypeColor } from "@/lib/status-helpers";
import { C, STYLE } from "@/lib/design-tokens";
import { formatDate } from "@/lib/format/date";

// Types
import type { Hospitalization } from "@/types";

const COLUMNS = [
  { header: "入院No", className: "w-[120px]" },
  { header: "飼主名" },
  { header: "ペット名" },
  { header: "種", className: "w-[80px] hidden lg:table-cell" },
  { header: "タイプ", className: "w-[100px]" },
  { header: "入院開始日", className: "w-[120px]" },
  { header: "退院予定日", className: "w-[120px] hidden lg:table-cell" },
  { header: "ステータス", className: "w-[100px]" },
  { header: "操作", className: "w-[100px]", align: "right" as const },
];

interface HospitalizationListViewProps {
  hospitalizations: Hospitalization[];
  onNavigate: (id: string) => void;
  canEdit: boolean;
}

export const HospitalizationListView = memo(function HospitalizationListView({ hospitalizations, onNavigate, canEdit }: HospitalizationListViewProps) {

  return (
    <DataTable
      headerRowClassName={DESIGN_TABLE_HEADER_ROW}
      headerCellClassName={DESIGN_TABLE_HEADER_CELL}
      columns={COLUMNS}
      data={hospitalizations}
      emptyMessage="入院データがありません"
      renderRow={(h) => (
        <DataTableRow
          key={h.id}
          onClick={h.petIsDeceased ? undefined : () => onNavigate(h.id)}
          className={h.petIsDeceased ? "opacity-40 cursor-default" : ""}
        >
          <TableCell className={`${STYLE.tableCellMono}`}>
            {h.hospitalizationNo}
          </TableCell>
          <TableCell className={STYLE.tableCell}>{h.ownerName}</TableCell>
          <TableCell className={STYLE.tableCell}>{h.petName}</TableCell>
          <TableCell className={`${STYLE.tableCell} hidden lg:table-cell`}>{h.species}</TableCell>
          <TableCell className="py-2">
            <StatusBadge colorClass={getHospitalizationTypeColor(h.hospitalizationType)}>
              {h.hospitalizationType}
            </StatusBadge>
          </TableCell>
          <TableCell className={`${STYLE.tableCellMono}`}>{formatDate(h.startDate)}</TableCell>
          <TableCell className={`${STYLE.tableCellMono} hidden lg:table-cell`}>{formatDate(h.endDate)}</TableCell>
          <TableCell className="py-2">
            <StatusBadge colorClass={getHospitalizationStatusColor(h.status)}>
              {h.status}
            </StatusBadge>
          </TableCell>
          <TableCell className="text-right py-2">
            {h.petIsDeceased ? (
              <span className={`text-xs ${C.text40} font-medium`}>死亡</span>
            ) : canEdit ? (
              <RowActionButton onClick={() => onNavigate(h.id)} />
            ) : null}
          </TableCell>
        </DataTableRow>
      )}
    />
  );
});
