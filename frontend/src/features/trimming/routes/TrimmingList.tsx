// React/Framework
import { useState, useCallback, useDeferredValue, useEffect, useMemo } from "react";
import { useNavigate, useSearchParams } from "react-router";

// Hooks
import { useSortableData } from "@/hooks/use-sortable-data";
import { useModalState } from "@/hooks/use-modal-state";
import { uniqueSortedOptions } from "@/lib/unique-sorted-options";

// External
import { toast } from "sonner";

// Types
import type { ActiveFilter, FilterProperty } from "@/components/shared/PropertyFilter/types";

// Internal
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { usePagination } from "@/hooks/use-pagination";
import { useUrlPageSync } from "@/hooks/use-url-page-sync";
import { useStaffValidation } from "@/hooks/use-staff-validation";
import type { TrimmingUI } from "@/types";
import { paths } from "@/config/paths";

// Relative
import { useFilterTrimmingRecords } from "../hooks/use-trimming-records";
import type { TrimmingFilters } from "../api/get-trimmings";
import { usePermission } from "@/hooks/use-permission";
import { handleApiError } from "@/lib/handle-api-error";
import { buildTrimmingDynamicFilterProperties } from "../components/trimming-list-table-model";
import { TrimmingListContent } from "./TrimmingListPanels";

export function TrimmingList() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { canCreate, canEdit, canDelete } = usePermission("trimming");
  const [searchKeyword, setSearchKeyword] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);

  const deferredKeyword = useDeferredValue(searchKeyword);
  const isFiltering = searchKeyword !== deferredKeyword;

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

  const urlPage = Number(searchParams.get("page") ?? 1);

  // FE-RC-028: totalPages を超える URL page をクランプして書き戻す（フィルタ変更等で母集団が縮んだ場合の空ページ対策）。
  useUrlPageSync({ urlPage, totalPages, isLoading, setSearchParams });

  useEffect(() => {
    const clampedPage = Math.max(1, Math.min(urlPage, totalPages));
    if (clampedPage !== currentPage) {
      goToPage(clampedPage);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps -- currentPage/goToPage は意図的に除外（URL変更時のみ同期する設計。FE-144）
  }, [urlPage, totalPages]);

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

  const handleSearchChange = useCallback((v: string) => {
    setSearchKeyword(v);
  }, []);

  if (isLoading) return <LoadingFallback />;
  if (error) return <ErrorFallback />;

  return (
    <TrimmingListContent
      canCreate={canCreate}
      onNew={handleNew}
      isTruncated={isTruncated}
      paginatedData={paginatedData}
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
      deleteOpen={deleteModal.isOpen}
      deleteLabel={deleteModal.item?.label}
      onDeleteClose={deleteModal.close}
      onDeleteConfirm={handleDeleteConfirm}
    />
  );
}
