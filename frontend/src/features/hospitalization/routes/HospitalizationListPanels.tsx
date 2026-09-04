import { LayoutGrid, List } from "lucide-react";
import type {
  ActiveFilter,
  ActiveSort,
  FilterProperty,
} from "@/components/shared/PropertyFilter/types";
import { UnifiedTabs } from "@/components/shared/UnifiedTabs";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { PropertyFilter } from "@/components/shared/PropertyFilter/PropertyFilter";
import { Pagination } from "@/components/shared/Pagination/Pagination";
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import type { Hospitalization, MasterItem } from "@/types";
import { Plus } from "lucide-react";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { ResourceHospitalization } from "@/types/generated/models";
import { HospitalizationBoard } from "../components/HospitalizationBoard";
import { HospitalizationListView } from "../components/HospitalizationListView";
import {
  HOSPITALIZATION_SORT_PROPERTIES,
  HOSPITALIZATION_TAB_ITEMS,
  isValidViewMode,
  type ServerPagePagination,
  type ViewMode,
} from "./hospitalization-list-model";

interface HospitalizationListContentProps {
  statusFilter: string;
  onStatusTabChange: (value: string) => void;
  filterProperties: FilterProperty[];
  activeFilters: ActiveFilter[];
  onFilterChange: (next: ActiveFilter[]) => void;
  searchTerm: string;
  onSearchChange: (value: string) => void;
  serverTotal: number;
  activeSorts: ActiveSort[];
  onSortChange: (sorts: ActiveSort[]) => void;
  viewMode: ViewMode;
  onViewModeChange: (mode: ViewMode) => void;
  cages: MasterItem[];
  boardHospitalizations: Hospitalization[];
  onNavigateToForm: (id?: string) => void;
  onMovePet: (hospitalizationId: string, targetCageId: string) => void;
  canCreate: boolean;
  canEdit: boolean;
  pagination: ServerPagePagination<Hospitalization>;
  onPageChange: (page: number) => void;
}

function HospitalizationListContent({
  statusFilter,
  onStatusTabChange,
  filterProperties,
  activeFilters,
  onFilterChange,
  searchTerm,
  onSearchChange,
  serverTotal,
  activeSorts,
  onSortChange,
  viewMode,
  onViewModeChange,
  cages,
  boardHospitalizations,
  onNavigateToForm,
  onMovePet,
  canCreate,
  canEdit,
  pagination,
  onPageChange,
}: HospitalizationListContentProps) {
  return (
    <div className="flex flex-col gap-4">
      <UnifiedTabs
        items={HOSPITALIZATION_TAB_ITEMS}
        value={statusFilter}
        onValueChange={onStatusTabChange}
        className="w-full"
      />

      <div className="flex items-center gap-4">
        <div className="flex-1">
          <PropertyFilter
            properties={filterProperties}
            activeFilters={activeFilters}
            onFilterChange={onFilterChange}
            searchTerm={searchTerm}
            onSearchChange={onSearchChange}
            searchPlaceholder="飼主名、ペット名、入院No..."
            count={serverTotal}
            sortProperties={HOSPITALIZATION_SORT_PROPERTIES}
            activeSorts={activeSorts}
            onSortChange={onSortChange}
          />
        </div>
        <div
          className={`${C.bgWhite} rounded-sm border ${C.borderMedium} p-1 h-11 flex items-center`}
        >
          <ToggleGroup
            type="single"
            value={viewMode}
            onValueChange={(v) => v && isValidViewMode(v) && onViewModeChange(v)}
          >
            <ToggleGroupItem
              value="board"
              size="sm"
              aria-label="Board View"
              className="-my-1 h-11 min-w-11"
            >
              <LayoutGrid className={ICON.action} />
            </ToggleGroupItem>
            <ToggleGroupItem
              value="list"
              size="sm"
              aria-label="List View"
              className="-my-1 h-11 min-w-11"
            >
              <List className={ICON.action} />
            </ToggleGroupItem>
          </ToggleGroup>
        </div>
      </div>

      {viewMode === "board" ? (
        <HospitalizationBoard
          cages={cages}
          hospitalizations={boardHospitalizations}
          onNavigateToForm={onNavigateToForm}
          onMovePet={onMovePet}
          canCreate={canCreate}
          canEdit={canEdit}
        />
      ) : (
        <>
          <HospitalizationListView
            hospitalizations={pagination.paginatedData}
            onNavigate={onNavigateToForm}
            canEdit={canEdit}
          />
          {pagination.totalPages > 1 ? (
            <Pagination
              currentPage={pagination.currentPage}
              totalPages={pagination.totalPages}
              totalCount={pagination.totalCount}
              startIndex={pagination.startIndex}
              endIndex={pagination.endIndex}
              onPageChange={onPageChange}
              onPrev={() => onPageChange(pagination.currentPage - 1)}
              onNext={() => onPageChange(pagination.currentPage + 1)}
            />
          ) : null}
        </>
      )}
    </div>
  );
}

interface HospitalizationListPageViewProps extends HospitalizationListContentProps {
  canCreateHeader: boolean;
  onCreate: () => void;
}

export function HospitalizationListPageView({
  canCreateHeader,
  onCreate,
  ...contentProps
}: HospitalizationListPageViewProps) {
  return (
    <PageLayout
      title="入院・ホテル管理"
      resource={ResourceHospitalization}
      headerAction={
        canCreateHeader ? (
          <PrimaryButton onClick={onCreate}>
            <Plus className={`mr-1.5 ${ICON.action}`} />
            新規入院登録
          </PrimaryButton>
        ) : null
      }
      maxWidth={LAYOUT.pageContentMaxWidth.full}
    >
      <HospitalizationListContent {...contentProps} />
    </PageLayout>
  );
}
