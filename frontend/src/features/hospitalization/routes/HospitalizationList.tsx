// React/Framework
import { useState, useCallback, useMemo } from "react";

// External
import { Plus, LayoutGrid, List, Building2, Calendar, PawPrint } from "lucide-react";

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
import { useGetHospitalizations } from "../api/get-hospitalizations";
import { HOSPITALIZATION_FILTER_STATUS, HOSPITALIZATION_STATUS } from "../constants";
import { usePermission } from "@/features/auth/hooks/use-permission";

// Types
import type { HospitalizationFilterStatus } from "../constants";
import type { Hospitalization } from "@/types";
import type { HospitalizationFilters } from "../api/get-hospitalizations";
import type {
  FilterProperty,
  ActiveFilter,
  SortProperty,
  ActiveSort,
} from "@/components/shared/NotionFilter/types";
import { CONDITIONS_NO_EMPTY } from "@/components/shared/NotionFilter/types";

type SortKey = "startDate" | "ownerName" | "petName" | "species" | "status";

const isValidFilterStatus = (v: string): v is HospitalizationFilterStatus =>
    Object.values(HOSPITALIZATION_FILTER_STATUS).includes(v as HospitalizationFilterStatus);

type ViewMode = "list" | "board";
const isValidViewMode = (v: string): v is ViewMode => v === "list" || v === "board";

// rendering-hoist-jsx: 静的フィルタプロパティ（種は動的オプションのためコンポーネント内で構築）
const HOSPITALIZATION_STATIC_FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "startDate",
    label: "入院日",
    type: "date-range",
    icon: Calendar,
  },
  {
    key: "hospitalizationType",
    label: "入院区分",
    type: "select",
    icon: Building2,
    // hospitalizations.hospitalization_type NOT NULL — 空値は存在しない
    conditions: CONDITIONS_NO_EMPTY,
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
    cages,
    movePet,
    handleNavigateToForm
  } = useHospitalizationList();
  const { canCreate } = usePermission("hospitalization");

  const [activeSorts, setActiveSorts] = useState<ActiveSort[]>([]);
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);

  // rerender-dependencies: activeFilters から日付フィルタを抽出してサーバーサイドフィルタに渡す
  const dateFilters = useMemo<HospitalizationFilters>(() => {
    const dateFilter = activeFilters.find((f) => f.key === "startDate")?.value as
      | { from?: string; to?: string }
      | undefined;
    return {
      startDate: dateFilter?.from,
      endDate: dateFilter?.to,
    };
  }, [activeFilters]);

  // サーバーサイド日付フィルタ適用済みデータを取得
  const { data: allHospitalizations = [] } = useGetHospitalizations(dateFilters);

  // js-cache-function-results: ロード済みデータから種の選択肢を動的生成
  const filterProperties = useMemo<FilterProperty[]>(() => {
    const speciesOptions = Array.from(new Set(allHospitalizations.map((h) => h.species).filter(Boolean)))
      .sort()
      .map((s) => ({ value: s, label: s }));
    return [
      ...HOSPITALIZATION_STATIC_FILTER_PROPERTIES,
      // pets.animal_species_id NOT NULL — 空値は存在しない
      { key: "species", label: "種", type: "select" as const, icon: PawPrint, conditions: CONDITIONS_NO_EMPTY, options: speciesOptions },
    ];
  }, [allHospitalizations]);

  const handleSortChange = useCallback((sorts: ActiveSort[]) => {
    setActiveSorts(sorts);
  }, []);

  // 入院区分・ステータス・テキスト検索・種フィルタ（クライアントサイド、日付フィルタはサーバーサイドに移行済み）
  const typeFilteredHospitalizations = useMemo(() => {
    let result = allHospitalizations;

    // Status フィルタ（タブ）
    if (statusFilter !== "all") {
      result = result.filter((h) => {
        if (statusFilter === "active") return h.status === HOSPITALIZATION_STATUS.ACTIVE || h.status === HOSPITALIZATION_STATUS.TEMP_DISCHARGE;
        if (statusFilter === "discharged") return h.status === HOSPITALIZATION_STATUS.DISCHARGED;
        if (statusFilter === "reserved") return h.status === HOSPITALIZATION_STATUS.RESERVED;
        return true;
      });
    }

    // テキスト検索
    if (searchTerm) {
      const lowerTerm = searchTerm.toLowerCase();
      result = result.filter(
        (h) =>
          h.ownerName.toLowerCase().includes(lowerTerm) ||
          h.petName.toLowerCase().includes(lowerTerm) ||
          h.hospitalizationNo.toLowerCase().includes(lowerTerm),
      );
    }

    // 入院区分フィルタ
    const typeFilter = activeFilters.find((f) => f.key === "hospitalizationType");
    if (typeFilter && typeof typeFilter.value === "string") {
      result = result.filter((h) => {
        switch (typeFilter.condition) {
          case "is":           return h.hospitalizationType === typeFilter.value;
          case "is_not":       return h.hospitalizationType !== typeFilter.value;
          default:             return h.hospitalizationType === typeFilter.value;
        }
      });
    }

    // species フィルタ（クライアントサイド）
    const speciesFilter = activeFilters.find((f) => f.key === "species");
    if (speciesFilter && typeof speciesFilter.value === "string") {
      result = result.filter((h) => {
        switch (speciesFilter.condition) {
          case "is":           return h.species === speciesFilter.value;
          case "is_not":       return h.species !== speciesFilter.value;
          case "is_empty":     return !h.species;
          case "is_not_empty": return !!h.species;
          default:             return h.species === speciesFilter.value;
        }
      });
    }

    return result;
  }, [allHospitalizations, statusFilter, searchTerm, activeFilters]);

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
        canCreate ? (
          <PrimaryButton onClick={() => handleNavigateToForm()}>
            <Plus className="mr-1.5 size-4" />
            新規入院登録
          </PrimaryButton>
        ) : null
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
                  properties={filterProperties}
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
