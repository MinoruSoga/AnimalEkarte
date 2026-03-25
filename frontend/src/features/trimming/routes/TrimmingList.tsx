// React/Framework
import { useState, useCallback, useDeferredValue, useMemo, memo } from "react";
import { useNavigate } from "react-router";

// Hooks
import { useSortableData } from "@/hooks/use-sortable-data";
import { useModalState } from "@/hooks/use-modal-state";

// External
import { Plus, Scissors, AlertTriangle, Edit, Trash2, Calendar, CircleDot, PawPrint, User } from "lucide-react";
import { toast } from "sonner";

// Types
import type {
  FilterProperty,
  ActiveFilter,
  SortProperty,
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
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates/DataStates";
import { Pagination } from "@/components/shared/Pagination";
import { FilteringIndicator } from "@/components/shared/FilteringIndicator/FilteringIndicator";
import { getTrimmingStatusColor } from "@/utils/status-helpers";
import { usePagination } from "@/hooks/use-pagination";
import { useStaffValidation } from "@/hooks/use-staff-validation";
import type { TrimmingUI } from "@/types";
import { paths } from "@/config/paths";

// Relative (direct file import, no barrel)
import { useFilterTrimmingRecords } from "../hooks/use-trimming-records";

// rerender-memo + js-cache-function-results: renderRow インライン closure を memo コンポーネントに抽出
interface TrimmingTableRowProps {
  record: TrimmingUI;
  isValidStaff: (staff: string) => boolean;
  onEdit: (id: string) => void;
  onDeleteClick: (record: TrimmingUI) => void;
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

// rendering-hoist-jsx: 静的フィルタプロパティ（種・担当は動的オプションのためコンポーネント内で構築）
const TRIMMING_STATIC_FILTER_PROPERTIES: FilterProperty[] = [
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
      { value: "予約", label: "予約" },
      { value: "進行中", label: "進行中" },
      { value: "完了", label: "完了" },
    ],
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

  // rerender-transitions: 検索は useDeferredValue で遅延
  const deferredKeyword = useDeferredValue(searchKeyword);
  const isFiltering = searchKeyword !== deferredKeyword;

  // activeFilters から日付を抽出
  const dateFilter = activeFilters.find((f) => f.key === "date")?.value as
    | { from?: string; to?: string }
    | undefined;
  const deferredDate = { from: dateFilter?.from ?? "", to: dateFilter?.to ?? "" };

  const { data: filteredRecords, allTrimmings, isLoading, error, deleteRecord } = useFilterTrimmingRecords(deferredKeyword, deferredDate, activeFilters);
  const { isValidStaff } = useStaffValidation();

  // js-cache-function-results: ロード済みデータから種・担当の選択肢を動的生成
  const filterProperties = useMemo<FilterProperty[]>(() => {
    const speciesOptions = Array.from(new Set(allTrimmings.map((r) => r.species).filter(Boolean)))
      .sort()
      .map((s) => ({ value: s, label: s }));
    const staffOptions = Array.from(new Set(allTrimmings.map((r) => r.staff).filter(Boolean)))
      .sort()
      .map((s) => ({ value: s, label: s }));
    return [
      ...TRIMMING_STATIC_FILTER_PROPERTIES,
      { key: "species", label: "種", type: "select" as const, icon: PawPrint, options: speciesOptions },
      { key: "staff", label: "担当", type: "select" as const, icon: User, options: staffOptions },
    ];
  }, [allTrimmings]);

  const { activeSorts, setActiveSorts, toggleSort, directionFor, sortedData } =
    useSortableData(filteredRecords);

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

  const deleteModal = useModalState<{ id: string; label: string }>();

  const handleEdit = useCallback((id: string) => {
    navigate(`/trimming/${id}`, { state: { from: "/trimming" } });
  }, [navigate]);

  // rerender-dependencies: deleteModal のメソッドを primitive に抽出して deps を安定化
  const openDeleteModal = deleteModal.open;
  const closeDeleteModal = deleteModal.close;
  const deleteTargetId = deleteModal.item?.id;
  const deleteTargetLabel = deleteModal.item?.label;

  const handleDeleteClick = useCallback((record: TrimmingUI) => {
    openDeleteModal({
      id: record.id,
      label: `${record.ownerName} - ${record.petName}`,
    });
  }, [openDeleteModal]);

  const handleDeleteConfirm = useCallback(() => {
    if (deleteTargetId && deleteTargetLabel) {
      deleteRecord(deleteTargetId);
      toast.success("削除しました", { description: deleteTargetLabel });
      closeDeleteModal();
    }
  }, [deleteTargetId, deleteTargetLabel, deleteRecord, closeDeleteModal]);

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

  if (isLoading) return <LoadingFallback />;
  if (error) return <ErrorFallback />;

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
          properties={filterProperties}
          activeFilters={activeFilters}
          onFilterChange={setActiveFilters}
          searchTerm={searchKeyword}
          onSearchChange={handleSearchChange}
          searchPlaceholder="飼主名、ペット名..."
          count={filteredRecords.length}
          sortProperties={TRIMMING_SORT_PROPERTIES}
          activeSorts={activeSorts}
          onSortChange={setActiveSorts}
        />

        {/* Table */}
        <FilteringIndicator isFiltering={isFiltering}>
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
        </FilteringIndicator>

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
        open={deleteModal.isOpen}
        onClose={deleteModal.close}
        title="削除確認"
        description={`${deleteModal.item?.label} を削除してもよろしいですか？`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={handleDeleteConfirm}
      />
    </PageLayout>
  );
}
