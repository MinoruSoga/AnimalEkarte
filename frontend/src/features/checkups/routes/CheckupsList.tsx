// React/Framework
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { useCallback, useDeferredValue, useEffect, useMemo, useState } from "react";
import { useNavigate, useSearchParams } from "react-router";

// External
import { ClipboardCheck, Plus } from "lucide-react";

// Internal
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { useSortableData } from "@/hooks/use-sortable-data";
import { usePermission } from "@/hooks/use-permission";
import { paths } from "@/config/paths";
import { useGetCheckups } from "../api/get-checkups";

// Types
import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import { ResourceCheckups, ResourceMedicalRecords } from "@/types/generated/models";
import { CheckupsListContent } from "./CheckupsListPanels";
import {
  PAGE_SIZE,
  buildCheckupListFilters,
  checkupChartHref,
  filterCheckupsBySearch,
  nextListSearchParamsWithPage,
} from "./checkups-list-model";

export function CheckupsList() {
  const navigate = useNavigate();
  const { canView, canCreate, canEdit } = usePermission(ResourceMedicalRecords);
  const canCreateCheckup = canCreate && canEdit;
  const [searchParams, setSearchParams] = useSearchParams();
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const deferredSearch = useDeferredValue(searchTerm);
  const isFiltering = searchTerm !== deferredSearch;

  const filters = useMemo(() => buildCheckupListFilters(activeFilters), [activeFilters]);

  const [currentPage, setCurrentPage] = useState(1);
  const filtersResetKey = JSON.stringify(filters);
  const [prevFiltersResetKey, setPrevFiltersResetKey] = useState(filtersResetKey);
  if (prevFiltersResetKey !== filtersResetKey) {
    setPrevFiltersResetKey(filtersResetKey);
    setCurrentPage(1);
  }

  const requestFilters = useMemo(
    () => ({ ...filters, page: currentPage, limit: PAGE_SIZE }),
    [filters, currentPage],
  );

  const { data: checkupsResult, isLoading, error } = useGetCheckups(requestFilters);
  const checkups = useMemo(
    () => checkupsResult?.data ?? [],
    [checkupsResult?.data],
  );
  const total = checkupsResult?.total ?? 0;
  const limit = checkupsResult?.limit ?? PAGE_SIZE;
  const totalPages = Math.max(1, Math.ceil(total / limit));
  const safePage = Math.min(currentPage, totalPages);

  const filteredRecords = useMemo(
    () => filterCheckupsBySearch(checkups, deferredSearch),
    [checkups, deferredSearch],
  );

  const { activeSorts, setActiveSorts, toggleSort, directionFor, sortedData } =
    useSortableData(filteredRecords);

  const startIndex = total === 0 ? 0 : (safePage - 1) * limit + 1;
  const endIndex = Math.min(safePage * limit, total);

  // FE-144: URLクエリパラメータからページ番号を読み取り、ローカル状態と同期
  // （URLが変わったときのみ。totalPages はサーバ応答後に確定するためクランプが必要）
  const urlPage = Number(searchParams.get("page") ?? 1);
  useEffect(() => {
    const clampedPage = Math.max(1, Math.min(urlPage, totalPages));
    if (clampedPage !== currentPage) {
      // URL/サーバ total 由来のページ同期。render 中 setState は不可のため effect で反映する。
      // eslint-disable-next-line react-hooks/set-state-in-effect -- FE-144 URL page clamp sync
      setCurrentPage(clampedPage);
    }
  // currentPage は比較対象のみ。URL/totalPages 変化時だけ同期する（自己ループ防止）
  // eslint-disable-next-line react-hooks/exhaustive-deps -- FE-144 URL page sync
  }, [urlPage, totalPages]);

  const handlePageChange = useCallback((page: number) => {
    setCurrentPage(page);
    setSearchParams(
      (prev) => nextListSearchParamsWithPage(prev, page),
      { replace: true },
    );
  }, [setSearchParams]);

  const handleCreate = useCallback(() => {
    navigate(paths.checkups.selectPet.getHref());
  }, [navigate]);

  const handleEdit = useCallback((medicalRecordId: string, checkupId: string) => {
    navigate(checkupChartHref(medicalRecordId, checkupId));
  }, [navigate]);

  if (isLoading) return <LoadingFallback />;
  if (error) return <ErrorFallback />;

  return (
    <PageLayout
      title="定期健診"
      resource={ResourceCheckups}
      icon={<ClipboardCheck className={`${ICON.page} ${C.text}`} />}
      maxWidth={LAYOUT.pageContentMaxWidth.full}
      headerAction={
        canCreateCheckup ? (
          <PrimaryButton colorVariant="primary" onClick={handleCreate}>
            <Plus className={`mr-1.5 ${ICON.action}`} />
            新規登録
          </PrimaryButton>
        ) : null
      }
    >
      <CheckupsListContent
        activeFilters={activeFilters}
        onFilterChange={setActiveFilters}
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        count={isLoading ? undefined : filteredRecords.length}
        activeSorts={activeSorts}
        onSortChange={setActiveSorts}
        directionFor={directionFor}
        toggleSort={toggleSort}
        isFiltering={isFiltering}
        records={sortedData}
        canView={canView}
        canEdit={canEdit}
        onEdit={handleEdit}
        totalPages={totalPages}
        safePage={safePage}
        total={total}
        startIndex={startIndex}
        endIndex={endIndex}
        onPageChange={handlePageChange}
      />
    </PageLayout>
  );
}
