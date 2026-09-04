// React/Framework
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { useCallback, useDeferredValue, useMemo, useState } from "react";
import { useNavigate, useSearchParams } from "react-router";

// External
import { ClipboardCheck, Plus } from "lucide-react";

// Internal
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { useSortableData } from "@/hooks/use-sortable-data";
import { usePermission } from "@/hooks/use-permission";
import { useUrlPageSync } from "@/hooks/use-url-page-sync";
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
  nextListSearchParamsWithoutPage,
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

  // FE-RC-028: ページ番号のソースはURLのみ。ローカルstateは持たない（ExaminationsListと同様の単一ソース化）。
  const urlPage = Math.max(1, Number(searchParams.get("page") ?? 1) || 1);

  const requestFilters = useMemo(
    () => ({ ...filters, page: urlPage, limit: PAGE_SIZE }),
    [filters, urlPage],
  );

  const { data: checkupsResult, isLoading, error } = useGetCheckups(requestFilters);
  const checkups = useMemo(() => checkupsResult?.data ?? [], [checkupsResult?.data]);
  const total = checkupsResult?.total ?? 0;
  const limit = checkupsResult?.limit ?? PAGE_SIZE;
  const totalPages = Math.max(1, Math.ceil(total / limit));
  const safePage = Math.min(urlPage, totalPages);

  const filteredRecords = useMemo(
    () => filterCheckupsBySearch(checkups, deferredSearch),
    [checkups, deferredSearch],
  );

  const { activeSorts, setActiveSorts, toggleSort, directionFor, sortedData } =
    useSortableData(filteredRecords);

  const startIndex = total === 0 ? 0 : (safePage - 1) * limit + 1;
  const endIndex = Math.min(safePage * limit, total);

  // FE-144 / FE-RC-028: URLの page がサーバ total から導いた totalPages を超えている場合はクランプする
  // （フィルタ変更等で母集団が縮んだ場合に空ページへ迷い込むのを防ぐ）。共通 hook に委譲。
  useUrlPageSync({
    urlPage,
    totalPages,
    isLoading,
    setSearchParams,
  });

  const handlePageChange = useCallback(
    (page: number) => {
      setSearchParams((prev) => nextListSearchParamsWithPage(prev, page), { replace: true });
    },
    [setSearchParams],
  );

  const handleFilterChange = useCallback(
    (next: ActiveFilter[]) => {
      setActiveFilters(next);
      // フィルタ変更時はページを1へ戻す（母集団が変わるため）
      setSearchParams((prev) => nextListSearchParamsWithoutPage(prev), { replace: true });
    },
    [setSearchParams],
  );

  const handleCreate = useCallback(() => {
    navigate(paths.checkups.selectPet.getHref());
  }, [navigate]);

  const handleEdit = useCallback(
    (medicalRecordId: string, checkupId: string) => {
      navigate(checkupChartHref(medicalRecordId, checkupId));
    },
    [navigate],
  );

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
        onFilterChange={handleFilterChange}
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
