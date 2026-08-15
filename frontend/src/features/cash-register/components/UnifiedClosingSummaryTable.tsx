import { memo, useMemo } from "react";
import { EmptyState } from "@/components/shared/DataStates";
import { TableCell, TableHead } from "@/components/ui/table";
import { C, STYLE } from "@/lib/design-tokens";
import { formatCurrency } from "@/lib/format/number";
import type { PaymentMethodMaster } from "@/types/generated/models";
import type { CloseBillingDetail } from "../api/get-cash-register-preview";
import {
  buildUnifiedClosingRows,
  buildUnifiedClosingTotals,
  formatClosingCount,
  type UnclassifiedOtherCountInput,
} from "../closing-summary";

interface UnifiedClosingSummaryTableProps {
  categories: Record<string, Record<string, number>>;
  paymentMethods: PaymentMethodMaster[];
  billingDetails: CloseBillingDetail[];
  /** DEC-40: other 行の会計 distinct 件数。null = 記録なし */
  unclassifiedOtherCount?: UnclassifiedOtherCountInput;
  /** #247: 部門ごとの会計 distinct 件数 */
  categoryCounts?: Record<string, number>;
}

/**
 * #153: レジ締めサマリーの統合テーブル。
 * 旧「部門別集計（CategoryPaymentMatrix）」に件数（会計件数）を統合し、
 * 部門 / 件数 / 支払方法別金額 / 合計 を 1 表で表示する。
 */
export const UnifiedClosingSummaryTable = memo(function UnifiedClosingSummaryTable({
  categories,
  paymentMethods,
  billingDetails,
  unclassifiedOtherCount,
  categoryCounts,
}: UnifiedClosingSummaryTableProps) {
  const rows = useMemo(
    () => buildUnifiedClosingRows(categories, billingDetails, unclassifiedOtherCount, categoryCounts),
    [categories, billingDetails, unclassifiedOtherCount, categoryCounts],
  );
  const totals = useMemo(() => buildUnifiedClosingTotals(rows), [rows]);

  if (rows.length === 0) {
    return <EmptyState message="対象期間の会計データがありません" />;
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-base">
        <thead>
          <tr className={`border-b ${C.borderLight} ${C.bgPage}`}>
            <TableHead className={`${C.text70} w-24`}>部門</TableHead>
            <TableHead className={`text-right ${C.text70}`}>件数</TableHead>
            {paymentMethods.map((pm) => (
              <TableHead key={pm.id} className={`text-right ${C.text70}`}>
                {pm.name}
              </TableHead>
            ))}
            <TableHead className={`text-right ${C.text70}`}>合計</TableHead>
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr key={row.label} className={`border-b ${C.borderLight} ${STYLE.tableRow}`}>
              <TableCell className={`font-medium ${C.text}`}>{row.label}</TableCell>
              <TableCell className={`text-right ${C.text60}`}>
                {formatClosingCount(row.count)}
              </TableCell>
              {paymentMethods.map((pm) => (
                <TableCell key={pm.id} className={`text-right ${C.text}`}>
                  {row.byMethod[pm.name] != null
                    ? formatCurrency(row.byMethod[pm.name])
                    : "—"}
                </TableCell>
              ))}
              <TableCell className={`text-right font-medium ${C.text}`}>
                {formatCurrency(row.rowTotal)}
              </TableCell>
            </tr>
          ))}
        </tbody>
        <tfoot>
          <tr className={`border-t-2 ${C.borderMedium} ${C.bgPage}`}>
            <TableCell className={`font-semibold ${C.text}`}>合計</TableCell>
            <TableCell className={`text-right font-semibold ${C.text}`}>
              {formatClosingCount(totals.count)}
            </TableCell>
            {paymentMethods.map((pm) => (
              <TableCell key={pm.id} className={`text-right font-semibold ${C.text}`}>
                {totals.byMethod[pm.name] != null
                  ? formatCurrency(totals.byMethod[pm.name])
                  : "—"}
              </TableCell>
            ))}
            <TableCell className={`text-right font-semibold ${C.text}`}>
              {formatCurrency(totals.grandTotal)}
            </TableCell>
          </tr>
        </tfoot>
      </table>
    </div>
  );
});
