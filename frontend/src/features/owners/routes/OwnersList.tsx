// React/Framework
import {
  useState,
  useMemo,
  useCallback,
  useTransition,
  useEffect,
  useLayoutEffect,
  useRef,
} from "react";
import {
  useNavigate,
  useLoaderData,
  useRevalidator,
  useSearchParams,
  useNavigation,
} from "react-router";

// Hooks
import { useModalState } from "@/hooks/use-modal-state";
import { useAuth } from "@/hooks/use-auth";
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
import { handleApiError } from "@/lib/handle-api-error";
import { openOwnerReport } from "@/lib/owner-report-window";
// bundle-barrel-imports: バレルindex経由ではなく直接ファイルからimport
import { deleteOwner } from "../api/delete-owner";
import { usePermission } from "@/hooks/use-permission";

// Types
import type { OwnersLoaderData } from "../loaders";
import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import { ResourceMedicalRecords, ResourceOwners } from "@/types/generated/models";
import { OwnersListTable } from "../components/OwnersListTable";
import { OwnersListScopeState } from "../components/OwnersListScopeState";
import { useOwnersClinicScope } from "../hooks/use-owners-clinic-scope";
import {
  activeFiltersToParams,
  paramsToActiveFilters,
  buildSpeciesFilterOptions,
  buildOwnerFilterProperties,
} from "../lib/owners-list-filters";

export function OwnersList() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { canCreate, canEdit, canDelete } = usePermission("owners");
  const canDeleteRef = useRef(canDelete);
  useLayoutEffect(() => {
    canDeleteRef.current = canDelete;
  }, [canDelete]);
  // #158: レポート導線は medical-records:view でゲートする（カルテ内容を横断表示するため）
  const { canView: canReport } = usePermission(ResourceMedicalRecords);
  const revalidator = useRevalidator();
  const navigation = useNavigation();
  const { pets, page, limit, total, listState, clinicScope } = useLoaderData<OwnersLoaderData>();
  const { currentClinicId } = useAuth();

  // #86 / EMR-463: 拠点横断表示 — 表示医院は loader が GET /me `owners_view_clinic_ids`
  // （現在の owners:view 許可）から確定した医院だけ。初期値は全許可医院、明示選択は
  // URL ?clinics= と保存先へ反映する。最終認可はサーバ側 (GET /pets) が行う。
  const { clinics, selectedClinicIds, isMultiClinic, clinicNameById, handleToggleClinic } =
    useOwnersClinicScope(clinicScope);

  // #266: 検索・フィルタ・ページはすべて URL 経由でサーバに転送する（loaders.ts 参照）。
  // EMR-247: searchTerm は PropertyFilter の確定操作（Enter / 検索ボタン / クリア）で
  // 届く確定済みの値であり、URL 反映（= loader 再フェッチ）は即時行う。
  const urlSearch = searchParams.get("search") ?? "";
  const [searchTerm, setSearchTerm] = useState(urlSearch);

  // #266: species フィルタは pets.animal_species_id（数値ID）契約のためマスタ取得が必要。
  const {
    activeSpecies,
    isLoading: isSpeciesLoading,
    isError: isSpeciesError,
  } = useAnimalSpecies();
  const speciesFilterOptions = useMemo(
    () => (isSpeciesError || isSpeciesLoading ? [] : buildSpeciesFilterOptions(activeSpecies)),
    [activeSpecies, isSpeciesError, isSpeciesLoading],
  );
  const filterProperties = useMemo(
    () => buildOwnerFilterProperties(speciesFilterOptions),
    [speciesFilterOptions],
  );

  // rerender-derived-state-no-effect: activeFilters は URL(searchParams) からの純粋な派生値のため
  // useState+resync ではなく useMemo で直接算出する（source of truth は常に searchParams）。
  const activeFilters = useMemo(
    () => paramsToActiveFilters(searchParams, speciesFilterOptions),
    [searchParams, speciesFilterOptions],
  );

  useEffect(() => {
    if (searchTerm === urlSearch) return;
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev);
        if (searchTerm) next.set("search", searchTerm);
        else next.delete("search");
        next.delete("page");
        return next;
      },
      { replace: true },
    );
  }, [searchTerm, urlSearch, setSearchParams]);

  // ブラウザの戻る/進むなど、こちら以外の経路で URL が変わった場合に検索入力欄を再同期する。
  // rerender-derived-state-no-effect: useEffect の代わりにレンダー中に derived state で処理
  // （use-pagination.ts の resetKey 比較と同型。URL 確定時の自分自身の書き込みでは値が
  // 一致するため no-op）。
  const searchParamsKey = searchParams.toString();
  const [prevSearchParamsKey, setPrevSearchParamsKey] = useState(searchParamsKey);
  if (searchParamsKey !== prevSearchParamsKey) {
    setPrevSearchParamsKey(searchParamsKey);
    setSearchTerm(urlSearch);
  }

  const handleFilterChange = useCallback(
    (filters: ActiveFilter[]) => {
      const { species, include_deceased, checkup_history } = activeFiltersToParams(filters);
      setSearchParams(
        (prev) => {
          const next = new URLSearchParams(prev);
          if (species) next.set("species", species);
          else next.delete("species");
          if (include_deceased) next.set("include_deceased", include_deceased);
          else next.delete("include_deceased");
          if (checkup_history) next.set("checkup_history", checkup_history);
          else next.delete("checkup_history");
          next.delete("page");
          return next;
        },
        { replace: true },
      );
    },
    [setSearchParams],
  );

  const deleteModal = useModalState<{ id: string; name: string }>();
  const [isDeleting, startDeleteTransition] = useTransition();

  const totalPages = Math.max(1, Math.ceil(total / limit));
  const startIndex = total === 0 ? 0 : (page - 1) * limit + 1;
  const endIndex = Math.min(page * limit, total);

  // BUG-049 踏襲: ページ変更時に URL クエリパラメータを更新（loader が再フェッチする）
  const handlePageChange = useCallback(
    (nextPage: number) => {
      const clamped = Math.max(1, Math.min(nextPage, totalPages));
      setSearchParams(
        (prev) => {
          const next = new URLSearchParams(prev);
          if (clamped === 1) {
            next.delete("page");
          } else {
            next.set("page", String(clamped));
          }
          return next;
        },
        { replace: true },
      );
    },
    [totalPages, setSearchParams],
  );

  const handleReload = useCallback(() => {
    void revalidator.revalidate();
  }, [revalidator]);

  // #266: loader 再フェッチ中（検索・フィルタ・ページ変更）の視覚フィードバック
  const isFiltering = navigation.state === "loading";

  const handleCreate = useCallback(() => {
    navigate(paths.owners.new.getHref());
  }, [navigate]);

  // rerender-functional-setstate: useCallback で安定した関数参照を維持
  const handleEdit = useCallback(
    (ownerId: string) => {
      navigate(paths.owners.detail.getHref(ownerId));
    },
    [navigate],
  );

  // #158: 飼主レポートを別ウィンドウで開く。初期ペットを ?petId= で指定する。
  const handleReport = useCallback((ownerId: string, petId: string) => {
    openOwnerReport(ownerId, petId);
  }, []);

  const openDeleteModal = deleteModal.open;
  const handleDeleteRequest = useCallback(
    (ownerId: string, ownerName: string) => {
      openDeleteModal({ id: ownerId, name: ownerName });
    },
    [openDeleteModal],
  );

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
      title="飼主・ペット管理"
      resource={ResourceOwners}
      headerAction={
        canCreate ? (
          <PrimaryButton colorVariant="primary" onClick={handleCreate}>
            <Plus className={`mr-1.5 ${ICON.action}`} />
            新規飼主登録
          </PrimaryButton>
        ) : null
      }
      maxWidth={LAYOUT.pageContentMaxWidth.full}
    >
      {/* #86 / EMR-463: 許可医院が2院以上のときだけ拠点横断フィルタを表示 */}
      {clinics.length >= 2 ? (
        <div className="mb-3">
          <ClinicScopeFilter
            clinics={clinics}
            selectedIds={selectedClinicIds}
            onToggle={handleToggleClinic}
          />
        </div>
      ) : null}
      {listState.kind === "ready" && listState.scopeRechecked ? (
        <p
          role="status"
          aria-live="polite"
          aria-atomic="true"
          className={`mb-3 text-sm ${C.text60}`}
        >
          表示できる医院が変わったため、現在表示できる医院で一覧を取得し直しました。
        </p>
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
          動物種類が登録されていません。
        </p>
      ) : null}
      {listState.kind !== "ready" ? (
        <OwnersListScopeState
          listState={listState}
          onReload={handleReload}
          isReloading={revalidator.state === "loading" || navigation.state === "loading"}
        />
      ) : (
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
      )}

      {/* Delete Confirm Dialog */}
      <ConfirmDialog
        open={deleteModal.isOpen}
        onClose={() => {
          if (!isDeleting) deleteModal.close();
        }}
        onConfirm={handleConfirmDelete}
        title="飼主を削除しますか？"
        description={`飼主「${deleteModal.item?.name}」を削除します。この操作は取り消すことができません。なお、ペットが紐付いている場合は削除できません（先にペットを削除してください）。`}
        confirmLabel={isDeleting ? "削除中..." : "削除"}
        cancelLabel="キャンセル"
        variant="destructive"
      />
    </PageLayout>
  );
}
