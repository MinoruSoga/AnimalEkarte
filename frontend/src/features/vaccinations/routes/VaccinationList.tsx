import {
  useState,
  useDeferredValue,
  useCallback,
  useLayoutEffect,
  useMemo,
  useRef,
  useTransition,
} from "react";
import { useLocation, useNavigate, useSearchParams } from "react-router";
import { Plus, Syringe } from "lucide-react";
import { toast } from "sonner";

import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { paths } from "@/config/paths";
import { useSortableData } from "@/hooks/use-sortable-data";
import { usePagination } from "@/hooks/use-pagination";
import { usePermission } from "@/hooks/use-permission";
import { useUrlPageSync } from "@/hooks/use-url-page-sync";
import { useGetPet } from "@/hooks/use-pet";
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { ResourceVaccinations } from "@/types/generated/models";

import { useFilterVaccinations } from "../hooks/use-vaccinations";
import { useDeleteVaccination } from "../api/delete-vaccination";
import { VaccinationListContent, VaccinationListDeleteDialog } from "./VaccinationListPanels";
import {
  buildVaccinationFilterProperties,
  buildVaccinationListQueryOptions,
  nextListSearchParamsWithPage,
  orderVaccinationListRows,
  vaccinationDateRange,
  vaccinationListDetailHref,
  vaccinationListDetailState,
} from "./vaccinations-list-model";

import type { VaccinationRecord } from "@/types";
import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";

export function VaccinationList() {
  const navigate = useNavigate();
  const location = useLocation();
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

  const petId = searchParams.get("petId")?.trim() ?? "";
  const filters = useMemo(() => {
    const options = buildVaccinationListQueryOptions({
      dateRange: vaccinationDateRange(activeFilters),
      search: deferredSearchTerm,
    });
    if (petId === "") return options;
    return { ...options, petId };
  }, [activeFilters, deferredSearchTerm, petId]);

  const {
    data: filteredRecords,
    allVaccinations,
    isLoading,
    error,
  } = useFilterVaccinations(deferredSearchTerm, filters, activeFilters);
  const pendingDeletePetId =
    allVaccinations.find((record) => record.id === pendingDeleteId)?.petId ?? "";
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

  const rawUrlPage = Number(searchParams.get("page") ?? 1);
  const urlPage = Number.isFinite(rawUrlPage) && rawUrlPage >= 1 ? Math.floor(rawUrlPage) : 1;
  const { currentPage, totalPages, goToPage } = pagination;

  // URL の page が正本。戻る・直接アクセスの有効な page を内部ページへ合わせ、範囲外はクランプする。
  useLayoutEffect(() => {
    if (isLoading) return;
    const clamped = Math.min(urlPage, Math.max(totalPages, 1));
    if (currentPage !== clamped) goToPage(clamped);
  }, [currentPage, goToPage, isLoading, totalPages, urlPage]);

  // FE-RC-028: totalPages を超える URL page をクランプして書き戻す（フィルタ変更等で母集団が縮んだ場合の空ページ対策）。
  useUrlPageSync({
    urlPage,
    totalPages,
    isLoading,
    setSearchParams,
    preserveNavigationState: true,
    navigationState: location.state,
  });

  const handleSearchChange = useCallback(
    (value: string) => {
      setSearchTerm(value);
      setSearchParams((prev) => nextListSearchParamsWithPage(prev, 1), {
        replace: true,
        state: location.state,
      });
    },
    [location.state, setSearchParams],
  );

  const handlePageChange = useCallback(
    (page: number) => {
      goToPage(page);
      setSearchParams((prev) => nextListSearchParamsWithPage(prev, page), {
        replace: true,
        state: location.state,
      });
    },
    [goToPage, location.state, setSearchParams],
  );

  const detailState = useMemo(
    () =>
      vaccinationListDetailState({
        pathname: location.pathname,
        search: location.search,
        locationState: location.state,
      }),
    [location.pathname, location.search, location.state],
  );

  const handleCreate = useCallback(() => {
    navigate(paths.vaccinations.selectPet.getHref());
  }, [navigate]);

  const handleEdit = useCallback(
    (record: VaccinationRecord) => {
      navigate(
        vaccinationListDetailHref({
          id: record.id,
          medicalRecordId: record.medicalRecordId,
        }),
        { state: detailState },
      );
    },
    [detailState, navigate],
  );

  const handleDeleteConfirm = useCallback(() => {
    if (canDeleteRef.current !== true || !pendingDeleteId || pendingDeletePet?.status !== "生存") {
      return;
    }
    startDeleteTransition(() => {
      // FE-RC-005: onError は指定しない。useDeleteVaccination の onError
      // （api/delete-vaccination.ts）が handleApiError 済みで、ここで重ねると二重 toast になる。
      deleteVaccinationFn(pendingDeleteId, {
        onSuccess: () => {
          toast.success("予防接種記録を削除しました");
          setPendingDeleteId(null);
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
          onSearchChange={handleSearchChange}
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
          detailState={detailState}
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
