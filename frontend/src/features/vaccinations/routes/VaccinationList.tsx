// React/Framework
import { useState, useDeferredValue, useCallback, useMemo } from "react";
import { useNavigate } from "react-router";

// External
import { Plus, Syringe, FileSpreadsheet, Calendar } from "lucide-react";

// Internal
import { paths } from "@/config/paths";
import { Button } from "@/components/ui/button";
import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { SortableHeader } from "@/components/shared/SortableHeader/SortableHeader";
import { usePagination } from "@/hooks/use-pagination";
import { Pagination } from "@/components/shared/Pagination/Pagination";

// Relative
import { useFilterVaccinations } from "../hooks/use-vaccinations";

// Types
import type {
  FilterProperty,
  ActiveFilter,
  SortProperty,
  ActiveSort,
} from "@/components/shared/NotionFilter/types";

type SortKey = "date" | "ownerName" | "petName" | "vaccineName" | "nextDate";

const FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "date",
    label: "日付",
    type: "date-range",
    icon: Calendar,
  },
];

// rendering-hoist-jsx: 静的ソートプロパティ定義
const VACCINATION_SORT_PROPERTIES: SortProperty[] = [
  { key: "date", label: "実施日" },
  { key: "ownerName", label: "飼主名" },
  { key: "petName", label: "ペット名" },
  { key: "vaccineName", label: "予防接種名" },
  { key: "nextDate", label: "次回予定" },
];

export function VaccinationList() {
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const [activeSorts, setActiveSorts] = useState<ActiveSort[]>([]);
  const deferredSearchTerm = useDeferredValue(searchTerm);
  const isFiltering = searchTerm !== deferredSearchTerm;

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

  const { data: filteredRecords } = useFilterVaccinations(deferredSearchTerm, filters);

  // ── Sort logic driven by activeSorts ──
  const handleSortChange = useCallback((sorts: ActiveSort[]) => {
    setActiveSorts(sorts);
  }, []);

  const toggleSort = useCallback((key: SortKey) => {
    setActiveSorts((prev) => {
      const existing = prev.find((s) => s.key === key);
      if (!existing) {
        return [{ key, direction: "asc" as const }];
      }
      if (existing.direction === "asc") {
        return prev.map((s) => s.key === key ? { ...s, direction: "desc" as const } : s);
      }
      return prev.filter((s) => s.key !== key);
    });
  }, []);

  const directionFor = useCallback(
    (key: SortKey): "ascending" | "descending" | "none" => {
      const sort = activeSorts.find((s) => s.key === key);
      if (!sort) return "none";
      return sort.direction === "asc" ? "ascending" : "descending";
    },
    [activeSorts],
  );

  const sortedData = useMemo(() => {
    if (activeSorts.length === 0) return [...filteredRecords];
    const sorted = [...filteredRecords];
    sorted.sort((a, b) => {
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
  }, [filteredRecords, activeSorts]);

  const pagination = usePagination(sortedData, {
    pageSize: 20,
    resetKey: deferredSearchTerm,
  });

  const handleCreate = useCallback(() => {
    navigate(paths.vaccinations.selectPet.getHref());
  }, [navigate]);

  const handleEdit = useCallback((id: string) => {
    navigate(`/vaccinations/${id}`);
  }, [navigate]);

  const columns = useMemo(() => [
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
          label="予防接種名"
          direction={directionFor("vaccineName")}
          onToggle={() => toggleSort("vaccineName")}
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
      className: "w-[140px]",
    },
    { header: "操作", className: "w-[100px]", align: "right" as const },
  ], [directionFor, toggleSort]);

  return (
    <PageLayout
      title="予防接種管理"
      icon={<Syringe className="size-5 text-[#37352F]" />}
      headerAction={
        <div className="flex items-center gap-2">
          <Button variant="outline" className="h-10 text-base gap-2 bg-white" onClick={() => {}}>
            <FileSpreadsheet className="size-4" />
            データ取込
          </Button>
          <PrimaryButton onClick={handleCreate}>
            <Plus className="mr-1.5 size-4" />
            新規登録
          </PrimaryButton>
        </div>
      }
      maxWidth="max-w-full"
    >
      <div className="flex flex-col gap-4">
        <NotionFilter
          properties={FILTER_PROPERTIES}
          activeFilters={activeFilters}
          onFilterChange={setActiveFilters}
          searchTerm={searchTerm}
          onSearchChange={setSearchTerm}
          searchPlaceholder="飼主名、ペット名、予防接種名..."
          count={filteredRecords.length}
          sortProperties={VACCINATION_SORT_PROPERTIES}
          activeSorts={activeSorts}
          onSortChange={handleSortChange}
        />

        <div className={isFiltering ? "opacity-60 transition-opacity duration-150" : "transition-opacity duration-150"}>
          <DataTable
            columns={columns}
            data={pagination.paginatedData}
            emptyMessage="データが見つかりません"
            renderRow={(r) => (
              <DataTableRow
                key={r.id}
                onClick={() => handleEdit(r.id)}
              >
                <TableCell className="font-mono text-base text-[#37352F] py-2">{r.date}</TableCell>
                <TableCell className="text-base text-[#37352F] py-2">{r.ownerName}</TableCell>
                <TableCell className="text-base text-[#37352F] py-2">{r.petName}</TableCell>
                <TableCell className="text-base font-medium text-[#37352F] py-2">{r.vaccineName}</TableCell>
                <TableCell className="font-mono text-base text-[#37352F] py-2">{r.nextDate}</TableCell>
                <TableCell className="text-right py-2">
                  <RowActionButton onClick={() => handleEdit(r.id)} />
                </TableCell>
              </DataTableRow>
            )}
          />
        </div>

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
