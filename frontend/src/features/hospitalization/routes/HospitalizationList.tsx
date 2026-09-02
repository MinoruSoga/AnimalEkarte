// React/Framework
import { useState, useCallback, useMemo, useDeferredValue, useEffect } from "react";
import { useSearchParams } from "react-router";

// Internal
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";

// Relative
import { useHospitalizationList } from "../hooks/use-hospitalization-list";
import { useGetHospitalizations } from "../api/get-hospitalizations";
import {
  HOSPITALIZATION_LIST_DEFAULT_LIMIT,
  HOSPITALIZATION_LIST_DEFAULT_PAGE,
} from "../constants";
import { usePermission } from "@/hooks/use-permission";

// Types
import type { ActiveFilter, ActiveSort } from "@/components/shared/PropertyFilter/types";
import { HospitalizationListPageView } from "./HospitalizationListPanels";
import {
  applyHospitalizationClientFilters,
  buildHospitalizationFilterProperties,
  buildHospitalizationListQueryFilters,
  buildServerPagePagination,
  isValidFilterStatus,
  nextListSearchParamsWithPage,
  nextListSearchParamsWithoutPage,
  sortHospitalizations,
} from "./hospitalization-list-model";

export function HospitalizationList() {
  const [searchParams, setSearchParams] = useSearchParams();
  const { canCreate, canEdit } = usePermission("hospitalization");
  const {
    searchTerm, setSearchTerm,
    statusFilter, setStatusFilter,
    viewMode, setViewMode,
    cages,
    movePet,
    handleNavigateToForm
  } = useHospitalizationList(canEdit);

  const [activeSorts, setActiveSorts] = useState<ActiveSort[]>([]);
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const deferredSearchTerm = useDeferredValue(searchTerm);

  const urlPage = Number(searchParams.get("page") ?? HOSPITALIZATION_LIST_DEFAULT_PAGE);
  const serverPage = Number.isFinite(urlPage) && urlPage >= 1 ? urlPage : HOSPITALIZATION_LIST_DEFAULT_PAGE;

  const listFilters = useMemo(
    () => buildHospitalizationListQueryFilters(activeFilters, statusFilter, serverPage),
    [activeFilters, statusFilter, serverPage],
  );

  const {
    data: hospitalizationsPage,
    isLoading: hospitalizationsLoading,
    isError: hospitalizationsError,
  } = useGetHospitalizations(listFilters);
  const allHospitalizations = useMemo(
    () => hospitalizationsPage?.data ?? [],
    [hospitalizationsPage?.data],
  );
  const serverTotal = hospitalizationsPage?.total ?? 0;

  const filterProperties = useMemo(
    () => buildHospitalizationFilterProperties(allHospitalizations),
    [allHospitalizations],
  );

  const handleSortChange = useCallback((sorts: ActiveSort[]) => {
    setActiveSorts(sorts);
  }, []);

  const typeFilteredHospitalizations = useMemo(
    () => applyHospitalizationClientFilters(allHospitalizations, deferredSearchTerm, activeFilters),
    [allHospitalizations, deferredSearchTerm, activeFilters],
  );

  const sortedHospitalizations = useMemo(
    () => sortHospitalizations(typeFilteredHospitalizations, activeSorts),
    [typeFilteredHospitalizations, activeSorts],
  );

  const pagination = buildServerPagePagination({
    rows: sortedHospitalizations,
    total: serverTotal,
    page: serverPage,
    limit: HOSPITALIZATION_LIST_DEFAULT_LIMIT,
  });

  useEffect(() => {
    if (hospitalizationsLoading) return;
    const clampedPage = Math.max(1, Math.min(urlPage, pagination.totalPages));
    if (clampedPage !== urlPage) {
      setSearchParams(
        (prev) => nextListSearchParamsWithPage(prev, clampedPage),
        { replace: true },
      );
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps -- setSearchParams は安定。urlPage/totalPages/loading 変化時のみ
  }, [urlPage, pagination.totalPages, hospitalizationsLoading]);

  const resetListPage = useCallback(() => {
    setSearchParams(
      (prev) => nextListSearchParamsWithoutPage(prev),
      { replace: true },
    );
  }, [setSearchParams]);

  const handlePageChange = useCallback((page: number) => {
    setSearchParams(
      (prev) => nextListSearchParamsWithPage(prev, page),
      { replace: true },
    );
  }, [setSearchParams]);

  const handleStatusTabChange = useCallback(
    (v: string) => {
      if (!isValidFilterStatus(v)) return;
      setStatusFilter(v);
      resetListPage();
    },
    [setStatusFilter, resetListPage],
  );

  const handleSearchChange = useCallback(
    (value: string) => {
      setSearchTerm(value);
      resetListPage();
    },
    [setSearchTerm, resetListPage],
  );

  const handleFilterChange = useCallback(
    (next: ActiveFilter[]) => {
      setActiveFilters(next);
      resetListPage();
    },
    [resetListPage],
  );

  if (hospitalizationsLoading) return <LoadingFallback />;
  if (hospitalizationsError) return <ErrorFallback />;

  return (
    <HospitalizationListPageView
      canCreateHeader={canCreate}
      onCreate={() => handleNavigateToForm()}
      statusFilter={statusFilter}
      onStatusTabChange={handleStatusTabChange}
      filterProperties={filterProperties}
      activeFilters={activeFilters}
      onFilterChange={handleFilterChange}
      searchTerm={searchTerm}
      onSearchChange={handleSearchChange}
      serverTotal={serverTotal}
      activeSorts={activeSorts}
      onSortChange={handleSortChange}
      viewMode={viewMode}
      onViewModeChange={setViewMode}
      cages={cages}
      boardHospitalizations={typeFilteredHospitalizations}
      onNavigateToForm={handleNavigateToForm}
      onMovePet={movePet}
      canCreate={canCreate}
      canEdit={canEdit}
      pagination={pagination}
      onPageChange={handlePageChange}
    />
  );
}
