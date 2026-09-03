import { memo } from "react";
import { H_STYLES } from "../lib/styles";
import { C } from "@/lib/design-tokens";

interface HospitalizationCostSummaryProps {
  totals: {
    subtotalBeforeDiscount: number;
    subtotalAfterDiscount: number;
    consumptionTax: number;
    total: number;
  };
}

/** Read-only cost summary from treatment-plan line items (no bulk discount inputs). */
export const HospitalizationCostSummary = memo(function HospitalizationCostSummary({
  totals,
}: HospitalizationCostSummaryProps) {
  return (
    <div className={`${C.bgWhite} rounded-lg border ${C.borderMedium} ${H_STYLES.padding.box}`}>
      <h2 className={`${H_STYLES.text.base} font-bold mb-3 ${C.text}`}>診療費計算</h2>

      <div className="space-y-2">
        {/* 小計 */}
        <div className={`flex items-center justify-between py-1.5 border-b ${C.borderLight}`}>
          <span className={`${H_STYLES.text.base} ${C.text60}`}>診療費 小計</span>
          <span className={`${H_STYLES.text.base} font-medium tabular-nums ${C.text}`}>
            ￥{totals.subtotalBeforeDiscount.toLocaleString()}
          </span>
        </div>

        {/* 消費税 */}
        <div className={`flex items-center justify-between py-1.5 border-b ${C.borderLight}`}>
          <span className={`${H_STYLES.text.base} ${C.text60}`}>消費税</span>
          <span className={`${H_STYLES.text.base} font-medium tabular-nums ${C.text}`}>
            ￥{totals.consumptionTax.toLocaleString()}
          </span>
        </div>

        {/* 請求額 */}
        <div className={`flex items-center justify-between py-2 ${C.bgPage} rounded-md px-3 mt-2`}>
          <span className={`font-medium ${H_STYLES.text.base} ${C.text}`}>請求額</span>
          <span className={`${H_STYLES.text.lg} font-semibold tabular-nums ${C.text}`}>
            ￥{totals.total.toLocaleString()}
          </span>
        </div>

        <p className={`mt-2 ${H_STYLES.text.sm} ${C.text60}`}>
          保険負担額と飼主負担額は、保険条件を確認したうえで会計時に確定します。
        </p>
      </div>
    </div>
  );
});
