import { Calendar, CircleDot, CreditCard, FileText, RotateCcw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { TableCell } from "@/components/ui/table";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { FilteringIndicator } from "@/components/shared/FilteringIndicator/FilteringIndicator";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { CONDITIONS_NO_EMPTY, CONDITIONS_WITH_EMPTY } from "@/components/shared/NotionFilter/types";
import type {
  ActiveFilter,
  ActiveSort,
  FilterProperty,
  SortProperty,
} from "@/components/shared/NotionFilter/types";
import { Pagination } from "@/components/shared/Pagination/Pagination";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { SortableHeader } from "@/components/shared/SortableHeader/SortableHeader";
import { StatusBadge } from "@/components/shared/StatusBadge/StatusBadge";
import { C, ICON } from "@/lib/design-tokens";
import { formatCurrency } from "@/utils/format/number";
import { getAccountingStatusColor } from "@/utils/status-helpers";
import type { Accounting as AccountingType, AccountingStatus, PaymentMethod } from "../types";

const PAYMENT_METHOD_LABELS: Record<PaymentMethod, string> = {
  cash: "現金",
  credit_card: "クレジットカード",
  electronic_money: "電子マネー",
};

const ACCOUNTING_STATUS_LABELS: Record<AccountingStatus, string> = {
  waiting: "会計待ち",
  pending: "会計待ち",
  completed: "会計済",
  cancelled: "キャンセル",
};

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
    options: [
      { value: "waiting", label: "会計待ち" },
      { value: "completed", label: "会計済" },
      { value: "cancelled", label: "キャンセル" },
    ],
  },
  {
    key: "paymentMethod",
    label: "支払方法",
    type: "select",
    icon: CreditCard,
    conditions: CONDITIONS_WITH_EMPTY,
    options: [
      { value: "cash", label: "現金" },
      { value: "credit_card", label: "クレジットカード" },
      { value: "electronic_money", label: "電子マネー" },
    ],
  },
];

const ACCOUNTING_SORT_PROPERTIES: SortProperty[] = [
  { key: "scheduledDate", label: "日時" },
  { key: "ownerName", label: "飼主名" },
  { key: "petName", label: "ペット名" },
  { key: "totalAmount", label: "請求金額" },
  { key: "status", label: "ステータス" },
];

export function calculateAccountingTotal(accounting: AccountingType) {
  if (accounting.payment) return accounting.payment.totalAmount;

  return accounting.items.reduce((sum: number, item) => {
    const price = item.unitPrice * item.quantity;
    const tax = Math.floor(price * item.taxRate);
    return sum + price + tax;
  }, 0);
}

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
}: AccountingListTableProps) {
  const columns = [
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
      className: "w-[100px]",
    },
    { header: "カルテ", className: "w-[80px] hidden lg:table-cell", align: "center" as const },
    { header: "操作", className: "w-[100px]", align: "right" as const },
  ];

  return (
    <>
      <NotionFilter
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
          columns={columns}
          data={pagination.paginatedData}
          emptyMessage="会計データが見つかりません"
          renderRow={(accounting) => (
            <AccountingListRow
              key={accounting.id}
              accounting={accounting}
              canEdit={canEdit}
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
  onEdit: (id: string) => void;
  onMedicalRecordOpen: (medicalRecordId: string) => void;
}

function AccountingListRow({
  accounting,
  canEdit,
  onEdit,
  onMedicalRecordOpen,
}: AccountingListRowProps) {
  const statusLabel = ACCOUNTING_STATUS_LABELS[accounting.status] ?? accounting.status;

  return (
    <DataTableRow key={accounting.id} onClick={() => onEdit(accounting.id)}>
      <TableCell className={`font-mono text-base ${C.text} py-2`}>{accounting.scheduledDate}</TableCell>
      <TableCell className={`text-base ${C.text} py-2`}>{accounting.ownerName}</TableCell>
      <TableCell className={`text-base ${C.text} py-2`}>{accounting.petName}</TableCell>
      <TableCell className={`text-right font-mono font-medium text-base ${C.text} py-2`}>
        {formatCurrency(calculateAccountingTotal(accounting))}
      </TableCell>
      <TableCell className={`text-center text-base ${C.text} py-2`}>
        {accounting.payment ? PAYMENT_METHOD_LABELS[accounting.payment.method] : "-"}
      </TableCell>
      <TableCell className="py-2">
        <div className="flex flex-wrap gap-1 items-center">
          <StatusBadge colorClass={getAccountingStatusColor(statusLabel)}>
            {statusLabel}
          </StatusBadge>
          {accounting.totalRefundedAmount > 0 ? (
            <span className={`inline-flex items-center gap-0.5 text-[10px] font-medium px-1.5 py-0.5 rounded ${C.bgDiscountLight} ${C.textDiscount} ${C.borderOrangeBadge}`}>
              <RotateCcw className={ICON.action} />
              返金あり
            </span>
          ) : null}
        </div>
      </TableCell>
      <TableCell className="text-center py-2 hidden lg:table-cell">
        {accounting.medicalRecordId ? (
          <Button
            variant="ghost"
            size="icon"
            className={`h-8 w-8 ${C.accent} ${C.hoverTextAccent} ${C.hoverBgAccent5}`}
            onClick={(event) => {
              event.stopPropagation();
              onMedicalRecordOpen(accounting.medicalRecordId!);
            }}
            aria-label="カルテを開く"
          >
            <FileText className={ICON.action} />
          </Button>
        ) : null}
      </TableCell>
      <TableCell className="text-right py-2">
        {canEdit ? <RowActionButton onClick={() => onEdit(accounting.id)} /> : null}
      </TableCell>
    </DataTableRow>
  );
}
