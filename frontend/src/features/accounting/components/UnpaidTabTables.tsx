import { Pagination } from "@/components/shared/Pagination/Pagination";
import { DataTableRowLink } from "@/components/shared/DataTable/DataTableRowLink";
import { EmptyState } from "@/components/shared/DataStates";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { C, STYLE } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import { formatCurrency } from "@/lib/format/number";
import { formatDate } from "@/lib/format/date";
import { daysSince } from "@/lib/jst-date";

import type { UnpaidOwner, MonthlyUnpaidResponse } from "../api/get-unpaid-billings";
import type { Accounting } from "../api/transforms";
import { unpaidBillingAmount } from "./unpaid-tab-model";

interface UnpaidOwnerTableProps {
  rows: UnpaidOwner[];
  endDate: string;
}

export function UnpaidOwnerTable({ rows, endDate }: UnpaidOwnerTableProps) {
  if (rows.length === 0) {
    return <p className={`text-sm ${C.text50} py-8 text-center`}>未納者はいません</p>;
  }

  return (
    <div className={`rounded-lg border ${C.borderLight} ${C.bgWhite} overflow-hidden`}>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>飼主名</TableHead>
            <TableHead className="text-right">件数</TableHead>
            <TableHead className="text-right">未納額合計</TableHead>
            <TableHead>最古未納日</TableHead>
            <TableHead>最新未納日</TableHead>
            <TableHead className="text-right">経過日数</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((row) => (
            <TableRow key={row.owner_id} className={STYLE.tableRowHover}>
              <TableCell className="font-medium">
                <DataTableRowLink
                  to={paths.owners.detail.getHref(String(row.owner_id))}
                  aria-label={`飼主詳細: ${row.owner_name} (ID ${row.owner_id})`}
                >
                  {row.owner_name}
                </DataTableRowLink>
              </TableCell>
              <TableCell className="text-right">{row.count}</TableCell>
              <TableCell className="text-right font-mono">
                {formatCurrency(row.total_amount)}
              </TableCell>
              <TableCell>{row.oldest_scheduled}</TableCell>
              <TableCell>{row.latest_scheduled}</TableCell>
              <TableCell className="text-right">
                {daysSince(row.oldest_scheduled, endDate)}日
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

interface UnpaidBillingTableProps {
  billings: Accounting[];
  endDate: string;
}

export function UnpaidBillingTable({ billings, endDate }: UnpaidBillingTableProps) {
  if (billings.length === 0) {
    return <EmptyState message="未納会計はありません" />;
  }

  return (
    <div className={`rounded-lg border ${C.borderLight} ${C.bgWhite} overflow-hidden`}>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>飼主名</TableHead>
            <TableHead>ペット名</TableHead>
            <TableHead>診療日</TableHead>
            <TableHead className="text-right">未納額</TableHead>
            <TableHead className="text-right">経過日数</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {billings.map((billing) => {
            const unpaidAmount = unpaidBillingAmount(billing);
            return (
              <TableRow key={billing.id} className={STYLE.tableRowHover}>
                <TableCell className="font-medium">
                  <DataTableRowLink
                    to={paths.accounting.detail.getHref(billing.id)}
                    aria-label={`会計詳細: ${billing.ownerName} / ${billing.petName} (ID ${billing.id})`}
                  >
                    {billing.ownerName}
                  </DataTableRowLink>
                </TableCell>
                <TableCell>{billing.petName}</TableCell>
                <TableCell>{formatDate(billing.scheduledDate)}</TableCell>
                <TableCell className="text-right font-mono">
                  {formatCurrency(unpaidAmount)}
                </TableCell>
                <TableCell className="text-right">
                  {billing.scheduledDate ? `${daysSince(billing.scheduledDate, endDate)}日` : "-"}
                </TableCell>
              </TableRow>
            );
          })}
        </TableBody>
      </Table>
    </div>
  );
}

interface UnpaidMonthlyTableProps {
  rows: MonthlyUnpaidResponse["data"];
}

export function UnpaidMonthlyTable({ rows }: UnpaidMonthlyTableProps) {
  if (rows.length === 0) {
    return <EmptyState message="対象月の未納データがありません" />;
  }

  return (
    <div className={`rounded-lg border ${C.borderLight} ${C.bgWhite} overflow-hidden`}>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>飼主名</TableHead>
            <TableHead>ペット名</TableHead>
            <TableHead className="text-right">前月繰越</TableHead>
            <TableHead className="text-right">当月未払い</TableHead>
            <TableHead className="text-right">次月繰越</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((row) => (
            <TableRow key={`${row.owner_id}-${row.pet_id ?? "none"}`} className={STYLE.tableRowHover}>
              <TableCell className="font-medium">
                <DataTableRowLink
                  to={paths.owners.detail.getHref(String(row.owner_id))}
                  aria-label={`飼主詳細: ${row.owner_name} (ID ${row.owner_id})`}
                >
                  {row.owner_name}
                </DataTableRowLink>
              </TableCell>
              <TableCell>{row.pet_name || "-"}</TableCell>
              <TableCell className="text-right font-mono">
                {formatCurrency(row.prev_month_carryover)}
              </TableCell>
              <TableCell className="text-right font-mono">
                {formatCurrency(row.current_month_unpaid)}
              </TableCell>
              <TableCell className="text-right font-mono">
                {formatCurrency(row.next_month_carryover)}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

interface UnpaidTabPaginationProps {
  total: number;
  page: number;
  limit: number;
  onPageChange: React.Dispatch<React.SetStateAction<number>>;
}

export function UnpaidTabPagination({
  total,
  page,
  limit,
  onPageChange,
}: UnpaidTabPaginationProps) {
  if (total <= limit) return null;
  const totalPages = Math.max(1, Math.ceil(total / limit));
  const startIndex = (page - 1) * limit + 1;
  const endIndex = Math.min(total, page * limit);
  return (
    <Pagination
      currentPage={page}
      totalPages={totalPages}
      totalCount={total}
      startIndex={startIndex}
      endIndex={endIndex}
      onPageChange={onPageChange}
      onPrev={() => onPageChange((currentPage) => Math.max(1, currentPage - 1))}
      onNext={() => onPageChange((currentPage) => Math.min(totalPages, currentPage + 1))}
    />
  );
}
