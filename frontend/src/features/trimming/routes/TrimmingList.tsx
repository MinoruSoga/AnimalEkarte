// React/Framework
import { useState, useCallback, useDeferredValue, useMemo, memo } from "react";
import { useNavigate } from "react-router";

// External
import { Plus, Scissors, AlertTriangle, Edit, Trash2, Calendar } from "lucide-react";
import { toast } from "sonner";

// Types
import type {
  FilterProperty,
  ActiveFilter,
  SortProperty,
  ActiveSort,
} from "@/components/shared/NotionFilter/types";

// Internal
import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { StatusBadge } from "@/components/shared/StatusBadge/StatusBadge";
import { RowActionDropdown } from "@/components/shared/RowActionDropdown";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { SortableHeader } from "@/components/shared/SortableHeader/SortableHeader";
import { Pagination } from "@/components/shared/Pagination";
import { getTrimmingStatusColor } from "@/utils/status-helpers";
import { usePagination } from "@/hooks/use-pagination";
import { useStaffValidation } from "@/hooks/use-staff-validation";
import type { TrimmingRecord } from "@/types";
import { paths } from "@/config/paths";

// Relative (direct file import, no barrel)
import { useTrimmingRecords } from "../hooks/use-trimming-records";

type SortKey = "date" | "ownerName" | "petName" | "species" | "staff" | "status";

// rerender-memo + js-cache-function-results: renderRow インライン closure を memo コンポーネントに抽出
interface TrimmingTableRowProps {
  record: TrimmingRecord;
  isValidStaff: (staff: string) => boolean;
  onEdit: (id: string) => void;
  onDeleteClick: (record: TrimmingRecord) => void;
}

const TrimmingTableRow = memo(function TrimmingTableRow({
  record,
  isValidStaff,
  onEdit,
  onDeleteClick,
}: TrimmingTableRowProps) {
  return (
    <DataTableRow onClick={() => onEdit(record.id)}>
      <TableCell className="font-mono text-base text-[#37352F] py-2">
        {record.date}
      </TableCell>
      <TableCell className="text-base text-[#37352F] py-2">{record.ownerName}</TableCell>
      <TableCell className="py-2">
        <div className="flex flex-col">
          <span className="text-base text-[#37352F]">{record.petName}</span>
          <span className="text-base text-[#37352F]/60">{record.petNumber}</span>
        </div>
      </TableCell>
      <TableCell className="text-base text-[#37352F] py-2">{record.species}</TableCell>
      <TableCell className="text-base text-[#37352F] py-2">{record.weight}</TableCell>
      <TableCell className="text-base text-[#37352F] truncate max-w-[200px] py-2">
        {record.styleRequest}
      </TableCell>
      <TableCell className="text-base text-[#37352F] py-2">
        <div className="flex items-center gap-1.5">
          {!isValidStaff(record.staff) ? (
            <AlertTriangle className="size-4 text-amber-500" />
          ) : null}
          {record.staff}
        </div>
      </TableCell>
      <TableCell className="py-2">
        <StatusBadge colorClass={getTrimmingStatusColor(record.status)}>
          {record.status}
        </StatusBadge>
      </TableCell>
      <TableCell className="text-right py-2">
        <RowActionDropdown
          actions={[
            {
              label: "編集",
              icon: Edit,
              onClick: () => onEdit(record.id),
            },
            {
              label: "削除",
              icon: Trash2,
              variant: "destructive",
              onClick: () => onDeleteClick(record),
            },
          ]}
        />
      </TableCell>
    </DataTableRow>
  );
});

const TRIMMING_FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "date",
    label: "日付",
    type: "date-range",
    icon: Calendar,
  },
];

// rendering-hoist-jsx: 静的ソートプロパティ定義
const TRIMMING_SORT_PROPERTIES: SortProperty[] = [
  { key: "date", label: "診療日" },
  { key: "ownerName", label: "飼主名" },
  { key: "petName", label: "ペット名" },
  { key: "species", label: "種" },
  { key: "staff", label: "担当" },
  { key: "status", label: "ステータス" },
];

export function TrimmingList() {
  const navigate = useNavigate();
  const [searchKeyword, setSearchKeyword] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const [activeSorts, setActiveSorts] = useState<ActiveSort[]>([]);

  // rerender-transitions: 検索は useDeferredValue で遅延
  const deferredKeyword = useDeferredValue(searchKeyword);
  const isFiltering = searchKeyword !== deferredKeyword;

  // activeFilters から日付を抽出
  const dateFilter = activeFilters.find((f) => f.key === "date")?.value as
    | { from?: string; to?: string }
    | undefined;
  const deferredDate = { from: dateFilter?.from ?? "", to: dateFilter?.to ?? "" };

  const { data: filteredRecords, isLoading, error, deleteRecord } = useTrimmingRecords(deferredKeyword, deferredDate);
  const { isValidStaff } = useStaffValidation();

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

  const {
    currentPage,
    paginatedData,
    totalPages,
    startIndex,
    endIndex,
    goToPage,
    prevPage,
    nextPage,
  } = usePagination(sortedData, { pageSize: 10 });

  const [deleteTarget, setDeleteTarget] = useState<{ id: string; label: string } | null>(null);

  const handleEdit = useCallback((id: string) => {
    navigate(`/trimming/${id}`, { state: { from: "/trimming" } });
  }, [navigate]);

  const handleDeleteClick = useCallback((record: TrimmingRecord) => {
    setDeleteTarget({
      id: record.id,
      label: `${record.ownerName} - ${record.petName}`,
    });
  }, []);

  // rerender-dependencies: deleteTarget (object) から primitive を抽出して deps を安定化
  const deleteTargetId = deleteTarget?.id;
  const deleteTargetLabel = deleteTarget?.label;

  const handleDeleteConfirm = useCallback(() => {
    if (deleteTargetId && deleteTargetLabel) {
      deleteRecord(deleteTargetId);
      toast.success("削除しました", { description: deleteTargetLabel });
      setDeleteTarget(null);
    }
  }, [deleteTargetId, deleteTargetLabel, deleteRecord]);

  const handleNew = useCallback(() => {
    navigate(paths.trimming.selectPet.getHref());
  }, [navigate]);

  // rerender-functional-setstate: setSearchKeyword を useCallback でラップして安定化
  const handleSearchChange = useCallback((v: string) => {
    setSearchKeyword(v);
  }, []);

  const columns = useMemo(() => [
    {
      header: (
        <SortableHeader
          label="診療日"
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
          label="種"
          direction={directionFor("species")}
          onToggle={() => toggleSort("species")}
        />
      ),
      className: "w-[80px]",
    },
    { header: "体重", className: "w-[80px]" },
    { header: "スタイル希望" },
    {
      header: (
        <SortableHeader
          label="担当"
          direction={directionFor("staff")}
          onToggle={() => toggleSort("staff")}
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
    { header: "操作", className: "w-[100px]", align: "right" as const },
  ], [directionFor, toggleSort]);

  if (isLoading) return (
    <div className="flex justify-center items-center p-8">
      <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-[#37352F]" />
    </div>
  );
  if (error) return <div className="p-4 text-red-600">データの取得に失敗しました</div>;

  return (
    <PageLayout
      title="トリミング管理"
      icon={<Scissors className="size-5 text-[#37352F]" />}
      headerAction={
        <PrimaryButton onClick={handleNew}>
          <Plus className="mr-1.5 size-4" />
          新規登録
        </PrimaryButton>
      }
      maxWidth="max-w-full"
    >
      <div className="flex flex-col gap-4">
        {/* Filters */}
        <NotionFilter
          properties={TRIMMING_FILTER_PROPERTIES}
          activeFilters={activeFilters}
          onFilterChange={setActiveFilters}
          searchTerm={searchKeyword}
          onSearchChange={handleSearchChange}
          searchPlaceholder="飼主名、ペット名..."
          count={filteredRecords.length}
          sortProperties={TRIMMING_SORT_PROPERTIES}
          activeSorts={activeSorts}
          onSortChange={handleSortChange}
        />

        {/* Table */}
        <div className={isFiltering ? "opacity-60 transition-opacity duration-150" : "transition-opacity duration-150"}>
          <DataTable
            columns={columns}
            data={paginatedData}
            renderRow={(record) => (
              <TrimmingTableRow
                key={record.id}
                record={record}
                isValidStaff={isValidStaff}
                onEdit={handleEdit}
                onDeleteClick={handleDeleteClick}
              />
            )}
          />
        </div>

        {/* Pagination */}
        <Pagination
          currentPage={currentPage}
          totalPages={totalPages}
          totalCount={filteredRecords.length}
          startIndex={startIndex}
          endIndex={endIndex}
          onPageChange={goToPage}
          onPrev={prevPage}
          onNext={nextPage}
        />
      </div>

      {/* Delete Confirmation Dialog */}
      <ConfirmDialog
        open={deleteTarget !== null}
        onClose={() => setDeleteTarget(null)}
        title="削除確認"
        description={`${deleteTarget?.label} を削除してもよろしいですか？`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={handleDeleteConfirm}
      />
    </PageLayout>
  );
}
