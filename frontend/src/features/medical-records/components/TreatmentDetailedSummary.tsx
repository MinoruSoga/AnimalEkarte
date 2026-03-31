// React/Framework
import React from "react";

// Internal
import { NumberInput } from "@/components/shared/NumberInput/NumberInput";

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

export const TreatmentDetailedSummary = React.memo(function TreatmentDetailedSummary({
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
    <div className="grid grid-cols-2 gap-0 mt-2 border border-[rgba(55,53,47,0.16)] rounded-lg bg-white text-sm overflow-hidden mb-8 shadow-sm">
      <div className="col-span-2">
        {/* Summary Headers */}
        <div className="grid grid-cols-5 border-b border-[rgba(55,53,47,0.16)] bg-[#F7F6F3]">
          <div className="p-2 font-bold text-[#37352F]/80 text-sm border-r border-[rgba(55,53,47,0.16)]">
            診療費 小計
          </div>
          <div className="p-2 font-bold text-[#37352F]/80 text-sm border-r border-[rgba(55,53,47,0.16)]">
            割引適用額
          </div>
          <div className="p-2 font-bold text-[#37352F]/80 text-sm border-r border-[rgba(55,53,47,0.16)]">
            値引適用額
          </div>
          <div className="p-2 font-bold text-[#37352F]/80 text-sm border-r border-[rgba(55,53,47,0.16)]">
            消費税
          </div>
          <div className="p-2 font-bold text-[#37352F]/80 text-sm">
            請求額
          </div>
        </div>

        {/* Summary Values */}
        <div className="grid grid-cols-5 border-b border-[rgba(55,53,47,0.16)] bg-white items-center h-12">
          <div className="p-2 text-right text-[#37352F] border-r border-[rgba(55,53,47,0.16)] h-full flex items-center justify-end font-mono font-medium">
            ￥{subtotal.toLocaleString()}
          </div>
          <div className="p-2 border-r border-[rgba(55,53,47,0.16)] h-full flex items-center justify-end gap-1">
            {isDiscountRateReadonly ? (
              <>
                <span className="text-xs text-[#37352F]/50">飼主割引</span>
                <span className="text-sm font-mono font-medium text-[#37352F]">{discountRate}</span>
                <span className="text-sm text-[#37352F]">%</span>
              </>
            ) : (
              <>
                <span className="text-sm text-[#37352F]/60">割引率</span>
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
          <div className="p-2 border-r border-[rgba(55,53,47,0.16)] h-full flex items-center justify-end gap-1">
            <span className="text-sm text-[#37352F]/60">値引額</span>
            <NumberInput
              className="w-20 h-10"
              value={discountAmount}
              onChange={(v) => onUpdateDiscountAmount(Number(v))}
              suffix="円"
              align="right"
              disabled={disabled}
            />
          </div>
          <div className="p-2 text-right text-[#37352F] border-r border-[rgba(55,53,47,0.16)] h-full flex items-center justify-end font-mono font-medium">
            ￥{tax.toLocaleString()}
          </div>
          <div className="p-2 text-right text-[#37352F] h-full flex items-center justify-end font-mono font-bold text-lg">
            ￥{total.toLocaleString()}
          </div>
        </div>

        {/* Final Totals */}
        <div className="grid grid-cols-2 bg-white">
          <div className="p-2 border-r border-[rgba(55,53,47,0.16)] flex justify-between items-center h-10">
            <span className="font-normal text-[#37352F] text-sm">
              保険請求額
            </span>
            <span className="text-[#37352F] font-mono font-medium">0</span>
          </div>
          <div className="p-2 flex justify-between items-center h-10">
            <span className="font-normal text-[#37352F] text-sm">
              飼主請求額
            </span>
            <span className="font-bold text-[#37352F] font-mono text-base">
              ￥{total.toLocaleString()}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
});
