import { memo, useMemo } from "react";
import { C, STYLE } from "@/lib/design-tokens";
import type { PaymentMethodMaster } from "@/types/generated/models";
import { DISPLAY_CATEGORIES } from "../constants";

interface CategoryPaymentMatrixProps {
  categories: Record<string, Record<string, number>>;
  paymentMethods: PaymentMethodMaster[];
}

export const CategoryPaymentMatrix = memo(function CategoryPaymentMatrix({
  categories,
  paymentMethods,
}: CategoryPaymentMatrixProps) {
  const rows = useMemo(() => {
    return DISPLAY_CATEGORIES.map((displayCat) => {
      const byMethod: Record<string, number> = {};
      let rowTotal = 0;
      for (const key of displayCat.keys) {
        const methodMap = categories[key] ?? {};
        for (const [methodName, amount] of Object.entries(methodMap)) {
          byMethod[methodName] = (byMethod[methodName] ?? 0) + amount;
          rowTotal += amount;
        }
      }
      return { label: displayCat.label, byMethod, rowTotal };
    }).filter((row) => row.rowTotal > 0);
  }, [categories]);

  const colTotals = useMemo(() => {
    const totals: Record<string, number> = {};
    for (const row of rows) {
      for (const [methodName, amount] of Object.entries(row.byMethod)) {
        totals[methodName] = (totals[methodName] ?? 0) + amount;
      }
    }
    return totals;
  }, [rows]);

  const grandTotal = useMemo(
    () => rows.reduce((sum, row) => sum + row.rowTotal, 0),
    [rows],
  );

  if (rows.length === 0) {
    return (
      <p className={`text-base ${C.text50} py-4 text-center`}>対象期間の会計データがありません</p>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-base">
        <thead>
          <tr className={`border-b ${C.borderLight} ${C.bgPage}`}>
            <th className={`text-left px-3 py-2 font-medium ${C.text70} w-24`}>部門</th>
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
              {paymentMethods.map((pm) => (
                <td key={pm.id} className={`text-right px-3 py-2 ${C.text}`}>
                  {row.byMethod[pm.name] != null
                    ? `¥${row.byMethod[pm.name].toLocaleString()}`
                    : "—"}
                </td>
              ))}
              <td className={`text-right px-3 py-2 font-medium ${C.text}`}>
                ¥{row.rowTotal.toLocaleString()}
              </td>
            </tr>
          ))}
        </tbody>
        <tfoot>
          <tr className={`border-t-2 ${C.borderMedium} ${C.bgPage}`}>
            <td className={`px-3 py-2 font-semibold ${C.text}`}>合計</td>
            {paymentMethods.map((pm) => (
              <td key={pm.id} className={`text-right px-3 py-2 font-semibold ${C.text}`}>
                {colTotals[pm.name] != null ? `¥${colTotals[pm.name].toLocaleString()}` : "—"}
              </td>
            ))}
            <td className={`text-right px-3 py-2 font-semibold ${C.text}`}>
              ¥{grandTotal.toLocaleString()}
            </td>
          </tr>
        </tfoot>
      </table>
    </div>
  );
});
