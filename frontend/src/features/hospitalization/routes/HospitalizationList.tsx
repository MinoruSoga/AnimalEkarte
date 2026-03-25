// React/Framework
import { useState, useCallback, useMemo } from "react";

// External
import { Plus, LayoutGrid, List, Building2 } from "lucide-react";

// Internal
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { usePagination } from "@/hooks/use-pagination";
import { Pagination } from "@/components/shared/Pagination/Pagination";

// Relative
import { HospitalizationBoard } from "../components/HospitalizationBoard";
import { HospitalizationListView } from "../components/HospitalizationListView";
import { useHospitalizationList } from "../hooks/use-hospitalization-list";
import { HOSPITALIZATION_FILTER_STATUS } from "../constants";

// Types
import type { HospitalizationFilterStatus } from "../constants";
import type { Hospitalization } from "@/types";
import type {
  FilterProperty,
  ActiveFilter,
  SortProperty,
  ActiveSort,
} from "@/components/shared/NotionFilter/types";

type SortKey = "startDate" | "ownerName" | "petName" | "species" | "status";

const isValidFilterStatus = (v: string): v is HospitalizationFilterStatus =>
    Object.values(HOSPITALIZATION_FILTER_STATUS).includes(v as HospitalizationFilterStatus);

type ViewMode = "list" | "board";
const isValidViewMode = (v: string): v is ViewMode => v === "list" || v === "board";

// rendering-hoist-jsx: 静的フィルタプロパティ定義
const HOSPITALIZATION_FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "hospitalizationType",
    label: "入院区分",
    type: "select",
    icon: Building2,
    options: [
      { value: "入院", label: "入院" },
      { value: "ホテル", label: "ホテル" },
    ],
  },
];

// rendering-hoist-jsx: 静的ソートプロパティ定義
const HOSPITALIZATION_SORT_PROPERTIES: SortProperty[] = [
  { key: "startDate", label: "入院日" },
  { key: "ownerName", label: "飼主名" },
  { key: "petName", label: "ペット名" },
  { key: "species", label: "種" },
  { key: "status", label: "ステータス" },
];

export function HospitalizationList() {
  const {
    searchTerm, setSearchTerm,
    statusFilter, setStatusFilter,
    viewMode, setViewMode,
    filteredHospitalizations,
    cages,
    movePet,
    handleNavigateToForm
  } = useHospitalizationList();

  const [activeSorts, setActiveSorts] = useState<ActiveSort[]>([]);
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);

  const handleSortChange = useCallback((sorts: ActiveSort[]) => {
    setActiveSorts(sorts);
  }, []);

  // 入院区分フィルタ（クライアントサイド）
  const typeFilteredHospitalizations = useMemo(() => {
    const typeFilter = activeFilters.find((f) => f.key === "hospitalizationType");
    if (!typeFilter || typeof typeFilter.value !== "string") return filteredHospitalizations;
    return filteredHospitalizations.filter((h) => {
      switch (typeFilter.condition) {
        case "is":           return h.hospitalizationType === typeFilter.value;
        case "is_not":       return h.hospitalizationType !== typeFilter.value;
        case "is_empty":     return !h.hospitalizationType;
        case "is_not_empty": return !!h.hospitalizationType;
        default:             return h.hospitalizationType === typeFilter.value;
      }
    });
  }, [filteredHospitalizations, activeFilters]);

  // Sort data for list view
  const sortedHospitalizations = useMemo(() => {
    if (activeSorts.length === 0) return [...typeFilteredHospitalizations];
    const sorted = [...typeFilteredHospitalizations];
    sorted.sort((a: Hospitalization, b: Hospitalization) => {
      for (const sort of activeSorts) {
        const key = sort.key as SortKey;
        const aVal = String(a[key] ?? "");
        const bVal = String(b[key] ?? "");
        const cmp = aVal.localeCompare(bVal, "ja");
        if (cmp !== 0) return sort.direction === "asc" ? cmp : -cmp;
      }
      return 0;
    });
    return sorted;
  }, [typeFilteredHospitalizations, activeSorts]);

  const pagination = usePagination(sortedHospitalizations, {
    pageSize: 20,
    resetKey: `${searchTerm}:${statusFilter}`,
  });

  return (
    <PageLayout
      title="入院・ホテル管理"
      headerAction={
        <PrimaryButton onClick={() => handleNavigateToForm()}>
          <Plus className="mr-1.5 size-4" />
          新規入院登録
        </PrimaryButton>
      }
      maxWidth="max-w-full"
    >
      <div className="flex flex-col gap-4">
        {/* Status Tabs */}
        <Tabs value={statusFilter} onValueChange={(v) => isValidFilterStatus(v) && setStatusFilter(v)} className="w-full">
            <TabsList className="grid w-[400px] grid-cols-4 h-11 p-[3px] rounded-xl">
                <TabsTrigger value={HOSPITALIZATION_FILTER_STATUS.ACTIVE} className="rounded-[10px]">入院中</TabsTrigger>
                <TabsTrigger value={HOSPITALIZATION_FILTER_STATUS.RESERVED} className="rounded-[10px]">予約</TabsTrigger>
                <TabsTrigger value={HOSPITALIZATION_FILTER_STATUS.DISCHARGED} className="rounded-[10px]">退院済</TabsTrigger>
                <TabsTrigger value={HOSPITALIZATION_FILTER_STATUS.ALL} className="rounded-[10px]">すべて</TabsTrigger>
            </TabsList>
        </Tabs>

        {/* Search & View Toggle */}
        <div className="flex items-center gap-4">
            <div className="flex-1">
                <NotionFilter
                  properties={HOSPITALIZATION_FILTER_PROPERTIES}
                  activeFilters={activeFilters}
                  onFilterChange={setActiveFilters}
                  searchTerm={searchTerm}
                  onSearchChange={setSearchTerm}
                  searchPlaceholder="飼主名、ペット名、入院No..."
                  count={typeFilteredHospitalizations.length}
                  sortProperties={HOSPITALIZATION_SORT_PROPERTIES}
                  activeSorts={activeSorts}
                  onSortChange={handleSortChange}
                />
            </div>
            <div className="bg-white rounded-[6px] border border-[rgba(55,53,47,0.16)] p-1 h-11 flex items-center">
                <ToggleGroup type="single" value={viewMode} onValueChange={(v) => v && isValidViewMode(v) && setViewMode(v)}>
                    <ToggleGroupItem value="board" size="sm" aria-label="Board View">
                        <LayoutGrid className="h-4 w-4" />
                    </ToggleGroupItem>
                    <ToggleGroupItem value="list" size="sm" aria-label="List View">
                        <List className="h-4 w-4" />
                    </ToggleGroupItem>
                </ToggleGroup>
            </div>
        </div>

        {/* Content */}
        {viewMode === "board" ? (
            <HospitalizationBoard
                cages={cages}
                hospitalizations={typeFilteredHospitalizations}
                onNavigateToForm={handleNavigateToForm}
                onMovePet={movePet}
            />
        ) : (
            <>
              <HospitalizationListView
                  hospitalizations={pagination.paginatedData}
                  onNavigate={handleNavigateToForm}
              />
              {pagination.totalPages > 1 ? (
                <Pagination
                  currentPage={pagination.currentPage}
                  totalPages={pagination.totalPages}
                  totalCount={pagination.totalCount}
                  startIndex={pagination.startIndex}
                  endIndex={pagination.endIndex}
                  onPageChange={pagination.goToPage}
                  onPrev={pagination.prevPage}
                  onNext={pagination.nextPage}
                />
              ) : null}
            </>
        )}
      </div>
    </PageLayout>
  );
}
