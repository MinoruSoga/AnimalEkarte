// React/Framework
import { useState, useMemo, useCallback, useTransition, useDeferredValue, lazy, Suspense } from "react";
import { useNavigate, useLoaderData, useRevalidator } from "react-router";

// Hooks
import { useSortableData } from "@/hooks/use-sortable-data";
import { useModalState } from "@/hooks/use-modal-state";

// External
import { Plus, Pencil, Trash2, PawPrint, Heart } from "lucide-react";
import { toast } from "sonner";

// Internal
import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { StatusBadge } from "@/components/shared/StatusBadge/StatusBadge";
import { RowActionDropdown } from "@/components/shared/RowActionDropdown";
import { Pagination } from "@/components/shared/Pagination";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { FilteringIndicator } from "@/components/shared/FilteringIndicator/FilteringIndicator";
import { SortableHeader } from "@/components/shared/SortableHeader";
import { getPetStatusColor } from "@/utils/status-helpers";
import { formatDate } from "@/utils/format/date";
import { formatWeight } from "@/utils/format/number";
import { usePagination } from "@/hooks/use-pagination";
import { STYLE } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import { transformUpdatePetRequest } from "@/lib/transforms/pet";
import { handleApiError } from "@/lib/handle-api-error";
// bundle-barrel-imports: バレルindex経由ではなく直接ファイルからimport
import { deleteOwner } from "../api/delete-owner";

// bundle-dynamic-imports: PetEditModal を遅延ロード
const PetEditModal = lazy(() =>
  import("../components/PetEditModal").then((m) => ({ default: m.PetEditModal }))
);

// Types
import type { Pet } from "@/types";
import type { OwnersLoaderData } from "../loaders";
import type { PetFormData } from "../types";
import type { UpdatePetRequest } from "@/types/pet";
import type {
  FilterProperty,
  ActiveFilter,
  SortProperty,
} from "@/components/shared/NotionFilter/types";
import { CONDITIONS_NO_EMPTY } from "@/components/shared/NotionFilter/types";

// rendering-hoist-jsx: 静的ソートプロパティ定義
const OWNER_SORT_PROPERTIES: SortProperty[] = [
  { key: "ownerNumber", label: "飼主No" },
  { key: "ownerName", label: "飼主名" },
  { key: "name", label: "ペット名" },
  { key: "species", label: "種" },
  { key: "birthDate", label: "生年月日" },
  { key: "lastVisit", label: "前回来院" },
];

// rendering-hoist-jsx: 静的フィルタ定義はモジュール定数に巻き上げ
const OWNER_FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "species",
    label: "種",
    type: "select",
    icon: PawPrint,
    // pets.animal_species_id NOT NULL — 空値は存在しない
    conditions: CONDITIONS_NO_EMPTY,
    options: [
      { value: "犬", label: "犬" },
      { value: "猫", label: "猫" },
      { value: "鳥", label: "鳥" },
      { value: "うさぎ", label: "うさぎ" },
      { value: "ハムスター", label: "ハムスター" },
      { value: "その他", label: "その他" },
    ],
  },
  {
    key: "status",
    label: "生死",
    type: "select",
    icon: Heart,
    // pets.status NOT NULL DEFAULT 'alive' — 空値は存在しない
    conditions: CONDITIONS_NO_EMPTY,
    options: [
      { value: "alive", label: "生存" },
      { value: "deceased", label: "死亡" },
    ],
  },
];

/** Pet 型 → PetEditModal 用の PetFormData に変換 */
function petToFormData(pet: Pet): PetFormData {
  return {
    id: pet.id,
    petNumber: pet.petNumber || "",
    petName: pet.name,
    petNameKana: pet.petNameKana || "",
    status: pet.status || "生存",
    species: pet.species,
    animalSpeciesId: pet.animalSpeciesId,
    gender: pet.gender || "",
    birthDate: pet.birthDate || "",
    color: pet.color || "",
    weight: pet.weight || "",
    food: pet.food || "",
    environment: pet.environment || "",
    neuteredDate: pet.neuteredDate || "",
    acquisitionType: (pet.acquisitionType as PetFormData["acquisitionType"]) || undefined,
    dangerLevel: (pet.dangerLevel as PetFormData["dangerLevel"]) || undefined,
    remarks: pet.remarks || "",
    breed: pet.breed,
    insuranceId: pet.insuranceId,
    insuranceName: pet.insuranceName,
    insuranceDetails: pet.insuranceDetails,
  };
}

interface OwnersListProps {
  onUpdatePet?: (id: string, req: UpdatePetRequest) => Promise<Pet>;
}

export function OwnersList({ onUpdatePet }: OwnersListProps = {}) {
  const navigate = useNavigate();
  const revalidator = useRevalidator();
  const { pets } = useLoaderData<OwnersLoaderData>();
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  // rerender-transitions: 入力は即座に反映しつつ、全件フィルタリングは
  // ブラウザがアイドル時まで遅延させてタイプ中の UI ブロッキングを防ぐ
  const deferredSearchTerm = useDeferredValue(searchTerm);
  const deleteModal = useModalState<{ id: string; name: string }>();
  const [isDeleting, startDeleteTransition] = useTransition();
  const petModal = useModalState<Pet>();
  const [_isPetSaving, startPetSaveTransition] = useTransition();

  const filteredPets = useMemo(() => {
    let result = pets;

    // ActiveFilter からフィルタ適用（condition 対応）
    for (const filter of activeFilters) {
      if (typeof filter.value !== "string") continue;

      if (filter.key === "species") {
        result = result.filter((p) => {
          switch (filter.condition) {
            case "is":
              return p.species === filter.value;
            case "is_not":
              return p.species !== filter.value;
            case "is_empty":
              return !p.species;
            case "is_not_empty":
              return !!p.species;
            default:
              return p.species === filter.value;
          }
        });
      }

      if (filter.key === "status") {
        result = result.filter((p) => {
          const isDeceased = p.status === "死亡";
          const matchesValue = filter.value === "deceased" ? isDeceased : !isDeceased;
          switch (filter.condition) {
            case "is":
              return matchesValue;
            case "is_not":
              return !matchesValue;
            case "is_empty":
              return !p.status;
            case "is_not_empty":
              return !!p.status;
            default:
              return matchesValue;
          }
        });
      }
    }

    // テキスト検索
    if (deferredSearchTerm) {
      const lowerTerm = deferredSearchTerm.toLowerCase();
      result = result.filter((pet) => {
        const ownerNumberStr = pet.ownerNumber?.toString() ?? "";
        return (
          pet.ownerName.toLowerCase().includes(lowerTerm) ||
          ownerNumberStr.includes(deferredSearchTerm) ||
          pet.name.toLowerCase().includes(lowerTerm) ||
          (pet.species && pet.species.toLowerCase().includes(lowerTerm))
        );
      });
    }

    return result;
  }, [pets, activeFilters, deferredSearchTerm]);

  const { activeSorts, setActiveSorts, toggleSort, directionFor, sortedData } =
    useSortableData(filteredPets, { numericKeys: ["ownerNumber"] });

  const pagination = usePagination(sortedData, {
    pageSize: 20,
    resetKey: deferredSearchTerm,
  });

  // フィルタ計算が遅延中（入力値 ≠ deferred 値）の視覚フィードバック
  const isFiltering = searchTerm !== deferredSearchTerm;

  const handleCreate = useCallback(() => {
    navigate(paths.owners.new.getHref());
  }, [navigate]);

  // rerender-functional-setstate: useCallback で安定した関数参照を維持
  const handleEdit = useCallback((ownerId: string) => {
    navigate(`/owners/${ownerId}`);
  }, [navigate]);

  // 行クリック → 飼主編集・ペット一覧ページに遷移
  const handleRowClick = useCallback((pet: Pet) => {
    navigate(paths.owners.detail.getHref(pet.ownerId));
  }, [navigate]);

  // PetEditModal の保存ハンドラ
  const handlePetSave = useCallback((formData: PetFormData) => {
    if (!petModal.item || !onUpdatePet) return;
    startPetSaveTransition(async () => {
      try {
        const req = transformUpdatePetRequest({
          name: formData.petName,
          petNameKana: formData.petNameKana,
          animalSpeciesId: formData.animalSpeciesId,
          gender: formData.gender,
          birthDate: formData.birthDate,
          breed: formData.breed,
          color: formData.color,
          weight: formData.weight,
          food: formData.food,
          environment: formData.environment,
          neuteredDate: formData.neuteredDate,
          acquisitionType: formData.acquisitionType,
          dangerLevel: formData.dangerLevel,
          status: formData.status === "死亡" ? "deceased" : "alive",
          insuranceId: formData.insuranceId,
          remarks: formData.remarks,
        });
        await onUpdatePet(petModal.item.id, req);
        toast.success("ペット情報を更新しました");
        petModal.close();
        revalidator.revalidate();
      } catch (error: unknown) {
        handleApiError(error, "更新");
      }
    });
  }, [petModal.item, petModal.close, onUpdatePet, revalidator]);

  const handleDeleteRequest = useCallback((ownerId: string, ownerName: string) => {
    deleteModal.open({ id: ownerId, name: ownerName });
  }, [deleteModal.open]);

  // rerender-dependencies: object依存を避け primitive の id のみを dep に使用
  const pendingDeleteOwnerId = deleteModal.item?.id ?? null;

  const handleConfirmDelete = useCallback(() => {
    if (!pendingDeleteOwnerId) return;

    startDeleteTransition(async () => {
      try {
        await deleteOwner(pendingDeleteOwnerId);
        toast.success("飼主を削除しました");
        deleteModal.close();
        revalidator.revalidate();
      } catch {
        toast.error("削除に失敗しました");
      }
    });
  }, [pendingDeleteOwnerId, deleteModal.close, revalidator]);

  // rerender-memo: renderRow を安定化してインラインクロージャ生成を排除
  const renderRow = useCallback((pet: (typeof filteredPets)[number]) => (
    <DataTableRow
      key={pet.id}
      onClick={() => handleRowClick(pet)}
    >
      <TableCell className={`${STYLE.tableCell} whitespace-nowrap hidden lg:table-cell`}>
        {pet.ownerNumber ?? "-"}
      </TableCell>
      <TableCell className={`${STYLE.tableCell} whitespace-nowrap`}>
        {pet.ownerName}
      </TableCell>
      <TableCell className={`${STYLE.tableCell} font-mono whitespace-nowrap hidden lg:table-cell`}>
        {pet.petNumber || "-"}
      </TableCell>
      <TableCell className={`${STYLE.tableCell} whitespace-nowrap`}>
        {pet.name}
      </TableCell>
      <TableCell className="whitespace-nowrap py-2 hidden lg:table-cell">
        {/* rendering-conditional-render: && は空文字をそのまま描画するためternaryを使用 */}
        {pet.status ? (
          <StatusBadge colorClass={getPetStatusColor(pet.status)}>
            {pet.status}
          </StatusBadge>
        ) : null}
      </TableCell>
      <TableCell className={`${STYLE.tableCell} whitespace-nowrap`}>
        {pet.species}
      </TableCell>
      <TableCell className={`${STYLE.tableCell} font-mono whitespace-nowrap hidden lg:table-cell`}>
        {formatDate(pet.birthDate)}
      </TableCell>
      <TableCell className={`${STYLE.tableCell} font-mono whitespace-nowrap hidden lg:table-cell`}>
        {formatWeight(pet.weight)}
      </TableCell>
      <TableCell className={`${STYLE.tableCell} whitespace-nowrap hidden lg:table-cell`}>
        {pet.environment || "-"}
      </TableCell>
      <TableCell className={`${STYLE.tableCell} font-mono whitespace-nowrap hidden lg:table-cell`}>
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
              onClick: () => handleDeleteRequest(pet.ownerId, pet.ownerName),
            },
          ]}
        />
      </TableCell>
    </DataTableRow>
  ), [handleRowClick, handleEdit, handleDeleteRequest]);

  const columns = useMemo(() => [
    {
      header: (
        <SortableHeader
          label="飼主No"
          direction={directionFor("ownerNumber")}
          onToggle={() => toggleSort("ownerNumber")}
        />
      ),
      className: "w-[100px] hidden lg:table-cell",
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
    { header: "ペット番号", className: "w-[100px] hidden lg:table-cell" },
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
    { header: "生死", className: "w-[60px] hidden lg:table-cell" },
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
      className: "w-[100px] hidden lg:table-cell",
    },
    { header: "体重", className: "w-[80px] hidden lg:table-cell" },
    { header: "環境", className: "w-[120px] hidden lg:table-cell" },
    {
      header: (
        <SortableHeader
          label="前回来院"
          direction={directionFor("lastVisit")}
          onToggle={() => toggleSort("lastVisit")}
        />
      ),
      className: "w-[100px] hidden lg:table-cell",
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
        <NotionFilter
          properties={OWNER_FILTER_PROPERTIES}
          activeFilters={activeFilters}
          onFilterChange={setActiveFilters}
          searchTerm={searchTerm}
          onSearchChange={setSearchTerm}
          searchPlaceholder="飼主名、ペット名、飼主No、種別..."
          count={filteredPets.length}
          sortProperties={OWNER_SORT_PROPERTIES}
          activeSorts={activeSorts}
          onSortChange={setActiveSorts}
        />

        {/* Table */}
        {/* rerender-memo: renderRow を useCallback で安定化し DataTable の不要な再レンダリングを防ぐ */}
        {/* rerender-transitions: isFiltering 中は opacity を落としてフィルタ遅延を視覚化 */}
        <FilteringIndicator isFiltering={isFiltering}>
          <DataTable
            columns={columns}
            data={pagination.paginatedData}
            emptyMessage="データが見つかりません"
            renderRow={renderRow}
          />
        </FilteringIndicator>

        {/* Pagination */}
        {pagination.totalPages > 1 ? (
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
        ) : null}
      </div>

      {/* Delete Confirm Dialog */}
      <ConfirmDialog
        open={deleteModal.isOpen}
        onClose={() => { if (!isDeleting) deleteModal.close(); }}
        onConfirm={handleConfirmDelete}
        title="飼主を削除しますか？"
        description={`飼主「${deleteModal.item?.name}」とこの飼主に関連するすべてのペット情報が削除されます。この操作は取り消すことができません。`}
        confirmLabel={isDeleting ? "削除中..." : "削除"}
        cancelLabel="キャンセル"
        variant="destructive"
      />

      {/* ペット編集モーダル */}
      {petModal.item ? (
        <Suspense fallback={null}>
          <PetEditModal
            open={petModal.isOpen}
            onOpenChange={(open) => {
              if (!open) petModal.close();
            }}
            ownerName={petModal.item.ownerName}
            petData={petToFormData(petModal.item)}
            onSave={handlePetSave}
          />
        </Suspense>
      ) : null}
    </PageLayout>
  );
}
