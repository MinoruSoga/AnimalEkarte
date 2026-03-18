// React/Framework
import { useState, useDeferredValue, useCallback, useMemo } from "react";
import { useNavigate } from "react-router";

// External
import { Plus, TestTube, FileSpreadsheet, Calendar, CircleDot } from "lucide-react";

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

// Relative
import { useExaminationRecords } from "../hooks/useExaminationRecords";
import { paths } from "@/config/paths";

// Types
import type {
  FilterProperty,
  ActiveFilter,
  SortProperty,
  ActiveSort,
} from "@/components/shared/NotionFilter/types";

type SortKey = "date" | "ownerName" | "petName" | "testType" | "doctor" | "status";

const FILTER_PROPERTIES: FilterProperty[] = [
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
    options: [
      { value: "pending", label: "依頼中" },
      { value: "in_progress", label: "検査中" },
      { value: "completed", label: "完了" },
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

export function Examinations() {
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const [activeSorts, setActiveSorts] = useState<ActiveSort[]>([]);
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

  const { data: filteredRecords, isLoading } = useExaminationRecords(deferredSearch, filters);

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

  const handleCreate = useCallback(() => {
    navigate(paths.examinations.selectPet.getHref());
  }, [navigate]);

  const handleEdit = useCallback((id: string) => {
    navigate(`/examinations/${id}`);
  }, [navigate]);

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
    { header: "結果概要" },
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
      className: "w-[80px]",
    },
    { header: "操作", className: "w-[80px]", align: "right" as const },
  ], [directionFor, toggleSort]);

  return (
    <PageLayout
      title="検査管理"
      icon={<TestTube className="size-5 text-[#37352F]" />}
      headerAction={
        <div className="flex items-center gap-2">
          <Button variant="outline" className="h-10 text-sm gap-2 bg-white" onClick={() => {}}>
            <FileSpreadsheet className="size-4" />
            検査データ取込
          </Button>
          <PrimaryButton onClick={handleCreate}>
            <Plus className="mr-1.5 size-4" />
            新規検査登録
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
          searchPlaceholder="飼主名、ペット名、検査種別..."
          count={isLoading ? undefined : filteredRecords.length}
          sortProperties={EXAMINATION_SORT_PROPERTIES}
          activeSorts={activeSorts}
          onSortChange={handleSortChange}
        />

        <div className={isFiltering ? "opacity-60 transition-opacity duration-150" : "transition-opacity duration-150"}>
          <DataTable
            columns={columns}
            data={sortedData}
            emptyMessage="検査データが見つかりません"
            renderRow={(r) => (
              <DataTableRow
                key={r.id}
                onClick={() => handleEdit(r.id)}
              >
                <TableCell className="font-mono text-sm text-[#37352F] py-2">{r.date}</TableCell>
                <TableCell className="text-sm text-[#37352F] py-2">{r.ownerName}</TableCell>
                <TableCell className="text-sm text-[#37352F] py-2">{r.petName}</TableCell>
                <TableCell className="text-sm font-medium text-[#37352F] py-2">{r.testType}</TableCell>
                <TableCell className="text-sm text-muted-foreground truncate max-w-[200px] py-2">
                  {r.resultSummary || "-"}
                </TableCell>
                <TableCell className="text-sm text-[#37352F] py-2">{r.doctor}</TableCell>
                <TableCell className="py-2">
                  <StatusBadge colorClass={getExaminationStatusColor(r.status)}>
                    {r.status}
                  </StatusBadge>
                </TableCell>
                <TableCell className="text-right py-2">
                  <RowActionButton onClick={() => handleEdit(r.id)} />
                </TableCell>
              </DataTableRow>
            )}
          />
        </div>
      </div>
    </PageLayout>
  );
}
