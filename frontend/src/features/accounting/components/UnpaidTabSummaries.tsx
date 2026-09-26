import { C } from "@/lib/design-tokens";
import { formatCurrency } from "@/lib/format/number";

import type { UnpaidByOwnerResponse, PeriodUnpaidResponse } from "../api/get-unpaid-billings";

interface UnpaidTabSummariesProps {
  groupBy: "owner" | "billing" | "period";
  summary: UnpaidByOwnerResponse["summary"] | undefined;
  periodSummary: PeriodUnpaidResponse["summary"] | undefined;
}

export function UnpaidTabSummaries({ groupBy, summary, periodSummary }: UnpaidTabSummariesProps) {
  return (
    <>
      {groupBy !== "period" && summary ? (
        <div className={`rounded-lg border ${C.borderLight} p-4 ${C.bgWhite}`}>
          <p className={`text-xs ${C.text50} mb-1`}>売掛金総額</p>
          <p className="text-heading-3 font-bold">{formatCurrency(summary.total_amount)}</p>
          <p className={`text-xs ${C.text60} mt-1`}>
            {summary.billing_count}件 / {summary.owner_count}名
          </p>
        </div>
      ) : null}

      {/* EMR-188: 月末未納者一覧 */}
      {groupBy === "period" && periodSummary ? (
        <div className={`rounded-lg border ${C.borderLight} p-4 ${C.bgWhite}`}>
          <div className="grid grid-cols-3 gap-4">
            <div>
              <p className={`text-xs ${C.text50} mb-1`}>期間前繰越</p>
              <p className="text-xl font-bold">
                {formatCurrency(periodSummary.prev_period_carryover)}
              </p>
            </div>
            <div>
              <p className={`text-xs ${C.text50} mb-1`}>期間内未納</p>
              <p className="text-xl font-bold">
                {formatCurrency(periodSummary.current_period_unpaid)}
              </p>
            </div>
            <div>
              <p className={`text-xs ${C.text50} mb-1`}>期末繰越</p>
              <p className="text-xl font-bold">
                {formatCurrency(periodSummary.period_end_carryover)}
              </p>
            </div>
          </div>
        </div>
      ) : null}
    </>
  );
}
