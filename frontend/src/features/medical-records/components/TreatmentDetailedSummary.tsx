// React/Framework
import { memo } from "react";

// Internal
import { NumberInput } from "@/components/shared/NumberInput/NumberInput";
import { C, STYLE } from "@/lib/design-tokens";

interface TreatmentDetailedSummaryProps {
  subtotal: number;
  tax: number;
  total: number;
  discountRate: number;
  discountAmount: number;
  onUpdateDiscountRate?: (value: number) => void;
  onUpdateDiscountAmount: (value: number) => void;
  isDiscountRateReadonly?: boolean;
  /** BUG-051: 確認済み等で全入力を無効化する */
  disabled?: boolean;
}

export const TreatmentDetailedSummary = memo(function TreatmentDetailedSummary({
  subtotal,
  tax,
  total,
  discountRate,
  discountAmount,
  onUpdateDiscountRate,
  onUpdateDiscountAmount,
  isDiscountRateReadonly = false,
  disabled = false,
}: TreatmentDetailedSummaryProps) {
  return (
    <div className={`grid grid-cols-2 gap-0 mt-2 border ${C.borderMedium} rounded-lg ${C.bgWhite} text-sm overflow-hidden mb-8`}>
      <div className="col-span-2">
        {/* Summary Headers — DESIGN.md ex-data-table-cell: canvas-soft 背景 + eyebrow 相当タイポグラフィ（STYLE.sectionLabel） */}
        <div className={`grid grid-cols-5 border-b ${C.borderMedium} ${C.bgPage}`}>
          <div className={`p-2 ${STYLE.sectionLabel} border-r ${C.borderMedium}`}>
            診療費 小計
          </div>
          <div className={`p-2 ${STYLE.sectionLabel} border-r ${C.borderMedium}`}>
            割引適用額
          </div>
          <div className={`p-2 ${STYLE.sectionLabel} border-r ${C.borderMedium}`}>
            値引適用額
          </div>
          <div className={`p-2 ${STYLE.sectionLabel} border-r ${C.borderMedium}`}>
            消費税
          </div>
          <div className={`p-2 ${STYLE.sectionLabel}`}>
            請求額
          </div>
        </div>

        {/* Summary Values */}
        <div className={`grid grid-cols-5 border-b ${C.borderMedium} ${C.bgWhite} items-center h-12`}>
          <div className={`p-2 text-right ${C.text} border-r ${C.borderMedium} h-full flex items-center justify-end font-mono font-medium`}>
            ￥{subtotal.toLocaleString()}
          </div>
          <div className={`p-2 border-r ${C.borderMedium} h-full flex items-center justify-end gap-1`}>
            {isDiscountRateReadonly ? (
              <>
                <span className={`text-xs ${C.text50}`}>飼主割引</span>
                <span className={`text-sm font-mono font-medium ${C.text}`}>{discountRate}</span>
                <span className={`text-sm ${C.text}`}>%</span>
              </>
            ) : (
              <>
                <span className={`text-sm ${C.text60}`}>割引率</span>
                <NumberInput
                  className="w-16 h-10"
                  value={discountRate}
                  onChange={(v) => onUpdateDiscountRate?.(Number(v))}
                  suffix="%"
                  align="right"
                  disabled={disabled}
                />
              </>
            )}
          </div>
          <div className={`p-2 border-r ${C.borderMedium} h-full flex items-center justify-end gap-1`}>
            <span className={`text-sm ${C.text60}`}>値引額</span>
            <NumberInput
              className="w-20 h-10"
              value={discountAmount}
              onChange={(v) => onUpdateDiscountAmount(Number(v))}
              suffix="円"
              align="right"
              disabled={disabled}
            />
          </div>
          <div className={`p-2 text-right ${C.text} border-r ${C.borderMedium} h-full flex items-center justify-end font-mono font-medium`}>
            ￥{tax.toLocaleString()}
          </div>
          <div className={`p-2 text-right ${C.text} h-full flex items-center justify-end font-mono font-bold text-xl`}>
            ￥{total.toLocaleString()}
          </div>
        </div>

        {/* Final Totals */}
        <div className={`grid grid-cols-2 ${C.bgWhite}`}>
          <div className={`p-2 border-r ${C.borderMedium} flex justify-between items-center h-10`}>
            <span className={`font-normal ${C.text} text-sm`}>
              保険請求額
            </span>
            <span className={`${C.text} font-mono font-medium`}>0</span>
          </div>
          <div className="p-2 flex justify-between items-center h-10">
            <span className={`font-normal ${C.text} text-sm`}>
              飼主請求額
            </span>
            <span className={`font-bold ${C.text} font-mono text-base`}>
              ￥{total.toLocaleString()}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
});
