import { memo, useMemo } from "react";
import { EmptyState } from "@/components/shared/DataStates";
import { TableCell, TableHead } from "@/components/ui/table";
import { C, STYLE } from "@/lib/design-tokens";
import { formatCurrency } from "@/lib/format/number";
import type { CategoryPaymentMatrix } from "../api/get-monthly-report";

/** DISPLAY_CATEGORIES と同じラベル集約（cash-register/constants と一致させる） */
const DISPLAY_GROUPS: { label: string; keys: string[] }[] = [
  { label: "診療", keys: ["examination", "test", "procedure", "medicine"] },
  { label: "外科", keys: ["surgery"] },
  { label: "RV", keys: ["vaccine"] },
  { label: "フード", keys: ["food"] },
  { label: "トリミング", keys: ["trimming"] },
  { label: "ホテル", keys: ["hotel"] },
  { label: "用品", keys: ["goods"] },
  { label: "トレセン", keys: ["training"] },
  { label: "未分類・要確認", keys: ["other"] },
];

interface DisplayRow {
  label: string;
  count: number;
  byMethod: Record<string, number>;
  rowTotal: number;
}

function buildDisplayRows(matrix: CategoryPaymentMatrix): DisplayRow[] {
  const byCat = new Map(matrix.rows.map((r) => [r.category, r]));
  return DISPLAY_GROUPS.map((group) => {
    const byMethod: Record<string, number> = {};
    let rowTotal = 0;
    let count = 0;
    for (const key of group.keys) {
      const src = byCat.get(key);
      if (!src) continue;
      count += src.count;
      rowTotal += src.rowTotal;
      for (const [method, amount] of Object.entries(src.byMethod)) {
        byMethod[method] = (byMethod[method] ?? 0) + amount;
      }
    }
    return { label: group.label, count, byMethod, rowTotal };
  }).filter((row) => row.rowTotal !== 0 || row.count > 0);
}

interface CategoryPaymentMatrixTableProps {
  matrix: CategoryPaymentMatrix;
  /** compact = print typography */
  compact?: boolean;
}

/**
 * #247: 部門×支払方法統合表。
 * 画面（MonthlySummaryCards 下）と印刷（MonthlyReportPrintArea）が同一 matrix を描画源とする。
 * 金額 = 支払実額基準（DEC-16⑥）。件数 = 会計 distinct。
 */
export const CategoryPaymentMatrixTable = memo(function CategoryPaymentMatrixTable({
  matrix,
  compact = false,
}: CategoryPaymentMatrixTableProps) {
  const rows = useMemo(() => buildDisplayRows(matrix), [matrix]);
  const methods = matrix.paymentMethods;
  const text = compact ? "text-[8pt]" : "text-base";

  if (rows.length === 0) {
    return <EmptyState message="対象期間の会計データがありません" />;
  }

  return (
    <div className="overflow-x-auto">
      <table className={`w-full ${text} border-collapse`}>
        <thead>
          <tr className={`border-b ${C.borderLight} ${C.bgPage}`}>
            <TableHead className={`${C.text70} w-24`}>部門</TableHead>
            <TableHead className={`text-right ${C.text70}`}>件数</TableHead>
            {methods.map((pm) => (
              <TableHead key={pm.name} className={`text-right ${C.text70}`}>
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
              <TableCell className={`text-right ${C.text60}`}>{row.count}件</TableCell>
              {methods.map((pm) => (
                <TableCell key={pm.name} className={`text-right ${C.text}`}>
                  {row.byMethod[pm.name] != null ? formatCurrency(row.byMethod[pm.name]) : "—"}
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
              {matrix.totals.count}件
            </TableCell>
            {methods.map((pm) => (
              <TableCell key={pm.name} className={`text-right font-semibold ${C.text}`}>
                {matrix.totals.byMethod[pm.name] != null
                  ? formatCurrency(matrix.totals.byMethod[pm.name])
                  : "—"}
              </TableCell>
            ))}
            <TableCell className={`text-right font-semibold ${C.text}`}>
              {formatCurrency(matrix.totals.grandTotal)}
            </TableCell>
          </tr>
        </tfoot>
      </table>
    </div>
  );
});
