// React/Framework
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { useState, useDeferredValue, useCallback, useEffect, useLayoutEffect, useMemo, useRef, useTransition } from "react";
import { useNavigate, useSearchParams } from "react-router";

// Hooks
import { useSortableData } from "@/hooks/use-sortable-data";

// External
import { Plus, Syringe } from "lucide-react";
import { toast } from "sonner";

// Internal
import { paths } from "@/config/paths";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { usePagination } from "@/hooks/use-pagination";
import { useGetPet } from "@/hooks/use-pet";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { handleApiError } from "@/lib/handle-api-error";

// Relative
import { useFilterVaccinations } from "../hooks/use-vaccinations";
import { useDeleteVaccination } from "../api/delete-vaccination";
import { usePermission } from "@/hooks/use-permission";

// Types
import type { VaccinationRecord } from "@/types";
import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import { ResourceVaccinations } from "@/types/generated/models";
import { VaccinationListContent, VaccinationListDeleteDialog } from "./VaccinationListPanels";
import {
  buildVaccinationFilterProperties,
  buildVaccinationListQueryOptions,
  nextListSearchParamsWithPage,
  orderVaccinationListRows,
  vaccinationDateRange,
  vaccinationListDetailHref,
} from "./vaccinations-list-model";

export function VaccinationList() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { canCreate, canEdit, canDelete } = usePermission("vaccinations");
  const [searchTerm, setSearchTerm] = useState("");
  const [pendingDeleteId, setPendingDeleteId] = useState<string | null>(null);
  const { mutate: deleteVaccinationFn } = useDeleteVaccination();
  const [isDeletePending, startDeleteTransition] = useTransition();
  const canDeleteRef = useRef(canDelete);
  useLayoutEffect(() => {
    canDeleteRef.current = canDelete;
  }, [canDelete]);
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const deferredSearchTerm = useDeferredValue(searchTerm);
  const isFiltering = searchTerm !== deferredSearchTerm;

  const filters = useMemo(() => {
    return buildVaccinationListQueryOptions({
      dateRange: vaccinationDateRange(activeFilters),
      search: deferredSearchTerm,
    });
  }, [activeFilters, deferredSearchTerm]);

  const { data: filteredRecords, allVaccinations, isLoading, error } = useFilterVaccinations(deferredSearchTerm, filters, activeFilters);
  const pendingDeletePetId = allVaccinations.find(
    (record) => record.id === pendingDeleteId,
  )?.petId ?? "";
  const { data: pendingDeletePet } = useGetPet(pendingDeletePetId);

  const filterProperties = useMemo(
    () => buildVaccinationFilterProperties(allVaccinations),
    [allVaccinations],
  );

  const defaultOrderedRecords = useMemo(
    () => orderVaccinationListRows(filteredRecords),
    [filteredRecords],
  );

  const { activeSorts, setActiveSorts, toggleSort, directionFor, sortedData } =
    useSortableData(defaultOrderedRecords);

  const pagination = usePagination(sortedData, {
    resetKey: deferredSearchTerm,
  });

  const urlPage = Number(searchParams.get("page") ?? 1);
  const { totalPages, currentPage, goToPage } = pagination;
  useEffect(() => {
    const clampedPage = Math.max(1, Math.min(urlPage, totalPages));
    if (clampedPage !== currentPage) {
      goToPage(clampedPage);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps -- FE-144: ページ変更時にURLクエリパラメータを更新
  }, [urlPage, totalPages]);

  const handlePageChange = useCallback((page: number) => {
    goToPage(page);
    setSearchParams(
      (prev) => nextListSearchParamsWithPage(prev, page),
      { replace: true },
    );
  }, [goToPage, setSearchParams]);

  const handleCreate = useCallback(() => {
    navigate(paths.vaccinations.selectPet.getHref());
  }, [navigate]);

  const handleEdit = useCallback((record: VaccinationRecord) => {
    navigate(vaccinationListDetailHref({
      id: record.id,
      medicalRecordId: record.medicalRecordId,
    }));
  }, [navigate]);

  const handleDeleteConfirm = useCallback(() => {
    if (
      canDeleteRef.current !== true ||
      !pendingDeleteId ||
      pendingDeletePet?.status !== "生存"
    ) {
      return;
    }
    startDeleteTransition(() => {
      deleteVaccinationFn(pendingDeleteId, {
        onSuccess: () => {
          toast.success("予防接種記録を削除しました");
          setPendingDeleteId(null);
        },
        onError: (error) => {
          handleApiError(error, "削除");
        },
      });
    });
  }, [pendingDeleteId, pendingDeletePet?.status, deleteVaccinationFn]);

  if (isLoading) return <LoadingFallback />;
  if (error) return <ErrorFallback />;

  return (
    <>
    <PageLayout
      title="予防接種管理"
      resource={ResourceVaccinations}
      icon={<Syringe className={`${ICON.page} ${C.text}`} />}
      headerAction={
        <div className="flex items-center gap-2">
          {canCreate ? (
            <PrimaryButton colorVariant="primary" onClick={handleCreate}>
              <Plus className={`mr-1.5 ${ICON.action}`} />
              新規登録
            </PrimaryButton>
          ) : null}
        </div>
      }
      maxWidth={LAYOUT.pageContentMaxWidth.full}
    >
      <VaccinationListContent
        filterProperties={filterProperties}
        activeFilters={activeFilters}
        onFilterChange={setActiveFilters}
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        count={filteredRecords.length}
        activeSorts={activeSorts}
        onSortChange={setActiveSorts}
        directionFor={directionFor}
        toggleSort={toggleSort}
        isFiltering={isFiltering}
        pagination={pagination}
        canEdit={canEdit}
        canDelete={canDelete}
        onEdit={handleEdit}
        onDelete={setPendingDeleteId}
        onPageChange={handlePageChange}
      />
    </PageLayout>
    <VaccinationListDeleteDialog
      open={pendingDeleteId !== null}
      isPending={isDeletePending}
      onClose={() => setPendingDeleteId(null)}
      onConfirm={handleDeleteConfirm}
    />
    </>
  );
}
