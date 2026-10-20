import { EmptyState } from "@/components/shared/DataStates";
import { TableCell, TableHead } from "@/components/ui/table";
import { C } from "@/lib/design-tokens";
import { formatCurrency } from "@/lib/format/number";
import { EMPTY_VALUE } from "@/constants/empty-value";
import { formatJSTTime } from "@/lib/jst-date";
import {
  DETAIL_DEPARTMENT_BUCKETS,
  projectBillingDetailRows,
  sumBillingDetailRowTotals,
  type BillingDetailDisplayRow,
} from "../lib/billing-detail-rows";
import type { CloseBillingDetail } from "../api/get-cash-register-preview";

function formatPaidAt(value: string): string {
  return value ? formatJSTTime(value) : "—";
}

function formatAmountCell(amount: number): string {
  return amount === 0 ? EMPTY_VALUE : formatCurrency(amount);
}

function DetailDataRow({ row }: { row: BillingDetailDisplayRow }) {
  const departments = row.departments;
  return (
    <tr className={`border-b ${C.borderLight}`}>
      <TableCell className={`whitespace-nowrap ${C.text60}`}>{formatPaidAt(row.paidAt)}</TableCell>
      <TableCell className={`whitespace-nowrap font-mono ${C.text}`}>{row.receiptNo}</TableCell>
      {departments === null ? (
        <TableCell colSpan={DETAIL_DEPARTMENT_BUCKETS.length} className={`text-center ${C.text60}`}>
          記録なし
        </TableCell>
      ) : (
        DETAIL_DEPARTMENT_BUCKETS.map((bucket) => (
          <TableCell key={bucket.key} className={`text-right whitespace-nowrap ${C.text}`}>
            {formatAmountCell(departments[bucket.key])}
          </TableCell>
        ))
      )}
      <TableCell className={`text-center ${C.text}`}>{row.paymentMethodLabel || "-"}</TableCell>
      <TableCell className={`text-right whitespace-nowrap ${C.text}`}>
        {formatCurrency(row.totalAmount)}
      </TableCell>
      <TableCell className={`text-right whitespace-nowrap ${C.text}`}>
        {row.refundAmount > 0 ? formatCurrency(-row.refundAmount) : EMPTY_VALUE}
      </TableCell>
      <TableCell className={`text-right whitespace-nowrap ${C.text}`}>
        {formatCurrency(row.netAmount)}
      </TableCell>
      <TableCell className={`${C.text}`}>
        {row.ownerName || "-"}
        {row.isHospitalization ? <span className={`ml-1 ${C.textStatusAmber}`}>(入院)</span> : null}
      </TableCell>
      <TableCell className={`${C.text}`}>{row.petName || "-"}</TableCell>
    </tr>
  );
}

interface BillingDetailTableProps {
  details: CloseBillingDetail[];
  /** 印刷フッタ（可視行の部門合計・合計・返金・純額を1回だけ） */
  showFooter?: boolean;
  footerLabel?: string;
  className?: string;
}

/**
 * EMR-354: 個別会計明細の共有テーブル（画面・印刷共通の単一 projection）。
 * split 明細は projectBillingDetailRows で 1 会計=1 行に集約済み。
 * 空時は共有 EmptyState。印刷時のタイポグラフィ/改ページ制御は
 * ClosePrintArea の [data-testid="close-print-area"] scoped CSS のみで行う。
 */
export function BillingDetailTable({
  details,
  showFooter = false,
  footerLabel = "合計",
  className,
}: BillingDetailTableProps) {
  const rows = projectBillingDetailRows(details);
  const totals = showFooter ? sumBillingDetailRowTotals(rows) : null;

  if (rows.length === 0) {
    return <EmptyState message="会計明細がありません" />;
  }

  const deptTotals = totals !== null ? totals.departments : null;

  return (
    <div className="overflow-x-auto">
      <table className={`w-full text-sm billing-detail-table${className ? ` ${className}` : ""}`}>
        <thead>
          <tr className={`border-b ${C.borderLight} ${C.bgPage}`}>
            <TableHead className={C.text70}>時刻</TableHead>
            <TableHead className={C.text70}>領収No</TableHead>
            {DETAIL_DEPARTMENT_BUCKETS.map((bucket) => (
              <TableHead key={bucket.key} className={`text-right ${C.text70}`}>
                {bucket.label}
              </TableHead>
            ))}
            <TableHead className={`text-center ${C.text70}`}>支払方法</TableHead>
            <TableHead className={`text-right ${C.text70}`}>合計</TableHead>
            <TableHead className={`text-right ${C.text70}`}>返金額</TableHead>
            <TableHead className={`text-right ${C.text70}`}>純額</TableHead>
            <TableHead className={`text-center ${C.text70}`}>飼主名</TableHead>
            <TableHead className={`text-center ${C.text70}`}>ペット名</TableHead>
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <DetailDataRow key={row.billingId} row={row} />
          ))}
        </tbody>
        {totals !== null ? (
          <tfoot>
            <tr className={`border-t-2 ${C.borderMedium} ${C.bgPage}`}>
              <TableCell colSpan={2} className={`font-semibold ${C.text}`}>
                {footerLabel}
              </TableCell>
              {deptTotals === null ? (
                <TableCell
                  colSpan={DETAIL_DEPARTMENT_BUCKETS.length}
                  className={`text-center ${C.text60}`}
                >
                  記録なし
                </TableCell>
              ) : (
                DETAIL_DEPARTMENT_BUCKETS.map((bucket) => (
                  <TableCell
                    key={bucket.key}
                    className={`text-right font-semibold whitespace-nowrap ${C.text}`}
                  >
                    {formatAmountCell(deptTotals[bucket.key])}
                  </TableCell>
                ))
              )}
              <TableCell />
              <TableCell className={`text-right font-semibold whitespace-nowrap ${C.text}`}>
                {formatCurrency(totals.totalAmount)}
              </TableCell>
              <TableCell className={`text-right font-semibold whitespace-nowrap ${C.text}`}>
                {totals.refundAmount > 0 ? formatCurrency(-totals.refundAmount) : EMPTY_VALUE}
              </TableCell>
              <TableCell className={`text-right font-semibold whitespace-nowrap ${C.text}`}>
                {formatCurrency(totals.netAmount)}
              </TableCell>
              <TableCell colSpan={2} />
            </tr>
          </tfoot>
        ) : null}
      </table>
    </div>
  );
}
