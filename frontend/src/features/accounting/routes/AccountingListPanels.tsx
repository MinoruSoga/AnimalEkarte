import type { ActiveFilter, ActiveSort } from "@/components/shared/PropertyFilter/types";
import type { ClinicMembership } from "@/types/auth";
import { UnifiedTabs, UnifiedTabsContent } from "@/components/shared/UnifiedTabs";
import { ClinicScopeFilter } from "@/components/shared/ClinicScopeFilter/ClinicScopeFilter";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { UnpaidTab } from "../components/UnpaidTab";
import { DailyAccountingTab } from "../components/DailyAccountingTab";
import { AccountingListTable } from "../components/AccountingListTable";
import type { Accounting as AccountingType } from "../types";
import { TABS, type AccountingListTab, type ServerPagePagination } from "./accounting-list-model";

interface AccountingListContentProps {
  assignedClinics: ClinicMembership[];
  selectedClinicIds: string[];
  onToggleClinic: (clinicId: string) => void;
  activeTab: AccountingListTab;
  onTabChange: (tab: string) => void;
  isMultiClinic: boolean;
  clinicNameById: Map<string, string>;
  isLoading: boolean;
  isError: boolean;
  serverTotal: number;
  pagination: ServerPagePagination<AccountingType>;
  searchTerm: string;
  activeFilters: ActiveFilter[];
  activeSorts: ActiveSort[];
  isFiltering: boolean;
  canEdit: boolean;
  directionFor: (key: string) => "ascending" | "descending" | "none";
  onSearchChange: (value: string) => void;
  onFilterChange: (next: ActiveFilter[]) => void;
  onSortChange: (sorts: ActiveSort[]) => void;
  onToggleSort: (key: string) => void;
  onEdit: (id: string) => void;
  onMedicalRecordOpen: (medicalRecordId: string) => void;
  onPageChange: (page: number) => void;
}

export function AccountingListContent({
  assignedClinics,
  selectedClinicIds,
  onToggleClinic,
  activeTab,
  onTabChange,
  isMultiClinic,
  clinicNameById,
  isLoading,
  isError,
  serverTotal,
  pagination,
  searchTerm,
  activeFilters,
  activeSorts,
  isFiltering,
  canEdit,
  directionFor,
  onSearchChange,
  onFilterChange,
  onSortChange,
  onToggleSort,
  onEdit,
  onMedicalRecordOpen,
  onPageChange,
}: AccountingListContentProps) {
  return (
    <div className="flex flex-col gap-4">
      {assignedClinics.length > 1 ? (
        <ClinicScopeFilter
          clinics={assignedClinics}
          selectedIds={selectedClinicIds}
          onToggle={onToggleClinic}
        />
      ) : null}
      <UnifiedTabs
        items={TABS}
        value={activeTab}
        onValueChange={onTabChange}
      >
        <UnifiedTabsContent value="daily" className="mt-4">
          <DailyAccountingTab
            selectedClinicIds={isMultiClinic ? selectedClinicIds : undefined}
            clinicNameById={clinicNameById}
          />
        </UnifiedTabsContent>
        <UnifiedTabsContent value="unpaid" className="mt-4">
          <UnpaidTab />
        </UnifiedTabsContent>
        <UnifiedTabsContent value="list" className="mt-4">
          {isLoading ? <LoadingFallback /> : null}
          {isError ? <ErrorFallback /> : null}
          {!isLoading && !isError ? (
            <AccountingListTable
              filteredCount={serverTotal}
              pagination={pagination}
              searchTerm={searchTerm}
              activeFilters={activeFilters}
              activeSorts={activeSorts}
              isFiltering={isFiltering}
              canEdit={canEdit}
              directionFor={directionFor}
              onSearchChange={onSearchChange}
              onFilterChange={onFilterChange}
              onSortChange={onSortChange}
              onToggleSort={onToggleSort}
              onEdit={onEdit}
              onMedicalRecordOpen={onMedicalRecordOpen}
              onPageChange={onPageChange}
              showClinicColumn={isMultiClinic}
              clinicNameById={clinicNameById}
            />
          ) : null}
        </UnifiedTabsContent>
      </UnifiedTabs>
    </div>
  );
}
