import { useCallback, useMemo } from "react";
import { AlertTriangle, Info } from "lucide-react";
import type { ActiveFilter, ActiveSort } from "@/components/shared/PropertyFilter/types";
import { TableCell } from "@/components/ui/table";
import { PropertyFilter } from "@/components/shared/PropertyFilter/PropertyFilter";
import { DataTable, DESIGN_TABLE_HEADER_ROW, DESIGN_TABLE_HEADER_CELL } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { DataTableRowLink } from "@/components/shared/DataTable/DataTableRowLink";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { StatusBadge } from "@/components/shared/StatusBadge/StatusBadge";
import { SortableHeader } from "@/components/shared/SortableHeader/SortableHeader";
import { Pagination } from "@/components/shared/Pagination/Pagination";
import { FilteringIndicator } from "@/components/shared/FilteringIndicator/FilteringIndicator";
import { getInventoryStatusColor, getInventoryStatusLabel } from "@/lib/status-helpers";
import { C, ICON } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import type { InventoryItem } from "@/types";
import {
  CATEGORY_LABELS,
  INVENTORY_FILTER_PROPERTIES,
  INVENTORY_SORT_PROPERTIES,
  type ServerPagePagination,
} from "./inventory-list-model";

interface InventoryListRowProps {
  item: InventoryItem;
  canEdit: boolean;
  onEdit: (id: string) => void;
}

export function InventoryListRow({ item, canEdit, onEdit }: InventoryListRowProps) {
  return (
    <DataTableRow key={item.id}>
      <TableCell className={`font-medium ${C.text}`}>
        {canEdit ? (
          <DataTableRowLink
            to={paths.inventory.detail.getHref(item.id)}
            aria-label={`在庫品「${item.name}」(ID: ${item.id}) の詳細を開く`}
          >
            {item.name}
          </DataTableRowLink>
        ) : item.name}
      </TableCell>
      <TableCell className={C.text}>
        {CATEGORY_LABELS[item.category]}
      </TableCell>
      <TableCell className={`${C.text} text-right font-mono`}>
        {item.quantity} {item.unit}
      </TableCell>
      <TableCell className={`${C.text60} text-right font-mono hidden lg:table-cell`}>
        {item.minStockLevel} {item.unit}
      </TableCell>
      <TableCell className={C.text}>
        {item.location ?? "-"}
      </TableCell>
      <TableCell className={`${C.text} font-mono hidden lg:table-cell`}>
        {item.expiryDate ?? "-"}
      </TableCell>
      <TableCell>
        <StatusBadge colorClass={getInventoryStatusColor(item.status)}>
          {getInventoryStatusLabel(item.status)}
        </StatusBadge>
      </TableCell>
      <TableCell className="text-right">
        {canEdit ? (
          <RowActionButton
            onClick={() => onEdit(item.id)}
            aria-label={`在庫品「${item.name}」(ID: ${item.id}) を編集`}
          />
        ) : null}
      </TableCell>
    </DataTableRow>
  );
}

interface InventorySummary {
  lowStock: number;
  outOfStock: number;
  isError: boolean;
}

interface InventoryListContentProps {
  summary: InventorySummary;
  activeFilters: ActiveFilter[];
  searchTerm: string;
  onSearchChange: (value: string) => void;
  onFilterChange: (next: ActiveFilter[]) => void;
  count: number;
  activeSorts: ActiveSort[];
  onSortChange: (sorts: ActiveSort[]) => void;
  directionFor: (key: string) => "ascending" | "descending" | "none";
  toggleSort: (key: string) => void;
  isFiltering: boolean;
  hasPageScopedFilter: boolean;
  pagination: ServerPagePagination<InventoryItem>;
  canEdit: boolean;
  onEdit: (id: string) => void;
  onPageChange: (page: number) => void;
}

export function InventoryListContent({
  summary,
  activeFilters,
  searchTerm,
  onSearchChange,
  onFilterChange,
  count,
  activeSorts,
  onSortChange,
  directionFor,
  toggleSort,
  isFiltering,
  hasPageScopedFilter,
  pagination,
  canEdit,
  onEdit,
  onPageChange,
}: InventoryListContentProps) {
  const columns = useMemo(() => [
    {
      header: (
        <SortableHeader
          label="品名"
          direction={directionFor("name")}
          onToggle={() => toggleSort("name")}
        />
      ),
      className: "min-w-[200px]",
    },
    {
      header: (
        <SortableHeader
          label="カテゴリ"
          direction={directionFor("category")}
          onToggle={() => toggleSort("category")}
        />
      ),
      className: "w-[100px]",
    },
    {
      header: (
        <SortableHeader
          label="在庫数"
          direction={directionFor("quantity")}
          onToggle={() => toggleSort("quantity")}
        />
      ),
      className: "w-[100px]",
      align: "right" as const,
    },
    { header: "最低在庫", className: "w-[100px] hidden lg:table-cell", align: "right" as const },
    { header: "保管場所", className: "w-[120px]" },
    { header: "有効期限", className: "w-[120px] hidden lg:table-cell" },
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
    { header: "操作", className: "w-[80px]", align: "right" as const },
  ], [directionFor, toggleSort]);

  const renderRow = useCallback((item: InventoryItem) => (
    <InventoryListRow item={item} canEdit={canEdit} onEdit={onEdit} />
  ), [canEdit, onEdit]);

  return (
    <div className="flex flex-col gap-4">
      {summary.isError ? (
        <div
          role="status"
          className={`flex items-center gap-4 p-3 ${C.bgWarning50} ${C.borderWarning20} border rounded-lg`}
        >
          <AlertTriangle aria-hidden="true" className={`${ICON.page} ${C.textWarningIcon}`} />
          <span className={`text-base ${C.textWarning} font-medium`}>
            低在庫・欠品件数を取得できませんでした。時間をおいて再度お試しください。
          </span>
        </div>
      ) : (summary.lowStock > 0 || summary.outOfStock > 0) ? (
        <div className={`flex items-center gap-4 p-3 ${C.bgWarning50} ${C.borderWarning20} border rounded-lg`}>
          <AlertTriangle className={`${ICON.page} ${C.textWarningIcon}`} />
          <div className="flex gap-4 text-base">
            {summary.outOfStock > 0 ? (
              <span className={`${C.danger} font-medium`}>
                在庫切れ: {summary.outOfStock}件
              </span>
            ) : null}
            {summary.lowStock > 0 ? (
              <span className={`${C.textWarning} font-medium`}>
                残少: {summary.lowStock}件
              </span>
            ) : null}
          </div>
        </div>
      ) : null}

      <PropertyFilter
        properties={INVENTORY_FILTER_PROPERTIES}
        activeFilters={activeFilters}
        onFilterChange={onFilterChange}
        searchTerm={searchTerm}
        onSearchChange={onSearchChange}
        searchPlaceholder="品名、保管場所、仕入先..."
        count={count}
        sortProperties={INVENTORY_SORT_PROPERTIES}
        activeSorts={activeSorts}
        onSortChange={onSortChange}
      />

      {hasPageScopedFilter && pagination.totalPages > 1 ? (
        <div
          role="status"
          className={`flex items-start gap-2 rounded-md px-3 py-2 ${C.bgWarning50} ${C.borderWarning20} border text-sm`}
        >
          <Info aria-hidden="true" className={`${ICON.action} ${C.textWarningIcon} shrink-0 mt-0.5`} />
          <span>
            検索・並び替えは現在表示中のページ内のみが対象です。他のページにある在庫はこの検索・並び替えでは見つかりません。
            全件を確認する場合はカテゴリ・ステータスでの絞り込みをご利用ください。
          </span>
        </div>
      ) : null}

      <FilteringIndicator isFiltering={isFiltering}>
        <DataTable
          headerRowClassName={DESIGN_TABLE_HEADER_ROW}
          headerCellClassName={DESIGN_TABLE_HEADER_CELL}
          columns={columns}
          data={pagination.paginatedData}
          emptyMessage="在庫データが見つかりません"
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
