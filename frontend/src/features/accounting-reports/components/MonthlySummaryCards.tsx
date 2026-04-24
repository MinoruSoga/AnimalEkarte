import { memo } from "react";
import { C } from "@/lib/design-tokens";
import type { MonthlyReportResponse } from "../api/get-monthly-report";

interface MonthlySummaryCardsProps {
  summary: MonthlyReportResponse["summary"];
}

export const MonthlySummaryCards = memo(function MonthlySummaryCards({
  summary,
}: MonthlySummaryCardsProps) {
  const topCards = [
    { label: "診療日数", value: `${summary.workingDays}日`, sub: null },
    { label: "会計件数", value: `${summary.totalBillings}件`, sub: null },
    {
      label: "売上合計",
      value: `¥${summary.totalAmount.toLocaleString()}`,
      sub: `返金: -¥${summary.totalRefund.toLocaleString()}`,
    },
    {
      label: "純売上",
      value: `¥${summary.netAmount.toLocaleString()}`,
      sub: null,
    },
  ];

  const { standard, reduced } = summary.taxBreakdown;

  return (
    <div className="space-y-4">
      {/* KPI 4枚 */}
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
        {topCards.map((card) => (
          <div
            key={card.label}
            className={`${C.bgWhite} rounded-lg border ${C.borderLight} p-4 flex flex-col gap-1`}
          >
            <span className={`text-base ${C.text60}`}>{card.label}</span>
            <span className={`text-xl font-semibold ${C.text}`}>{card.value}</span>
            {card.sub !== null ? (
              <span className={`text-sm ${C.text50}`}>{card.sub}</span>
            ) : null}
          </div>
        ))}
      </div>

      {/* 支払方法別・部門別・消費税内訳 */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        {/* 支払方法別合計 */}
        <div className={`${C.bgWhite} rounded-lg border ${C.borderLight} p-4`}>
          <p className={`text-base font-medium ${C.text70} mb-2`}>支払方法別合計</p>
          {Object.keys(summary.byPaymentMethod).length === 0 ? (
            <p className={`text-sm ${C.text40}`}>データなし</p>
          ) : (
            <ul className="space-y-1">
              {Object.entries(summary.byPaymentMethod).map(([method, amount]) => (
                <li key={method} className="flex justify-between text-sm">
                  <span className={C.text60}>{method}</span>
                  <span className={`font-medium ${C.text}`}>¥{amount.toLocaleString()}</span>
                </li>
              ))}
            </ul>
          )}
        </div>

        {/* 部門別合計 */}
        <div className={`${C.bgWhite} rounded-lg border ${C.borderLight} p-4`}>
          <p className={`text-base font-medium ${C.text70} mb-2`}>部門別合計</p>
          {Object.keys(summary.byCategory).length === 0 ? (
            <p className={`text-sm ${C.text40}`}>データなし</p>
          ) : (
            <ul className="space-y-1">
              {Object.entries(summary.byCategory).map(([cat, amount]) => (
                <li key={cat} className="flex justify-between text-sm">
                  <span className={C.text60}>{cat}</span>
                  <span className={`font-medium ${C.text}`}>¥{amount.toLocaleString()}</span>
                </li>
              ))}
            </ul>
          )}
        </div>

        {/* 消費税内訳 */}
        <div className={`${C.bgWhite} rounded-lg border ${C.borderLight} p-4`}>
          <p className={`text-base font-medium ${C.text70} mb-2`}>消費税内訳</p>
          <ul className="space-y-2">
            <li>
              <p className={`text-xs ${C.text40} mb-0.5`}>標準税率（10%）</p>
              <div className="flex justify-between text-sm">
                <span className={C.text60}>課税対象</span>
                <span className={C.text}>¥{standard.taxableAmount.toLocaleString()}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className={C.text60}>消費税</span>
                <span className={C.text}>¥{standard.taxAmount.toLocaleString()}</span>
              </div>
            </li>
            <li>
              <p className={`text-xs ${C.text40} mb-0.5`}>軽減税率（8%）</p>
              <div className="flex justify-between text-sm">
                <span className={C.text60}>課税対象</span>
                <span className={C.text}>¥{reduced.taxableAmount.toLocaleString()}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className={C.text60}>消費税</span>
                <span className={C.text}>¥{reduced.taxAmount.toLocaleString()}</span>
              </div>
            </li>
          </ul>
        </div>
      </div>
    </div>
  );
});
