// React/Framework
import { type ReactNode, useState, useCallback, useDeferredValue, useMemo } from "react";
import { useNavigate } from "react-router";

// Hooks
import { useSortableData } from "@/hooks/use-sortable-data";
import { useModalState } from "@/hooks/use-modal-state";

// External
import { Plus, FileText, Edit, Trash2, Receipt, AlertTriangle, Calendar } from "lucide-react";

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
import { C, STYLE } from "@/lib/design-tokens";
import { getMedicalRecordStatusColor } from "@/utils/status-helpers";
import { usePagination } from "@/hooks/use-pagination";
import { useStaffValidation } from "@/hooks/use-staff-validation";

// Relative
import { useFilterMedicalRecords } from "../hooks/use-medical-records";
import { useDeleteMedicalRecord } from "../api/delete-medical-record";

// Types
import type {
  FilterProperty,
  ActiveFilter,
  SortProperty,
} from "@/components/shared/NotionFilter/types";
import type { MedicalRecordFilters } from "../api/get-medical-records";

// rendering-hoist-jsx: 静的フィルタプロパティ定義
const FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "date",
    label: "診療日",
    type: "date-range",
    icon: Calendar,
  },
];

// rendering-hoist-jsx: 静的ソートプロパティ定義
const MEDICAL_RECORD_SORT_PROPERTIES: SortProperty[] = [
  { key: "date", label: "診療日" },
  { key: "ownerName", label: "飼主名" },
  { key: "petName", label: "ペット名" },
  { key: "species", label: "種" },
  { key: "doctor", label: "担当医" },
  { key: "status", label: "ステータス" },
];

export function MedicalRecords() {
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const deferredSearch = useDeferredValue(searchTerm);

  // activeFilters から日付フィルタのみを抽出してAPIに渡す
  const apiFilters = useMemo<MedicalRecordFilters>(() => {
    const dateFilter = activeFilters.find((f) => f.key === "date")?.value as
      | { from?: string; to?: string }
      | undefined;
    return {
      startDate: dateFilter?.from,
      endDate: dateFilter?.to,
    };
  }, [activeFilters]);

  const { data: filteredRecords, isLoading, isError } = useFilterMedicalRecords(deferredSearch, apiFilters);
  const deleteModal = useModalState<{ id: string; label: string }>();
  const { mutate: deleteRecord } = useDeleteMedicalRecord();
  const { isValidStaff } = useStaffValidation();

  const { activeSorts, setActiveSorts, toggleSort, directionFor, sortedData } =
    useSortableData(filteredRecords);

  const {
    paginatedData,
    currentPage,
    totalPages,
    totalCount,
    startIndex,
    endIndex,
    goToPage,
    nextPage,
    prevPage,
  } = usePagination(sortedData, { pageSize: 20, resetKey: searchTerm });

  const isFiltering = searchTerm !== deferredSearch;

  const handleNavigateToForm = useCallback((recordId?: string) => {
    navigate(
      recordId ? `/medical-records/${recordId}` : "/medical-records/select-pet",
      { state: { from: "/medical-records" } },
    );
  }, [navigate]);

  if (isLoading) return <LoadingFallback />;
  if (isError) return <ErrorFallback />;

  const COLUMNS: { header: ReactNode; className?: string; align?: "left" | "center" | "right" }[] = [
    {
      header: <SortableHeader label="診療日" direction={directionFor("date")} onToggle={() => toggleSort("date")} />,
      className: "w-[120px]",
    },
    {
      header: <SortableHeader label="飼主名" direction={directionFor("ownerName")} onToggle={() => toggleSort("ownerName")} />,
    },
    {
      header: <SortableHeader label="ペット名" direction={directionFor("petName")} onToggle={() => toggleSort("petName")} />,
    },
    {
      header: <SortableHeader label="種" direction={directionFor("species")} onToggle={() => toggleSort("species")} />,
      className: "w-[80px]",
    },
    { header: "主訴" },
    { header: "関連", className: "w-[100px]" },
    {
      header: <SortableHeader label="担当医" direction={directionFor("doctor")} onToggle={() => toggleSort("doctor")} />,
      className: "w-[100px]",
    },
    {
      header: <SortableHeader label="ステータス" direction={directionFor("status")} onToggle={() => toggleSort("status")} />,
      className: "w-[100px]",
    },
    { header: "操作", className: "w-[100px]", align: "right" as const },
  ];

  return (
    <PageLayout
      title="カルテ管理"
      icon={<FileText className={`size-5 ${C.text}`} />}
      headerAction={
        <PrimaryButton onClick={() => handleNavigateToForm()}>
          <Plus className="size-4" />
          新規カルテ作成
        </PrimaryButton>
      }
      maxWidth="max-w-full"
    >
      <div className="flex flex-col gap-4 flex-1 min-h-0">
        {/* Search */}
        <NotionFilter
          properties={FILTER_PROPERTIES}
          activeFilters={activeFilters}
          onFilterChange={setActiveFilters}
          searchTerm={searchTerm}
          onSearchChange={setSearchTerm}
          searchPlaceholder="飼主名、ペット名、カルテNo、主訴で検索..."
          count={filteredRecords.length}
          sortProperties={MEDICAL_RECORD_SORT_PROPERTIES}
          activeSorts={activeSorts}
          onSortChange={setActiveSorts}
        />

        {/* Table */}
        <FilteringIndicator isFiltering={isFiltering}>
          <DataTable
            columns={COLUMNS}
            data={paginatedData}
            emptyMessage="カルテデータが見つかりません"
            renderRow={(r) => (
              <DataTableRow
                key={r.id}
                onClick={() => handleNavigateToForm(r.id)}
              >
                <TableCell className={STYLE.tableCellMono}>{r.date}</TableCell>
                <TableCell className={STYLE.tableCell}>{r.ownerName}</TableCell>
                <TableCell className={STYLE.tableCell}>{r.petName}</TableCell>
                <TableCell className={STYLE.tableCell}>{r.species}</TableCell>
                <TableCell className={`text-sm ${C.text} max-w-[200px] truncate py-2`} title={r.chiefComplaint}>
                  {r.chiefComplaint}
                </TableCell>
                <TableCell className="py-2">
                  {r.accountingId ? (
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        navigate(`/accounting/${r.accountingId}`);
                      }}
                      className={`inline-flex items-center gap-1 text-xs font-medium px-2 py-1 rounded-[3px] border ${C.textSuccess} bg-[#10B981]/10 border-[#10B981]/30 hover:bg-[#10B981]/20 transition-colors`}
                    >
                      <Receipt className="size-3" />
                      会計
                    </button>
                  ) : (
                    <span className={`text-sm ${C.text40}`}>—</span>
                  )}
                </TableCell>
                <TableCell className={STYLE.tableCell}>
                  <div className="flex items-center gap-1">
                    <span className={!isValidStaff(r.doctor) ? "text-red-500 font-medium" : ""}>
                      {r.doctor}
                    </span>
                    {!isValidStaff(r.doctor) ? (
                      <span title="担当医が無効（退職等）に設定されています"><AlertTriangle className="size-3.5 text-red-500" /></span>
                    ) : null}
                  </div>
                </TableCell>
                <TableCell className="py-2">
                  <StatusBadge colorClass={getMedicalRecordStatusColor(r.status)}>
                    {r.status}
                  </StatusBadge>
                </TableCell>
                <TableCell className="text-right py-2">
                  <RowActionDropdown
                    actions={[
                      {
                        label: "編集",
                        icon: Edit,
                        onClick: () => handleNavigateToForm(r.id),
                      },
                      {
                        label: "削除",
                        icon: Trash2,
                        onClick: () =>
                          deleteModal.open({
                            id: r.id,
                            label: `${r.recordNo} ${r.petName}`,
                          }),
                        variant: "destructive",
                      },
                    ]}
                  />
                </TableCell>
              </DataTableRow>
            )}
          />
        </FilteringIndicator>

        {totalPages > 1 ? (
          <Pagination
            currentPage={currentPage}
            totalPages={totalPages}
            totalCount={totalCount}
            startIndex={startIndex}
            endIndex={endIndex}
            onPageChange={goToPage}
            onPrev={prevPage}
            onNext={nextPage}
          />
        ) : null}
      </div>

      <ConfirmDialog
        open={deleteModal.isOpen}
        onClose={deleteModal.close}
        onConfirm={() => {
          if (deleteModal.item) {
            deleteRecord(deleteModal.item.id);
          }
          deleteModal.close();
        }}
        title="カルテを削除しますか？"
        description={`「${deleteModal.item?.label ?? ""}」を削除します。関連する治療・検査データも削除されます。この操作は元に戻せません。`}
        confirmLabel="削除"
        variant="destructive"
      />
    </PageLayout>
  );
}
