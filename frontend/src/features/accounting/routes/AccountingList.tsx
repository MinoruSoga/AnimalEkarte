// React/Framework
import { ICON, C, LAYOUT } from "@/lib/design-tokens";
import { useCallback, useDeferredValue, useMemo, useState } from "react";
import { useNavigate, useSearchParams } from "react-router";

// Hooks
import { useSortableData } from "@/hooks/use-sortable-data";
import { useClinicScope } from "@/hooks/use-clinic-scope";
import { useUrlPageSync } from "@/hooks/use-url-page-sync";

// External
import { Plus, CreditCard } from "lucide-react";

// Internal
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { paths } from "@/config/paths";

// Relative
import { calculateAccountingTotal } from "../components/accounting-list-table-model";
import { useGetAccountingsPage } from "../api/get-accountings";
import { usePermission } from "@/hooks/use-permission";

// Types
import type { Accounting as AccountingType } from "../types";
import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import { ResourceAccounting } from "@/types/generated/models";
import { AccountingListContent } from "./AccountingListPanels";
import {
  ACCOUNTING_PAGE_SIZE,
  CLINIC_TOGGLE_RESET_PARAMS,
  accountingListTabFromParam,
  buildAccountingListPageFilters,
  buildServerPagePagination,
  nextAccountingListTabSearchParams,
  nextListSearchParamsWithPage,
  nextListSearchParamsWithoutPage,
} from "./accounting-list-model";

const EMPTY_ACCOUNTINGS: AccountingType[] = [];

export function AccountingList() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { canCreate, canEdit } = usePermission("accounting");

  const {
    assignedClinics,
    selectedClinicIds,
    isMultiClinic,
    clinicNameById,
    handleToggleClinic,
  } = useClinicScope({ resetParamsOnToggle: CLINIC_TOGGLE_RESET_PARAMS });

  const activeTab = accountingListTabFromParam(searchParams.get("tab"));
  const handleTabChange = useCallback((tab: string) => {
    setSearchParams(
      (prev) => nextAccountingListTabSearchParams(prev, tab),
      { replace: true },
    );
  }, [setSearchParams]);

  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const deferredSearch = useDeferredValue(searchTerm);
  const isFiltering = searchTerm !== deferredSearch;

  const urlPage = Math.max(1, Number(searchParams.get("page") ?? 1) || 1);

  const apiFilters = useMemo(
    () => buildAccountingListPageFilters({
      activeFilters,
      deferredSearch,
      isMultiClinic,
      selectedClinicIds,
      urlPage,
    }),
    [activeFilters, deferredSearch, isMultiClinic, selectedClinicIds, urlPage],
  );

  const { data: pageResult, isLoading, isError } = useGetAccountingsPage(apiFilters);
  const accountings = pageResult?.data ?? EMPTY_ACCOUNTINGS;

  const getSortValue = useCallback((item: AccountingType, key: string) => {
    if (key === "totalAmount") return calculateAccountingTotal(item);
    return String(item[key as keyof AccountingType] ?? "");
  }, []);

  const { activeSorts, setActiveSorts, toggleSort, directionFor, sortedData } =
    useSortableData(accountings, { getSortValue });

  const serverTotal = pageResult?.total ?? 0;
  const pagination = buildServerPagePagination({
    rows: sortedData,
    total: serverTotal,
    page: pageResult?.page ?? urlPage,
    limit: pageResult?.limit ?? ACCOUNTING_PAGE_SIZE,
  });

  useUrlPageSync({ urlPage, totalPages: pagination.totalPages, isLoading, setSearchParams });

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
    navigate(paths.accounting.selectPet.getHref());
  }, [navigate]);

  const handleEdit = useCallback((id: string) => {
    navigate(paths.accounting.detail.getHref(id));
  }, [navigate]);

  const handleMedicalRecordOpen = useCallback((medicalRecordId: string) => {
    navigate(paths.medicalRecords.detail.getHref(medicalRecordId));
  }, [navigate]);

  return (
    <PageLayout
      title="会計管理"
      resource={ResourceAccounting}
      icon={<CreditCard className={`${ICON.page} ${C.text}`} />}
      headerAction={
        activeTab === "list" && canCreate ? (
          <PrimaryButton onClick={handleCreate}>
            <Plus className={`mr-1.5 ${ICON.action}`} />
            新規会計登録
          </PrimaryButton>
        ) : null
      }
      maxWidth={LAYOUT.pageContentMaxWidth.full}
    >
      <AccountingListContent
        assignedClinics={assignedClinics}
        selectedClinicIds={selectedClinicIds}
        onToggleClinic={handleToggleClinic}
        activeTab={activeTab}
        onTabChange={handleTabChange}
        isMultiClinic={isMultiClinic}
        clinicNameById={clinicNameById}
        isLoading={isLoading}
        isError={isError}
        serverTotal={serverTotal}
        pagination={pagination}
        searchTerm={searchTerm}
        activeFilters={activeFilters}
        activeSorts={activeSorts}
        isFiltering={isFiltering}
        canEdit={canEdit}
        directionFor={directionFor}
        onSearchChange={handleSearchChange}
        onFilterChange={handleFilterChange}
        onSortChange={setActiveSorts}
        onToggleSort={toggleSort}
        onEdit={handleEdit}
        onMedicalRecordOpen={handleMedicalRecordOpen}
        onPageChange={handlePageChange}
      />
    </PageLayout>
  );
}
