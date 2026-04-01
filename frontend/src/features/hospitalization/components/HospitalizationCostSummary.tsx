import { memo } from "react";
import { NumberInput } from "@/components/shared/NumberInput/NumberInput";
import { H_STYLES } from "../styles";
import { C } from "@/lib/design-tokens";

interface HospitalizationCostSummaryProps {
    totals: {
        subtotalBeforeDiscount: number;
        subtotalAfterDiscount: number;
        consumptionTax: number;
        total: number;
    };
    globalDiscount: number;
    setGlobalDiscount: (val: number) => void;
    globalDiscountAmount: number;
    setGlobalDiscountAmount: (val: number) => void;
}

export const HospitalizationCostSummary = memo(function HospitalizationCostSummary({
    totals,
    globalDiscount,
    setGlobalDiscount,
    globalDiscountAmount,
    setGlobalDiscountAmount
}: HospitalizationCostSummaryProps) {
  return (
    <div className={`bg-white rounded-lg shadow-sm border border-[rgba(55,53,47,0.16)] ${H_STYLES.padding.box}`}>
      <h2 className={`${H_STYLES.text.base} font-bold mb-3 ${C.text}`}>診療費計算</h2>
      
      <div className="space-y-2">
        {/* 小計 */}
        <div className="flex items-center justify-between py-1.5 border-b border-[rgba(55,53,47,0.09)]">
          <span className={`${H_STYLES.text.base} ${C.text60}`}>診療費 小計</span>
          <span className={`${H_STYLES.text.base} font-medium tabular-nums ${C.text}`}>
            ￥{totals.subtotalBeforeDiscount.toLocaleString()}
          </span>
        </div>

        {/* 割引 */}
        <div className="flex items-center justify-between py-1.5 border-b border-[rgba(55,53,47,0.09)]">
          <div className="flex items-center gap-3">
            <span className={`${H_STYLES.text.base} ${C.text60}`}>割引適用額</span>
            <div className="flex items-center gap-2">
              <NumberInput
                value={globalDiscount}
                onChange={(v) => setGlobalDiscount(parseInt(v) || 0)}
                suffix="%"
                align="right"
                className={`w-16 h-10 ${H_STYLES.text.base}`}
              />
            </div>
          </div>
          <span className={`${H_STYLES.text.base} font-medium tabular-nums ${C.text}`}>
            ￥{totals.subtotalAfterDiscount.toLocaleString()}
          </span>
        </div>

        {/* 値引 */}
        <div className="flex items-center justify-between py-1.5 border-b border-[rgba(55,53,47,0.09)]">
          <div className="flex items-center gap-3">
            <span className={`${H_STYLES.text.base} ${C.text60}`}>値引適用額</span>
            <div className="flex items-center gap-2">
              <NumberInput
                value={globalDiscountAmount}
                onChange={(v) => setGlobalDiscountAmount(parseInt(v) || 0)}
                suffix="円"
                align="right"
                className={`w-20 h-10 ${H_STYLES.text.base}`}
              />
            </div>
          </div>
          <span className={`${H_STYLES.text.base} font-medium tabular-nums ${C.text}`}>
            ￥{totals.subtotalAfterDiscount.toLocaleString()}
          </span>
        </div>

        {/* 消費税 */}
        <div className="flex items-center justify-between py-1.5 border-b border-[rgba(55,53,47,0.09)]">
          <span className={`${H_STYLES.text.base} ${C.text60}`}>消費税</span>
          <span className={`${H_STYLES.text.base} font-medium tabular-nums ${C.text}`}>
            ￥{totals.consumptionTax.toLocaleString()}
          </span>
        </div>

        {/* 請求額 */}
        <div className="flex items-center justify-between py-2 bg-[#F7F6F3] rounded-md px-3 mt-2">
          <span className={`font-medium ${H_STYLES.text.base} ${C.text}`}>請求額</span>
          <span className={`${H_STYLES.text.lg} font-semibold tabular-nums ${C.text}`}>
            ￥{totals.total.toLocaleString()}
          </span>
        </div>

        {/* 保険・飼主請求額 */}
        <div className="grid grid-cols-2 gap-3 mt-2">
          <div className="flex items-center justify-between py-1.5 px-3 bg-[#E3F2FD] rounded-md">
            <span className={`${H_STYLES.text.base} ${C.text60}`}>保険請求額</span>
            <span className={`${H_STYLES.text.base} font-medium tabular-nums text-[#1565C0]`}>￥0</span>
          </div>
          <div className="flex items-center justify-between py-1.5 px-3 bg-[#E8F5E9] rounded-md">
            <span className={`${H_STYLES.text.base} ${C.text60}`}>飼主請求額</span>
            <span className={`${H_STYLES.text.base} font-medium tabular-nums text-[#2E7D32]`}>
              ￥{totals.total.toLocaleString()}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
});
