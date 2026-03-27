// React/Framework
import { ICON } from "@/lib/design-tokens";
import { useDeferredValue, useMemo, useState } from "react";

// External
import { Calendar, ClipboardCheck } from "lucide-react";

// Internal
import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { SortableHeader } from "@/components/shared/SortableHeader/SortableHeader";
import { FilteringIndicator } from "@/components/shared/FilteringIndicator/FilteringIndicator";
import { Pagination } from "@/components/shared/Pagination/Pagination";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates/DataStates";
import { useSortableData } from "@/hooks/use-sortable-data";
import { usePagination } from "@/hooks/use-pagination";
import { formatDate } from "@/utils/format/date";
import { useGetCheckups } from "../api/get-checkups";

// Types
import type { FilterProperty, ActiveFilter, SortProperty } from "@/components/shared/NotionFilter/types";

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
    ],
    [directionFor, toggleSort],
  );

  if (isLoading) return <LoadingFallback />;
  if (error) return <ErrorFallback />;

  return (
    <PageLayout
      title="定期健診"
      icon={<ClipboardCheck className={`${ICON.page} text-[#37352F]`} />}
      maxWidth="max-w-full"
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
              <DataTableRow key={c.id}>
                <TableCell className="font-mono text-base text-[#37352F] py-2">
                  {c.date ? formatDate(c.date) : "-"}
                </TableCell>
                <TableCell className="text-base text-[#37352F] py-2">{c.ownerName || "-"}</TableCell>
                <TableCell className="text-base text-[#37352F] py-2">{c.petName || "-"}</TableCell>
                <TableCell className="text-base text-[#37352F] py-2">{c.checkupTypeName || "-"}</TableCell>
                <TableCell className="font-mono text-base text-[#37352F] py-2 hidden lg:table-cell">
                  {c.nextDate ? formatDate(c.nextDate) : "-"}
                </TableCell>
                <TableCell className="text-base text-[#37352F] py-2 max-w-xs truncate hidden lg:table-cell">
                  {c.result || "-"}
                </TableCell>
                <TableCell className="text-base text-[#37352F] py-2">{c.doctorName || "-"}</TableCell>
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
            onPageChange={pagination.goToPage}
            onPrev={pagination.prevPage}
            onNext={pagination.nextPage}
          />
        ) : null}
      </div>
    </PageLayout>
  );
}
