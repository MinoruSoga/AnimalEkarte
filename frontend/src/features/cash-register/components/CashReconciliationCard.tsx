import { memo } from "react";
import { C } from "@/lib/design-tokens";
import { formatCurrency } from "@/lib/format/number";

interface CashReconciliationCardProps {
  theoreticalCash: number;
  actualCash: number;
}

export const CashReconciliationCard = memo(function CashReconciliationCard({
  theoreticalCash,
  actualCash,
}: CashReconciliationCardProps) {
  const difference = actualCash - theoreticalCash;

  return (
    <div className={`rounded-lg border ${C.borderLight} ${C.bgWhite} p-4 space-y-3`}>
      <div className="flex justify-between items-center">
        <span className={`text-base ${C.text70}`}>理論現金</span>
        <span className={`text-base font-semibold ${C.text}`}>
          {formatCurrency(theoreticalCash)}
        </span>
      </div>
      <div className="flex justify-between items-center">
        <span className={`text-base ${C.text70}`}>実際の現金</span>
        <span className={`text-base font-semibold ${C.text}`}>
          {formatCurrency(actualCash)}
        </span>
      </div>
      <div className={`flex justify-between items-center border-t pt-3 ${C.borderLight}`}>
        <span className={`text-base font-medium ${C.text}`}>差額</span>
        <span
          className={`text-base font-semibold ${difference === 0 ? C.textStatusGreen : C.danger}`}
        >
          {difference >= 0 ? "+" : ""}
          {difference.toLocaleString()}
        </span>
      </div>
    </div>
  );
});
