// React/Framework
import { C, ICON } from "@/lib/design-tokens";
import { useState, useDeferredValue, useCallback, useEffect, useMemo } from "react";
import { useNavigate, useSearchParams } from "react-router";

// Hooks
import { useSortableData } from "@/hooks/use-sortable-data";

// External
import { Plus, TestTube, FileSpreadsheet, Calendar, CircleDot, FlaskConical, User } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { StatusBadge } from "@/components/shared/StatusBadge/StatusBadge";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { SortableHeader } from "@/components/shared/SortableHeader/SortableHeader";
import { getExaminationStatusColor } from "@/utils/status-helpers";
import { usePagination } from "@/hooks/use-pagination";
import { Pagination } from "@/components/shared/Pagination/Pagination";
import { FilteringIndicator } from "@/components/shared/FilteringIndicator/FilteringIndicator";

// Relative
import { useFilterExaminationRecords } from "../hooks/use-examination-records";
import { paths } from "@/config/paths";
import { usePermission } from "@/features/auth";

// Types
import type {
  FilterProperty,
  ActiveFilter,
  SortProperty,
} from "@/components/shared/NotionFilter/types";
import { CONDITIONS_NO_EMPTY, CONDITIONS_WITH_EMPTY } from "@/components/shared/NotionFilter/types";
import type { ExaminationRecord } from "@/types";

// rendering-hoist-jsx: 静的フィルタプロパティ（検査種別は動的オプションのためコンポーネント内で構築）
const STATIC_FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "date",
    label: "日付",
    type: "date-range",
    icon: Calendar,
  },
  {
    key: "status",
    label: "ステータス",
    type: "select",
    icon: CircleDot,
    // exams.status DEFAULT 'pending' — 空値は存在しない
    // transform で EN→JA 変換済のためフィルタ値も日本語で指定
    conditions: CONDITIONS_NO_EMPTY,
    options: [
      { value: "依頼中", label: "依頼中" },
      { value: "検査中", label: "検査中" },
      { value: "完了", label: "完了" },
    ],
  },
];

// rendering-hoist-jsx: 静的ソートプロパティ定義
const EXAMINATION_SORT_PROPERTIES: SortProperty[] = [
  { key: "date", label: "日時" },
  { key: "ownerName", label: "飼主名" },
  { key: "petName", label: "ペット名" },
  { key: "testType", label: "検査種別" },
  { key: "doctor", label: "担当医" },
  { key: "status", label: "ステータス" },
];

export function ExaminationsList() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { canCreate, canEdit } = usePermission("examinations");
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const deferredSearch = useDeferredValue(searchTerm);
  const isFiltering = searchTerm !== deferredSearch;

  // activeFilters から日付フィルタを抽出
  const filters = useMemo(() => {
    const dateFilter = activeFilters.find((f) => f.key === "date")?.value as
      | { from?: string; to?: string }
      | undefined;
    return {
      startDate: dateFilter?.from,
      endDate: dateFilter?.to,
    };
  }, [activeFilters]);

  const { data: filteredRecords, allExaminations, isLoading } = useFilterExaminationRecords(deferredSearch, filters, activeFilters);

  // js-cache-function-results: ロード済みデータから検査種別・担当医の選択肢を動的生成
  const filterProperties = useMemo<FilterProperty[]>(() => {
    const testTypeOptions = Array.from(new Set(allExaminations.map((r) => r.testType).filter(Boolean)))
      .sort()
      .map((t) => ({ value: t, label: t }));
    const doctorOptions = Array.from(new Set(allExaminations.map((r) => r.doctor).filter(Boolean)))
      .sort()
      .map((d) => ({ value: d, label: d }));
    return [
      ...STATIC_FILTER_PROPERTIES,
      // testType は nullable（未設定の検査依頼あり）
      { key: "testType", label: "検査種別", type: "select" as const, icon: FlaskConical, conditions: CONDITIONS_WITH_EMPTY, options: testTypeOptions },
      // exams.doctor_id nullable（未割当あり）
      { key: "doctor", label: "担当医", type: "select" as const, icon: User, conditions: CONDITIONS_WITH_EMPTY, options: doctorOptions },
    ];
  }, [allExaminations]);

  const { activeSorts, setActiveSorts, toggleSort, directionFor, sortedData } =
    useSortableData(filteredRecords);

  const pagination = usePagination(sortedData, {
    pageSize: 20,
    resetKey: deferredSearch,
  });

  // FE-144: URLクエリパラメータからページ番号を読み取る
  const urlPage = Number(searchParams.get("page") ?? 1);

  // FE-144: URLのページ番号とローカル状態を同期（URLが変わったときのみ）
  useEffect(() => {
    const clampedPage = Math.max(1, Math.min(urlPage, pagination.totalPages));
    if (clampedPage !== pagination.currentPage) {
      pagination.goToPage(clampedPage);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [urlPage, pagination.totalPages]);

  // FE-144: ページ変更時にURLクエリパラメータを更新
  const handlePageChange = useCallback((page: number) => {
    pagination.goToPage(page);
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev);
      if (page === 1) {
        next.delete("page");
      } else {
        next.set("page", String(page));
      }
      return next;
    }, { replace: true });
  }, [pagination, setSearchParams]);

  const handleCreate = useCallback(() => {
    navigate(paths.examinations.selectPet.getHref());
  }, [navigate]);

  const handleEdit = useCallback((id: string) => {
    navigate(`/examinations/${id}`);
  }, [navigate]);

  // rerender-memo: renderRow を useCallback でメモ化（DataTable への参照を安定化）
  const renderRow = useCallback((r: ExaminationRecord) => (
    <DataTableRow key={r.id} onClick={() => handleEdit(r.id)}>
      <TableCell className={`font-mono text-base ${C.text} py-2.5`}>{r.date}</TableCell>
      <TableCell className={`text-base ${C.text} py-2.5`}>{r.ownerName}</TableCell>
      <TableCell className={`text-base ${C.text} py-2.5`}>{r.petName}</TableCell>
      <TableCell className={`text-base font-medium ${C.text} py-2.5`}>{r.testType}</TableCell>
      <TableCell className="text-base text-muted-foreground truncate max-w-[200px] py-2.5 hidden lg:table-cell">
        {r.resultSummary || "-"}
      </TableCell>
      <TableCell className={`text-base ${C.text} py-2.5`}>{r.doctor}</TableCell>
      <TableCell className="py-2.5">
        <StatusBadge colorClass={getExaminationStatusColor(r.status)}>
          {r.status}
        </StatusBadge>
      </TableCell>
      <TableCell className="text-right py-2.5">
        {canEdit ? <RowActionButton onClick={() => handleEdit(r.id)} /> : null}
      </TableCell>
    </DataTableRow>
  ), [handleEdit, canEdit]);

  const columns = useMemo(() => [
    {
      header: (
        <SortableHeader
          label="日時"
          direction={directionFor("date")}
          onToggle={() => toggleSort("date")}
        />
      ),
      className: "w-[120px]",
    },
    {
      header: (
        <SortableHeader
          label="飼主名"
          direction={directionFor("ownerName")}
          onToggle={() => toggleSort("ownerName")}
        />
      ),
    },
    {
      header: (
        <SortableHeader
          label="ペット名"
          direction={directionFor("petName")}
          onToggle={() => toggleSort("petName")}
        />
      ),
    },
    {
      header: (
        <SortableHeader
          label="検査種別"
          direction={directionFor("testType")}
          onToggle={() => toggleSort("testType")}
        />
      ),
    },
    { header: "結果概要", className: "hidden lg:table-cell" },
    {
      header: (
        <SortableHeader
          label="担当医"
          direction={directionFor("doctor")}
          onToggle={() => toggleSort("doctor")}
        />
      ),
      className: "w-[100px]",
    },
    {
      header: (
        <SortableHeader
          label="ステータス"
          direction={directionFor("status")}
          onToggle={() => toggleSort("status")}
        />
      ),
      className: "w-[100px]",
    },
    { header: "操作", className: "w-[80px]", align: "right" as const },
  ], [directionFor, toggleSort]);

  return (
    <PageLayout
      title="検査管理"
      icon={<TestTube className={`${ICON.page} ${C.text}`} />}
      headerAction={
        <div className="flex items-center gap-2">
          <Button variant="outline" className="h-10 text-sm gap-2 bg-white" onClick={() => {}}>
            <FileSpreadsheet className={ICON.action} />
            検査データ取込
          </Button>
          {canCreate ? (
            <PrimaryButton onClick={handleCreate}>
              <Plus className={`mr-1.5 ${ICON.action}`} />
              新規検査登録
            </PrimaryButton>
          ) : null}
        </div>
      }
      maxWidth="max-w-full"
    >
      <div className="flex flex-col gap-4">
        <NotionFilter
          properties={filterProperties}
          activeFilters={activeFilters}
          onFilterChange={setActiveFilters}
          searchTerm={searchTerm}
          onSearchChange={setSearchTerm}
          searchPlaceholder="飼主名、ペット名、検査種別..."
          count={isLoading ? undefined : filteredRecords.length}
          sortProperties={EXAMINATION_SORT_PROPERTIES}
          activeSorts={activeSorts}
          onSortChange={setActiveSorts}
        />

        <FilteringIndicator isFiltering={isFiltering}>
          <DataTable
            columns={columns}
            data={pagination.paginatedData}
            emptyMessage="検査データが見つかりません"
            renderRow={renderRow}
          />
        </FilteringIndicator>

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
      </div>
    </PageLayout>
  );
}
