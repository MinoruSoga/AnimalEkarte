// React/Framework
import { useState, useMemo, useCallback, useTransition } from "react";
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
import { SortableHeader } from "@/components/shared/SortableHeader";
import { getPetStatusColor } from "@/utils/status-helpers";
import { formatDate } from "@/utils/format/date";
import { formatWeight } from "@/utils/format/number";
import { usePagination } from "@/hooks/usePagination";
import { useTableSort } from "@/hooks/useTableSort";
import { STYLE } from "@/lib/design-tokens";
import { deleteOwner } from "../api";

// Types
import type { OwnersLoaderData } from "../loaders";

type SortKey = "ownerNumber" | "ownerName" | "name" | "species" | "birthDate" | "lastVisit";

export function OwnersList() {
  const navigate = useNavigate();
  const { pets } = useLoaderData<OwnersLoaderData>();
  const [searchTerm, setSearchTerm] = useState("");
  const [pendingDeleteOwner, setPendingDeleteOwner] = useState<{
    id: string;
    name: string;
  } | null>(null);
  const [isDeleting, startDeleteTransition] = useTransition();

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

  const accessor = useCallback(
    (item: (typeof filteredPets)[number], key: SortKey): string =>
      String(item[key] ?? ""),
    [],
  );

  const { sortedData, directionFor, toggleSort } = useTableSort<
    (typeof filteredPets)[number],
    SortKey
  >(filteredPets, { accessor });

  const pagination = usePagination(sortedData, {
    pageSize: 20,
    resetKey: searchTerm,
  });

  const handleCreate = () => {
    navigate("/owners/new");
  };

  const handleEdit = (ownerId: string) => {
    navigate(`/owners/${ownerId}`);
  };

  const handleDeleteRequest = (ownerId: string, ownerName: string) => {
    setPendingDeleteOwner({ id: ownerId, name: ownerName });
  };

  const handleConfirmDelete = useCallback(() => {
    if (!pendingDeleteOwner) return;

    startDeleteTransition(async () => {
      try {
        await deleteOwner(pendingDeleteOwner.id);
        toast.success("飼主を削除しました");
        setPendingDeleteOwner(null);
      } catch {
        toast.error("削除に失敗しました");
      }
    });
  }, [pendingDeleteOwner]);

  const columns = useMemo(() => [
    {
      header: (
        <SortableHeader
          label="飼主No"
          direction={directionFor("ownerNumber")}
          onToggle={() => toggleSort("ownerNumber")}
        />
      ),
      className: "w-[100px]",
    },
    {
      header: (
        <SortableHeader
          label="飼主名"
          direction={directionFor("ownerName")}
          onToggle={() => toggleSort("ownerName")}
        />
      ),
      className: "w-[180px]",
    },
    { header: "ペット番号", className: "w-[100px]" },
    {
      header: (
        <SortableHeader
          label="ペット名"
          direction={directionFor("name")}
          onToggle={() => toggleSort("name")}
        />
      ),
      className: "w-[120px]",
    },
    { header: "生死", className: "w-[60px]" },
    {
      header: (
        <SortableHeader
          label="種"
          direction={directionFor("species")}
          onToggle={() => toggleSort("species")}
        />
      ),
      className: "w-[60px]",
    },
    {
      header: (
        <SortableHeader
          label="生年月日"
          direction={directionFor("birthDate")}
          onToggle={() => toggleSort("birthDate")}
        />
      ),
      className: "w-[100px]",
    },
    { header: "体重", className: "w-[80px]" },
    { header: "環境", className: "w-[120px]" },
    {
      header: (
        <SortableHeader
          label="前回来院"
          direction={directionFor("lastVisit")}
          onToggle={() => toggleSort("lastVisit")}
        />
      ),
      className: "w-[100px]",
    },
    { header: "操作", className: "w-[100px]", align: "right" as const },
  ], [directionFor, toggleSort]);

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
      <div className="flex flex-col gap-4 flex-1 min-h-0">
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
              <TableCell className={`${STYLE.tableCell} whitespace-nowrap`}>
                {pet.ownerNumber ?? "-"}
              </TableCell>
              <TableCell className={`${STYLE.tableCell} whitespace-nowrap`}>
                {pet.ownerName}
              </TableCell>
              <TableCell className={`${STYLE.tableCell} font-mono whitespace-nowrap`}>
                {pet.petNumber || "-"}
              </TableCell>
              <TableCell className={`${STYLE.tableCell} whitespace-nowrap`}>
                {pet.name}
              </TableCell>
              <TableCell className="whitespace-nowrap py-2">
                {pet.status && (
                  <StatusBadge colorClass={getPetStatusColor(pet.status)}>
                    {pet.status}
                  </StatusBadge>
                )}
              </TableCell>
              <TableCell className={`${STYLE.tableCell} whitespace-nowrap`}>
                {pet.species}
              </TableCell>
              <TableCell className={`${STYLE.tableCell} font-mono whitespace-nowrap`}>
                {formatDate(pet.birthDate)}
              </TableCell>
              <TableCell className={`${STYLE.tableCell} font-mono whitespace-nowrap`}>
                {formatWeight(pet.weight)}
              </TableCell>
              <TableCell className={`${STYLE.tableCell} whitespace-nowrap`}>
                {pet.environment || "-"}
              </TableCell>
              <TableCell className={`${STYLE.tableCell} font-mono whitespace-nowrap`}>
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
                        handleDeleteRequest(pet.ownerId, pet.ownerName),
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
        open={!!pendingDeleteOwner}
        onClose={() => !isDeleting && setPendingDeleteOwner(null)}
        onConfirm={handleConfirmDelete}
        title="飼主を削除しますか？"
        description={`飼主「${pendingDeleteOwner?.name}」とこの飼主に関連するすべてのペット情報が削除されます。この操作は取り消すことができません。`}
        confirmLabel={isDeleting ? "削除中..." : "削除"}
        cancelLabel="キャンセル"
        variant="destructive"
      />
    </PageLayout>
  );
}
