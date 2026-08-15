// React/Framework
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { useCallback, useDeferredValue, useEffect, useMemo, useState } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { normalizeKana } from "@/lib/normalize-kana";

// External
import { AlertCircle, Calendar, ClipboardCheck, Plus } from "lucide-react";

// Internal
import { TableCell } from "@/components/ui/table";
import { CheckupAlertBadge } from "@/components/shared/CheckupAlertBadge/CheckupAlertBadge";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PropertyFilter } from "@/components/shared/PropertyFilter/PropertyFilter";
import { DataTable, DESIGN_TABLE_HEADER_ROW, DESIGN_TABLE_HEADER_CELL } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { DataTableRowLink } from "@/components/shared/DataTable/DataTableRowLink";
import { SortableHeader } from "@/components/shared/SortableHeader/SortableHeader";
import { FilteringIndicator } from "@/components/shared/FilteringIndicator/FilteringIndicator";
import { Pagination } from "@/components/shared/Pagination/Pagination";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { useSortableData } from "@/hooks/use-sortable-data";
import { usePermission } from "@/hooks/use-permission";
import { formatDate } from "@/lib/format/date";
import { paths } from "@/config/paths";
import { useGetCheckups } from "../api/get-checkups";
import { todayISODate, addDaysISO } from "@/lib/iso-date";

// Types
import type { FilterProperty, ActiveFilter, SortProperty } from "@/components/shared/PropertyFilter/types";
import { ResourceCheckups, ResourceMedicalRecords } from "@/types/generated/models";

// rendering-hoist-jsx: 静的定数をモジュールスコープに
const FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "alertStatus",
    label: "期限状態",
    type: "select",
    icon: AlertCircle,
    options: [
      { value: "overdue", label: "期限切れ" },
      { value: "upcoming30", label: "期限間近 (30日以内)" },
    ],
  },
  {
    key: "date",
    label: "日付",
    type: "date-range",
    icon: Calendar,
  },
];

// X-16②: BE 既定 limit と揃える（query_helpers.go parsePagination の既定値）
const PAGE_SIZE = 20;

const CHECKUPS_SORT_PROPERTIES: SortProperty[] = [
  { key: "date", label: "実施日" },
  { key: "ownerName", label: "飼主名" },
  { key: "petName", label: "ペット名" },
  { key: "checkupTypeName", label: "健診種別" },
  { key: "nextDate", label: "次回予定" },
];

export function CheckupsList() {
  const navigate = useNavigate();
  const { canView, canCreate, canEdit } = usePermission(ResourceMedicalRecords);
  const canCreateCheckup = canCreate && canEdit;
  const [searchParams, setSearchParams] = useSearchParams();
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const deferredSearch = useDeferredValue(searchTerm);
  const isFiltering = searchTerm !== deferredSearch;

  // activeFilters から日付フィルタ・アラートフィルタを抽出してAPIに渡す
  const filters = useMemo(() => {
    const today = todayISODate();
    const dateFilter = activeFilters.find((f) => f.key === "date")?.value as
      | { from?: string; to?: string }
      | undefined;
    const alertStatus = activeFilters.find((f) => f.key === "alertStatus")?.value as
      | string
      | undefined;

    let nextStartDate: string | undefined;
    let nextEndDate: string | undefined;

    if (alertStatus === "overdue") {
      nextEndDate = addDaysISO(today, -1);
    } else if (alertStatus === "upcoming30") {
      nextStartDate = today;
      nextEndDate = addDaysISO(today, 30);
    }

    return {
      startDate: dateFilter?.from,
      endDate: dateFilter?.to,
      nextStartDate,
      nextEndDate,
    };
  }, [activeFilters]);

  // X-16②: 実サーバページング。ページ変更は state で管理し、日付/アラートフィルタが
  // 変わったら 1 ページ目へ戻す（rerender-derived-state-no-effect: レンダー中に derived state で処理）。
  const [currentPage, setCurrentPage] = useState(1);
  const filtersResetKey = JSON.stringify(filters);
  const [prevFiltersResetKey, setPrevFiltersResetKey] = useState(filtersResetKey);
  if (prevFiltersResetKey !== filtersResetKey) {
    setPrevFiltersResetKey(filtersResetKey);
    setCurrentPage(1);
  }

  const requestFilters = useMemo(
    () => ({ ...filters, page: currentPage, limit: PAGE_SIZE }),
    [filters, currentPage],
  );

  const { data: checkupsResult, isLoading, error } = useGetCheckups(requestFilters);
  const checkups = useMemo(
    () => checkupsResult?.data ?? [],
    [checkupsResult?.data],
  );
  const total = checkupsResult?.total ?? 0;
  const limit = checkupsResult?.limit ?? PAGE_SIZE;
  const totalPages = Math.max(1, Math.ceil(total / limit));
  const safePage = Math.min(currentPage, totalPages);

  // テキスト検索・ソートは BE がサーバ側 search/sort パラメータを持たないため、
  // 取得済みの現在ページ内でクライアントサイドに行う（X-16②の既知トレードオフ）。
  const filteredRecords = useMemo(() => {
    if (!deferredSearch) return checkups;
    const normalizedTerm = normalizeKana(deferredSearch).toLowerCase();
    return checkups.filter(
      (c) =>
        normalizeKana(c.petName).toLowerCase().includes(normalizedTerm) ||
        normalizeKana(c.ownerName).toLowerCase().includes(normalizedTerm) ||
        normalizeKana(c.checkupTypeName).toLowerCase().includes(normalizedTerm) ||
        c.result.toLowerCase().includes(normalizedTerm),
    );
  }, [checkups, deferredSearch]);

  const { activeSorts, setActiveSorts, toggleSort, directionFor, sortedData } =
    useSortableData(filteredRecords);

  const startIndex = total === 0 ? 0 : (safePage - 1) * limit + 1;
  const endIndex = Math.min(safePage * limit, total);

  // FE-144: URLクエリパラメータからページ番号を読み取り、ローカル状態と同期
  // （URLが変わったときのみ。totalPages はサーバ応答後に確定するためクランプが必要）
  const urlPage = Number(searchParams.get("page") ?? 1);
  useEffect(() => {
    const clampedPage = Math.max(1, Math.min(urlPage, totalPages));
    if (clampedPage !== currentPage) {
      // URL/サーバ total 由来のページ同期。render 中 setState は不可のため effect で反映する。
      // eslint-disable-next-line react-hooks/set-state-in-effect -- FE-144 URL page clamp sync
      setCurrentPage(clampedPage);
    }
  // currentPage は比較対象のみ。URL/totalPages 変化時だけ同期する（自己ループ防止）
  // eslint-disable-next-line react-hooks/exhaustive-deps -- FE-144 URL page sync
  }, [urlPage, totalPages]);

  // FE-144: ページ変更時にURLクエリパラメータを更新
  const handlePageChange = useCallback((page: number) => {
    setCurrentPage(page);
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

  const columns = useMemo(
    () => [
      {
        header: (
          <SortableHeader
            label="実施日"
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
            label="健診種別"
            direction={directionFor("checkupTypeName")}
            onToggle={() => toggleSort("checkupTypeName")}
          />
        ),
        className: "hidden md:table-cell",
      },
      {
        header: (
          <SortableHeader
            label="次回予定"
            direction={directionFor("nextDate")}
            onToggle={() => toggleSort("nextDate")}
          />
        ),
        // Clinical deadline cue: keep reachable via md+; full hide only under md (BUG-458).
        className: "w-[120px] hidden md:table-cell",
      },
      { header: "結果・所見", className: "hidden lg:table-cell" },
      { header: "担当医", className: "w-[100px] hidden md:table-cell" },
      { header: "操作", className: "w-[80px]", align: "right" as const },
    ],
    [directionFor, toggleSort],
  );

  const handleCreate = useCallback(() => {
    navigate(paths.checkups.selectPet.getHref());
  }, [navigate]);

  const handleEdit = useCallback((medicalRecordId: string) => {
    navigate(paths.medicalRecords.detail.getHref(medicalRecordId));
  }, [navigate]);

  if (isLoading) return <LoadingFallback />;
  if (error) return <ErrorFallback />;

  return (
    <PageLayout
      title="定期健診"
      resource={ResourceCheckups}
      icon={<ClipboardCheck className={`${ICON.page} ${C.text}`} />}
      maxWidth={LAYOUT.pageContentMaxWidth.full}
      headerAction={
        canCreateCheckup ? (
          <PrimaryButton colorVariant="primary" onClick={handleCreate}>
            <Plus className={`mr-1.5 ${ICON.action}`} />
            新規登録
          </PrimaryButton>
        ) : null
      }
    >
      <div className="flex flex-col gap-4">
        <PropertyFilter
          properties={FILTER_PROPERTIES}
          activeFilters={activeFilters}
          onFilterChange={setActiveFilters}
          searchTerm={searchTerm}
          onSearchChange={setSearchTerm}
          searchPlaceholder="ペット名・飼主名・種別で検索..."
          count={isLoading ? undefined : filteredRecords.length}
          sortProperties={CHECKUPS_SORT_PROPERTIES}
          activeSorts={activeSorts}
          onSortChange={setActiveSorts}
        />

        <FilteringIndicator isFiltering={isFiltering}>
          <DataTable
            headerRowClassName={DESIGN_TABLE_HEADER_ROW}
            headerCellClassName={DESIGN_TABLE_HEADER_CELL}
            columns={columns}
            data={sortedData}
            emptyMessage="定期健診の記録がありません"
            renderRow={(c) => (
              <DataTableRow key={c.id}>
                <TableCell className={`font-mono ${C.text}`}>
                  {c.date ? formatDate(c.date) : "-"}
                </TableCell>
                <TableCell className={C.text}>{c.ownerName || "-"}</TableCell>
                <TableCell className={C.text}>
                  {canView && c.medicalRecordId ? (
                    <DataTableRowLink
                      to={paths.medicalRecords.detail.getHref(c.medicalRecordId)}
                      aria-label={`カルテ詳細: ${c.petName || "-"} ${c.date || "-"} 健診ID ${c.id}`}
                    >
                      {c.petName || "-"}
                    </DataTableRowLink>
                  ) : (c.petName || "-")}
                </TableCell>
                <TableCell className={`${C.text} hidden md:table-cell`}>{c.checkupTypeName || "-"}</TableCell>
                <TableCell className={`font-mono ${C.text} hidden md:table-cell`}>
                  <div className="flex items-center gap-1.5">
                    {c.nextDate ? formatDate(c.nextDate) : "-"}
                    <CheckupAlertBadge nextDate={c.nextDate} />
                  </div>
                </TableCell>
                <TableCell className={`${C.text} max-w-xs truncate hidden lg:table-cell`}>
                  {c.result || "-"}
                </TableCell>
                <TableCell className={`${C.text} hidden md:table-cell`}>{c.doctorName || "-"}</TableCell>
                <TableCell className="text-right">
                  {canView && canEdit ? (
                    <RowActionButton
                      onClick={() => handleEdit(c.medicalRecordId)}
                      aria-label={`健診操作: ${c.petName || "-"} ${c.date || "-"} ID ${c.id}`}
                    />
                  ) : null}
                </TableCell>
              </DataTableRow>
            )}
          />
        </FilteringIndicator>

        {totalPages > 1 ? (
          <Pagination
            currentPage={safePage}
            totalPages={totalPages}
            totalCount={total}
            startIndex={startIndex}
            endIndex={endIndex}
            onPageChange={handlePageChange}
            onPrev={() => handlePageChange(safePage - 1)}
            onNext={() => handlePageChange(safePage + 1)}
          />
        ) : null}
      </div>
    </PageLayout>
  );
}
