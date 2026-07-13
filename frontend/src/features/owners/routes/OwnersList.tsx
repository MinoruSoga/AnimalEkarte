// React/Framework
import { useState, useMemo, useCallback, useTransition, useDeferredValue, lazy, Suspense, useEffect } from "react";
import { useNavigate, useLoaderData, useRevalidator, useSearchParams } from "react-router";

// Hooks
import { useSortableData } from "@/hooks/use-sortable-data";
import { useModalState } from "@/hooks/use-modal-state";
import { useClinicScope } from "@/hooks/use-clinic-scope";

// External
import { Plus } from "lucide-react";
import { toast } from "sonner";

// Internal
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { ClinicScopeFilter } from "@/components/shared/ClinicScopeFilter/ClinicScopeFilter";
import { usePagination } from "@/hooks/use-pagination";
import { ICON } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import { transformUpdatePetRequest } from "@/lib/transforms/pet";
import { handleApiError } from "@/lib/handle-api-error";
import { openOwnerReport } from "@/lib/owner-report-window";
// bundle-barrel-imports: バレルindex経由ではなく直接ファイルからimport
import { deleteOwner } from "../api/delete-owner";
import { usePermission } from "@/hooks/use-permission";
import { matchesPetSearch } from "@/lib/pet-search";

// bundle-dynamic-imports: PetEditModal を遅延ロード
const PetEditModal = lazy(() =>
  import("../components/PetEditModal").then((m) => ({ default: m.PetEditModal }))
);

// Types
import type { Pet } from "@/types";
import type { OwnersLoaderData } from "../loaders";
import type { PetFormData } from "../types";
import type { UpdatePetRequest } from "@/types/pet";
import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import { ResourceMedicalRecords, ResourceOwners } from "@/types/generated/models";
import { OwnersListTable } from "../components/OwnersListTable";

// Pet → ペット編集フォーム初期値の変換。本ルートのモーダルでのみ使用するため
// コンポーネントファイルからの export はせずここに置く (react-refresh/only-export-components)。
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

const CLINIC_TOGGLE_RESET_PARAMS = ["page"] as const;

export function OwnersList({ onUpdatePet }: OwnersListProps = {}) {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { canCreate, canEdit, canDelete } = usePermission("owners");
  // #158: レポート導線は medical-records:view でゲートする（カルテ内容を横断表示するため）
  const { canView: canReport } = usePermission(ResourceMedicalRecords);
  const revalidator = useRevalidator();
  const { pets } = useLoaderData<OwnersLoaderData>();

  // #86: 拠点横断表示 — URL の ?clinics=1,2 が表示拠点。未指定は現在の医院のみ（従来挙動）。
  // 選択変更で loader が再実行され、サーバ側 (resolveListClinicIDs) で所属検証される。
  const {
    assignedClinics,
    selectedClinicIds,
    isMultiClinic,
    clinicNameById,
    currentClinicId,
    handleToggleClinic,
  } = useClinicScope({ resetParamsOnToggle: CLINIC_TOGGLE_RESET_PARAMS });
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

    // テキスト検索 — カナ正規化済み (ひらがな↔カタカナ区別なし)
    if (deferredSearchTerm) {
      result = result.filter((pet) => matchesPetSearch(pet, deferredSearchTerm));
    }

    return result;
  }, [pets, activeFilters, deferredSearchTerm]);

  const { activeSorts, setActiveSorts, toggleSort, directionFor, sortedData } =
    useSortableData(filteredPets, { numericKeys: ["ownerNumber"] });

  // BUG-049: URLクエリパラメータからページ番号を読み取る
  const urlPage = Number(searchParams.get("page") ?? 1);

  const pagination = usePagination(sortedData, {
    resetKey: deferredSearchTerm,
  });

  // BUG-049: URLのページ番号とローカル状態を同期（URLが変わったときのみ）
  // rerender-dependencies: pagination（オブジェクト）を destructure し primitive を deps に使用
  const { totalPages, currentPage, goToPage } = pagination;
  useEffect(() => {
    const clampedPage = Math.max(1, Math.min(urlPage, totalPages));
    if (clampedPage !== currentPage) {
      goToPage(clampedPage);
    }
  // goToPage は安定した参照、totalPages は依存として必要
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [urlPage, totalPages]);

  // BUG-049: ページ変更時にURLクエリパラメータを更新
  const handlePageChange = useCallback((page: number) => {
    goToPage(page);
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev);
      if (page === 1) {
        next.delete("page");
      } else {
        next.set("page", String(page));
      }
      return next;
    }, { replace: true });
  }, [goToPage, setSearchParams]);

  // フィルタ計算が遅延中（入力値 ≠ deferred 値）の視覚フィードバック
  const isFiltering = searchTerm !== deferredSearchTerm;

  const handleCreate = useCallback(() => {
    navigate(paths.owners.new.getHref());
  }, [navigate]);

  // rerender-functional-setstate: useCallback で安定した関数参照を維持
  const handleEdit = useCallback((ownerId: string) => {
    navigate(paths.owners.detail.getHref(ownerId));
  }, [navigate]);

  // #158: 飼主レポートを別ウィンドウで開く。初期ペットを ?petId= で指定する。
  const handleReport = useCallback((ownerId: string, petId: string) => {
    openOwnerReport(ownerId, petId);
  }, []);

  // 行クリック → 飼主編集・ペット一覧ページに遷移
  // #86: 別医院の行は詳細 API が現在医院スコープで 404 になるため遷移させない（閲覧のみ）
  const handleRowClick = useCallback((pet: Pet) => {
    if (pet.clinicId && currentClinicId && pet.clinicId !== currentClinicId) {
      toast.info("別医院のデータです。医院を切り替えると詳細を表示できます");
      return;
    }
    navigate(paths.owners.detail.getHref(pet.ownerId));
  }, [navigate, currentClinicId]);

  // rerender-dependencies: object 依存を避け stable な変数に抽出してから deps に渡す
  const petModalItem = petModal.item;
  const closePetModal = petModal.close;

  // PetEditModal の保存ハンドラ
  const handlePetSave = useCallback((formData: PetFormData) => {
    if (!petModalItem || !onUpdatePet) return;
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
        await onUpdatePet(petModalItem.id, req);
        toast.success("ペット情報を更新しました");
        closePetModal();
        revalidator.revalidate();
      } catch (error: unknown) {
        handleApiError(error, "更新");
      }
    });
  }, [petModalItem, closePetModal, onUpdatePet, revalidator]);

  const openDeleteModal = deleteModal.open;
  const handleDeleteRequest = useCallback((ownerId: string, ownerName: string) => {
    openDeleteModal({ id: ownerId, name: ownerName });
  }, [openDeleteModal]);

  // rerender-dependencies: object依存を避け primitive の id のみを dep に使用
  const pendingDeleteOwnerId = deleteModal.item?.id ?? null;
  const closeDeleteModal = deleteModal.close;

  const handleConfirmDelete = useCallback(() => {
    if (!pendingDeleteOwnerId) return;

    startDeleteTransition(async () => {
      try {
        await deleteOwner(pendingDeleteOwnerId);
        toast.success("飼主を削除しました");
        closeDeleteModal();
        revalidator.revalidate();
      } catch (error) {
        handleApiError(error, "削除");
      }
    });
  }, [pendingDeleteOwnerId, closeDeleteModal, revalidator]);

  return (
    <PageLayout
      title="飼主・ペット一覧"
      resource={ResourceOwners}
      headerAction={
        canCreate ? (
          <PrimaryButton colorVariant="brand" onClick={handleCreate}>
            <Plus className={`mr-1.5 ${ICON.action}`} />
            新規登録
          </PrimaryButton>
        ) : null
      }
      maxWidth="max-w-full"
    >
      {/* #86: 複数所属ユーザーのみ拠点横断フィルタを表示 */}
      {assignedClinics.length >= 2 ? (
        <div className="mb-3">
          <ClinicScopeFilter
            clinics={assignedClinics}
            selectedIds={selectedClinicIds}
            onToggle={handleToggleClinic}
          />
        </div>
      ) : null}
      <OwnersListTable
        filteredCount={filteredPets.length}
        pagination={pagination}
        searchTerm={searchTerm}
        activeFilters={activeFilters}
        activeSorts={activeSorts}
        isFiltering={isFiltering}
        canEdit={canEdit}
        canDelete={canDelete}
        canReport={canReport}
        showClinicColumn={isMultiClinic}
        clinicNameById={clinicNameById}
        currentClinicId={currentClinicId}
        directionFor={directionFor}
        onSearchChange={setSearchTerm}
        onFilterChange={setActiveFilters}
        onSortChange={setActiveSorts}
        onToggleSort={toggleSort}
        onRowClick={handleRowClick}
        onEdit={handleEdit}
        onDeleteRequest={handleDeleteRequest}
        onReport={handleReport}
        onPageChange={handlePageChange}
      />

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
