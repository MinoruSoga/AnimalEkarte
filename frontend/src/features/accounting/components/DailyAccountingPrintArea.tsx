import { createPortal } from "react-dom";

import { Z } from "@/lib/design-tokens";
import { PAYMENT_METHOD_LABELS } from "@/constants/payment-method";
import { formatCurrency, formatCurrencyOrDash } from "@/lib/format/number";
import { CatCell } from "./DailyAccountingTabParts";
import { formatReceiptNo } from "./daily-accounting-utils";
import type { RowData, TotalsData } from "./daily-accounting-utils";

interface DailyPrintAreaProps {
  date: string;
  rows: RowData[];
  totals: TotalsData;
}

export function DailyPrintArea({ date, rows, totals }: DailyPrintAreaProps) {
  const hospitalTotal = totals.medical + totals.surgery + totals.rv + totals.food + totals.goods;
  const trimmingTotal = totals.trimming + totals.hotel;

  return createPortal(
    <div
      hidden
      className="bg-white"
      data-testid="daily-print-area"
      style={{ position: "fixed", inset: 0, zIndex: Z.overlay, overflow: "auto", padding: "8mm" }}
    >
      <style type="text/css">
        {`
          @media print {
            [data-testid="daily-print-area"] {
              display: block !important;
            }
            body > :not([data-testid="daily-print-area"]) {
              display: none !important;
            }
          }
          @page { size: A4 landscape; margin: 8mm; }
          @media print { body { margin: 0; -webkit-print-color-adjust: exact; print-color-adjust: exact; } }
        `}
      </style>

      {/* ヘッダー */}
      <div className="mb-3 text-center">
        <h1 className="text-[14pt] font-bold">日次集計一覧表</h1>
        <p className="text-[10pt]">対象日: {date} 件数: {rows.length}件</p>
      </div>

      {/* メインテーブル */}
      <table className="w-full text-[8pt] border-collapse">
        <thead>
          <tr className="bg-gray-100">
            <th className="border border-gray-400 px-1 py-0.5 text-left whitespace-nowrap">領収No</th>
            <th className="border border-gray-400 px-1 py-0.5 text-left whitespace-nowrap">飼主名</th>
            <th className="border border-gray-400 px-1 py-0.5 text-left whitespace-nowrap">ペット名</th>
            <th className="border border-gray-400 px-1 py-0.5 text-right whitespace-nowrap">診療</th>
            <th className="border border-gray-400 px-1 py-0.5 text-right whitespace-nowrap">外科</th>
            <th className="border border-gray-400 px-1 py-0.5 text-right whitespace-nowrap">RV</th>
            <th className="border border-gray-400 px-1 py-0.5 text-right whitespace-nowrap">フード</th>
            <th className="border border-gray-400 px-1 py-0.5 text-right whitespace-nowrap">トリミング</th>
            <th className="border border-gray-400 px-1 py-0.5 text-right whitespace-nowrap">ホテル</th>
            <th className="border border-gray-400 px-1 py-0.5 text-right whitespace-nowrap">用品他</th>
            <th className="border border-gray-400 px-1 py-0.5 text-center whitespace-nowrap">支払方法</th>
            <th className="border border-gray-400 px-1 py-0.5 text-right whitespace-nowrap">合計</th>
          </tr>
        </thead>
        <tbody>
          {rows.map(({ accounting: a, detailedBreakdown, total }) => {
            const isMixed = Boolean(a.paymentSplits && a.paymentSplits.length > 1);
            const paymentLabel = isMixed
              ? a.paymentSplits!.map((s) => PAYMENT_METHOD_LABELS[s.method] ?? s.method).join("/")
              : a.payment
                ? (PAYMENT_METHOD_LABELS[a.payment.method] ?? a.payment.method)
                : "-";

            return (
              <tr key={a.id}>
                <td className="border border-gray-300 px-1 py-0.5 text-[9pt] font-mono">
                  {formatReceiptNo(a.id)}
                </td>
                <td className="border border-gray-300 px-1 py-0.5 text-[9pt]">{a.ownerName}</td>
                <td className="border border-gray-300 px-1 py-0.5 text-[9pt]">{a.petName}</td>
                <CatCell detail={detailedBreakdown.medical} isMixed={isMixed} />
                <CatCell detail={detailedBreakdown.surgery} isMixed={isMixed} />
                <CatCell detail={detailedBreakdown.rv} isMixed={isMixed} />
                <CatCell detail={detailedBreakdown.food} isMixed={isMixed} />
                <CatCell detail={detailedBreakdown.trimming} isMixed={isMixed} />
                <CatCell detail={detailedBreakdown.hotel} isMixed={isMixed} />
                <CatCell detail={detailedBreakdown.goods} isMixed={isMixed} />
                <td className="border border-gray-300 px-1 py-0.5 text-[9pt] text-center whitespace-nowrap">
                  {paymentLabel}
                </td>
                <td className="border border-gray-300 px-1 py-0.5 text-right text-[9pt] font-semibold">
                  {formatCurrency(total)}
                </td>
              </tr>
            );
          })}
        </tbody>
        <tfoot>
          {/* 病院合計行 */}
          <tr className="bg-gray-50 font-semibold">
            <td colSpan={3} className="border border-gray-400 px-1 py-0.5 text-[9pt]">病院合計</td>
            <td className="border border-gray-400 px-1 py-0.5 text-right text-[9pt]">
              {formatCurrencyOrDash(totals.medical)}
            </td>
            <td className="border border-gray-400 px-1 py-0.5 text-right text-[9pt]">
              {formatCurrencyOrDash(totals.surgery)}
            </td>
            <td className="border border-gray-400 px-1 py-0.5 text-right text-[9pt]">
              {formatCurrencyOrDash(totals.rv)}
            </td>
            <td className="border border-gray-400 px-1 py-0.5 text-right text-[9pt]">
              {formatCurrencyOrDash(totals.food)}
            </td>
            <td className="border border-gray-400 px-1 py-0.5 text-center text-[9pt]">-</td>
            <td className="border border-gray-400 px-1 py-0.5 text-center text-[9pt]">-</td>
            <td className="border border-gray-400 px-1 py-0.5 text-right text-[9pt]">
              {formatCurrencyOrDash(totals.goods)}
            </td>
            <td className="border border-gray-400 px-1 py-0.5 text-[9pt]" />
            <td className="border border-gray-400 px-1 py-0.5 text-right text-[9pt] font-bold">
              {formatCurrency(hospitalTotal)}
            </td>
          </tr>
          {/* トリミング合計行 */}
          <tr className="bg-gray-50 font-semibold">
            <td colSpan={3} className="border border-gray-400 px-1 py-0.5 text-[9pt]">トリミング合計</td>
            <td className="border border-gray-400 px-1 py-0.5 text-center text-[9pt]">-</td>
            <td className="border border-gray-400 px-1 py-0.5 text-center text-[9pt]">-</td>
            <td className="border border-gray-400 px-1 py-0.5 text-center text-[9pt]">-</td>
            <td className="border border-gray-400 px-1 py-0.5 text-center text-[9pt]">-</td>
            <td className="border border-gray-400 px-1 py-0.5 text-right text-[9pt]">
              {formatCurrencyOrDash(totals.trimming)}
            </td>
            <td className="border border-gray-400 px-1 py-0.5 text-right text-[9pt]">
              {formatCurrencyOrDash(totals.hotel)}
            </td>
            <td className="border border-gray-400 px-1 py-0.5 text-center text-[9pt]">-</td>
            <td className="border border-gray-400 px-1 py-0.5 text-[9pt]" />
            <td className="border border-gray-400 px-1 py-0.5 text-right text-[9pt] font-bold">
              {formatCurrency(trimmingTotal)}
            </td>
          </tr>
          {/* 全体合計行 */}
          <tr className="bg-gray-200 font-bold">
            <td colSpan={3} className="border border-gray-400 px-1 py-1 text-[9pt]">全体合計</td>
            <td className="border border-gray-400 px-1 py-1 text-right text-[9pt]">
              {formatCurrencyOrDash(totals.medical)}
            </td>
            <td className="border border-gray-400 px-1 py-1 text-right text-[9pt]">
              {formatCurrencyOrDash(totals.surgery)}
            </td>
            <td className="border border-gray-400 px-1 py-1 text-right text-[9pt]">
              {formatCurrencyOrDash(totals.rv)}
            </td>
            <td className="border border-gray-400 px-1 py-1 text-right text-[9pt]">
              {formatCurrencyOrDash(totals.food)}
            </td>
            <td className="border border-gray-400 px-1 py-1 text-right text-[9pt]">
              {formatCurrencyOrDash(totals.trimming)}
            </td>
            <td className="border border-gray-400 px-1 py-1 text-right text-[9pt]">
              {formatCurrencyOrDash(totals.hotel)}
            </td>
            <td className="border border-gray-400 px-1 py-1 text-right text-[9pt]">
              {formatCurrencyOrDash(totals.goods)}
            </td>
            <td className="border border-gray-400 px-1 py-1 text-[9pt]" />
            <td className="border border-gray-400 px-1 py-1 text-right text-[10pt] font-bold">
              {formatCurrency(totals.total)}
            </td>
          </tr>
        </tfoot>
      </table>
    </div>,
    document.body,
  );
}
