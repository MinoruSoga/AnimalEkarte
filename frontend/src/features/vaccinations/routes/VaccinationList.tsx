// React/Framework
import { useState } from "react";
import { useNavigate } from "react-router";

// External
import { Plus, Syringe, FileSpreadsheet } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar";
import { DataTable } from "@/components/shared/DataTable";
import { PrimaryButton } from "@/components/shared/Form";
import { DataTableRow } from "@/components/shared/DataTable";
import { RowActionButton } from "@/components/shared/RowActionButton";

// Relative
import { useVaccinations } from "../hooks/useVaccinations";

export const VaccinationList = () => {
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState("");
  const { data: filteredRecords } = useVaccinations(searchTerm);

  const handleCreate = () => {
    navigate("/vaccinations/select-pet");
  };

  const handleEdit = (id: string) => {
    navigate(`/vaccinations/${id}`);
  };

  const columns = [
    { header: "実施日", className: "w-[120px]" },
    { header: "飼主名" },
    { header: "ペット名" },
    { header: "予防接種名" },
    { header: "次回予定", className: "w-[140px]" },
    { header: "操作", className: "w-[100px]", align: "right" as const },
  ];

  return (
    <PageLayout
      title="予防接種管理"
      icon={<Syringe className="size-5 text-[#37352F]" />}
      headerAction={
        <div className="flex items-center gap-2">
             <Button variant="outline" className="h-10 text-sm gap-2 bg-white" onClick={() => {}}>
                <FileSpreadsheet className="size-4" />
                データ取込
            </Button>
            <PrimaryButton onClick={handleCreate}>
                <Plus className="mr-1.5 size-4" />
                新規登録
            </PrimaryButton>
        </div>
      }
      maxWidth="max-w-full"
    >
      <div className="flex flex-col gap-4">
        {/* Search */}
        <SearchFilterBar
          searchTerm={searchTerm}
          onSearchChange={setSearchTerm}
          placeholder="飼主名、ペット名、予防接種名..."
          count={filteredRecords.length}
        />

        {/* Table */}
        <DataTable
            columns={columns}
            data={filteredRecords}
            emptyMessage="データが見つかりません"
            renderRow={(r) => (
                <DataTableRow 
                    key={r.id} 
                    onClick={() => handleEdit(r.id)}
                >
                    <TableCell className="font-mono text-sm text-[#37352F] py-2">{r.date}</TableCell>
                    <TableCell className="text-sm text-[#37352F] py-2">{r.ownerName}</TableCell>
                    <TableCell className="text-sm text-[#37352F] py-2">{r.petName}</TableCell>
                    <TableCell className="text-sm font-medium text-[#37352F] py-2">{r.vaccineName}</TableCell>
                    <TableCell className="font-mono text-sm text-[#37352F] py-2">{r.nextDate}</TableCell>
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
