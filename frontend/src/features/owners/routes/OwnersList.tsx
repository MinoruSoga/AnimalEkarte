// React/Framework
import { useState, useMemo, useCallback, useTransition, useDeferredValue, lazy, Suspense } from "react";
import { useNavigate, useLoaderData, useRevalidator } from "react-router";

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
  ActiveSort,
} from "@/components/shared/NotionFilter/types";

type SortKey = "ownerNumber" | "ownerName" | "name" | "species" | "birthDate" | "lastVisit";

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
  const [activeSorts, setActiveSorts] = useState<ActiveSort[]>([]);
  // rerender-transitions: 入力は即座に反映しつつ、全件フィルタリングは
  // ブラウザがアイドル時まで遅延させてタイプ中の UI ブロッキングを防ぐ
  const deferredSearchTerm = useDeferredValue(searchTerm);
  const [pendingDeleteOwner, setPendingDeleteOwner] = useState<{
    id: string;
    name: string;
  } | null>(null);
  const [isDeleting, startDeleteTransition] = useTransition();
  const [selectedPet, setSelectedPet] = useState<Pet | null>(null);
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

  // ── Sort logic driven by activeSorts ──
  const handleSortChange = useCallback((sorts: ActiveSort[]) => {
    setActiveSorts(sorts);
  }, []);

  // SortableHeader integration: toggle sort via activeSorts
  const toggleSort = useCallback((key: SortKey) => {
    setActiveSorts((prev) => {
      const existing = prev.find((s) => s.key === key);
      if (!existing) {
        // Add new sort (replace all - single sort for table header clicks)
        return [{ key, direction: "asc" as const }];
      }
      if (existing.direction === "asc") {
        return prev.map((s) => s.key === key ? { ...s, direction: "desc" as const } : s);
      }
      // Remove sort (was desc -> none)
      return prev.filter((s) => s.key !== key);
    });
  }, []);

  const directionFor = useCallback(
    (key: SortKey): "ascending" | "descending" | "none" => {
      const sort = activeSorts.find((s) => s.key === key);
      if (!sort) return "none";
      return sort.direction === "asc" ? "ascending" : "descending";
    },
    [activeSorts],
  );

  const sortedData = useMemo(() => {
    if (activeSorts.length === 0) return [...filteredPets];
    const sorted = [...filteredPets];
    sorted.sort((a, b) => {
      for (const sort of activeSorts) {
        const key = sort.key as SortKey;
        let cmp: number;
        if (key === "ownerNumber") {
          cmp = Number(a[key] ?? 0) - Number(b[key] ?? 0);
        } else {
          const aVal = String(a[key] ?? "");
          const bVal = String(b[key] ?? "");
          cmp = aVal.localeCompare(bVal, "ja");
        }
        if (cmp !== 0) return sort.direction === "asc" ? cmp : -cmp;
      }
      return 0;
    });
    return sorted;
  }, [filteredPets, activeSorts]);

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
    if (!selectedPet || !onUpdatePet) return;
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
        await onUpdatePet(selectedPet.id, req);
        toast.success("ペット情報を更新しました");
        setSelectedPet(null);
        revalidator.revalidate();
      } catch (error: unknown) {
        handleApiError(error, "更新");
      }
    });
  }, [selectedPet, onUpdatePet, revalidator]);

  const handleDeleteRequest = useCallback((ownerId: string, ownerName: string) => {
    setPendingDeleteOwner({ id: ownerId, name: ownerName });
  }, []);

  // rerender-dependencies: object依存を避け primitive の id のみを dep に使用
  const pendingDeleteOwnerId = pendingDeleteOwner?.id ?? null;

  const handleConfirmDelete = useCallback(() => {
    if (!pendingDeleteOwnerId) return;

    startDeleteTransition(async () => {
      try {
        await deleteOwner(pendingDeleteOwnerId);
        toast.success("飼主を削除しました");
        setPendingDeleteOwner(null);
        revalidator.revalidate();
      } catch {
        toast.error("削除に失敗しました");
      }
    });
  }, [pendingDeleteOwnerId, revalidator]);

  // rerender-memo: renderRow を安定化してインラインクロージャ生成を排除
  const renderRow = useCallback((pet: (typeof filteredPets)[number]) => (
    <DataTableRow
      key={pet.id}
      onClick={() => handleRowClick(pet)}
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
          onSortChange={handleSortChange}
        />

        {/* Table */}
        {/* rerender-memo: renderRow を useCallback で安定化し DataTable の不要な再レンダリングを防ぐ */}
        {/* rerender-transitions: isFiltering 中は opacity を落としてフィルタ遅延を視覚化 */}
        <div className={isFiltering ? "opacity-60 transition-opacity duration-150" : "transition-opacity duration-150"}>
          <DataTable
            columns={columns}
            data={pagination.paginatedData}
            emptyMessage="データが見つかりません"
            renderRow={renderRow}
          />
        </div>

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
        open={!!pendingDeleteOwner}
        onClose={() => !isDeleting && setPendingDeleteOwner(null)}
        onConfirm={handleConfirmDelete}
        title="飼主を削除しますか？"
        description={`飼主「${pendingDeleteOwner?.name}」とこの飼主に関連するすべてのペット情報が削除されます。この操作は取り消すことができません。`}
        confirmLabel={isDeleting ? "削除中..." : "削除"}
        cancelLabel="キャンセル"
        variant="destructive"
      />

      {/* ペット編集モーダル */}
      {selectedPet ? (
        <Suspense fallback={null}>
          <PetEditModal
            open={!!selectedPet}
            onOpenChange={(open) => {
              if (!open) setSelectedPet(null);
            }}
            ownerName={selectedPet.ownerName}
            petData={petToFormData(selectedPet)}
            onSave={handlePetSave}
          />
        </Suspense>
      ) : null}
    </PageLayout>
  );
}
