// Internal
import { TableCell } from "../../../components/ui/table";
import { DataTable } from "../../../components/shared/DataTable";
import { StatusBadge } from "../../../components/shared/StatusBadge";
import { DataTableRow } from "../../../components/shared/DataTableRow";
import { RowActionButton } from "../../../components/shared/RowActionButton";
import { getHospitalizationStatusColor, getHospitalizationTypeColor } from "../../../lib/status-helpers";

// Relative
import { H_STYLES } from "../styles";

// Types
import type { Hospitalization } from "../../../types";

interface HospitalizationListViewProps {
  hospitalizations: Hospitalization[];
  onNavigate: (id: string) => void;
}

export const HospitalizationListView = ({ hospitalizations, onNavigate }: HospitalizationListViewProps) => {
  const columns = [
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

  return (
    <DataTable
      columns={columns}
      data={hospitalizations}
      emptyMessage="入院データがありません"
      renderRow={(h) => (
        <DataTableRow key={h.id} onClick={() => onNavigate(h.id)}>
          <TableCell className={`font-mono ${H_STYLES.text.base} text-[#37352F] py-2`}>
            {h.hospitalizationNo}
          </TableCell>
          <TableCell className={`${H_STYLES.text.base} text-[#37352F] py-2`}>{h.ownerName}</TableCell>
          <TableCell className={`${H_STYLES.text.base} text-[#37352F] py-2`}>{h.petName}</TableCell>
          <TableCell className={`${H_STYLES.text.base} text-[#37352F] py-2`}>{h.species}</TableCell>
          <TableCell className="py-2">
            <StatusBadge colorClass={getHospitalizationTypeColor(h.hospitalizationType)}>
              {h.hospitalizationType}
            </StatusBadge>
          </TableCell>
          <TableCell className={`font-mono ${H_STYLES.text.base} text-[#37352F] py-2`}>{h.startDate}</TableCell>
          <TableCell className={`font-mono ${H_STYLES.text.base} text-[#37352F] py-2`}>{h.endDate}</TableCell>
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
};
