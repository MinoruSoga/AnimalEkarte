// React/Framework
import {
  useState,
  useMemo,
  useCallback,
  useTransition,
  lazy,
  Suspense,
  useEffect,
  useLayoutEffect,
  useRef,
} from "react";
import { useNavigate, useLoaderData, useRevalidator, useSearchParams, useNavigation } from "react-router";

// Hooks
import { useModalState } from "@/hooks/use-modal-state";
import { useClinicScope } from "@/hooks/use-clinic-scope";
import { useAnimalSpecies } from "@/hooks/use-animal-species";

// External
import { Plus } from "lucide-react";
import { toast } from "sonner";

// Internal
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { ClinicScopeFilter } from "@/components/shared/ClinicScopeFilter/ClinicScopeFilter";
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import { transformUpdatePetRequest } from "@/lib/transforms/pet";
import { handleApiError } from "@/lib/handle-api-error";
import { openOwnerReport } from "@/lib/owner-report-window";
// bundle-barrel-imports: バレルindex経由ではなく直接ファイルからimport
import { deleteOwner } from "../api/delete-owner";
import { usePermission } from "@/hooks/use-permission";

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
import {
  activeFiltersToParams,
  paramsToActiveFilters,
  buildSpeciesFilterOptions,
  buildOwnerFilterProperties,
} from "../lib/owners-list-filters";

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
    dangerReason: pet.dangerReason || "",
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

const SEARCH_DEBOUNCE_MS = 300;

export function OwnersList({ onUpdatePet }: OwnersListProps = {}) {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { canCreate, canEdit, canDelete } = usePermission("owners");
  const canEditRef = useRef(canEdit);
  const canDeleteRef = useRef(canDelete);
  useLayoutEffect(() => {
    canEditRef.current = canEdit;
    canDeleteRef.current = canDelete;
  }, [canDelete, canEdit]);
  // #158: レポート導線は medical-records:view でゲートする（カルテ内容を横断表示するため）
  const { canView: canReport } = usePermission(ResourceMedicalRecords);
  const revalidator = useRevalidator();
  const navigation = useNavigation();
  const { pets, page, limit, total } = useLoaderData<OwnersLoaderData>();

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

  // #266: 検索・フィルタ・ページはすべて URL 経由でサーバに転送する（loaders.ts 参照）。
  // 検索語は即時入力を受けつつ、URL 反映（= loader 再フェッチ）はデバウンスする。
  const urlSearch = searchParams.get("search") ?? "";
  const [searchTerm, setSearchTerm] = useState(urlSearch);

  // #266: species フィルタは pets.animal_species_id（数値ID）契約のためマスタ取得が必要。
  const {
    activeSpecies,
    isLoading: isSpeciesLoading,
    isError: isSpeciesError,
  } = useAnimalSpecies();
  const speciesFilterOptions = useMemo(
    () => isSpeciesError || isSpeciesLoading ? [] : buildSpeciesFilterOptions(activeSpecies),
    [activeSpecies, isSpeciesError, isSpeciesLoading],
  );
  const filterProperties = useMemo(() => buildOwnerFilterProperties(speciesFilterOptions), [speciesFilterOptions]);

  // rerender-derived-state-no-effect: activeFilters は URL(searchParams) からの純粋な派生値のため
  // useState+resync ではなく useMemo で直接算出する（source of truth は常に searchParams）。
  const activeFilters = useMemo(
    () => paramsToActiveFilters(searchParams, speciesFilterOptions),
    [searchParams, speciesFilterOptions],
  );

  useEffect(() => {
    if (searchTerm === urlSearch) return;
    const timer = setTimeout(() => {
      setSearchParams((prev) => {
        const next = new URLSearchParams(prev);
        if (searchTerm) next.set("search", searchTerm); else next.delete("search");
        next.delete("page");
        return next;
      }, { replace: true });
    }, SEARCH_DEBOUNCE_MS);
    return () => clearTimeout(timer);
  }, [searchTerm, urlSearch, setSearchParams]);

  // ブラウザの戻る/進むなど、こちら以外の経路で URL が変わった場合に検索入力欄を再同期する。
  // rerender-derived-state-no-effect: useEffect の代わりにレンダー中に derived state で処理
  // （use-pagination.ts の resetKey 比較と同型。デバウンス中の自分自身の書き込みでは値が
  // 一致するため no-op）。
  const searchParamsKey = searchParams.toString();
  const [prevSearchParamsKey, setPrevSearchParamsKey] = useState(searchParamsKey);
  if (searchParamsKey !== prevSearchParamsKey) {
    setPrevSearchParamsKey(searchParamsKey);
    setSearchTerm(urlSearch);
  }

  const handleFilterChange = useCallback((filters: ActiveFilter[]) => {
    const { species, include_deceased } = activeFiltersToParams(filters);
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev);
      if (species) next.set("species", species); else next.delete("species");
      if (include_deceased) next.set("include_deceased", include_deceased); else next.delete("include_deceased");
      next.delete("page");
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  const deleteModal = useModalState<{ id: string; name: string }>();
  const [isDeleting, startDeleteTransition] = useTransition();
  const petModal = useModalState<Pet>();
  // PetEditModal は保存呼び出し後すぐ閉じるため pending state を UI で使わない（FE-RC-083）
  const [, startPetSaveTransition] = useTransition();

  const totalPages = Math.max(1, Math.ceil(total / limit));
  const startIndex = total === 0 ? 0 : (page - 1) * limit + 1;
  const endIndex = Math.min(page * limit, total);

  // BUG-049 踏襲: ページ変更時に URL クエリパラメータを更新（loader が再フェッチする）
  const handlePageChange = useCallback((nextPage: number) => {
    const clamped = Math.max(1, Math.min(nextPage, totalPages));
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev);
      if (clamped === 1) {
        next.delete("page");
      } else {
        next.set("page", String(clamped));
      }
      return next;
    }, { replace: true });
  }, [totalPages, setSearchParams]);

  // #266: loader 再フェッチ中（検索・フィルタ・ページ変更）の視覚フィードバック
  const isFiltering = navigation.state === "loading";

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
          dangerReason: formData.dangerReason,
          originalDangerReason: petModalItem.dangerReason,
          // status は渡さない(BUG-415): transformUpdatePetRequest は status を無視する。
          insuranceId: formData.insuranceId,
          remarks: formData.remarks,
        });
        if (canEditRef.current !== true) return;
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
    if (canDeleteRef.current !== true || !pendingDeleteOwnerId) return;

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
          <PrimaryButton colorVariant="primary" onClick={handleCreate}>
            <Plus className={`mr-1.5 ${ICON.action}`} />
            新規登録
          </PrimaryButton>
        ) : null
      }
      maxWidth={LAYOUT.pageContentMaxWidth.full}
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
      {isSpeciesError ? (
        <p role="alert" aria-atomic="true" className={`mb-3 text-sm ${C.danger}`}>
          動物種の取得に失敗しました。
        </p>
      ) : isSpeciesLoading ? (
        <p
          role="status"
          aria-live="polite"
          aria-atomic="true"
          className={`mb-3 text-sm ${C.text50}`}
        >
          動物種を読み込み中です。
        </p>
      ) : activeSpecies.length === 0 ? (
        <p
          role="status"
          aria-live="polite"
          aria-atomic="true"
          className={`mb-3 text-sm ${C.text50}`}
        >
          動物種マスタが登録されていません。
        </p>
      ) : null}
      <OwnersListTable
        pets={pets}
        pagination={{ currentPage: page, totalPages, totalCount: total, startIndex, endIndex }}
        searchTerm={searchTerm}
        activeFilters={activeFilters}
        filterProperties={filterProperties}
        isFiltering={isFiltering}
        canEdit={canEdit}
        canDelete={canDelete}
        canReport={canReport}
        showClinicColumn={isMultiClinic}
        clinicNameById={clinicNameById}
        currentClinicId={currentClinicId}
        onSearchChange={setSearchTerm}
        onFilterChange={handleFilterChange}
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
        description={`飼主「${deleteModal.item?.name}」を削除します。この操作は取り消すことができません。なお、ペットが紐付いている場合は削除できません（先にペットを削除してください）。`}
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
