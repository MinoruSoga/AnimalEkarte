// React/Framework
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { useState, useCallback, useMemo, useDeferredValue, useEffect } from "react";
import { useSearchParams } from "react-router";
import { normalizeKana } from "@/lib/normalize-kana";

// External
import { Plus, LayoutGrid, List, Building2, Calendar, PawPrint } from "lucide-react";

// Internal
import { uniqueSortedOptions } from "@/lib/unique-sorted-options";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { UnifiedTabs } from "@/components/shared/UnifiedTabs";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PropertyFilter } from "@/components/shared/PropertyFilter/PropertyFilter";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { Pagination } from "@/components/shared/Pagination/Pagination";

// Relative
import { HospitalizationBoard } from "../components/HospitalizationBoard";
import { HospitalizationListView } from "../components/HospitalizationListView";
import { useHospitalizationList } from "../hooks/use-hospitalization-list";
import { useGetHospitalizations } from "../api/get-hospitalizations";
import {
  HOSPITALIZATION_FILTER_STATUS,
  HOSPITALIZATION_LIST_DEFAULT_LIMIT,
  HOSPITALIZATION_LIST_DEFAULT_PAGE,
} from "../constants";
import { usePermission } from "@/hooks/use-permission";

// Types
import type { HospitalizationFilterStatus } from "../constants";
import type { Hospitalization } from "@/types";
import type { HospitalizationFilters } from "../api/get-hospitalizations";
import type {
  FilterProperty,
  ActiveFilter,
  SortProperty,
  ActiveSort,
} from "@/components/shared/PropertyFilter/types";
import { CONDITIONS_NO_EMPTY } from "@/components/shared/PropertyFilter/types";
import { ResourceHospitalization } from "@/types/generated/models";

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
  const [searchParams, setSearchParams] = useSearchParams();
  const { canCreate, canEdit } = usePermission("hospitalization");
  const {
    searchTerm, setSearchTerm,
    statusFilter, setStatusFilter,
    viewMode, setViewMode,
    cages,
    movePet,
    handleNavigateToForm
  } = useHospitalizationList(canEdit);

  const tabItems = [
    { value: HOSPITALIZATION_FILTER_STATUS.ACTIVE, label: "入院中" },
    { value: HOSPITALIZATION_FILTER_STATUS.RESERVED, label: "予約" },
    { value: HOSPITALIZATION_FILTER_STATUS.DISCHARGED, label: "退院済" },
    { value: HOSPITALIZATION_FILTER_STATUS.ALL, label: "すべて" },
  ];

  const [activeSorts, setActiveSorts] = useState<ActiveSort[]>([]);
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);

  // rerender-transitions: 検索ワードを useDeferredValue で遅延させ、入力中のリスト再計算コストを抑制
  const deferredSearchTerm = useDeferredValue(searchTerm);

  // FE-144: URLクエリパラメータからページ番号を読み取る（サーバ page の正本）
  const urlPage = Number(searchParams.get("page") ?? HOSPITALIZATION_LIST_DEFAULT_PAGE);
  const serverPage = Number.isFinite(urlPage) && urlPage >= 1 ? urlPage : HOSPITALIZATION_LIST_DEFAULT_PAGE;

  // rerender-dependencies: 日付 + タブ status + page/limit をサーバ query の正本へ寄せる（BUG-009）
  const listFilters = useMemo<HospitalizationFilters>(() => {
    const dateFilter = activeFilters.find((f) => f.key === "startDate")?.value as
      | { from?: string; to?: string }
      | undefined;
    return {
      startDate: dateFilter?.from,
      endDate: dateFilter?.to,
      statusFilter,
      page: serverPage,
      limit: HOSPITALIZATION_LIST_DEFAULT_LIMIT,
    };
  }, [activeFilters, statusFilter, serverPage]);

  // サーバ status/date/page 適用済みデータを取得（total を捨てない）
  const {
    data: hospitalizationsPage,
    isLoading: hospitalizationsLoading,
    isError: hospitalizationsError,
  } = useGetHospitalizations(listFilters);
  const allHospitalizations = useMemo(
    () => hospitalizationsPage?.data ?? [],
    [hospitalizationsPage?.data],
  );
  // 件数表示の正本は server total（page-window で欠落しない）
  const serverTotal = hospitalizationsPage?.total ?? 0;

  // js-cache-function-results: ロード済みデータから種の選択肢を動的生成
  const filterProperties = useMemo<FilterProperty[]>(() => {
    const speciesOptions = uniqueSortedOptions(allHospitalizations, (h) => h.species);
    return [
      ...HOSPITALIZATION_STATIC_FILTER_PROPERTIES,
      // pets.animal_species_id NOT NULL — 空値は存在しない
      { key: "species", label: "種", type: "select" as const, icon: PawPrint, conditions: CONDITIONS_NO_EMPTY, options: speciesOptions },
    ];
  }, [allHospitalizations]);

  const handleSortChange = useCallback((sorts: ActiveSort[]) => {
    setActiveSorts(sorts);
  }, []);

  // 入院区分・テキスト検索・種フィルタのみクライアント（status は server 正本 — 二重 filter 禁止）
  // board / list は同一 typeFilteredHospitalizations を参照する（実装分岐なし）
  const typeFilteredHospitalizations = useMemo(() => {
    let result = allHospitalizations;

    // テキスト検索（deferredSearchTerm で遅延評価）
    if (deferredSearchTerm) {
      const normalizedTerm = normalizeKana(deferredSearchTerm).toLowerCase();
      result = result.filter(
        (h) =>
          normalizeKana(h.ownerName).toLowerCase().includes(normalizedTerm) ||
          normalizeKana(h.petName).toLowerCase().includes(normalizedTerm) ||
          h.hospitalizationNo.toLowerCase().includes(normalizedTerm),
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
  }, [allHospitalizations, deferredSearchTerm, activeFilters]);

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

  // BUG-009 review: usePagination クライアントスライスを撤去し server total/page/limit を正本にする
  // （ExaminationsList BUG-411 同型）。list 行はサーバが既に返す page 分を client filter/sort した集合。
  const serverLimit = HOSPITALIZATION_LIST_DEFAULT_LIMIT;
  const totalPages = Math.max(1, Math.ceil(serverTotal / serverLimit));
  const pagination = {
    paginatedData: sortedHospitalizations,
    totalPages,
    totalCount: serverTotal,
    startIndex: serverTotal === 0 ? 0 : (serverPage - 1) * serverLimit + 1,
    endIndex: Math.min(serverPage * serverLimit, serverTotal),
    currentPage: serverPage,
  };

  // FE-144: URL page が totalPages を超えたら clamp（母集団縮小で空 page に迷わない）
  useEffect(() => {
    if (hospitalizationsLoading) return;
    const clampedPage = Math.max(1, Math.min(urlPage, totalPages));
    if (clampedPage !== urlPage) {
      setSearchParams((prev) => {
        const next = new URLSearchParams(prev);
        if (clampedPage === 1) {
          next.delete("page");
        } else {
          next.set("page", String(clampedPage));
        }
        return next;
      }, { replace: true });
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps -- setSearchParams は安定。urlPage/totalPages/loading 変化時のみ
  }, [urlPage, totalPages, hospitalizationsLoading]);

  const resetListPage = useCallback(() => {
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev);
      next.delete("page");
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  // FE-144: ページ変更時に URL を更新し、server re-fetch を誘発する
  const handlePageChange = useCallback((page: number) => {
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev);
      if (page === 1) {
        next.delete("page");
      } else {
        next.set("page", String(page));
      }
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  const handleStatusTabChange = useCallback(
    (v: string) => {
      if (!isValidFilterStatus(v)) return;
      setStatusFilter(v);
      resetListPage();
    },
    [setStatusFilter, resetListPage],
  );

  const handleSearchChange = useCallback(
    (value: string) => {
      setSearchTerm(value);
      resetListPage();
    },
    [setSearchTerm, resetListPage],
  );

  const handleFilterChange = useCallback(
    (next: ActiveFilter[]) => {
      setActiveFilters(next);
      resetListPage();
    },
    [resetListPage],
  );

  if (hospitalizationsLoading) return <LoadingFallback />;
  if (hospitalizationsError) return <ErrorFallback />;

  return (
    <PageLayout
      title="入院・ホテル管理"
      resource={ResourceHospitalization}
      headerAction={
        canCreate ? (
          <PrimaryButton onClick={() => handleNavigateToForm()}>
            <Plus className={`mr-1.5 ${ICON.action}`} />
            新規入院登録
          </PrimaryButton>
        ) : null
      }
      maxWidth={LAYOUT.pageContentMaxWidth.full}
    >
      <div className="flex flex-col gap-4">
        {/* Status Tabs */}
        <UnifiedTabs
          items={tabItems}
          value={statusFilter}
          onValueChange={handleStatusTabChange}
          className="w-full"
        />

        {/* Search & View Toggle */}
        <div className="flex items-center gap-4">
            <div className="flex-1">
                <PropertyFilter
                  properties={filterProperties}
                  activeFilters={activeFilters}
                  onFilterChange={handleFilterChange}
                  searchTerm={searchTerm}
                  onSearchChange={handleSearchChange}
                  searchPlaceholder="飼主名、ペット名、入院No..."
                  count={serverTotal}
                  sortProperties={HOSPITALIZATION_SORT_PROPERTIES}
                  activeSorts={activeSorts}
                  onSortChange={handleSortChange}
                />
            </div>
            <div className={`${C.bgWhite} rounded-sm border ${C.borderMedium} p-1 h-11 flex items-center`}>
                <ToggleGroup type="single" value={viewMode} onValueChange={(v) => v && isValidViewMode(v) && setViewMode(v)}>
                    <ToggleGroupItem value="board" size="sm" aria-label="Board View" className="-my-1 h-11 min-w-11">
                        <LayoutGrid className={ICON.action} />
                    </ToggleGroupItem>
                    <ToggleGroupItem value="list" size="sm" aria-label="List View" className="-my-1 h-11 min-w-11">
                        <List className={ICON.action} />
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
                canCreate={canCreate}
                canEdit={canEdit}
            />
        ) : (
            <>
              {/* board と同じ typeFiltered 集合。list は server page 分をそのまま表示 */}
              <HospitalizationListView
                  hospitalizations={pagination.paginatedData}
                  onNavigate={handleNavigateToForm}
                  canEdit={canEdit}
              />
              {pagination.totalPages > 1 ? (
                <Pagination
                  currentPage={pagination.currentPage}
                  totalPages={pagination.totalPages}
                  totalCount={pagination.totalCount}
                  startIndex={pagination.startIndex}
                  endIndex={pagination.endIndex}
                  onPageChange={handlePageChange}
                  onPrev={() => handlePageChange(pagination.currentPage - 1)}
                  onNext={() => handlePageChange(pagination.currentPage + 1)}
                />
              ) : null}
            </>
        )}
      </div>
    </PageLayout>
  );
}
