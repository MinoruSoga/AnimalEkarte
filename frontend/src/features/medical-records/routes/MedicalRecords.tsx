// React/Framework
import {
  useState,
  useCallback,
  useDeferredValue,
  useMemo,
  useEffect,
  useLayoutEffect,
  useRef,
} from "react";
import { useNavigate, useSearchParams } from "react-router";

// Auth
import { useClinicScope } from "@/hooks/use-clinic-scope";

// Hooks
import { useModalState } from "@/hooks/use-modal-state";
import { useGetStaffs } from "@/hooks/use-staffs";

// External
import { Plus, FileText, Edit, Trash2, Receipt, AlertTriangle, Calendar, CircleDot, User, PawPrint } from "lucide-react";

// Internal
import { paths } from "@/config/paths";
import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PropertyFilter } from "@/components/shared/PropertyFilter/PropertyFilter";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { LIST_TABLE_COL } from "@/components/shared/DataTable/list-table-col";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { DataTableRowLink } from "@/components/shared/DataTable/DataTableRowLink";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { StatusBadge } from "@/components/shared/StatusBadge/StatusBadge";
import { RowActionDropdown } from "@/components/shared/RowActionDropdown";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { Pagination } from "@/components/shared/Pagination";
import { FilteringIndicator } from "@/components/shared/FilteringIndicator/FilteringIndicator";
import { ClinicScopeFilter } from "@/components/shared/ClinicScopeFilter/ClinicScopeFilter";
import { C, STYLE, ICON, LAYOUT } from "@/lib/design-tokens";
import { getMedicalRecordStatusColor } from "@/lib/status-helpers";
import { useStaffValidation } from "@/hooks/use-staff-validation";

// Relative
import { useMedicalRecordsList } from "../hooks/use-medical-records";
import { useMedicalRecordsUrlState } from "../hooks/use-medical-records-url-state";
import { useDeleteMedicalRecord } from "../api/delete-medical-record";
import { usePermission } from "@/hooks/use-permission";
import { useAnimalSpecies } from "@/hooks/use-animal-species";
import { useMedicalRecordsColumns } from "./medical-records-columns";

// Types
import type {
  FilterProperty,
  ActiveFilter,
  FilterCondition,
} from "@/components/shared/PropertyFilter/types";
import { ResourceAccounting, ResourceMedicalRecords } from "@/types/generated/models";

const PAGE_SIZE = 20;

// BE は status/doctor_id/animal_species_id を単一値の完全一致でのみ受け付けるため、
// これらのフィルタは "is" 条件のみ UI で選択可能にする（is_not/is_empty は server 未対応）。
// 詳細: bug.md B-1 follow-up
const SERVER_EQUALITY_ONLY: FilterCondition[] = ["is"];

// rendering-hoist-jsx: 静的フィルタプロパティ（担当医・種は master 取得後に動的追加）
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
    conditions: SERVER_EQUALITY_ONLY,
    options: [
      { value: "作成中", label: "作成中" },
      { value: "確定済", label: "確定済" },
    ],
  },
];

const CLINIC_TOGGLE_RESET_PARAMS = ["page"] as const;

// DESIGN.md ex-data-table-cell: header は canvas-soft 背景 + eyebrow 相当タイポグラフィ（12px/600/tracking）。
// STYLE.tableHeaderRow/tableHeaderCell（他画面と共有）を直接変更すると影響範囲が広がるため、
// DataTable の headerRowClassName/headerCellClassName で本テーブルのみ上書きする。
const MEDICAL_RECORDS_HEADER_ROW = `border-b ${C.borderLight} ${C.bgPage} h-11`;
const MEDICAL_RECORDS_HEADER_CELL = `${STYLE.sectionLabel} h-11`;

export function MedicalRecords() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("pet_id") || undefined;
  const { canCreate, canEdit, canDelete } = usePermission(ResourceMedicalRecords);
  const { canView: canViewAccounting } = usePermission(ResourceAccounting);
  const {
    assignedClinics,
    selectedClinicIds,
    isMultiClinic,
    clinicNameById,
    currentClinicId,
    handleToggleClinic,
  } = useClinicScope({ resetParamsOnToggle: CLINIC_TOGGLE_RESET_PARAMS });
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const deferredSearch = useDeferredValue(searchTerm);

  const { data: staffs } = useGetStaffs();
  const {
    activeSpecies,
    isLoading: isSpeciesLoading,
    isError: isSpeciesError,
  } = useAnimalSpecies();

  // js-cache-function-results: staff/species master から担当医・種の選択肢を動的生成（allRecords 非依存）
  const filterProperties = useMemo<FilterProperty[]>(() => {
    const doctorOptions = (staffs ?? [])
      .filter((s) => s.isActive)
      .map((s) => ({ value: s.id, label: s.name }));
    const speciesOptions =
      isSpeciesError || isSpeciesLoading
        ? []
        : activeSpecies.map((s) => ({ value: String(s.id), label: s.name }));
    return [
      ...STATIC_FILTER_PROPERTIES,
      { key: "doctor", label: "担当医", type: "select" as const, icon: User, conditions: SERVER_EQUALITY_ONLY, options: doctorOptions },
      { key: "species", label: "種", type: "select" as const, icon: PawPrint, conditions: SERVER_EQUALITY_ONLY, options: speciesOptions },
    ];
  }, [staffs, activeSpecies, isSpeciesError, isSpeciesLoading]);

  // rerender-derived-state-no-effect: 検索/フィルタが変わったら1ページ目へリセット（useEffect不使用）
  const resetKey = `${deferredSearch}|${JSON.stringify(activeFilters)}|${petId ?? ""}`;
  const {
    currentPage,
    sortKey,
    sortOrder,
    handleSortToggle,
    directionForSort,
    handlePageChange,
  } = useMedicalRecordsUrlState(resetKey);

  const clinicIdsForApi = isMultiClinic ? selectedClinicIds : undefined;
  const { records, total, isLoading, isError } = useMedicalRecordsList({
    searchTerm: deferredSearch,
    activeFilters,
    clinicIds: clinicIdsForApi,
    petId,
    page: currentPage,
    limit: PAGE_SIZE,
    sort: sortKey,
    order: sortKey ? sortOrder : undefined,
  });

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  // BUG-B1: server-side フィルタで件数が減り currentPage が範囲外になった場合、最終ページへ補正
  useEffect(() => {
    if (total > 0 && currentPage > totalPages) {
      handlePageChange(totalPages);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [totalPages, total]);

  const deleteModal = useModalState<{
    id: string;
    label: string;
    petIsDeceased: boolean;
  }>();
  const { mutate: deleteRecord } = useDeleteMedicalRecord();
  const { isValidStaff } = useStaffValidation();
  const canDeleteRef = useRef(canDelete);
  const recordsByIdRef = useRef(
    new Map(records.map((record) => [record.id, record])),
  );

  useLayoutEffect(() => {
    canDeleteRef.current = canDelete;
  }, [canDelete]);
  useLayoutEffect(() => {
    recordsByIdRef.current = new Map(
      records.map((record) => [record.id, record]),
    );
  }, [records]);

  const startIndex = total === 0 ? 0 : (currentPage - 1) * PAGE_SIZE + 1;
  const endIndex = Math.min(currentPage * PAGE_SIZE, total);

  const isFiltering = searchTerm !== deferredSearch;

  const handleNavigateToForm = useCallback((recordId?: string) => {
    navigate(
      recordId ? paths.medicalRecords.detail.getHref(recordId) : paths.medicalRecords.selectPet.getHref(),
      { state: { from: paths.medicalRecords.getHref() } },
    );
  }, [navigate]);

  const showClinicColumn = isMultiClinic;

  const COLUMNS = useMedicalRecordsColumns({
    showClinicColumn,
    directionForSort,
    onSortToggle: handleSortToggle,
  });

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
      maxWidth={LAYOUT.pageContentMaxWidth.full}
    >
      <div className="flex flex-col gap-4 flex-1 min-h-0">
        {/* #86: 拠点横断フィルター — 複数所属医院がある場合のみ表示 */}
        <ClinicScopeFilter
          clinics={assignedClinics}
          selectedIds={selectedClinicIds}
          onToggle={handleToggleClinic}
        />

        {isSpeciesError ? (
          <p
            role="alert"
            aria-atomic="true"
            className={`text-sm ${C.danger}`}
          >
            動物種の取得に失敗しました。
          </p>
        ) : isSpeciesLoading ? (
          <p
            role="status"
            aria-live="polite"
            aria-atomic="true"
            className={`text-sm ${C.text50}`}
          >
            動物種を読み込み中です。
          </p>
        ) : activeSpecies.length === 0 ? (
          <p
            role="status"
            aria-live="polite"
            aria-atomic="true"
            className={`text-sm ${C.text50}`}
          >
            動物種マスタが登録されていません。
          </p>
        ) : null}

        {/* Search */}
        <PropertyFilter
          properties={filterProperties}
          activeFilters={activeFilters}
          onFilterChange={setActiveFilters}
          searchTerm={searchTerm}
          onSearchChange={setSearchTerm}
          searchPlaceholder="飼主名、ペット名、カルテNo、主訴、治療内容・メモ、処置・薬剤・診察・在庫品名で検索..."
          count={total}
        />

        {/* Table */}
        <FilteringIndicator isFiltering={isFiltering}>
          <DataTable
            columns={COLUMNS}
            data={records}
            emptyMessage="カルテデータが見つかりません"
            headerRowClassName={MEDICAL_RECORDS_HEADER_ROW}
            headerCellClassName={MEDICAL_RECORDS_HEADER_CELL}
            renderRow={(r) => {
              const isOtherClinic = showClinicColumn && r.clinicId !== currentClinicId;
              const accountingId = r.accountingId;
              return (
                <DataTableRow key={r.id}>
                <TableCell className={STYLE.tableCellMono}>{r.date}</TableCell>
                <TableCell className={STYLE.tableCell}>{r.ownerName}</TableCell>
                <TableCell className={STYLE.tableCell}>
                  {isOtherClinic ? r.petName : (
                    <DataTableRowLink
                      to={paths.medicalRecords.detail.getHref(r.id)}
                      state={{ from: paths.medicalRecords.getHref() }}
                      aria-label={`カルテ詳細: ${r.petName} ${r.date} ID ${r.id}`}
                    >
                      {r.petName}
                    </DataTableRowLink>
                  )}
                </TableCell>
                <TableCell className={`${STYLE.tableCell} hidden lg:table-cell`}>{r.species}</TableCell>
                <TableCell className={`${STYLE.tableCell} max-w-[200px] truncate hidden md:table-cell`} title={r.chiefComplaint}>
                  {r.chiefComplaint}
                </TableCell>
                <TableCell className="hidden lg:table-cell">
                  {!isOtherClinic && canViewAccounting && accountingId ? (
                    <button type="button"
                      onClick={(e) => {
                        e.stopPropagation();
                        navigate(paths.accounting.detail.getHref(accountingId));
                      }}
                      className={`inline-flex min-h-11 min-w-11 items-center gap-1 rounded-xxs border px-2 text-xs font-medium ${C.textSuccess} ${C.bgSuccess10} ${C.borderSuccess30} ${C.hoverBgSuccess20} transition-colors`}
                      aria-label={`会計詳細: ${r.petName} ${r.date} カルテID ${r.id}`}
                    >
                      <Receipt className={ICON.xs} aria-hidden="true" />
                      会計
                    </button>
                  ) : (
                    <span className={`text-sm ${C.text40}`}>—</span>
                  )}
                </TableCell>
                <TableCell className={`${STYLE.tableCell} hidden md:table-cell`}>
                  <div className="flex items-center gap-1">
                    <span className={!isValidStaff(r.doctor) ? `${C.danger} font-medium` : ""}>
                      {r.doctor}
                    </span>
                    {!isValidStaff(r.doctor) ? (
                      <span
                        role="img"
                        aria-label={`無効な担当医: ${r.doctor}（退職等）`}
                        title="担当医が無効（退職等）に設定されています"
                      >
                        <AlertTriangle className={`${ICON.xs} ${C.danger}`} aria-hidden="true" />
                      </span>
                    ) : null}
                  </div>
                </TableCell>
                <TableCell className={LIST_TABLE_COL.status}>
                  <StatusBadge colorClass={getMedicalRecordStatusColor(r.status)}>
                    {r.status}
                  </StatusBadge>
                </TableCell>
                {showClinicColumn ? (
                  <TableCell className={`${STYLE.tableCell} hidden lg:table-cell`} data-testid="mr-row-clinic">
                    <span className={isOtherClinic ? `text-xs ${C.text40}` : "text-xs"}>
                      {r.clinicId ? (clinicNameById.get(r.clinicId) ?? r.clinicId) : "—"}
                    </span>
                  </TableCell>
                ) : null}
                <TableCell className="text-right">
                  {(canEdit || canDelete) && !isOtherClinic && !r.petIsDeceased ? (
                    <RowActionDropdown
                      ariaLabel={`カルテ操作: ${r.petName} ${r.date} ID ${r.id}`}
                      actions={[
                        ...(canEdit ? [{
                          label: "編集",
                          icon: Edit,
                          onClick: () => {
                            if (r.petIsDeceased) return;
                            handleNavigateToForm(r.id);
                          },
                        }] : []),
                        ...(canDelete ? [{
                          label: "削除",
                          icon: Trash2,
                          onClick: () =>
                            deleteModal.open({
                              id: r.id,
                              label: `${r.recordNo} ${r.petName}`,
                              petIsDeceased: r.petIsDeceased,
                            }),
                          variant: "destructive" as const,
                        }] : []),
                      ]}
                    />
                  ) : null}
                </TableCell>
              </DataTableRow>
            );
            }}
          />
        </FilteringIndicator>

        {totalPages > 1 ? (
          <Pagination
            currentPage={currentPage}
            totalPages={totalPages}
            totalCount={total}
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
          const item = deleteModal.item;
          const currentRecord = item
            ? recordsByIdRef.current.get(item.id)
            : undefined;
          if (
            canDeleteRef.current === true
            && item?.petIsDeceased === false
            && currentRecord?.petIsDeceased === false
          ) {
            deleteRecord(item.id);
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
