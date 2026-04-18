// React/Framework
import { C, ICON } from "@/lib/design-tokens";
import { useCallback, useDeferredValue, useEffect, useMemo, useState } from "react";
import { useNavigate, useSearchParams } from "react-router";

// External
import { Calendar, ClipboardCheck, Plus } from "lucide-react";

// Internal
import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { SortableHeader } from "@/components/shared/SortableHeader/SortableHeader";
import { FilteringIndicator } from "@/components/shared/FilteringIndicator/FilteringIndicator";
import { Pagination } from "@/components/shared/Pagination/Pagination";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { useSortableData } from "@/hooks/use-sortable-data";
import { usePagination } from "@/hooks/use-pagination";
import { usePermission } from "@/hooks/use-permission";
import { formatDate } from "@/utils/format/date";
import { paths } from "@/config/paths";
import { useGetCheckups } from "../api/get-checkups";

// Types
import type { FilterProperty, ActiveFilter, SortProperty } from "@/components/shared/NotionFilter/types";
import { ResourceCheckups } from "@/types/generated/models";

// rendering-hoist-jsx: 静的定数をモジュールスコープに
const FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "date",
    label: "日付",
    type: "date-range",
    icon: Calendar,
  },
];

const CHECKUPS_SORT_PROPERTIES: SortProperty[] = [
  { key: "date", label: "実施日" },
  { key: "ownerName", label: "飼主名" },
  { key: "petName", label: "ペット名" },
  { key: "checkupTypeName", label: "健診種別" },
  { key: "nextDate", label: "次回予定" },
];

export function CheckupsList() {
  const navigate = useNavigate();
  const { canCreate, canEdit } = usePermission(ResourceCheckups);
  const [searchParams, setSearchParams] = useSearchParams();
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const deferredSearch = useDeferredValue(searchTerm);
  const isFiltering = searchTerm !== deferredSearch;

  // activeFilters から日付フィルタを抽出してAPIに渡す
  const filters = useMemo(() => {
    const dateFilter = activeFilters.find((f) => f.key === "date")?.value as
      | { from?: string; to?: string }
      | undefined;
    return {
      startDate: dateFilter?.from,
      endDate: dateFilter?.to,
    };
  }, [activeFilters]);

  const { data: checkups = [], isLoading, error } = useGetCheckups(filters);

  // テキスト検索はクライアントサイドで行う
  const filteredRecords = useMemo(() => {
    if (!deferredSearch) return checkups;
    const searchQuery = deferredSearch.toLowerCase();
    return checkups.filter(
      (c) =>
        c.petName.toLowerCase().includes(searchQuery) ||
        c.ownerName.toLowerCase().includes(searchQuery) ||
        c.checkupTypeName.toLowerCase().includes(searchQuery) ||
        c.result.toLowerCase().includes(searchQuery),
    );
  }, [checkups, deferredSearch]);

  const { activeSorts, setActiveSorts, toggleSort, directionFor, sortedData } =
    useSortableData(filteredRecords);

  const pagination = usePagination(sortedData, {
    pageSize: 20,
    resetKey: deferredSearch,
  });

  // FE-144: URLクエリパラメータからページ番号を読み取る
  const urlPage = Number(searchParams.get("page") ?? 1);

  // FE-144: URLのページ番号とローカル状態を同期（URLが変わったときのみ）
  // rerender-dependencies: pagination（オブジェクト）を destructure し primitive を deps に使用
  const { totalPages, currentPage, goToPage } = pagination;
  useEffect(() => {
    const clampedPage = Math.max(1, Math.min(urlPage, totalPages));
    if (clampedPage !== currentPage) {
      goToPage(clampedPage);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [urlPage, totalPages]);

  // FE-144: ページ変更時にURLクエリパラメータを更新
  const handlePageChange = useCallback((page: number) => {
    goToPage(page);
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev);
      if (page === 1) {
        next.delete("page");
      } else {
        next.set("page", String(page));
      }
      return next;
    }, { replace: true });
  }, [goToPage, setSearchParams]);

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
      },
      {
        header: (
          <SortableHeader
            label="次回予定"
            direction={directionFor("nextDate")}
            onToggle={() => toggleSort("nextDate")}
          />
        ),
        className: "w-[120px] hidden lg:table-cell",
      },
      { header: "結果・所見", className: "hidden lg:table-cell" },
      { header: "担当医", className: "w-[100px]" },
      { header: "操作", className: "w-[80px]", align: "right" as const },
    ],
    [directionFor, toggleSort],
  );

  const handleCreate = useCallback(() => {
    navigate(paths.medicalRecords.selectPet.getHref());
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
      maxWidth="max-w-full"
      headerAction={
        canCreate ? (
          <PrimaryButton onClick={handleCreate}>
            <Plus className={`mr-1.5 ${ICON.action}`} />
            新規登録
          </PrimaryButton>
        ) : null
      }
    >
      <div className="flex flex-col gap-4">
        <NotionFilter
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
            columns={columns}
            data={pagination.paginatedData}
            emptyMessage="定期健診の記録がありません"
            renderRow={(c) => (
              <DataTableRow key={c.id} onClick={canEdit ? () => handleEdit(c.medicalRecordId) : undefined}>
                <TableCell className={`font-mono text-base ${C.text} py-2`}>
                  {c.date ? formatDate(c.date) : "-"}
                </TableCell>
                <TableCell className={`text-base ${C.text} py-2`}>{c.ownerName || "-"}</TableCell>
                <TableCell className={`text-base ${C.text} py-2`}>{c.petName || "-"}</TableCell>
                <TableCell className={`text-base ${C.text} py-2`}>{c.checkupTypeName || "-"}</TableCell>
                <TableCell className={`font-mono text-base ${C.text} py-2 hidden lg:table-cell`}>
                  {c.nextDate ? formatDate(c.nextDate) : "-"}
                </TableCell>
                <TableCell className={`text-base ${C.text} py-2 max-w-xs truncate hidden lg:table-cell`}>
                  {c.result || "-"}
                </TableCell>
                <TableCell className={`text-base ${C.text} py-2`}>{c.doctorName || "-"}</TableCell>
                <TableCell className="text-right py-2">
                  {canEdit ? <RowActionButton onClick={() => handleEdit(c.medicalRecordId)} /> : null}
                </TableCell>
              </DataTableRow>
            )}
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
