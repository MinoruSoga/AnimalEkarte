import { useCallback, useMemo } from "react";
import type { ActiveFilter, ActiveSort } from "@/components/shared/PropertyFilter/types";
import { TableCell } from "@/components/ui/table";
import { CheckupAlertBadge } from "@/components/shared/CheckupAlertBadge/CheckupAlertBadge";
import { PropertyFilter } from "@/components/shared/PropertyFilter/PropertyFilter";
import { DataTable, DESIGN_TABLE_HEADER_ROW, DESIGN_TABLE_HEADER_CELL } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { DataTableRowLink } from "@/components/shared/DataTable/DataTableRowLink";
import { SortableHeader } from "@/components/shared/SortableHeader/SortableHeader";
import { FilteringIndicator } from "@/components/shared/FilteringIndicator/FilteringIndicator";
import { Pagination } from "@/components/shared/Pagination/Pagination";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { C } from "@/lib/design-tokens";
import { formatDate } from "@/lib/format/date";
import type { CheckupRecord } from "../api/transforms";
import {
  CHECKUPS_SORT_PROPERTIES,
  FILTER_PROPERTIES,
  checkupChartHref,
} from "./checkups-list-model";

interface CheckupsListRowProps {
  record: CheckupRecord;
  canView: boolean;
  canEdit: boolean;
  onEdit: (medicalRecordId: string, checkupId: string) => void;
}

function CheckupsListRow({ record, canView, canEdit, onEdit }: CheckupsListRowProps) {
  return (
    <DataTableRow key={record.id}>
      <TableCell className={`font-mono ${C.text}`}>
        {record.date ? formatDate(record.date) : "-"}
      </TableCell>
      <TableCell className={C.text}>{record.ownerName || "-"}</TableCell>
      <TableCell className={C.text}>
        {canView && record.medicalRecordId ? (
          <DataTableRowLink
            to={checkupChartHref(record.medicalRecordId, record.id)}
            aria-label={`カルテ詳細: ${record.petName || "-"} ${record.date || "-"} 健診ID ${record.id}`}
          >
            {record.petName || "-"}
          </DataTableRowLink>
        ) : (record.petName || "-")}
      </TableCell>
      <TableCell className={`${C.text} hidden md:table-cell`}>{record.checkupTypeName || "-"}</TableCell>
      <TableCell className={`font-mono ${C.text} hidden md:table-cell`}>
        <div className="flex items-center gap-1.5">
          {record.nextDate ? formatDate(record.nextDate) : "-"}
          <CheckupAlertBadge nextDate={record.nextDate} />
        </div>
      </TableCell>
      <TableCell className={`${C.text} max-w-xs truncate hidden lg:table-cell`}>
        {record.result || "-"}
      </TableCell>
      <TableCell className={`${C.text} hidden md:table-cell`}>{record.doctorName || "-"}</TableCell>
      <TableCell className="text-right">
        {canView && canEdit ? (
          <RowActionButton
            onClick={() => onEdit(record.medicalRecordId, record.id)}
            aria-label={`健診操作: ${record.petName || "-"} ${record.date || "-"} ID ${record.id}`}
          />
        ) : null}
      </TableCell>
    </DataTableRow>
  );
}

interface CheckupsListContentProps {
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
  records: CheckupRecord[];
  canView: boolean;
  canEdit: boolean;
  onEdit: (medicalRecordId: string, checkupId: string) => void;
  totalPages: number;
  safePage: number;
  total: number;
  startIndex: number;
  endIndex: number;
  onPageChange: (page: number) => void;
}

export function CheckupsListContent({
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
  records,
  canView,
  canEdit,
  onEdit,
  totalPages,
  safePage,
  total,
  startIndex,
  endIndex,
  onPageChange,
}: CheckupsListContentProps) {
  const columns = useMemo(
    () => [
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
            label="健診種別"
            direction={directionFor("checkupTypeName")}
            onToggle={() => toggleSort("checkupTypeName")}
          />
        ),
        className: "hidden md:table-cell",
      },
      {
        header: (
          <SortableHeader
            label="次回予定"
            direction={directionFor("nextDate")}
            onToggle={() => toggleSort("nextDate")}
          />
        ),
        className: "w-[120px] hidden md:table-cell",
      },
      { header: "結果・所見", className: "hidden lg:table-cell" },
      { header: "担当医", className: "w-[100px] hidden md:table-cell" },
      { header: "操作", className: "w-[80px]", align: "right" as const },
    ],
    [directionFor, toggleSort],
  );

  const renderRow = useCallback((c: CheckupRecord) => (
    <CheckupsListRow record={c} canView={canView} canEdit={canEdit} onEdit={onEdit} />
  ), [canView, canEdit, onEdit]);

  return (
    <div className="flex flex-col gap-4">
      <PropertyFilter
        properties={FILTER_PROPERTIES}
        activeFilters={activeFilters}
        onFilterChange={onFilterChange}
        searchTerm={searchTerm}
        onSearchChange={onSearchChange}
        searchPlaceholder="ペット名・飼主名・種別で検索..."
        count={count}
        sortProperties={CHECKUPS_SORT_PROPERTIES}
        activeSorts={activeSorts}
        onSortChange={onSortChange}
      />

      <FilteringIndicator isFiltering={isFiltering}>
        <DataTable
          headerRowClassName={DESIGN_TABLE_HEADER_ROW}
          headerCellClassName={DESIGN_TABLE_HEADER_CELL}
          columns={columns}
          data={records}
          emptyMessage="定期健診の記録がありません"
          renderRow={renderRow}
        />
      </FilteringIndicator>

      {totalPages > 1 ? (
        <Pagination
          currentPage={safePage}
          totalPages={totalPages}
          totalCount={total}
          startIndex={startIndex}
          endIndex={endIndex}
          onPageChange={onPageChange}
          onPrev={() => onPageChange(safePage - 1)}
          onNext={() => onPageChange(safePage + 1)}
        />
      ) : null}
    </div>
  );
}
