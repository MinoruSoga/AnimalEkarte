// React/Framework
import { useState } from "react";
import { useNavigate } from "react-router";

// External
import { Plus, FileText } from "lucide-react";

// Internal
import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar";
import { DataTable, DataTableRow } from "@/components/shared/DataTable";
import { PrimaryButton } from "@/components/shared/Form";
import { StatusBadge } from "@/components/shared/StatusBadge";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { getMedicalRecordStatusColor } from "@/utils/status-helpers";

// Relative
import { useMedicalRecords } from "../hooks/useMedicalRecords";

export const MedicalRecords = () => {
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState("");
  const { data: filteredRecords } = useMedicalRecords(searchTerm);

  const handleNavigateToForm = (recordId?: string) => {
    navigate(recordId ? `/medical-records/${recordId}` : "/medical-records/select-pet");
  };

  const columns = [
    { header: "カルテNo", className: "w-[120px]" },
    { header: "診療日", className: "w-[120px]" },
    { header: "飼主名" },
    { header: "ペット名" },
    { header: "種", className: "w-[80px]" },
    { header: "主訴" },
    { header: "担当医", className: "w-[100px]" },
    { header: "ステータス", className: "w-[100px]" },
    { header: "操作", className: "w-[100px]", align: "right" as const },
  ];

  return (
    <PageLayout
      title="カルテ管理"
      icon={<FileText className="size-4 text-[#37352F]" />}
      headerAction={
        <PrimaryButton onClick={() => handleNavigateToForm()}>
          <Plus className="mr-1.5 size-4" />
          新規カルテ作成
        </PrimaryButton>
      }
      maxWidth="max-w-full"
    >
      <div className="flex flex-col gap-4">
        {/* Search */}
        <SearchFilterBar
          searchTerm={searchTerm}
          onSearchChange={setSearchTerm}
          placeholder="飼主名、ペット名、カルテNo、主訴で検索..."
          count={filteredRecords.length}
        />

        {/* Table */}
        <DataTable
          columns={columns}
          data={filteredRecords}
          emptyMessage="カルテデータが見つかりません"
          renderRow={(r) => (
            <DataTableRow 
              key={r.id} 
              onClick={() => handleNavigateToForm(r.id)}
            >
              <TableCell className="font-mono text-sm text-[#37352F] py-2">
                {r.recordNo}
              </TableCell>
              <TableCell className="text-sm text-[#37352F] py-2 font-mono">{r.date}</TableCell>
              <TableCell className="text-sm text-[#37352F] py-2">{r.ownerName}</TableCell>
              <TableCell className="text-sm text-[#37352F] py-2">{r.petName}</TableCell>
              <TableCell className="text-sm text-[#37352F] py-2">{r.species}</TableCell>
              <TableCell className="text-sm text-[#37352F] max-w-[200px] truncate py-2" title={r.chiefComplaint}>
                {r.chiefComplaint}
              </TableCell>
              <TableCell className="text-sm text-[#37352F] py-2">{r.doctor}</TableCell>
              <TableCell className="py-2">
                <StatusBadge colorClass={getMedicalRecordStatusColor(r.status)}>
                  {r.status}
                </StatusBadge>
              </TableCell>
              <TableCell className="text-right py-2">
                <RowActionButton onClick={() => handleNavigateToForm(r.id)} />
              </TableCell>
            </DataTableRow>
          )}
        />
      </div>
    </PageLayout>
  );
}
