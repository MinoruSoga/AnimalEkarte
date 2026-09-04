import { Edit, Trash2, Receipt, AlertTriangle, Plus, FileText } from "lucide-react";
import type { ReactNode } from "react";
import type { ActiveFilter, FilterProperty } from "@/components/shared/PropertyFilter/types";
import { TableCell } from "@/components/ui/table";
import { PropertyFilter } from "@/components/shared/PropertyFilter/PropertyFilter";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { LIST_TABLE_COL } from "@/components/shared/DataTable/list-table-col";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { DataTableRowLink } from "@/components/shared/DataTable/DataTableRowLink";
import { StatusBadge } from "@/components/shared/StatusBadge/StatusBadge";
import { RowActionDropdown } from "@/components/shared/RowActionDropdown";
import { Pagination } from "@/components/shared/Pagination";
import { FilteringIndicator } from "@/components/shared/FilteringIndicator/FilteringIndicator";
import { ClinicScopeFilter } from "@/components/shared/ClinicScopeFilter/ClinicScopeFilter";
import { C, STYLE, ICON, LAYOUT } from "@/lib/design-tokens";
import { getMedicalRecordStatusColor } from "@/lib/status-helpers";
import { paths } from "@/config/paths";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { ResourceMedicalRecords } from "@/types/generated/models";
import type { ClinicMembership } from "@/types/auth";
import type { MedicalRecord } from "../api/transforms";
import {
  MEDICAL_RECORDS_HEADER_CELL,
  MEDICAL_RECORDS_HEADER_ROW,
} from "./medical-records-list-model";

interface MedicalRecordsSpeciesStatusProps {
  isSpeciesError: boolean;
  isSpeciesLoading: boolean;
  hasSpecies: boolean;
}

function MedicalRecordsSpeciesStatus({
  isSpeciesError,
  isSpeciesLoading,
  hasSpecies,
}: MedicalRecordsSpeciesStatusProps) {
  if (isSpeciesError) {
    return (
      <p role="alert" aria-atomic="true" className={`text-sm ${C.danger}`}>
        動物種の取得に失敗しました。
      </p>
    );
  }
  if (isSpeciesLoading) {
    return (
      <p role="status" aria-live="polite" aria-atomic="true" className={`text-sm ${C.text50}`}>
        動物種を読み込み中です。
      </p>
    );
  }
  if (!hasSpecies) {
    return (
      <p role="status" aria-live="polite" aria-atomic="true" className={`text-sm ${C.text50}`}>
        動物種マスタが登録されていません。
      </p>
    );
  }
  return null;
}

interface MedicalRecordsListRowProps {
  record: MedicalRecord;
  showClinicColumn: boolean;
  currentClinicId: string | undefined;
  clinicNameById: Map<string, string>;
  canViewAccounting: boolean;
  canEdit: boolean;
  canDelete: boolean;
  isValidStaff: (name: string) => boolean;
  onAccountingOpen: (accountingId: string) => void;
  onEdit: (recordId: string) => void;
  onDeleteRequest: (input: { id: string; label: string; petIsDeceased: boolean }) => void;
}

function MedicalRecordsListRow({
  record,
  showClinicColumn,
  currentClinicId,
  clinicNameById,
  canViewAccounting,
  canEdit,
  canDelete,
  isValidStaff,
  onAccountingOpen,
  onEdit,
  onDeleteRequest,
}: MedicalRecordsListRowProps) {
  const isOtherClinic = showClinicColumn && record.clinicId !== currentClinicId;
  const accountingId = record.accountingId;
  return (
    <DataTableRow key={record.id}>
      <TableCell className={STYLE.tableCellMono}>{record.date}</TableCell>
      <TableCell className={STYLE.tableCell}>{record.ownerName}</TableCell>
      <TableCell className={STYLE.tableCell}>
        {isOtherClinic ? (
          record.petName
        ) : (
          <DataTableRowLink
            to={paths.medicalRecords.detail.getHref(record.id)}
            state={{ from: paths.medicalRecords.getHref() }}
            aria-label={`カルテ詳細: ${record.petName} ${record.date} ID ${record.id}`}
          >
            {record.petName}
          </DataTableRowLink>
        )}
      </TableCell>
      <TableCell className={`${STYLE.tableCell} hidden lg:table-cell`}>{record.species}</TableCell>
      <TableCell
        className={`${STYLE.tableCell} max-w-[200px] truncate hidden md:table-cell`}
        title={record.chiefComplaint}
      >
        {record.chiefComplaint}
      </TableCell>
      <TableCell className="hidden lg:table-cell">
        {!isOtherClinic && canViewAccounting && accountingId ? (
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation();
              onAccountingOpen(accountingId);
            }}
            className={`inline-flex min-h-11 min-w-11 items-center gap-1 rounded-xxs border px-2 text-xs font-medium ${C.textSuccess} ${C.bgSuccess10} ${C.borderSuccess30} ${C.hoverBgSuccess20} transition-colors`}
            aria-label={`会計詳細: ${record.petName} ${record.date} カルテID ${record.id}`}
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
          <span className={!isValidStaff(record.doctor) ? `${C.danger} font-medium` : ""}>
            {record.doctor}
          </span>
          {!isValidStaff(record.doctor) ? (
            <span
              role="img"
              aria-label={`無効な担当医: ${record.doctor}（退職等）`}
              title="担当医が無効（退職等）に設定されています"
            >
              <AlertTriangle className={`${ICON.xs} ${C.danger}`} aria-hidden="true" />
            </span>
          ) : null}
        </div>
      </TableCell>
      <TableCell className={LIST_TABLE_COL.status}>
        <StatusBadge colorClass={getMedicalRecordStatusColor(record.status)}>
          {record.status}
        </StatusBadge>
      </TableCell>
      {showClinicColumn ? (
        <TableCell
          className={`${STYLE.tableCell} hidden lg:table-cell`}
          data-testid="mr-row-clinic"
        >
          <span className={isOtherClinic ? `text-xs ${C.text40}` : "text-xs"}>
            {record.clinicId ? (clinicNameById.get(record.clinicId) ?? record.clinicId) : "—"}
          </span>
        </TableCell>
      ) : null}
      <TableCell className="text-right">
        {(canEdit || canDelete) && !isOtherClinic && !record.petIsDeceased ? (
          <RowActionDropdown
            ariaLabel={`カルテ操作: ${record.petName} ${record.date} ID ${record.id}`}
            actions={[
              ...(canEdit
                ? [
                    {
                      label: "編集",
                      icon: Edit,
                      onClick: () => {
                        if (record.petIsDeceased) return;
                        onEdit(record.id);
                      },
                    },
                  ]
                : []),
              ...(canDelete
                ? [
                    {
                      label: "削除",
                      icon: Trash2,
                      onClick: () =>
                        onDeleteRequest({
                          id: record.id,
                          label: `${record.recordNo} ${record.petName}`,
                          petIsDeceased: record.petIsDeceased,
                        }),
                      variant: "destructive" as const,
                    },
                  ]
                : []),
            ]}
          />
        ) : null}
      </TableCell>
    </DataTableRow>
  );
}

interface MedicalRecordsListContentProps {
  assignedClinics: ClinicMembership[];
  selectedClinicIds: string[];
  onToggleClinic: (clinicId: string) => void;
  isSpeciesError: boolean;
  isSpeciesLoading: boolean;
  hasSpecies: boolean;
  filterProperties: FilterProperty[];
  activeFilters: ActiveFilter[];
  onFilterChange: (next: ActiveFilter[]) => void;
  searchTerm: string;
  onSearchChange: (value: string) => void;
  total: number;
  isFiltering: boolean;
  columns: { header: ReactNode; className?: string; align?: "left" | "center" | "right" }[];
  records: MedicalRecord[];
  showClinicColumn: boolean;
  currentClinicId: string | undefined;
  clinicNameById: Map<string, string>;
  canViewAccounting: boolean;
  canEdit: boolean;
  canDelete: boolean;
  isValidStaff: (name: string) => boolean;
  onAccountingOpen: (accountingId: string) => void;
  onEdit: (recordId: string) => void;
  onDeleteRequest: (input: { id: string; label: string; petIsDeceased: boolean }) => void;
  totalPages: number;
  currentPage: number;
  startIndex: number;
  endIndex: number;
  onPageChange: (page: number) => void;
}

function MedicalRecordsListContent({
  assignedClinics,
  selectedClinicIds,
  onToggleClinic,
  isSpeciesError,
  isSpeciesLoading,
  hasSpecies,
  filterProperties,
  activeFilters,
  onFilterChange,
  searchTerm,
  onSearchChange,
  total,
  isFiltering,
  columns,
  records,
  showClinicColumn,
  currentClinicId,
  clinicNameById,
  canViewAccounting,
  canEdit,
  canDelete,
  isValidStaff,
  onAccountingOpen,
  onEdit,
  onDeleteRequest,
  totalPages,
  currentPage,
  startIndex,
  endIndex,
  onPageChange,
}: MedicalRecordsListContentProps) {
  return (
    <div className="flex flex-col gap-4 flex-1 min-h-0">
      <ClinicScopeFilter
        clinics={assignedClinics}
        selectedIds={selectedClinicIds}
        onToggle={onToggleClinic}
      />

      <MedicalRecordsSpeciesStatus
        isSpeciesError={isSpeciesError}
        isSpeciesLoading={isSpeciesLoading}
        hasSpecies={hasSpecies}
      />

      <PropertyFilter
        properties={filterProperties}
        activeFilters={activeFilters}
        onFilterChange={onFilterChange}
        searchTerm={searchTerm}
        onSearchChange={onSearchChange}
        searchPlaceholder="飼主名、ペット名、カルテNo、主訴、治療内容・メモ、処置・薬剤・診察・在庫品名で検索..."
        count={total}
      />

      <FilteringIndicator isFiltering={isFiltering}>
        <DataTable
          columns={columns}
          data={records}
          emptyMessage="カルテデータが見つかりません"
          headerRowClassName={MEDICAL_RECORDS_HEADER_ROW}
          headerCellClassName={MEDICAL_RECORDS_HEADER_CELL}
          renderRow={(r) => (
            <MedicalRecordsListRow
              record={r}
              showClinicColumn={showClinicColumn}
              currentClinicId={currentClinicId}
              clinicNameById={clinicNameById}
              canViewAccounting={canViewAccounting}
              canEdit={canEdit}
              canDelete={canDelete}
              isValidStaff={isValidStaff}
              onAccountingOpen={onAccountingOpen}
              onEdit={onEdit}
              onDeleteRequest={onDeleteRequest}
            />
          )}
        />
      </FilteringIndicator>

      {totalPages > 1 ? (
        <Pagination
          currentPage={currentPage}
          totalPages={totalPages}
          totalCount={total}
          startIndex={startIndex}
          endIndex={endIndex}
          onPageChange={onPageChange}
          onPrev={() => onPageChange(currentPage - 1)}
          onNext={() => onPageChange(currentPage + 1)}
        />
      ) : null}
    </div>
  );
}

interface MedicalRecordsListPanelsProps extends MedicalRecordsListContentProps {
  canCreate: boolean;
  onCreate: () => void;
  deleteOpen: boolean;
  deleteLabel: string;
  onDeleteClose: () => void;
  onDeleteConfirm: () => void;
}

export function MedicalRecordsListPanels({
  canCreate,
  onCreate,
  deleteOpen,
  deleteLabel,
  onDeleteClose,
  onDeleteConfirm,
  ...contentProps
}: MedicalRecordsListPanelsProps) {
  return (
    <PageLayout
      title="カルテ管理"
      icon={<FileText className={`${ICON.page} ${C.text}`} />}
      resource={ResourceMedicalRecords}
      headerAction={
        canCreate ? (
          <PrimaryButton onClick={onCreate}>
            <Plus className={ICON.action} />
            新規カルテ登録
          </PrimaryButton>
        ) : null
      }
      maxWidth={LAYOUT.pageContentMaxWidth.full}
    >
      <MedicalRecordsListContent {...contentProps} />
      <ConfirmDialog
        open={deleteOpen}
        onClose={onDeleteClose}
        onConfirm={onDeleteConfirm}
        title="カルテを削除しますか？"
        description={`「${deleteLabel}」を削除します。関連する治療・検査データも削除されます。この操作は元に戻せません。`}
        confirmLabel="削除"
        variant="destructive"
      />
    </PageLayout>
  );
}
