// React/Framework
import { type ReactNode, useState, useCallback, useDeferredValue, useMemo, useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router";

// Hooks
import { useSortableData } from "@/hooks/use-sortable-data";
import { useModalState } from "@/hooks/use-modal-state";

// External
import { Plus, FileText, Edit, Trash2, Receipt, AlertTriangle, Calendar, CircleDot, User, PawPrint } from "lucide-react";

// Internal
import { paths } from "@/config/paths";
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
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { Pagination } from "@/components/shared/Pagination";
import { FilteringIndicator } from "@/components/shared/FilteringIndicator/FilteringIndicator";
import { C, STYLE, ICON } from "@/lib/design-tokens";
import { getMedicalRecordStatusColor } from "@/utils/status-helpers";
import { usePagination } from "@/hooks/use-pagination";
import { useStaffValidation } from "@/hooks/use-staff-validation";

// Relative
import { useFilterMedicalRecords } from "../hooks/use-medical-records";
import { useDeleteMedicalRecord } from "../api/delete-medical-record";
import { usePermission } from "@/features/auth";

// Types
import type {
  FilterProperty,
  ActiveFilter,
  SortProperty,
} from "@/components/shared/NotionFilter/types";
import { CONDITIONS_NO_EMPTY, CONDITIONS_WITH_EMPTY } from "@/components/shared/NotionFilter/types";
import type { MedicalRecordFilters } from "../api/get-medical-records";
import { ResourceMedicalRecords } from "@/types/generated/models";

// rendering-hoist-jsx: 静的フィルタプロパティ（担当医・種は動的オプションのためコンポーネント内で構築）
const STATIC_FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "date",
    label: "診療日",
    type: "date-range",
    icon: Calendar,
  },
  {
    key: "status",
    label: "ステータス",
    type: "select",
    icon: CircleDot,
    // medical_records.status DEFAULT 'draft' — 空値は存在しない
    conditions: CONDITIONS_NO_EMPTY,
    options: [
      { value: "作成中", label: "作成中" },
      { value: "確定済", label: "確定済" },
    ],
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
  const [searchParams, setSearchParams] = useSearchParams();
  const { canCreate, canEdit, canDelete } = usePermission("medical-records");
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

  const { data: filteredRecords, allRecords, isLoading, isError } = useFilterMedicalRecords(deferredSearch, apiFilters, activeFilters);

  // js-cache-function-results: ロード済みレコードから担当医・種の選択肢を動的生成
  const filterProperties = useMemo<FilterProperty[]>(() => {
    const doctorOptions = Array.from(new Set(allRecords.map((r) => r.doctor).filter(Boolean)))
      .sort()
      .map((d) => ({ value: d, label: d }));
    const speciesOptions = Array.from(new Set(allRecords.map((r) => r.species).filter(Boolean)))
      .sort()
      .map((s) => ({ value: s, label: s }));
    return [
      ...STATIC_FILTER_PROPERTIES,
      // medical_records.doctor_id nullable（未割当あり）
      { key: "doctor", label: "担当医", type: "select" as const, icon: User, conditions: CONDITIONS_WITH_EMPTY, options: doctorOptions },
      // pets.animal_species_id NOT NULL — 空値は存在しない
      { key: "species", label: "種", type: "select" as const, icon: PawPrint, conditions: CONDITIONS_NO_EMPTY, options: speciesOptions },
    ];
  }, [allRecords]);
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
  } = usePagination(sortedData, { pageSize: 20, resetKey: searchTerm });

  // FE-144: URLクエリパラメータからページ番号を読み取る
  const urlPage = Number(searchParams.get("page") ?? 1);

  // FE-144: URLのページ番号とローカル状態を同期（URLが変わったときのみ）
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

  const isFiltering = searchTerm !== deferredSearch;

  const handleNavigateToForm = useCallback((recordId?: string) => {
    navigate(
      recordId ? `/medical-records/${recordId}` : "/medical-records/select-pet",
      { state: { from: "/medical-records" } },
    );
  }, [navigate]);

  // js-cache-function-results: directionFor/toggleSort に依存するカラム定義をメモ化
  const COLUMNS = useMemo<{ header: ReactNode; className?: string; align?: "left" | "center" | "right" }[]>(() => [
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
      className: "w-[80px] hidden lg:table-cell",
    },
    { header: "主訴" },
    { header: "関連", className: "w-[100px] hidden lg:table-cell" },
    {
      header: <SortableHeader label="担当医" direction={directionFor("doctor")} onToggle={() => toggleSort("doctor")} />,
      className: "w-[100px]",
    },
    {
      header: <SortableHeader label="ステータス" direction={directionFor("status")} onToggle={() => toggleSort("status")} />,
      className: "w-[100px]",
    },
    { header: "操作", className: "w-[100px]", align: "right" as const },
  ], [directionFor, toggleSort]);

  if (isLoading) return <LoadingFallback />;
  if (isError) return <ErrorFallback />;

  return (
    <PageLayout
      title="カルテ管理"
      icon={<FileText className={`${ICON.page} ${C.text}`} />}
      resource={ResourceMedicalRecords}
      headerAction={
        canCreate ? (
          <PrimaryButton onClick={() => handleNavigateToForm()}>
            <Plus className={ICON.action} />
            新規カルテ登録
          </PrimaryButton>
        ) : null
      }
      maxWidth="max-w-full"
    >
      <div className="flex flex-col gap-4 flex-1 min-h-0">
        {/* Search */}
        <NotionFilter
          properties={filterProperties}
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
                <TableCell className={`${STYLE.tableCell} hidden lg:table-cell`}>{r.species}</TableCell>
                <TableCell className={`${STYLE.tableCell} max-w-[200px] truncate`} title={r.chiefComplaint}>
                  {r.chiefComplaint}
                </TableCell>
                <TableCell className="py-2.5 hidden lg:table-cell">
                  {r.accountingId ? (
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        navigate(paths.accounting.detail.getHref(r.accountingId ?? ""));
                      }}
                      className={`inline-flex items-center gap-1 text-xs font-medium px-2 py-1 rounded-[3px] border ${C.textSuccess} ${C.bgSuccess10} ${C.borderSuccess30} ${C.hoverBgSuccess20} transition-colors`}
                    >
                      <Receipt className={ICON.xs} />
                      会計
                    </button>
                  ) : (
                    <span className={`text-sm ${C.text40}`}>—</span>
                  )}
                </TableCell>
                <TableCell className={STYLE.tableCell}>
                  <div className="flex items-center gap-1">
                    <span className={!isValidStaff(r.doctor) ? `${C.danger} font-medium` : ""}>
                      {r.doctor}
                    </span>
                    {!isValidStaff(r.doctor) ? (
                      <span title="担当医が無効（退職等）に設定されています"><AlertTriangle className={`${ICON.xs} ${C.danger}`} /></span>
                    ) : null}
                  </div>
                </TableCell>
                <TableCell className="py-2.5">
                  <StatusBadge colorClass={getMedicalRecordStatusColor(r.status)}>
                    {r.status}
                  </StatusBadge>
                </TableCell>
                <TableCell className="text-right py-2.5">
                  {(canEdit || canDelete) ? (
                    <RowActionDropdown
                      actions={[
                        ...(canEdit ? [{
                          label: "編集",
                          icon: Edit,
                          onClick: () => handleNavigateToForm(r.id),
                        }] : []),
                        ...(canDelete ? [{
                          label: "削除",
                          icon: Trash2,
                          onClick: () =>
                            deleteModal.open({
                              id: r.id,
                              label: `${r.recordNo} ${r.petName}`,
                            }),
                          variant: "destructive" as const,
                        }] : []),
                      ]}
                    />
                  ) : null}
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
            onPageChange={handlePageChange}
            onPrev={() => handlePageChange(currentPage - 1)}
            onNext={() => handlePageChange(currentPage + 1)}
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
