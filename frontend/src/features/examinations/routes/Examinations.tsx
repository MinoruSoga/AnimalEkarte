// React/Framework
import { useState } from "react";
import { useNavigate } from "react-router-dom";

// External
import { Plus, TestTube, FileSpreadsheet } from "lucide-react";

// Internal
import { Button } from "../../../components/ui/button";
import { TableCell } from "../../../components/ui/table";
import { PageLayout } from "../../../components/shared/PageLayout";
import { SearchFilterBar } from "../../../components/shared/SearchFilterBar";
import { DataTable } from "../../../components/shared/DataTable";
import { PrimaryButton } from "../../../components/shared/PrimaryButton";
import { StatusBadge } from "../../../components/shared/StatusBadge";
import { DataTableRow } from "../../../components/shared/DataTableRow";
import { RowActionButton } from "../../../components/shared/RowActionButton";
import { getExaminationStatusColor } from "../../../lib/status-helpers";

// Relative
import { useExaminationRecords } from "../hooks/useExaminationRecords";

export const Examinations = () => {
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState("");
  const { data: filteredRecords } = useExaminationRecords(searchTerm);

  const handleCreate = () => {
    navigate("/examinations/select-pet");
  };

  const handleEdit = (id: string) => {
    navigate(`/examinations/${id}`);
  };

  const columns = [
    { header: "日時", className: "w-[120px]" },
    { header: "飼主名" },
    { header: "ペット名" },
    { header: "検査種別" },
    { header: "結果概要" },
    { header: "担当医", className: "w-[100px]" },
    { header: "ステータス", className: "w-[80px]" },
    { header: "操作", className: "w-[80px]", align: "right" as const },
  ];

  return (
    <PageLayout
      title="検査管理"
      icon={<TestTube className="size-5 text-[#37352F]" />}
      headerAction={
        <div className="flex items-center gap-2">
             <Button variant="outline" className="h-10 text-sm gap-2 bg-white" onClick={() => {}}>
                <FileSpreadsheet className="size-4" />
                検査データ取込
            </Button>
            <PrimaryButton onClick={handleCreate}>
                <Plus className="mr-1.5 size-4" />
                新規検査登録
            </PrimaryButton>
        </div>
      }
      maxWidth="max-w-full"
    >
      <div className="flex flex-col gap-4">
         {/* Filter & Search */}
         <SearchFilterBar
            searchTerm={searchTerm}
            onSearchChange={setSearchTerm}
            placeholder="飼主名、ペット名、検査種別..."
            count={filteredRecords.length}
         />

        {/* Table */}
        <DataTable
            columns={columns}
            data={filteredRecords}
            emptyMessage="検査データが見つかりません"
            renderRow={(r) => (
                <DataTableRow 
                    key={r.id} 
                    onClick={() => handleEdit(r.id)}
                >
                    <TableCell className="font-mono text-sm text-[#37352F] py-2">{r.date}</TableCell>
                    <TableCell className="text-sm text-[#37352F] py-2">{r.ownerName}</TableCell>
                    <TableCell className="text-sm text-[#37352F] py-2">{r.petName}</TableCell>
                    <TableCell className="text-sm font-medium text-[#37352F] py-2">{r.testType}</TableCell>
                    <TableCell className="text-sm text-muted-foreground truncate max-w-[200px] py-2">
                    {r.resultSummary || "-"}
                    </TableCell>
                    <TableCell className="text-sm text-[#37352F] py-2">{r.doctor}</TableCell>
                    <TableCell className="py-2">
                    <StatusBadge colorClass={getExaminationStatusColor(r.status)}>
                        {r.status}
                    </StatusBadge>
                    </TableCell>
                    <TableCell className="text-right py-2">
                        <RowActionButton onClick={() => handleEdit(r.id)} />
                    </TableCell>
                </DataTableRow>
            )}
        />
      </div>
    </PageLayout>
  );
}
