import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Plus } from "lucide-react";
import { TableCell } from "../../../components/ui/table";
import { RowActionButton } from "../../../components/shared/RowActionButton";
import { PageLayout } from "../../../components/shared/PageLayout";
import { SearchFilterBar } from "../../../components/shared/SearchFilterBar";
import { DataTable } from "../../../components/shared/DataTable";
import { PrimaryButton } from "../../../components/shared/PrimaryButton";
import { StatusBadge } from "../../../components/shared/StatusBadge";
import { DataTableRow } from "../../../components/shared/DataTableRow";
import { usePets } from "../hooks/usePets";
import { getPetStatusColor } from "../../../lib/status-helpers";

export const OwnersList = () => {
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState("");
  const { data: filteredPets } = usePets(searchTerm);

  const handleCreate = () => {
    navigate("/owners/new");
  };

  const handleEdit = (ownerId: string) => {
    navigate(`/owners/${ownerId}`);
  };

  const columns = [
    { header: "飼主No", className: "w-[100px]" },
    { header: "飼主名", className: "w-[180px]" },
    { header: "ペット番号", className: "w-[100px]" },
    { header: "ペット名", className: "w-[120px]" },
    { header: "生死", className: "w-[60px]" },
    { header: "種", className: "w-[60px]" },
    { header: "生年月日", className: "w-[100px]" },
    { header: "体重", className: "w-[80px]" },
    { header: "環境", className: "w-[120px]" },
    { header: "前回来院", className: "w-[100px]" },
    { header: "操作", className: "w-[100px]", align: "right" as const },
  ];

  return (
    <PageLayout
      title="飼主・ペット一覧"
      headerAction={
        <PrimaryButton onClick={handleCreate}>
          <Plus className="mr-1.5 size-4" />
          新規登録
        </PrimaryButton>
      }
      maxWidth="max-w-full"
    >
      <div className="flex flex-col gap-4">
        {/* Search */}
        <SearchFilterBar
          searchTerm={searchTerm}
          onSearchChange={setSearchTerm}
          placeholder="飼主名、ペット名、飼主No、種別..."
          count={filteredPets.length}
        />

        {/* Table */}
        <DataTable
          columns={columns}
          data={filteredPets}
          emptyMessage="データが見つかりません"
          renderRow={(pet) => (
            <DataTableRow 
              key={pet.id}
              onClick={() => handleEdit(pet.ownerId)}
            >
              <TableCell className="text-sm text-[#37352F] whitespace-nowrap py-2">{pet.ownerId}</TableCell>
              <TableCell className="text-sm text-[#37352F] whitespace-nowrap py-2">{pet.ownerName}</TableCell>
              <TableCell className="font-mono text-sm text-[#37352F] whitespace-nowrap py-2">{pet.petNumber}</TableCell>
              <TableCell className="text-sm text-[#37352F] whitespace-nowrap py-2">{pet.name}</TableCell>
              <TableCell className="whitespace-nowrap py-2">
                {pet.status && (
                    <StatusBadge 
                        colorClass={getPetStatusColor(pet.status)} 
                    >
                      {pet.status}
                    </StatusBadge>
                )}
              </TableCell>
              <TableCell className="text-sm text-[#37352F] whitespace-nowrap py-2">{pet.species}</TableCell>
              <TableCell className="font-mono text-sm text-[#37352F] whitespace-nowrap py-2">{pet.birthDate}</TableCell>
              <TableCell className="font-mono text-sm text-[#37352F] whitespace-nowrap py-2">{pet.weight}</TableCell>
              <TableCell className="text-sm text-[#37352F] whitespace-nowrap py-2">{pet.environment}</TableCell>
              <TableCell className="font-mono text-sm text-[#37352F] whitespace-nowrap py-2">{pet.lastVisit}</TableCell>
              <TableCell className="whitespace-nowrap py-2 text-right">
                <RowActionButton onClick={() => handleEdit(pet.ownerId)} />
              </TableCell>
            </DataTableRow>
          )}
        />
      </div>
    </PageLayout>
  );
}
