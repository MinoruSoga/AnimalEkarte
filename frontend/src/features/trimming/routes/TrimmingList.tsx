// React/Framework
import { ICON, C, LAYOUT } from "@/lib/design-tokens";
import { useState, useCallback, useDeferredValue, useEffect, useMemo } from "react";
import { useNavigate, useSearchParams } from "react-router";

// Hooks
import { useSortableData } from "@/hooks/use-sortable-data";
import { useModalState } from "@/hooks/use-modal-state";
import { uniqueSortedOptions } from "@/lib/unique-sorted-options";

// External
import { Plus, Scissors } from "lucide-react";
import { toast } from "sonner";

// Types
import type { ActiveFilter, FilterProperty } from "@/components/shared/PropertyFilter/types";

// Internal
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { usePagination } from "@/hooks/use-pagination";
import { useStaffValidation } from "@/hooks/use-staff-validation";
import type { TrimmingUI } from "@/types";
import { paths } from "@/config/paths";
import { HISTORY_FETCH_LIMIT } from "@/config/fetch-limits";

// Relative (direct file import, no barrel)
import { useFilterTrimmingRecords } from "../hooks/use-trimming-records";
import type { TrimmingFilters } from "../api/get-trimmings";
import { usePermission } from "@/hooks/use-permission";
import { ResourceTrimming } from "@/types/generated/models";
import { handleApiError } from "@/lib/handle-api-error";
import { TrimmingListTable } from "../components/TrimmingListTable";
import { buildTrimmingDynamicFilterProperties } from "../components/trimming-list-table-model";

export function TrimmingList() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { canCreate, canEdit, canDelete } = usePermission("trimming");
  const [searchKeyword, setSearchKeyword] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);

  // rerender-transitions: 検索は useDeferredValue で遅延
  const deferredKeyword = useDeferredValue(searchKeyword);
  const isFiltering = searchKeyword !== deferredKeyword;

  // rerender-dependencies: activeFilters から日付フィルタを抽出してサーバーサイドフィルタに渡す
  const filters = useMemo<TrimmingFilters>(() => {
    const dateFilter = activeFilters.find((f) => f.key === "date")?.value as
      | { from?: string; to?: string }
      | undefined;
    return {
      startDate: dateFilter?.from,
      endDate: dateFilter?.to,
    };
  }, [activeFilters]);

  const {
    data: filteredRecords,
    allTrimmings,
    isTruncated,
    isLoading,
    error,
    deleteRecord,
  } = useFilterTrimmingRecords(deferredKeyword, filters, activeFilters);
  const { isValidStaff } = useStaffValidation();

  // js-cache-function-results: ロード済みデータから種・担当の選択肢を動的生成
  const filterProperties = useMemo<FilterProperty[]>(() => {
    const speciesOptions = uniqueSortedOptions(allTrimmings, (r) => r.species);
    const staffOptions = uniqueSortedOptions(allTrimmings, (r) => r.staff);
    return buildTrimmingDynamicFilterProperties(speciesOptions, staffOptions);
  }, [allTrimmings]);

  const { activeSorts, setActiveSorts, toggleSort, directionFor, sortedData } =
    useSortableData(filteredRecords);

  const {
    currentPage,
    paginatedData,
    totalPages,
    startIndex,
    endIndex,
    goToPage,
  } = usePagination(sortedData, { pageSize: 10, resetKey: [deferredKeyword, JSON.stringify(activeFilters)].join("|") });

  // FE-144: URLクエリパラメータからページ番号を読み取る
  const urlPage = Number(searchParams.get("page") ?? 1);

  // FE-144: URLのページ番号とローカル状態を同期（URLが変わったときのみ）
  useEffect(() => {
    const clampedPage = Math.max(1, Math.min(urlPage, totalPages));
    if (clampedPage !== currentPage) {
      goToPage(clampedPage);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps -- currentPage/goToPage は意図的に除外（URL変更時のみ同期する設計。FE-144）
  }, [urlPage, totalPages]);

  // FE-144: ページ変更時にURLクエリパラメータを更新
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

  const deleteModal = useModalState<{ id: string; label: string }>();

  const handleEdit = useCallback((id: string) => {
    navigate(paths.trimming.detail.getHref(id), { state: { from: paths.trimming.getHref() } });
  }, [navigate]);

  // rerender-dependencies: deleteModal のメソッドを primitive に抽出して deps を安定化
  const openDeleteModal = deleteModal.open;
  const closeDeleteModal = deleteModal.close;
  const deleteTargetId = deleteModal.item?.id;
  const deleteTargetLabel = deleteModal.item?.label;

  const handleDeleteClick = useCallback((record: TrimmingUI) => {
    openDeleteModal({
      id: record.id,
      label: `${record.ownerName} - ${record.petName}`,
    });
  }, [openDeleteModal]);

  const handleDeleteConfirm = useCallback(() => {
    if (deleteTargetId && deleteTargetLabel) {
      deleteRecord(deleteTargetId, {
        onSuccess: () => {
          toast.success("削除しました", { description: deleteTargetLabel });
          closeDeleteModal();
        },
        onError: (error) => {
          handleApiError(error, "トリミング削除");
        },
      });
    }
  }, [deleteTargetId, deleteTargetLabel, deleteRecord, closeDeleteModal]);

  const handleNew = useCallback(() => {
    navigate(paths.trimming.selectPet.getHref());
  }, [navigate]);

  // rerender-functional-setstate: setSearchKeyword を useCallback でラップして安定化
  const handleSearchChange = useCallback((v: string) => {
    setSearchKeyword(v);
  }, []);

  if (isLoading) return <LoadingFallback />;
  if (error) return <ErrorFallback />;

  return (
    <PageLayout
      title="トリミング管理"
      icon={<Scissors className={`${ICON.page} ${C.text}`} />}
      resource={ResourceTrimming}
      headerAction={
        canCreate ? (
          <PrimaryButton onClick={handleNew}>
            <Plus className={`mr-1.5 ${ICON.action}`} />
            新規登録
          </PrimaryButton>
        ) : null
      }
      maxWidth={LAYOUT.pageContentMaxWidth.full}
    >
      <div className="flex flex-col gap-4">
        {isTruncated ? (
          <p className={`text-xs ${C.text50}`} role="status">
            取得上限の{HISTORY_FETCH_LIMIT}件を対象に検索・絞り込みしています
          </p>
        ) : null}
        <TrimmingListTable
          records={paginatedData}
          filteredCount={filteredRecords.length}
          currentPage={currentPage}
          totalPages={totalPages}
          startIndex={startIndex}
          endIndex={endIndex}
          searchKeyword={searchKeyword}
          activeFilters={activeFilters}
          activeSorts={activeSorts}
          filterProperties={filterProperties}
          isFiltering={isFiltering}
          canEdit={canEdit}
          canDelete={canDelete}
          isValidStaff={isValidStaff}
          directionFor={directionFor}
          onSearchChange={handleSearchChange}
          onFilterChange={setActiveFilters}
          onSortChange={setActiveSorts}
          onToggleSort={toggleSort}
          onEdit={handleEdit}
          onDeleteClick={handleDeleteClick}
          onPageChange={handlePageChange}
        />
      </div>

      {/* Delete Confirmation Dialog */}
      <ConfirmDialog
        open={deleteModal.isOpen}
        onClose={deleteModal.close}
        title="削除確認"
        description={`${deleteModal.item?.label} を削除してもよろしいですか？`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={handleDeleteConfirm}
      />
    </PageLayout>
  );
}
