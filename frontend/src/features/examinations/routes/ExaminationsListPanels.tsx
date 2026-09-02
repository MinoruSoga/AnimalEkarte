import { useCallback, useMemo } from "react";
import { Info } from "lucide-react";
import type { ActiveFilter, ActiveSort, FilterProperty } from "@/components/shared/PropertyFilter/types";
import { TableCell } from "@/components/ui/table";
import { PropertyFilter } from "@/components/shared/PropertyFilter/PropertyFilter";
import { DataTable, DESIGN_TABLE_HEADER_ROW, DESIGN_TABLE_HEADER_CELL } from "@/components/shared/DataTable/DataTable";
import { LIST_TABLE_COL } from "@/components/shared/DataTable/list-table-col";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { DataTableRowLink } from "@/components/shared/DataTable/DataTableRowLink";
import { StatusBadge } from "@/components/shared/StatusBadge/StatusBadge";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { SortableHeader } from "@/components/shared/SortableHeader/SortableHeader";
import { Pagination } from "@/components/shared/Pagination/Pagination";
import { FilteringIndicator } from "@/components/shared/FilteringIndicator/FilteringIndicator";
import { getExaminationStatusColor } from "@/lib/status-helpers";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import type { ExaminationRecord } from "../api/transforms";
import {
  EXAMINATION_SORT_PROPERTIES,
  examinationListDetailHref,
  type ServerPagePagination,
} from "./examinations-list-model";

interface ExaminationsListRowProps {
  record: ExaminationRecord;
  canEdit: boolean;
  canViewMedicalRecords: boolean;
  onEdit: (record: ExaminationRecord) => void;
}

export function ExaminationsListRow({
  record,
  canEdit,
  canViewMedicalRecords,
  onEdit,
}: ExaminationsListRowProps) {
  return (
    <DataTableRow key={record.id}>
      <TableCell className={STYLE.tableCellMono}>{record.date}</TableCell>
      <TableCell className={STYLE.tableCell}>{record.ownerName}</TableCell>
      <TableCell className={STYLE.tableCell}>
        <DataTableRowLink
          to={examinationListDetailHref({
            id: record.id,
            medicalRecordId: canViewMedicalRecords ? record.medicalRecordId : undefined,
          })}
          aria-label={
            canViewMedicalRecords && record.medicalRecordId
              ? `カルテ検査: ${record.petName} ${record.date} ID ${record.id}`
              : `検査詳細: ${record.petName} ${record.date} ID ${record.id}`
          }
        >
          {record.petName}
        </DataTableRowLink>
      </TableCell>
      <TableCell className={`${STYLE.tableCell} font-medium hidden sm:table-cell`}>{record.testType}</TableCell>
      <TableCell className={`${C.text60} truncate max-w-[200px] hidden lg:table-cell`}>
        {record.resultSummary || "-"}
      </TableCell>
      <TableCell className={`${STYLE.tableCell} hidden md:table-cell`}>{record.doctor}</TableCell>
      <TableCell className={LIST_TABLE_COL.status}>
        <StatusBadge colorClass={getExaminationStatusColor(record.status)}>
          {record.status}
        </StatusBadge>
      </TableCell>
      <TableCell className="text-right">
        {canEdit ? (
          <RowActionButton
            onClick={() => onEdit(record)}
            aria-label={`検査操作: ${record.petName} ${record.date} ID ${record.id}`}
          />
        ) : null}
      </TableCell>
    </DataTableRow>
  );
}

interface ExaminationsListContentProps {
  filterProperties: FilterProperty[];
  activeFilters: ActiveFilter[];
  onFilterChange: (next: ActiveFilter[]) => void;
  searchTerm: string;
  onSearchChange: (value: string) => void;
  count: number | undefined;
  activeSorts: ActiveSort[];
  onSortChange: (sorts: ActiveSort[]) => void;
  directionFor: (key: string) => "ascending" | "descending" | "none";
  toggleSort: (key: string) => void;
  isFiltering: boolean;
  hasPageScopedFilter: boolean;
  pagination: ServerPagePagination<ExaminationRecord>;
  canEdit: boolean;
  canViewMedicalRecords: boolean;
  onEdit: (record: ExaminationRecord) => void;
  onPageChange: (page: number) => void;
}

export function ExaminationsListContent({
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
  hasPageScopedFilter,
  pagination,
  canEdit,
  canViewMedicalRecords,
  onEdit,
  onPageChange,
}: ExaminationsListContentProps) {
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
      className: "hidden sm:table-cell",
    },
    { header: "結果概要", className: "hidden lg:table-cell" },
    {
      header: (
        <SortableHeader
          label="担当医"
          direction={directionFor("doctor")}
          onToggle={() => toggleSort("doctor")}
        />
      ),
      className: "w-[100px] hidden md:table-cell",
    },
    {
      header: (
        <SortableHeader
          label="ステータス"
          direction={directionFor("status")}
          onToggle={() => toggleSort("status")}
        />
      ),
      className: LIST_TABLE_COL.status,
    },
    { header: "操作", className: "w-[80px]", align: "right" as const },
  ], [directionFor, toggleSort]);

  const renderRow = useCallback((r: ExaminationRecord) => (
    <ExaminationsListRow
      record={r}
      canEdit={canEdit}
      canViewMedicalRecords={canViewMedicalRecords}
      onEdit={onEdit}
    />
  ), [canEdit, canViewMedicalRecords, onEdit]);

  return (
    <div className="flex flex-col gap-4">
      <PropertyFilter
        properties={filterProperties}
        activeFilters={activeFilters}
        onFilterChange={onFilterChange}
        searchTerm={searchTerm}
        onSearchChange={onSearchChange}
        searchPlaceholder="飼主名、ペット名、検査種別..."
        count={count}
        sortProperties={EXAMINATION_SORT_PROPERTIES}
        activeSorts={activeSorts}
        onSortChange={onSortChange}
      />

      {hasPageScopedFilter && pagination.totalPages > 1 ? (
        <div className={`flex items-start gap-2 rounded-md px-3 py-2 text-sm ${C.textWarning}`}>
          <Info className={`${ICON.action} ${C.textWarningIcon} shrink-0 mt-0.5`} />
          <span>
            検索・ステータス・検査種別・担当医での絞り込みは現在表示中のページ内のみが対象です。
            他のページにある検査記録はこの検索では見つかりません。全期間から探す場合は日付での絞り込みをご利用ください。
          </span>
        </div>
      ) : null}

      <FilteringIndicator isFiltering={isFiltering}>
        <DataTable
          headerRowClassName={DESIGN_TABLE_HEADER_ROW}
          headerCellClassName={DESIGN_TABLE_HEADER_CELL}
          columns={columns}
          data={pagination.paginatedData}
          emptyMessage="検査データが見つかりません"
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
