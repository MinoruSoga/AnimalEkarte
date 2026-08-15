import { PrintPortal } from "@/components/shared/PrintPortal";
import { formatTaxRatePercent } from "@/hooks/use-clinic-tax-rates";
import { formatCurrency } from "@/lib/format/number";
import type {
  CategoryPaymentMatrix,
  MonthlyReportResponse,
} from "../api/get-monthly-report";
import { CategoryPaymentMatrixTable } from "./CategoryPaymentMatrixTable";

interface MonthlyReportPrintAreaProps {
  periodLabel: string;
  clinicName: string;
  summary: MonthlyReportResponse["summary"];
  dailyDetails: MonthlyReportResponse["dailyDetails"];
  /** #179 ②: 病院マスタ設定の標準税率。印刷面でも固定「10%」を撤廃。 */
  standardTaxRate: number;
  /** #179 ②: 病院マスタ設定の軽減税率。印刷面でも固定「8%」を撤廃。 */
  reducedTaxRate: number;
  /** #247: 部門×支払方法（画面と同一データ） */
  categoryPaymentMatrix?: CategoryPaymentMatrix | null;
}

/**
 * #184: 月次集計レポートの印刷／PDF出力（ブラウザ「PDFとして保存」）用ビュー。
 *
 * 画面表示（MonthlySummaryCards / DailyBreakdownTable）と同一の `summary` /
 * `dailyDetails` を描画源とし、印刷値が画面値と一致することを構造的に保証する。
 * A4 横（日次明細が横長のため）で対象年月・病院名・サマリー・日次明細を出力する。
 */
export function MonthlyReportPrintArea({
  periodLabel,
  clinicName,
  summary,
  dailyDetails,
  standardTaxRate,
  reducedTaxRate,
  categoryPaymentMatrix = null,
}: MonthlyReportPrintAreaProps) {
  const { standard, reduced } = summary.taxBreakdown;
  const paymentEntries = Object.entries(summary.byPaymentMethod);

  return (
    <PrintPortal testId="monthly-report-print-area" orientation="landscape">
      {/* ヘッダー */}
      <div className="mb-3 text-center">
        <h1 className="text-[14pt] font-bold">月次集計レポート</h1>
        {clinicName ? <p className="text-[10pt]">{clinicName}</p> : null}
        <p className="text-[10pt]">対象期間: {periodLabel}</p>
      </div>

      {/* サマリー KPI */}
      <table className="w-full text-[9pt] border-collapse mb-3">
        <tbody>
          <tr>
            <th className="border border-gray-400 bg-gray-100 px-2 py-1 text-left">診療日数</th>
            <td className="border border-gray-300 px-2 py-1 text-right">{summary.workingDays}日</td>
            <th className="border border-gray-400 bg-gray-100 px-2 py-1 text-left">会計件数</th>
            <td className="border border-gray-300 px-2 py-1 text-right">{summary.totalBillings}件</td>
            <th className="border border-gray-400 bg-gray-100 px-2 py-1 text-left">売上合計</th>
            <td className="border border-gray-300 px-2 py-1 text-right">{formatCurrency(summary.totalAmount)}</td>
            <th className="border border-gray-400 bg-gray-100 px-2 py-1 text-left">返金</th>
            <td className="border border-gray-300 px-2 py-1 text-right">-{formatCurrency(summary.totalRefund)}</td>
            <th className="border border-gray-400 bg-gray-100 px-2 py-1 text-left">純売上</th>
            <td className="border border-gray-300 px-2 py-1 text-right font-bold">{formatCurrency(summary.netAmount)}</td>
          </tr>
        </tbody>
      </table>

      {/* 支払方法別 + 消費税内訳 */}
      <div className="flex gap-4 mb-3 text-[9pt]">
        <div className="flex-1">
          <p className="font-semibold mb-1">支払方法別集計</p>
          <table className="w-full border-collapse">
            <tbody>
              {paymentEntries.length === 0 ? (
                <tr>
                  <td className="border border-gray-300 px-2 py-1 text-gray-500">データなし</td>
                </tr>
              ) : (
                paymentEntries.map(([method, amount]) => (
                  <tr key={method}>
                    <td className="border border-gray-300 px-2 py-1">{method}</td>
                    <td className="border border-gray-300 px-2 py-1 text-right">{formatCurrency(amount)}</td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
        <div className="flex-1">
          <p className="font-semibold mb-1">消費税内訳</p>
          <table className="w-full border-collapse">
            <thead>
              <tr className="bg-gray-100">
                <th className="border border-gray-400 px-2 py-1 text-left">区分</th>
                <th className="border border-gray-400 px-2 py-1 text-right">課税対象額</th>
                <th className="border border-gray-400 px-2 py-1 text-right">税額</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td className="border border-gray-300 px-2 py-1">
                  標準税率（{formatTaxRatePercent(standardTaxRate)}）
                </td>
                <td className="border border-gray-300 px-2 py-1 text-right">{formatCurrency(standard.taxableAmount)}</td>
                <td className="border border-gray-300 px-2 py-1 text-right">{formatCurrency(standard.taxAmount)}</td>
              </tr>
              <tr>
                <td className="border border-gray-300 px-2 py-1">
                  軽減税率（{formatTaxRatePercent(reducedTaxRate)}）
                </td>
                <td className="border border-gray-300 px-2 py-1 text-right">{formatCurrency(reduced.taxableAmount)}</td>
                <td className="border border-gray-300 px-2 py-1 text-right">{formatCurrency(reduced.taxAmount)}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      {/* #247 部門×支払方法（画面と同一 matrix） */}
      {categoryPaymentMatrix ? (
        <div className="mb-3">
          <p className="font-semibold text-[9pt] mb-1">部門×支払方法</p>
          <CategoryPaymentMatrixTable matrix={categoryPaymentMatrix} compact />
        </div>
      ) : null}

      {/* 日次明細 */}
      <p className="font-semibold text-[9pt] mb-1">日次明細</p>
      <table className="w-full text-[8pt] border-collapse">
        <thead>
          <tr className="bg-gray-100">
            <th className="border border-gray-400 px-1 py-0.5 text-left">日付</th>
            <th className="border border-gray-400 px-1 py-0.5 text-left">曜日</th>
            <th className="border border-gray-400 px-1 py-0.5 text-right">午前件数</th>
            <th className="border border-gray-400 px-1 py-0.5 text-right">午前売上</th>
            <th className="border border-gray-400 px-1 py-0.5 text-right">午後件数</th>
            <th className="border border-gray-400 px-1 py-0.5 text-right">午後売上</th>
            <th className="border border-gray-400 px-1 py-0.5 text-right">日計</th>
            <th className="border border-gray-400 px-1 py-0.5 text-right">返金</th>
          </tr>
        </thead>
        <tbody>
          {dailyDetails.map((d) => (
            <tr key={d.date} className={d.isHoliday ? "bg-gray-50" : undefined}>
              <td className="border border-gray-300 px-1 py-0.5">{d.date}</td>
              <td className="border border-gray-300 px-1 py-0.5">{d.weekday}</td>
              <td className="border border-gray-300 px-1 py-0.5 text-right">{d.amCount}</td>
              <td className="border border-gray-300 px-1 py-0.5 text-right">{formatCurrency(d.amNet)}</td>
              <td className="border border-gray-300 px-1 py-0.5 text-right">{d.pmCount}</td>
              <td className="border border-gray-300 px-1 py-0.5 text-right">{formatCurrency(d.pmNet)}</td>
              <td className="border border-gray-300 px-1 py-0.5 text-right font-medium">{formatCurrency(d.dayNet)}</td>
              <td className="border border-gray-300 px-1 py-0.5 text-right">
                {d.refund > 0 ? `-${formatCurrency(d.refund)}` : "—"}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </PrintPortal>
  );
}
