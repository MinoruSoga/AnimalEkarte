// React/Framework
import { C, ICON, STYLE, LAYOUT } from "@/lib/design-tokens";
import { useState, useDeferredValue, useCallback, useEffect, useMemo } from "react";
import { useNavigate, useSearchParams } from "react-router";

// Hooks
import { useSortableData } from "@/hooks/use-sortable-data";
import { uniqueSortedOptions } from "@/lib/unique-sorted-options";

// External
import { Plus, TestTube, Calendar, CircleDot, FlaskConical, User, Info } from "lucide-react";

// Internal
import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { PropertyFilter } from "@/components/shared/PropertyFilter/PropertyFilter";
import { DataTable, DESIGN_TABLE_HEADER_ROW, DESIGN_TABLE_HEADER_CELL } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { DataTableRowLink } from "@/components/shared/DataTable/DataTableRowLink";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { StatusBadge } from "@/components/shared/StatusBadge/StatusBadge";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { SortableHeader } from "@/components/shared/SortableHeader/SortableHeader";
import { getExaminationStatusColor } from "@/lib/status-helpers";
import { Pagination } from "@/components/shared/Pagination/Pagination";
import { FilteringIndicator } from "@/components/shared/FilteringIndicator/FilteringIndicator";

// Relative
import { useFilterExaminationRecords } from "../hooks/use-examination-records";
import { paths } from "@/config/paths";
import { usePermission } from "@/hooks/use-permission";

// Types
import type {
  FilterProperty,
  ActiveFilter,
  SortProperty,
} from "@/components/shared/PropertyFilter/types";
import { CONDITIONS_NO_EMPTY, CONDITIONS_WITH_EMPTY } from "@/components/shared/PropertyFilter/types";
import type { ExaminationRecord } from "../api/transforms";
import { ResourceExaminations } from "@/types/generated/models";

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
    // SD-12: backend ExaminationStatus 全5値（pending/in_progress/result_entered/completed/confirmed）に追従
    conditions: CONDITIONS_NO_EMPTY,
    options: [
      { value: "依頼中", label: "依頼中" },
      { value: "検査中", label: "検査中" },
      { value: "結果入力済み", label: "結果入力済み" },
      { value: "完了", label: "完了" },
      { value: "確定", label: "確定" },
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

// BUG-411: サーバサイドページネーションのページサイズ。backend parsePagination の既定 limit(20) を踏襲。
const EXAMINATIONS_PAGE_SIZE = 20;

// BUG-411 react-reviewer指摘(HIGH): backend が受け付けないため client-only のまま残るフィルタキー。
// これらが有効な間は「表示件数」と Pagination の「総件数」が別ソース（前者=現在ページの絞込結果、
// 後者=サーバ生total）になり乖離しうる。ユーザーへの明示的な告知で対応する（案B）。
const CLIENT_ONLY_FILTER_KEYS = ["status", "testType", "doctor"];

export function ExaminationsList() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { canCreate, canEdit } = usePermission("examinations");
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const deferredSearch = useDeferredValue(searchTerm);
  const isFiltering = searchTerm !== deferredSearch;

  // FE-144: URLクエリパラメータからページ番号を読み取る（BUG-411: サーバフェッチのキーそのものになる）
  const urlPage = Math.max(1, Number(searchParams.get("page") ?? 1) || 1);

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

  const { data: filteredRecords, allExaminations, isLoading, error, total, page: serverPage, limit: serverLimit } =
    useFilterExaminationRecords(deferredSearch, { page: urlPage, limit: EXAMINATIONS_PAGE_SIZE }, filters, activeFilters);

  // js-cache-function-results: ロード済みデータから検査種別・担当医の選択肢を動的生成
  const filterProperties = useMemo<FilterProperty[]>(() => {
    const testTypeOptions = uniqueSortedOptions(allExaminations, (r) => r.testType);
    const doctorOptions = uniqueSortedOptions(allExaminations, (r) => r.doctor);
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

  // BUG-411 react-reviewer指摘(HIGH): 検索語 or client-only フィルタ（ステータス・検査種別・担当医）が
  // 有効な間は、表示行数（現在ページ内で絞り込み済み）と Pagination の総件数（サーバ生total）が
  // 一致しない。ユーザーに明示する。
  const hasPageScopedFilter =
    deferredSearch !== "" || activeFilters.some((f) => CLIENT_ONLY_FILTER_KEYS.includes(f.key));

  // BUG-411: usePagination() のクライアントスライスを撤去し、サーバの total/page/limit をそのまま使う。
  // paginatedData はクライアントフィルタ・ソート済みの「現在ページの行」（サーバは既にページ分だけ返す）。
  const totalPages = Math.max(1, Math.ceil(total / serverLimit));
  const pagination = {
    paginatedData: sortedData,
    totalPages,
    totalCount: total,
    startIndex: total === 0 ? 0 : (serverPage - 1) * serverLimit + 1,
    endIndex: Math.min(serverPage * serverLimit, total),
    currentPage: serverPage,
  };

  // FE-144 / BUG-411: URLの page がサーバ total から導いた totalPages を超えている場合はクランプする
  // （日付フィルタ変更等で母集団が縮んだ場合に空ページへ迷い込むのを防ぐ。既存クランプ effect を踏襲）。
  useEffect(() => {
    if (isLoading) return;
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
  // eslint-disable-next-line react-hooks/exhaustive-deps -- setSearchParams は安定参照。urlPage/totalPages/isLoading の変化時のみ再評価する設計（FE-144踏襲）。
  }, [urlPage, totalPages, isLoading]);

  // FE-144: ページ変更時にURLクエリパラメータを更新
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

  // BUG-411 react-reviewer指摘: 旧 usePagination() の resetKey（検索・フィルタ変更時に page を1へ戻す）
  // を撤去したままだと、ページ3閲覧中にフィルタ変更→表示件数とPaginationのtotalが食い違ったまま
  // page=3に留まる退行になる。検索・フィルタ変更時は常に page を1へ戻す（旧挙動を維持）。
  const handleSearchChange = useCallback((value: string) => {
    setSearchTerm(value);
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev);
      next.delete("page");
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  const handleFilterChange = useCallback((next: ActiveFilter[]) => {
    setActiveFilters(next);
    setSearchParams((prev) => {
      const params = new URLSearchParams(prev);
      params.delete("page");
      return params;
    }, { replace: true });
  }, [setSearchParams]);

  const handleCreate = useCallback(() => {
    navigate(paths.examinations.selectPet.getHref());
  }, [navigate]);

  const handleEdit = useCallback((id: string) => {
    navigate(paths.examinations.detail.getHref(id));
  }, [navigate]);

  // rerender-memo: renderRow を useCallback でメモ化（DataTable への参照を安定化）
  const renderRow = useCallback((r: ExaminationRecord) => (
    <DataTableRow key={r.id}>
      <TableCell className={STYLE.tableCellMono}>{r.date}</TableCell>
      <TableCell className={STYLE.tableCell}>{r.ownerName}</TableCell>
      <TableCell className={STYLE.tableCell}>
        <DataTableRowLink
          to={paths.examinations.detail.getHref(r.id)}
          aria-label={`検査詳細: ${r.petName} ${r.date} ID ${r.id}`}
        >
          {r.petName}
        </DataTableRowLink>
      </TableCell>
      <TableCell className={`${STYLE.tableCell} font-medium hidden sm:table-cell`}>{r.testType}</TableCell>
      <TableCell className={`${C.text60} truncate max-w-[200px] hidden lg:table-cell`}>
        {r.resultSummary || "-"}
      </TableCell>
      <TableCell className={`${STYLE.tableCell} hidden md:table-cell`}>{r.doctor}</TableCell>
      <TableCell>
        <StatusBadge colorClass={getExaminationStatusColor(r.status)}>
          {r.status}
        </StatusBadge>
      </TableCell>
      <TableCell className="text-right">
        {canEdit ? (
          <RowActionButton
            onClick={() => handleEdit(r.id)}
            aria-label={`検査操作: ${r.petName} ${r.date} ID ${r.id}`}
          />
        ) : null}
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
      className: "hidden sm:table-cell",
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
      className: "w-[100px] hidden md:table-cell",
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
      <div className="flex flex-col gap-4">
        <PropertyFilter
          properties={filterProperties}
          activeFilters={activeFilters}
          onFilterChange={handleFilterChange}
          searchTerm={searchTerm}
          onSearchChange={handleSearchChange}
          searchPlaceholder="飼主名、ペット名、検査種別..."
          count={isLoading ? undefined : filteredRecords.length}
          sortProperties={EXAMINATION_SORT_PROPERTIES}
          activeSorts={activeSorts}
          onSortChange={setActiveSorts}
        />

        {hasPageScopedFilter && pagination.totalPages > 1 ? (
          <div className={`flex items-start gap-2 rounded-md px-3 py-2 text-sm ${C.textWarning}`}>
            <Info className={`${ICON.action} ${C.textWarningIcon} shrink-0 mt-0.5`} />
            <span>
              検索・ステータス・検査種別・担当医での絞り込みは現在表示中のページ内のみが対象です。
              他のページにある検査記録はこの検索では見つかりません。全期間から探す場合は日付での絞り込みをご利用ください。
            </span>
          </div>
        ) : null}

        <FilteringIndicator isFiltering={isFiltering}>
          <DataTable
            headerRowClassName={DESIGN_TABLE_HEADER_ROW}
            headerCellClassName={DESIGN_TABLE_HEADER_CELL}
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
