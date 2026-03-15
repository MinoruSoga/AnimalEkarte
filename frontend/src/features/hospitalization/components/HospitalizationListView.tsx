// Internal
import { TableCell } from "@/components/ui/table";
import { DataTable, DataTableRow } from "@/components/shared/DataTable";
import { StatusBadge } from "@/components/shared/StatusBadge";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { getHospitalizationStatusColor, getHospitalizationTypeColor } from "@/utils/status-helpers";
import { STYLE } from "@/lib/design-tokens";

// Types
import type { Hospitalization } from "@/types";

const COLUMNS = [
  { header: "入院No", className: "w-[120px]" },
  { header: "飼主名" },
  { header: "ペット名" },
  { header: "種", className: "w-[80px]" },
  { header: "タイプ", className: "w-[100px]" },
  { header: "入院開始日", className: "w-[120px]" },
  { header: "退院予定日", className: "w-[120px]" },
  { header: "ステータス", className: "w-[100px]" },
  { header: "操作", className: "w-[100px]", align: "right" as const },
];

interface HospitalizationListViewProps {
  hospitalizations: Hospitalization[];
  onNavigate: (id: string) => void;
}

export function HospitalizationListView({ hospitalizations, onNavigate }: HospitalizationListViewProps) {

  return (
    <DataTable
      columns={COLUMNS}
      data={hospitalizations}
      emptyMessage="入院データがありません"
      renderRow={(h) => (
        <DataTableRow key={h.id} onClick={() => onNavigate(h.id)}>
          <TableCell className={`${STYLE.tableCellMono}`}>
            {h.hospitalizationNo}
          </TableCell>
          <TableCell className={STYLE.tableCell}>{h.ownerName}</TableCell>
          <TableCell className={STYLE.tableCell}>{h.petName}</TableCell>
          <TableCell className={STYLE.tableCell}>{h.species}</TableCell>
          <TableCell className="py-2">
            <StatusBadge colorClass={getHospitalizationTypeColor(h.hospitalizationType)}>
              {h.hospitalizationType}
            </StatusBadge>
          </TableCell>
          <TableCell className={`${STYLE.tableCellMono}`}>{h.startDate}</TableCell>
          <TableCell className={`${STYLE.tableCellMono}`}>{h.endDate}</TableCell>
          <TableCell className="py-2">
            <StatusBadge colorClass={getHospitalizationStatusColor(h.status)}>
              {h.status}
            </StatusBadge>
          </TableCell>
          <TableCell className="text-right py-2">
            <RowActionButton onClick={() => onNavigate(h.id)} />
          </TableCell>
        </DataTableRow>
      )}
    />
  );
}
