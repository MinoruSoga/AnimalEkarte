// React/Framework
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { useState, useDeferredValue, useCallback, useMemo } from "react";
import { useNavigate, useSearchParams } from "react-router";

// Hooks
import { useSortableData } from "@/hooks/use-sortable-data";
import { useUrlPageSync } from "@/hooks/use-url-page-sync";

// External
import { Plus, TestTube } from "lucide-react";

// Internal
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";

// Relative
import { useFilterExaminationRecords } from "../hooks/use-examination-records";
import { paths } from "@/config/paths";
import { usePermission } from "@/hooks/use-permission";

// Types
import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import type { ExaminationRecord } from "../api/transforms";
import { ResourceExaminations, ResourceMedicalRecords } from "@/types/generated/models";
import { ExaminationsListContent } from "./ExaminationsListPanels";
import {
  EXAMINATIONS_PAGE_SIZE,
  buildExaminationFilterProperties,
  buildServerPagePagination,
  examinationDateFilters,
  examinationListDetailHref,
  hasExaminationPageScopedFilter,
  nextListSearchParamsWithPage,
  nextListSearchParamsWithoutPage,
} from "./examinations-list-model";

export function ExaminationsList() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { canCreate, canEdit } = usePermission(ResourceExaminations);
  const { canView: canViewMedicalRecords } = usePermission(ResourceMedicalRecords);
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const deferredSearch = useDeferredValue(searchTerm);
  const isFiltering = searchTerm !== deferredSearch;

  // FE-144: URLクエリパラメータからページ番号を読み取る（BUG-411: サーバフェッチのキーそのものになる）
  const urlPage = Math.max(1, Number(searchParams.get("page") ?? 1) || 1);

  const filters = useMemo(() => examinationDateFilters(activeFilters), [activeFilters]);

  const { data: filteredRecords, allExaminations, isLoading, error, total, page: serverPage, limit: serverLimit } =
    useFilterExaminationRecords(deferredSearch, { page: urlPage, limit: EXAMINATIONS_PAGE_SIZE }, filters, activeFilters);

  const filterProperties = useMemo(
    () => buildExaminationFilterProperties(allExaminations),
    [allExaminations],
  );

  const { activeSorts, setActiveSorts, toggleSort, directionFor, sortedData } =
    useSortableData(filteredRecords);

  const hasPageScopedFilter = hasExaminationPageScopedFilter(deferredSearch, activeFilters);

  const pagination = buildServerPagePagination({
    rows: sortedData,
    total,
    page: serverPage,
    limit: serverLimit,
  });

  // FE-144 / BUG-411 / FE-RC-028: URLの page がサーバ total から導いた totalPages を超えている場合はクランプする
  // （日付フィルタ変更等で母集団が縮んだ場合に空ページへ迷い込むのを防ぐ）。共通 hook に委譲。
  useUrlPageSync({
    urlPage,
    totalPages: pagination.totalPages,
    isLoading,
    setSearchParams,
  });

  const handlePageChange = useCallback((page: number) => {
    setSearchParams(
      (prev) => nextListSearchParamsWithPage(prev, page),
      { replace: true },
    );
  }, [setSearchParams]);

  const handleSearchChange = useCallback((value: string) => {
    setSearchTerm(value);
    setSearchParams(
      (prev) => nextListSearchParamsWithoutPage(prev),
      { replace: true },
    );
  }, [setSearchParams]);

  const handleFilterChange = useCallback((next: ActiveFilter[]) => {
    setActiveFilters(next);
    setSearchParams(
      (prev) => nextListSearchParamsWithoutPage(prev),
      { replace: true },
    );
  }, [setSearchParams]);

  const handleCreate = useCallback(() => {
    navigate(paths.examinations.selectPet.getHref());
  }, [navigate]);

  const handleEdit = useCallback((record: ExaminationRecord) => {
    navigate(
      examinationListDetailHref({
        id: record.id,
        medicalRecordId: canViewMedicalRecords ? record.medicalRecordId : undefined,
      }),
    );
  }, [canViewMedicalRecords, navigate]);

  if (isLoading) return <LoadingFallback />;
  if (error) return <ErrorFallback />;

  return (
    <PageLayout
      title="検査管理"
      resource={ResourceExaminations}
      icon={<TestTube className={`${ICON.page} ${C.text}`} />}
      headerAction={
        <div className="flex items-center gap-2">
          {canCreate ? (
            <PrimaryButton colorVariant="primary" onClick={handleCreate}>
              <Plus className={`mr-1.5 ${ICON.action}`} />
              新規検査登録
            </PrimaryButton>
          ) : null}
        </div>
      }
      maxWidth={LAYOUT.pageContentMaxWidth.full}
    >
      <ExaminationsListContent
        filterProperties={filterProperties}
        activeFilters={activeFilters}
        onFilterChange={handleFilterChange}
        searchTerm={searchTerm}
        onSearchChange={handleSearchChange}
        count={isLoading ? undefined : filteredRecords.length}
        activeSorts={activeSorts}
        onSortChange={setActiveSorts}
        directionFor={directionFor}
        toggleSort={toggleSort}
        isFiltering={isFiltering}
        hasPageScopedFilter={hasPageScopedFilter}
        pagination={pagination}
        canEdit={canEdit}
        canViewMedicalRecords={canViewMedicalRecords}
        onEdit={handleEdit}
        onPageChange={handlePageChange}
      />
    </PageLayout>
  );
}
