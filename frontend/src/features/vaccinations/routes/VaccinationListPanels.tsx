import { useCallback, useMemo } from "react";
import { AlertTriangle, Pencil, Trash2 } from "lucide-react";
import type { ActiveFilter, ActiveSort, FilterProperty } from "@/components/shared/PropertyFilter/types";
import { TableCell } from "@/components/ui/table";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { PropertyFilter } from "@/components/shared/PropertyFilter/PropertyFilter";
import { DataTable, DESIGN_TABLE_HEADER_ROW, DESIGN_TABLE_HEADER_CELL } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { DataTableRowLink } from "@/components/shared/DataTable/DataTableRowLink";
import { RowActionDropdown } from "@/components/shared/RowActionDropdown/RowActionDropdown";
import { SortableHeader } from "@/components/shared/SortableHeader/SortableHeader";
import { Pagination } from "@/components/shared/Pagination/Pagination";
import { FilteringIndicator } from "@/components/shared/FilteringIndicator/FilteringIndicator";
import { useGetPet } from "@/hooks/use-pet";
import { isPastJSTDate } from "@/lib/jst-date";
import { C, ICON } from "@/lib/design-tokens";
import type { VaccinationRecord } from "@/types";
import {
  VACCINATION_SORT_PROPERTIES,
  vaccinationListDetailHref,
} from "./vaccinations-list-model";

interface VaccinationRowActionsProps {
  record: VaccinationRecord;
  canEdit: boolean;
  canDelete: boolean;
  onEdit: (record: VaccinationRecord) => void;
  onDelete: (id: string) => void;
}

export function VaccinationRowActions({
  record,
  canEdit,
  canDelete,
  onEdit,
  onDelete,
}: VaccinationRowActionsProps) {
  const { data: pet } = useGetPet(record.petId ?? "");
  const actions = [
    ...(canEdit && pet?.status === "生存" ? [{
      label: "編集",
      icon: Pencil,
      onClick: () => onEdit(record),
    }] : []),
    ...(canDelete ? [{
      label: "削除",
      icon: Trash2,
      onClick: () => onDelete(record.id),
      variant: "destructive" as const,
    }] : []),
  ];

  return actions.length > 0 ? (
    <RowActionDropdown
      actions={actions}
      ariaLabel={`予防接種操作: ${record.petName} ${record.date} ID ${record.id}`}
    />
  ) : null;
}

interface VaccinationListRowProps {
  record: VaccinationRecord;
  canEdit: boolean;
  canDelete: boolean;
  onEdit: (record: VaccinationRecord) => void;
  onDelete: (id: string) => void;
}

export function VaccinationListRow({
  record,
  canEdit,
  canDelete,
  onEdit,
  onDelete,
}: VaccinationListRowProps) {
  const overdue = isPastJSTDate(record.nextDate);
  return (
    <DataTableRow key={record.id}>
      <TableCell className={`font-mono ${C.text}`}>{record.date}</TableCell>
      <TableCell className={C.text}>{record.ownerName}</TableCell>
      <TableCell className={C.text}>
        <DataTableRowLink
          to={vaccinationListDetailHref({ id: record.id, medicalRecordId: record.medicalRecordId })}
          aria-label={`予防接種詳細: ${record.petName} ${record.date} ID ${record.id}`}
        >
          {record.petName}
        </DataTableRowLink>
      </TableCell>
      <TableCell className={`font-medium ${C.text}`}>{record.vaccineName}</TableCell>
      <TableCell className={`font-mono ${overdue ? C.danger : C.text}`}>
        {overdue ? (
          <span className="inline-flex items-center gap-1.5">
            <AlertTriangle className={`${ICON.xs} shrink-0`} />
            <span>
              {record.nextDate}
              <span className="ml-1.5 text-xs font-medium">（期限超過）</span>
            </span>
          </span>
        ) : record.nextDate}
      </TableCell>
      <TableCell className="text-right">
        {canEdit || canDelete ? (
          <VaccinationRowActions
            record={record}
            canEdit={canEdit}
            canDelete={canDelete}
            onEdit={onEdit}
            onDelete={onDelete}
          />
        ) : null}
      </TableCell>
    </DataTableRow>
  );
}

interface VaccinationPaginationView {
  paginatedData: VaccinationRecord[];
  totalPages: number;
  totalCount: number;
  startIndex: number;
  endIndex: number;
  currentPage: number;
}

interface VaccinationListContentProps {
  filterProperties: FilterProperty[];
  activeFilters: ActiveFilter[];
  onFilterChange: (next: ActiveFilter[]) => void;
  searchTerm: string;
  onSearchChange: (value: string) => void;
  count: number;
  activeSorts: ActiveSort[];
  onSortChange: (sorts: ActiveSort[]) => void;
  directionFor: (key: string) => "ascending" | "descending" | "none";
  toggleSort: (key: string) => void;
  isFiltering: boolean;
  pagination: VaccinationPaginationView;
  canEdit: boolean;
  canDelete: boolean;
  onEdit: (record: VaccinationRecord) => void;
  onDelete: (id: string) => void;
  onPageChange: (page: number) => void;
}

export function VaccinationListContent({
  filterProperties,
  activeFilters,
  onFilterChange,
  searchTerm,
  onSearchChange,
  count,
  activeSorts,
  onSortChange,
  directionFor,
  toggleSort,
  isFiltering,
  pagination,
  canEdit,
  canDelete,
  onEdit,
  onDelete,
  onPageChange,
}: VaccinationListContentProps) {
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

  const renderRow = useCallback((r: VaccinationRecord) => (
    <VaccinationListRow
      record={r}
      canEdit={canEdit}
      canDelete={canDelete}
      onEdit={onEdit}
      onDelete={onDelete}
    />
  ), [canEdit, canDelete, onEdit, onDelete]);

  return (
    <div className="flex flex-col gap-4">
      <PropertyFilter
        properties={filterProperties}
        activeFilters={activeFilters}
        onFilterChange={onFilterChange}
        searchTerm={searchTerm}
        onSearchChange={onSearchChange}
        searchPlaceholder="飼主名、ペット名、予防接種名..."
        count={count}
        sortProperties={VACCINATION_SORT_PROPERTIES}
        activeSorts={activeSorts}
        onSortChange={onSortChange}
      />

      <FilteringIndicator isFiltering={isFiltering}>
        <DataTable
          headerRowClassName={DESIGN_TABLE_HEADER_ROW}
          headerCellClassName={DESIGN_TABLE_HEADER_CELL}
          columns={columns}
          data={pagination.paginatedData}
          emptyMessage="データが見つかりません"
          renderRow={renderRow}
        />
      </FilteringIndicator>

      {pagination.totalPages > 1 ? (
        <Pagination
          currentPage={pagination.currentPage}
          totalPages={pagination.totalPages}
          totalCount={pagination.totalCount}
          startIndex={pagination.startIndex}
          endIndex={pagination.endIndex}
          onPageChange={onPageChange}
          onPrev={() => onPageChange(pagination.currentPage - 1)}
          onNext={() => onPageChange(pagination.currentPage + 1)}
        />
      ) : null}
    </div>
  );
}

export function VaccinationListDeleteDialog({
  open,
  isPending,
  onClose,
  onConfirm,
}: {
  open: boolean;
  isPending: boolean;
  onClose: () => void;
  onConfirm: () => void;
}) {
  return (
    <ConfirmDialog
      open={open}
      onClose={onClose}
      title="予防接種記録を削除しますか？"
      description="この操作は取り消せません。"
      confirmLabel="削除"
      variant="destructive"
      onConfirm={onConfirm}
      isPending={isPending}
    />
  );
}
