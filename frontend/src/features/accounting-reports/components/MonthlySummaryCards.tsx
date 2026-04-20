import { memo } from "react";
import { C } from "@/lib/design-tokens";
import type { MonthlyReportResponse } from "@/types/generated/models";

interface MonthlySummaryCardsProps {
  summary: MonthlyReportResponse["summary"];
}

export const MonthlySummaryCards = memo(function MonthlySummaryCards({
  summary,
}: MonthlySummaryCardsProps) {
  const cards = [
    { label: "診療日数", value: `${summary.working_days}日`, sub: null },
    { label: "会計件数", value: `${summary.total_billings}件`, sub: null },
    {
      label: "売上合計",
      value: `¥${summary.total_amount.toLocaleString()}`,
      sub: `返金: -¥${summary.total_refund.toLocaleString()}`,
    },
    {
      label: "純売上",
      value: `¥${summary.net_amount.toLocaleString()}`,
      sub: null,
    },
  ];

  return (
    <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
      {cards.map((card) => (
        <div
          key={card.label}
          className={`bg-white rounded-lg border ${C.borderLight} p-4 flex flex-col gap-1`}
        >
          <span className={`text-base ${C.text60}`}>{card.label}</span>
          <span className={`text-xl font-semibold ${C.text}`}>{card.value}</span>
          {card.sub !== null ? (
            <span className={`text-sm ${C.text50}`}>{card.sub}</span>
          ) : null}
        </div>
      ))}
    </div>
  );
});
