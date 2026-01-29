import { Input } from "@/components/ui/input";
import { H_STYLES } from "../styles";

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

export const HospitalizationCostSummary = ({ 
    totals, 
    globalDiscount, 
    setGlobalDiscount, 
    globalDiscountAmount, 
    setGlobalDiscountAmount 
}: HospitalizationCostSummaryProps) => {
  return (
    <div className={`bg-white rounded-lg shadow-sm border border-[rgba(55,53,47,0.16)] ${H_STYLES.padding.box}`}>
      <h2 className={`${H_STYLES.text.base} font-bold mb-3 text-[#37352F]`}>診療費計算</h2>
      
      <div className="space-y-2">
        {/* 小計 */}
        <div className="flex items-center justify-between py-1.5 border-b border-[rgba(55,53,47,0.09)]">
          <span className={`${H_STYLES.text.base} text-[#37352F]/60`}>診療費 小計</span>
          <span className={`${H_STYLES.text.base} font-medium tabular-nums text-[#37352F]`}>
            ￥{totals.subtotalBeforeDiscount.toLocaleString()}
          </span>
        </div>

        {/* 割引 */}
        <div className="flex items-center justify-between py-1.5 border-b border-[rgba(55,53,47,0.09)]">
          <div className="flex items-center gap-3">
            <span className={`${H_STYLES.text.base} text-[#37352F]/60`}>割引適用額</span>
            <div className="flex items-center gap-2">
              <Input
                type="number"
                value={globalDiscount}
                onChange={(e) =>
                  setGlobalDiscount(parseInt(e.target.value) || 0)
                }
                className={`w-16 h-10 ${H_STYLES.text.base} text-right bg-white border-[rgba(55,53,47,0.16)] focus-visible:ring-[#2EAADC]`}
                placeholder="0"
              />
              <span className={`${H_STYLES.text.base} text-[#37352F]/60`}>%</span>
            </div>
          </div>
          <span className={`${H_STYLES.text.base} font-medium tabular-nums text-[#37352F]`}>
            ￥{totals.subtotalAfterDiscount.toLocaleString()}
          </span>
        </div>

        {/* 値引 */}
        <div className="flex items-center justify-between py-1.5 border-b border-[rgba(55,53,47,0.09)]">
          <div className="flex items-center gap-3">
            <span className={`${H_STYLES.text.base} text-[#37352F]/60`}>値引適用額</span>
            <div className="flex items-center gap-2">
              <span className={`${H_STYLES.text.base} text-[#37352F]/60`}>￥</span>
              <Input
                type="number"
                value={globalDiscountAmount}
                onChange={(e) =>
                  setGlobalDiscountAmount(parseInt(e.target.value) || 0)
                }
                className={`w-20 h-10 ${H_STYLES.text.base} text-right bg-white border-[rgba(55,53,47,0.16)] focus-visible:ring-[#2EAADC]`}
                placeholder="0"
              />
            </div>
          </div>
          <span className={`${H_STYLES.text.base} font-medium tabular-nums text-[#37352F]`}>
            ￥{totals.subtotalAfterDiscount.toLocaleString()}
          </span>
        </div>

        {/* 消費税 */}
        <div className="flex items-center justify-between py-1.5 border-b border-[rgba(55,53,47,0.09)]">
          <span className={`${H_STYLES.text.base} text-[#37352F]/60`}>消費税</span>
          <span className={`${H_STYLES.text.base} font-medium tabular-nums text-[#37352F]`}>
            ￥{totals.consumptionTax.toLocaleString()}
          </span>
        </div>

        {/* 請求額 */}
        <div className="flex items-center justify-between py-2 bg-[#F7F6F3] rounded-md px-3 mt-2">
          <span className={`font-medium ${H_STYLES.text.base} text-[#37352F]`}>請求額</span>
          <span className={`${H_STYLES.text.lg} font-semibold tabular-nums text-[#37352F]`}>
            ￥{totals.total.toLocaleString()}
          </span>
        </div>

        {/* 保険・飼主請求額 */}
        <div className="grid grid-cols-2 gap-3 mt-2">
          <div className="flex items-center justify-between py-1.5 px-3 bg-[#E3F2FD] rounded-md">
            <span className={`${H_STYLES.text.base} text-[#37352F]/60`}>保険請求額</span>
            <span className={`${H_STYLES.text.base} font-medium tabular-nums text-[#1565C0]`}>￥0</span>
          </div>
          <div className="flex items-center justify-between py-1.5 px-3 bg-[#E8F5E9] rounded-md">
            <span className={`${H_STYLES.text.base} text-[#37352F]/60`}>飼主請求額</span>
            <span className={`${H_STYLES.text.base} font-medium tabular-nums text-[#2E7D32]`}>
              ￥{totals.total.toLocaleString()}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
};
