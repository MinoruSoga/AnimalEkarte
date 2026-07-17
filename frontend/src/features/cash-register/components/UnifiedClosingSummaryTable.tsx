import { memo, useMemo } from "react";
import { EmptyState } from "@/components/shared/DataStates";
import { C, STYLE } from "@/lib/design-tokens";
import { formatCurrency } from "@/utils/format/number";
import type { PaymentMethodMaster } from "@/types/generated/models";
import type { CloseBillingDetail } from "../api/get-cash-register-preview";
import { buildUnifiedClosingRows, buildUnifiedClosingTotals } from "../closing-summary";

interface UnifiedClosingSummaryTableProps {
  categories: Record<string, Record<string, number>>;
  paymentMethods: PaymentMethodMaster[];
  billingDetails: CloseBillingDetail[];
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
}: UnifiedClosingSummaryTableProps) {
  const rows = useMemo(
    () => buildUnifiedClosingRows(categories, billingDetails),
    [categories, billingDetails],
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
            <th className={`text-left px-3 py-2 font-medium ${C.text70} w-24`}>部門</th>
            <th className={`text-right px-3 py-2 font-medium ${C.text70}`}>件数</th>
            {paymentMethods.map((pm) => (
              <th key={pm.id} className={`text-right px-3 py-2 font-medium ${C.text70}`}>
                {pm.name}
              </th>
            ))}
            <th className={`text-right px-3 py-2 font-medium ${C.text70}`}>合計</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr key={row.label} className={`border-b ${C.borderLight} ${STYLE.tableRow}`}>
              <td className={`px-3 py-2 font-medium ${C.text}`}>{row.label}</td>
              <td className={`text-right px-3 py-2 ${C.text60}`}>{row.count}件</td>
              {paymentMethods.map((pm) => (
                <td key={pm.id} className={`text-right px-3 py-2 ${C.text}`}>
                  {row.byMethod[pm.name] != null
                    ? `¥${row.byMethod[pm.name].toLocaleString()}`
                    : "—"}
                </td>
              ))}
              <td className={`text-right px-3 py-2 font-medium ${C.text}`}>
                {formatCurrency(row.rowTotal)}
              </td>
            </tr>
          ))}
        </tbody>
        <tfoot>
          <tr className={`border-t-2 ${C.borderMedium} ${C.bgPage}`}>
            <td className={`px-3 py-2 font-semibold ${C.text}`}>合計</td>
            <td className={`text-right px-3 py-2 font-semibold ${C.text}`}>{totals.count}件</td>
            {paymentMethods.map((pm) => (
              <td key={pm.id} className={`text-right px-3 py-2 font-semibold ${C.text}`}>
                {totals.byMethod[pm.name] != null
                  ? `¥${totals.byMethod[pm.name].toLocaleString()}`
                  : "—"}
              </td>
            ))}
            <td className={`text-right px-3 py-2 font-semibold ${C.text}`}>
              {formatCurrency(totals.grandTotal)}
            </td>
          </tr>
        </tfoot>
      </table>
    </div>
  );
});
