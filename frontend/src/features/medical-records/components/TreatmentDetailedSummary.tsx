// React/Framework
import React from "react";

// Internal
import { Input } from "@/components/ui/input";

interface TreatmentDetailedSummaryProps {
  subtotal: number;
  tax: number;
  total: number;
  discountRate: number;
  discountAmount: number;
  onUpdateDiscountRate: (value: number) => void;
  onUpdateDiscountAmount: (value: number) => void;
}

export const TreatmentDetailedSummary = React.memo(function TreatmentDetailedSummary({
  subtotal,
  tax,
  total,
  discountRate,
  discountAmount,
  onUpdateDiscountRate,
  onUpdateDiscountAmount,
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
            <span className="text-sm text-[#37352F]/60">割引率</span>
            <Input
              className="w-16 h-10 text-right text-sm bg-white border-[rgba(55,53,47,0.16)]"
              value={discountRate}
              onChange={(e) => onUpdateDiscountRate(Number(e.target.value))}
              type="number"
            />
            <span className="text-sm text-[#37352F]">%</span>
          </div>
          <div className="p-2 border-r border-[rgba(55,53,47,0.16)] h-full flex items-center justify-end gap-1">
            <span className="text-sm text-[#37352F]/60">値引額 ￥</span>
            <Input
              className="w-20 h-10 text-right text-sm bg-white border-[rgba(55,53,47,0.16)]"
              value={discountAmount}
              onChange={(e) => onUpdateDiscountAmount(Number(e.target.value))}
              type="number"
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
