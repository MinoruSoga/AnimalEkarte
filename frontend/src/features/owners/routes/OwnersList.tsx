// React/Framework
import { useState, useMemo } from "react";
import { useNavigate, useLoaderData } from "react-router";

// External
import { Plus, Pencil, Trash2 } from "lucide-react";
import { toast } from "sonner";

// Internal
import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar";
import { DataTable, DataTableRow } from "@/components/shared/DataTable";
import { PrimaryButton } from "@/components/shared/Form";
import { StatusBadge } from "@/components/shared/StatusBadge";
import { RowActionDropdown } from "@/components/shared/RowActionDropdown";
import { Pagination } from "@/components/shared/Pagination";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { getPetStatusColor } from "@/utils/status-helpers";
import { formatDate } from "@/utils/format/date";
import { formatWeight } from "@/utils/format/number";
import { usePagination } from "@/hooks/usePagination";
import { deleteOwner } from "../api";

// Types
import type { OwnersLoaderData } from "../loaders";

export function OwnersList() {
  const navigate = useNavigate();
  const { pets } = useLoaderData<OwnersLoaderData>();
  const [searchTerm, setSearchTerm] = useState("");
  const [deleteTarget, setDeleteTarget] = useState<{
    id: string;
    name: string;
  } | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const filteredPets = useMemo(() => {
    if (!searchTerm) return pets;
    const lowerTerm = searchTerm.toLowerCase();
    return pets.filter((pet) => {
      const ownerNumberStr = pet.ownerNumber?.toString() ?? "";
      return (
        pet.ownerName.toLowerCase().includes(lowerTerm) ||
        ownerNumberStr.includes(searchTerm) ||
        pet.name.toLowerCase().includes(lowerTerm) ||
        (pet.species && pet.species.toLowerCase().includes(lowerTerm))
      );
    });
  }, [pets, searchTerm]);

  const pagination = usePagination(filteredPets, {
    pageSize: 20,
    resetKey: searchTerm,
  });

  const handleCreate = () => {
    navigate("/owners/new");
  };

  const handleEdit = (ownerId: string) => {
    navigate(`/owners/${ownerId}`);
  };

  const handleDeleteClick = (ownerId: string, ownerName: string) => {
    setDeleteTarget({ id: ownerId, name: ownerName });
  };

  const handleConfirmDelete = async () => {
    if (!deleteTarget) return;

    setIsDeleting(true);
    try {
      await deleteOwner(deleteTarget.id);
      toast.success("飼主を削除しました");
      setDeleteTarget(null);
    } catch {
      toast.error("削除に失敗しました");
    } finally {
      setIsDeleting(false);
    }
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
          data={pagination.paginatedData}
          emptyMessage="データが見つかりません"
          renderRow={(pet) => (
            <DataTableRow
              key={pet.id}
              onClick={() => handleEdit(pet.ownerId)}
            >
              <TableCell className="font-mono text-sm whitespace-nowrap py-2">
                {pet.ownerNumber ?? "-"}
              </TableCell>
              <TableCell className="text-sm whitespace-nowrap py-2">
                {pet.ownerName}
              </TableCell>
              <TableCell className="font-mono text-sm whitespace-nowrap py-2">
                {pet.petNumber || "-"}
              </TableCell>
              <TableCell className="text-sm whitespace-nowrap py-2">
                {pet.name}
              </TableCell>
              <TableCell className="whitespace-nowrap py-2">
                {pet.status && (
                  <StatusBadge
                    colorClass={getPetStatusColor(pet.status)}
                  >
                    {pet.status}
                  </StatusBadge>
                )}
              </TableCell>
              <TableCell className="text-sm whitespace-nowrap py-2">
                {pet.species}
              </TableCell>
              <TableCell className="font-mono text-sm whitespace-nowrap py-2">
                {formatDate(pet.birthDate)}
              </TableCell>
              <TableCell className="font-mono text-sm whitespace-nowrap py-2">
                {formatWeight(pet.weight)}
              </TableCell>
              <TableCell className="text-sm whitespace-nowrap py-2">
                {pet.environment || "-"}
              </TableCell>
              <TableCell className="font-mono text-sm whitespace-nowrap py-2">
                {formatDate(pet.lastVisit)}
              </TableCell>
              <TableCell className="whitespace-nowrap py-2 text-right">
                <RowActionDropdown
                  actions={[
                    {
                      label: "編集",
                      icon: Pencil,
                      onClick: () => handleEdit(pet.ownerId),
                    },
                    {
                      label: "削除",
                      icon: Trash2,
                      variant: "destructive",
                      onClick: () =>
                        handleDeleteClick(pet.ownerId, pet.ownerName),
                    },
                  ]}
                />
              </TableCell>
            </DataTableRow>
          )}
        />

        {/* Pagination */}
        {pagination.totalPages > 1 && (
          <Pagination
            currentPage={pagination.currentPage}
            totalPages={pagination.totalPages}
            totalCount={pagination.totalCount}
            startIndex={pagination.startIndex}
            endIndex={pagination.endIndex}
            onPageChange={pagination.goToPage}
            onPrev={pagination.prevPage}
            onNext={pagination.nextPage}
          />
        )}
      </div>

      {/* Delete Confirm Dialog */}
      <ConfirmDialog
        open={!!deleteTarget}
        onClose={() => !isDeleting && setDeleteTarget(null)}
        onConfirm={handleConfirmDelete}
        title="飼主を削除しますか？"
        description={`飼主「${deleteTarget?.name}」とこの飼主に関連するすべてのペット情報が削除されます。この操作は取り消すことができません。`}
        confirmLabel={isDeleting ? "削除中..." : "削除"}
        cancelLabel="キャンセル"
        variant="destructive"
      />
    </PageLayout>
  );
}
