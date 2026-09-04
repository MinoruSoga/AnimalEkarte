import { Calendar, CircleDot, CreditCard, FileText, RotateCcw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { TableCell } from "@/components/ui/table";
import {
  DataTable,
  DESIGN_TABLE_HEADER_ROW,
  DESIGN_TABLE_HEADER_CELL,
} from "@/components/shared/DataTable/DataTable";
import { LIST_TABLE_COL } from "@/components/shared/DataTable/list-table-col";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { DataTableRowLink } from "@/components/shared/DataTable/DataTableRowLink";
import { FilteringIndicator } from "@/components/shared/FilteringIndicator/FilteringIndicator";
import { PropertyFilter } from "@/components/shared/PropertyFilter/PropertyFilter";
import {
  CONDITIONS_NO_EMPTY,
  CONDITIONS_WITH_EMPTY,
} from "@/components/shared/PropertyFilter/types";
import type {
  ActiveFilter,
  ActiveSort,
  FilterProperty,
  SortProperty,
} from "@/components/shared/PropertyFilter/types";
import { Pagination } from "@/components/shared/Pagination/Pagination";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { SortableHeader } from "@/components/shared/SortableHeader/SortableHeader";
import { StatusBadge } from "@/components/shared/StatusBadge/StatusBadge";
import { paths } from "@/config/paths";
import { C, ICON } from "@/lib/design-tokens";
import { PAYMENT_METHOD_LABELS } from "@/constants/payment-method";
import { ACCOUNTING_STATUS_LABELS } from "@/constants/accounting-status";
import { formatCurrency } from "@/lib/format/number";
import { formatDate } from "@/lib/format/date";
import { getAccountingStatusColor } from "@/lib/status-helpers";
import type { Accounting as AccountingType } from "../types";
import { calculateAccountingTotal } from "../lib/accounting-list-table-model";

const FILTER_PROPERTIES: FilterProperty[] = [
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
    conditions: CONDITIONS_NO_EMPTY,
    options: Object.entries(ACCOUNTING_STATUS_LABELS).map(([value, label]) => ({ value, label })),
  },
  {
    key: "paymentMethod",
    label: "支払方法",
    type: "select",
    icon: CreditCard,
    conditions: CONDITIONS_WITH_EMPTY,
    options: Object.entries(PAYMENT_METHOD_LABELS).map(([value, label]) => ({ value, label })),
  },
];

const ACCOUNTING_SORT_PROPERTIES: SortProperty[] = [
  { key: "scheduledDate", label: "日時" },
  { key: "ownerName", label: "飼主名" },
  { key: "petName", label: "ペット名" },
  { key: "totalAmount", label: "請求金額" },
  { key: "status", label: "ステータス" },
];

interface AccountingPaginationView {
  paginatedData: AccountingType[];
  totalPages: number;
  totalCount: number;
  startIndex: number;
  endIndex: number;
  currentPage: number;
}

interface AccountingListTableProps {
  filteredCount: number;
  pagination: AccountingPaginationView;
  searchTerm: string;
  activeFilters: ActiveFilter[];
  activeSorts: ActiveSort[];
  isFiltering: boolean;
  canEdit: boolean;
  directionFor: (key: string) => "ascending" | "descending" | "none";
  onSearchChange: (value: string) => void;
  onFilterChange: (filters: ActiveFilter[]) => void;
  onSortChange: (sorts: ActiveSort[]) => void;
  onToggleSort: (key: string) => void;
  onEdit: (id: string) => void;
  onMedicalRecordOpen: (medicalRecordId: string) => void;
  onPageChange: (page: number) => void;
  /** 拠点横断表示 (#86 段階3) */
  showClinicColumn?: boolean;
  clinicNameById?: Map<string, string>;
}

export function AccountingListTable({
  filteredCount,
  pagination,
  searchTerm,
  activeFilters,
  activeSorts,
  isFiltering,
  canEdit,
  directionFor,
  onSearchChange,
  onFilterChange,
  onSortChange,
  onToggleSort,
  onEdit,
  onMedicalRecordOpen,
  onPageChange,
  showClinicColumn,
  clinicNameById,
}: AccountingListTableProps) {
  const clinicColumn = showClinicColumn ? [{ header: "拠点", className: "w-[100px]" }] : [];
  const columns = [
    ...clinicColumn,
    {
      header: (
        <SortableHeader
          label="日時"
          direction={directionFor("scheduledDate")}
          onToggle={() => onToggleSort("scheduledDate")}
        />
      ),
      className: "w-[120px]",
    },
    {
      header: (
        <SortableHeader
          label="飼主名"
          direction={directionFor("ownerName")}
          onToggle={() => onToggleSort("ownerName")}
        />
      ),
    },
    {
      header: (
        <SortableHeader
          label="ペット名"
          direction={directionFor("petName")}
          onToggle={() => onToggleSort("petName")}
        />
      ),
    },
    {
      header: (
        <SortableHeader
          label="請求金額"
          direction={directionFor("totalAmount")}
          onToggle={() => onToggleSort("totalAmount")}
        />
      ),
      align: "right" as const,
    },
    { header: "支払方法", align: "center" as const },
    {
      header: (
        <SortableHeader
          label="ステータス"
          direction={directionFor("status")}
          onToggle={() => onToggleSort("status")}
        />
      ),
      className: LIST_TABLE_COL.status,
    },
    { header: "カルテ", className: "w-[80px] hidden lg:table-cell", align: "center" as const },
    { header: "操作", className: "w-[100px]", align: "right" as const },
  ];

  return (
    <>
      <PropertyFilter
        properties={FILTER_PROPERTIES}
        activeFilters={activeFilters}
        onFilterChange={onFilterChange}
        searchTerm={searchTerm}
        onSearchChange={onSearchChange}
        searchPlaceholder="飼主名、ペット名..."
        count={filteredCount}
        sortProperties={ACCOUNTING_SORT_PROPERTIES}
        activeSorts={activeSorts}
        onSortChange={onSortChange}
      />

      <FilteringIndicator isFiltering={isFiltering}>
        <DataTable
          headerRowClassName={DESIGN_TABLE_HEADER_ROW}
          headerCellClassName={DESIGN_TABLE_HEADER_CELL}
          columns={columns}
          data={pagination.paginatedData}
          emptyMessage="会計データが見つかりません"
          renderRow={(accounting) => (
            <AccountingListRow
              key={accounting.id}
              accounting={accounting}
              canEdit={canEdit}
              showClinicColumn={showClinicColumn}
              clinicNameById={clinicNameById}
              onEdit={onEdit}
              onMedicalRecordOpen={onMedicalRecordOpen}
            />
          )}
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
    </>
  );
}

interface AccountingListRowProps {
  accounting: AccountingType;
  canEdit: boolean;
  showClinicColumn?: boolean;
  clinicNameById?: Map<string, string>;
  onEdit: (id: string) => void;
  onMedicalRecordOpen: (medicalRecordId: string) => void;
}

function AccountingListRow({
  accounting,
  canEdit,
  showClinicColumn,
  clinicNameById,
  onEdit,
  onMedicalRecordOpen,
}: AccountingListRowProps) {
  const statusLabel = ACCOUNTING_STATUS_LABELS[accounting.status] ?? accounting.status;
  const recordedTotal = calculateAccountingTotal(accounting);

  return (
    <DataTableRow key={accounting.id}>
      {showClinicColumn ? (
        <TableCell className={`text-sm ${C.text60} whitespace-nowrap`}>
          {clinicNameById?.get(accounting.clinicId) ?? accounting.clinicId}
        </TableCell>
      ) : null}
      <TableCell className={`font-mono ${C.text}`}>
        <DataTableRowLink
          to={paths.accounting.detail.getHref(accounting.id)}
          aria-label={`会計「${accounting.scheduledDate} ${accounting.ownerName} / ${accounting.petName}」(ID: ${accounting.id}) の詳細を開く`}
        >
          {formatDate(accounting.scheduledDate)}
        </DataTableRowLink>
      </TableCell>
      <TableCell className={C.text}>{accounting.ownerName}</TableCell>
      <TableCell className={C.text}>{accounting.petName}</TableCell>
      <TableCell className={`text-right font-mono font-medium ${C.text}`}>
        {formatCurrency(recordedTotal)}
      </TableCell>
      <TableCell className={`text-center ${C.text}`}>
        {accounting.payment ? PAYMENT_METHOD_LABELS[accounting.payment.method] : "-"}
      </TableCell>
      <TableCell className={LIST_TABLE_COL.status}>
        <div className="flex flex-wrap gap-1 items-center">
          <StatusBadge colorClass={getAccountingStatusColor(statusLabel)}>
            {statusLabel}
          </StatusBadge>
          {accounting.totalRefundedAmount > 0 ? (
            <span
              className={`inline-flex items-center gap-0.5 text-2xs font-medium px-1.5 py-0.5 rounded ${C.bgDiscountLight} ${C.textDiscount} ${C.borderOrangeBadge}`}
            >
              <RotateCcw className={ICON.action} />
              返金あり
            </span>
          ) : null}
        </div>
      </TableCell>
      <TableCell className="text-center hidden lg:table-cell">
        {accounting.medicalRecordId ? (
          <Button
            variant="ghost"
            size="icon"
            className={`h-11 w-11 ${C.textBrand} ${C.hoverTextBrand} ${C.hoverBgBrand5}`}
            onClick={() => onMedicalRecordOpen(accounting.medicalRecordId!)}
            aria-label={`カルテ「${accounting.ownerName} / ${accounting.petName}」(ID: ${accounting.medicalRecordId}) を開く`}
          >
            <FileText className={ICON.action} />
          </Button>
        ) : null}
      </TableCell>
      <TableCell className="text-right">
        {canEdit ? (
          <RowActionButton
            onClick={() => onEdit(accounting.id)}
            aria-label={`会計「${accounting.scheduledDate} ${accounting.ownerName} / ${accounting.petName}」(ID: ${accounting.id}) を編集`}
          />
        ) : null}
      </TableCell>
    </DataTableRow>
  );
}
