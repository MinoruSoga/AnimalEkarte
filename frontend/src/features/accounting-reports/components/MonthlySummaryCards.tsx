import { memo, type ReactNode } from "react";
import { C } from "@/lib/design-tokens";
import { formatTaxRatePercent } from "@/hooks/use-clinic-tax-rates";
import { formatCurrency } from "@/lib/format/number";
import type { MonthlyReportResponse } from "../api/get-monthly-report";

interface MonthlySummaryCardsProps {
  summary: MonthlyReportResponse["summary"];
  /** #179 ②: 病院マスタ設定の標準税率（例: 0.1）。固定「10%」表記を撤廃。 */
  standardTaxRate: number;
  /** #179 ②: 病院マスタ設定の軽減税率（例: 0.08）。固定「8%」表記を撤廃。 */
  reducedTaxRate: number;
  /** #179 ②: 税率設定導線。ページ側で権限判定して渡す（Router 依存をカードに持ち込まない） */
  taxSettingsLink?: ReactNode;
}

export const MonthlySummaryCards = memo(function MonthlySummaryCards({
  summary,
  standardTaxRate,
  reducedTaxRate,
  taxSettingsLink,
}: MonthlySummaryCardsProps) {
  const topCards = [
    { label: "診療日数", value: `${summary.workingDays}日`, sub: null },
    { label: "会計件数", value: `${summary.totalBillings}件`, sub: null },
    {
      label: "売上合計",
      value: formatCurrency(summary.totalAmount),
      sub: `返金: -${formatCurrency(summary.totalRefund)}`,
    },
    {
      label: "純売上",
      value: formatCurrency(summary.netAmount),
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
                  <span className={`font-medium ${C.text}`}>{formatCurrency(amount)}</span>
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
                  <span className={`font-medium ${C.text}`}>{formatCurrency(amount)}</span>
                </li>
              ))}
            </ul>
          )}
        </div>

        {/* 消費税内訳 */}
        <div className={`${C.bgWhite} rounded-lg border ${C.borderLight} p-4`}>
          <div className="flex items-center justify-between gap-3 mb-2">
            <p className={`text-base font-medium ${C.text70}`}>消費税内訳</p>
            {taxSettingsLink ? (
              <div className="print:hidden">{taxSettingsLink}</div>
            ) : null}
          </div>
          <ul className="space-y-2">
            <li>
              <p className={`text-xs ${C.text40} mb-0.5`}>
                標準税率（{formatTaxRatePercent(standardTaxRate)}）
              </p>
              <div className="flex justify-between text-sm">
                <span className={C.text60}>課税対象</span>
                <span className={C.text}>{formatCurrency(standard.taxableAmount)}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className={C.text60}>消費税</span>
                <span className={C.text}>{formatCurrency(standard.taxAmount)}</span>
              </div>
            </li>
            <li>
              <p className={`text-xs ${C.text40} mb-0.5`}>
                軽減税率（{formatTaxRatePercent(reducedTaxRate)}）
              </p>
              <div className="flex justify-between text-sm">
                <span className={C.text60}>課税対象</span>
                <span className={C.text}>{formatCurrency(reduced.taxableAmount)}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className={C.text60}>消費税</span>
                <span className={C.text}>{formatCurrency(reduced.taxAmount)}</span>
              </div>
            </li>
          </ul>
        </div>
      </div>
    </div>
  );
});
