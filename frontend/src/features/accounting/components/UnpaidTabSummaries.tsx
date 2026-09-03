import { C } from "@/lib/design-tokens";
import { formatCurrency } from "@/lib/format/number";

import type { UnpaidByOwnerResponse, MonthlyUnpaidResponse } from "../api/get-unpaid-billings";

interface UnpaidTabSummariesProps {
  groupBy: "owner" | "billing" | "monthly";
  summary: UnpaidByOwnerResponse["summary"] | undefined;
  monthlySummary: MonthlyUnpaidResponse["summary"] | undefined;
}

export function UnpaidTabSummaries({ groupBy, summary, monthlySummary }: UnpaidTabSummariesProps) {
  return (
    <>
      {groupBy !== "monthly" && summary ? (
        <div className={`rounded-lg border ${C.borderLight} p-4 ${C.bgWhite}`}>
          <p className={`text-xs ${C.text50} mb-1`}>売掛金総額</p>
          <p className="text-heading-3 font-bold">{formatCurrency(summary.total_amount)}</p>
          <p className={`text-xs ${C.text60} mt-1`}>
            {summary.billing_count}件 / {summary.owner_count}名
          </p>
        </div>
      ) : null}

      {groupBy === "monthly" && monthlySummary ? (
        <div className={`rounded-lg border ${C.borderLight} p-4 ${C.bgWhite}`}>
          <div className="grid grid-cols-3 gap-4">
            <div>
              <p className={`text-xs ${C.text50} mb-1`}>前月繰越</p>
              <p className="text-xl font-bold">
                {formatCurrency(monthlySummary.prev_month_carryover)}
              </p>
            </div>
            <div>
              <p className={`text-xs ${C.text50} mb-1`}>当月未払い</p>
              <p className="text-xl font-bold">
                {formatCurrency(monthlySummary.current_month_unpaid)}
              </p>
            </div>
            <div>
              <p className={`text-xs ${C.text50} mb-1`}>次月繰越</p>
              <p className="text-xl font-bold">
                {formatCurrency(monthlySummary.next_month_carryover)}
              </p>
            </div>
          </div>
        </div>
      ) : null}
    </>
  );
}
