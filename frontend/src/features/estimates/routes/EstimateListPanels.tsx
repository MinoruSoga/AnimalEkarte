import { useCallback } from "react";
import { FileText, Trash2, ExternalLink } from "lucide-react";
import type { ActiveFilter, ActiveSort } from "@/components/shared/PropertyFilter/types";
import { TableCell } from "@/components/ui/table";
import { PropertyFilter } from "@/components/shared/PropertyFilter/PropertyFilter";
import { DataTable, DESIGN_TABLE_HEADER_ROW, DESIGN_TABLE_HEADER_CELL } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { DataTableRowLink } from "@/components/shared/DataTable/DataTableRowLink";
import { Pagination } from "@/components/shared/Pagination/Pagination";
import { RowActionDropdown } from "@/components/shared/RowActionDropdown";
import { EstimateStatusBadge } from "../components/EstimateStatusBadge/EstimateStatusBadge";
import { paths } from "@/config/paths";
import { formatCurrency } from "@/lib/format/number";
import { formatDate } from "@/lib/format/date";
import { C } from "@/lib/design-tokens";
import { isEstimateLockedStatus } from "../lib/is-estimate-locked-status";
import type { Estimate } from "../types";
import { COLUMNS, FILTER_PROPERTIES, SORT_PROPERTIES } from "./estimate-list-model";

interface EstimateListRowProps {
  estimate: Estimate;
  canEdit: boolean;
  canDelete: boolean;
  onOpenDetail: (id: string) => void;
  onOpenEdit: (id: string) => void;
  onDelete: (id: string) => void;
}

function EstimateListRow({
  estimate,
  canEdit,
  canDelete,
  onOpenDetail,
  onOpenEdit,
  onDelete,
}: EstimateListRowProps) {
  const isLocked = isEstimateLockedStatus(estimate.status);
  const showEdit = canEdit && !isLocked;
  const showDelete = canDelete && !isLocked;

  return (
    <DataTableRow key={estimate.id}>
      <TableCell className={`font-mono ${C.text60}`}>
        <DataTableRowLink
          to={paths.estimates.detail.getHref(estimate.id)}
          aria-label={`見積書「${estimate.estimateNo} / ${estimate.title}」(ID: ${estimate.id}) の詳細を開く`}
        >
          {estimate.estimateNo}
        </DataTableRowLink>
      </TableCell>
      <TableCell className={`${C.text} font-medium`}>{estimate.title}</TableCell>
      <TableCell className={C.text}>
        {estimate.ownerName ?? "-"}
        {estimate.petName ? (
          <span className={`block text-xs ${C.text50}`}>{estimate.petName}</span>
        ) : null}
      </TableCell>
      <TableCell className={C.text60}>
        {formatDate(estimate.validUntil)}
      </TableCell>
      <TableCell className={`text-right font-mono font-medium ${C.text}`}>
        {formatCurrency(estimate.totalAmount)}
      </TableCell>
      <TableCell>
        <EstimateStatusBadge status={estimate.status} />
      </TableCell>
      <TableCell className="text-right">
        {(canEdit || canDelete) ? (
          <RowActionDropdown
            ariaLabel={`見積書「${estimate.estimateNo} / ${estimate.title}」(ID: ${estimate.id}) の操作`}
            actions={[
              {
                label: "詳細",
                icon: ExternalLink,
                onClick: () => onOpenDetail(estimate.id),
              },
              ...(showEdit ? [{
                label: "編集",
                icon: FileText,
                onClick: () => onOpenEdit(estimate.id),
              }] : []),
              ...(showDelete ? [{
                label: "削除",
                icon: Trash2,
                variant: "destructive" as const,
                onClick: () => onDelete(estimate.id),
              }] : []),
            ]}
          />
        ) : null}
      </TableCell>
    </DataTableRow>
  );
}

interface EstimatePaginationView {
  paginatedData: Estimate[];
  currentPage: number;
  totalPages: number;
  totalCount: number;
  startIndex: number;
  endIndex: number;
  goToPage: (page: number) => void;
  prevPage: () => void;
  nextPage: () => void;
}

interface EstimateListContentProps {
  searchTerm: string;
  onSearchChange: (value: string) => void;
  activeFilters: ActiveFilter[];
  onFilterChange: (next: ActiveFilter[]) => void;
  activeSorts: ActiveSort[];
  onSortChange: (sorts: ActiveSort[]) => void;
  filteredCount: number;
  pagination: EstimatePaginationView;
  canEdit: boolean;
  canDelete: boolean;
  onOpenDetail: (id: string) => void;
  onOpenEdit: (id: string) => void;
  onDelete: (id: string) => void;
}

export function EstimateListContent({
  searchTerm,
  onSearchChange,
  activeFilters,
  onFilterChange,
  activeSorts,
  onSortChange,
  filteredCount,
  pagination,
  canEdit,
  canDelete,
  onOpenDetail,
  onOpenEdit,
  onDelete,
}: EstimateListContentProps) {
  const renderRow = useCallback((estimate: Estimate) => (
    <EstimateListRow
      estimate={estimate}
      canEdit={canEdit}
      canDelete={canDelete}
      onOpenDetail={onOpenDetail}
      onOpenEdit={onOpenEdit}
      onDelete={onDelete}
    />
  ), [canEdit, canDelete, onOpenDetail, onOpenEdit, onDelete]);

  return (
    <div className="flex flex-col gap-4">
      <PropertyFilter
        properties={FILTER_PROPERTIES}
        activeFilters={activeFilters}
        onFilterChange={onFilterChange}
        searchTerm={searchTerm}
        onSearchChange={onSearchChange}
        searchPlaceholder="見積番号、タイトル、飼主名..."
        count={filteredCount}
        sortProperties={SORT_PROPERTIES}
        activeSorts={activeSorts}
        onSortChange={onSortChange}
      />

      <DataTable
        headerRowClassName={DESIGN_TABLE_HEADER_ROW}
        headerCellClassName={DESIGN_TABLE_HEADER_CELL}
        columns={COLUMNS}
        data={pagination.paginatedData}
        emptyMessage="見積書が見つかりません"
        renderRow={renderRow}
      />

      {filteredCount > 0 ? (
        <Pagination
          currentPage={pagination.currentPage}
          totalPages={pagination.totalPages}
          totalCount={pagination.totalCount}
          startIndex={pagination.startIndex}
          endIndex={pagination.endIndex}
          onPageChange={pagination.goToPage}
          onPrev={pagination.prevPage}
          onNext={pagination.nextPage}
        />
      ) : null}
    </div>
  );
}
