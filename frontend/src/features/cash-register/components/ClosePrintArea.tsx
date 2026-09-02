import { PrintPortal } from "@/components/shared/PrintPortal";
import { C } from "@/lib/design-tokens";
import { formatJSTTime } from "@/lib/jst-date";
import { formatCurrency } from "@/lib/format/number";
import type { PaymentMethodMaster } from "@/types/generated/models";
import type { CloseBillingDetail } from "../api/get-cash-register-preview";
import {
  buildUnifiedClosingRows,
  buildUnifiedClosingTotals,
  formatClosingCount,
  type UnclassifiedOtherCountInput,
} from "../closing-summary";
import { CATEGORY_LABELS, PERIOD_LABELS, type CashRegisterPeriod } from "../constants";

interface TaxEntry {
  taxableAmount: number;
  taxAmount: number;
}

interface ClosePrintAreaProps {
  date: string;
  period: CashRegisterPeriod;
  clinicName: string;
  categories: Record<string, Record<string, number>>;
  paymentMethods: PaymentMethodMaster[];
  billingDetails: CloseBillingDetail[];
  taxBreakdown: { standard: TaxEntry; reduced: TaxEntry };
  theoreticalCash: number;
  /** 実際のレジ現金。未入力時は null（差額は出力しない）。 */
  actualCash: number | null;
  /** DEC-40: other 行の会計 distinct 件数。null = 記録なし */
  unclassifiedOtherCount?: UnclassifiedOtherCountInput;
  /** #247: 部門ごとの会計 distinct 件数 */
  categoryCounts?: Record<string, number>;
}

/**
 * #153: レジ締めサマリーの印刷／PDF出力（ブラウザ「PDFとして保存」）用ビュー。
 *
 * #184 と同じ PrintPortal（共通印刷基盤）を再利用し、A4 横 1 面に
 * 統合テーブル（部門別集計）＋消費税内訳＋個別会計明細＋現金照合を縦に並べる。
 * 画面表示（UnifiedClosingSummaryTable / BillingDetailTable）と同一データ源
 * （buildUnifiedClosingRows / billingDetails）を描画し、表示値とPDF値を一致させる。
 */
export function ClosePrintArea({
  date,
  period,
  clinicName,
  categories,
  paymentMethods,
  billingDetails,
  taxBreakdown,
  theoreticalCash,
  actualCash,
  unclassifiedOtherCount,
  categoryCounts,
}: ClosePrintAreaProps) {
  const rows = buildUnifiedClosingRows(
    categories,
    billingDetails,
    unclassifiedOtherCount,
    categoryCounts,
  );
  const totals = buildUnifiedClosingTotals(rows);
  const { standard, reduced } = taxBreakdown;
  const difference = actualCash !== null ? actualCash - theoreticalCash : null;

  return (
    <PrintPortal testId="close-print-area" orientation="landscape">
      {/* ヘッダー */}
      <div className="mb-3 text-center">
        <h1 className="text-[14pt] font-bold">レジ締めサマリー</h1>
        {clinicName ? <p className="text-[10pt]">{clinicName}</p> : null}
        <p className="text-[10pt]">
          対象日: {date} 区分: {PERIOD_LABELS[period]}
        </p>
      </div>

      {/* 部門別集計（統合テーブル） */}
      <p className="font-semibold text-[9pt] mb-1">部門別集計</p>
      <table className="w-full text-[9pt] border-collapse mb-3">
        <thead>
          <tr className={`${C.bgGray100}`}>
            <th className={`border ${C.borderGray300} px-1 py-0.5 text-left`}>部門</th>
            <th className={`border ${C.borderGray300} px-1 py-0.5 text-right`}>件数</th>
            {paymentMethods.map((pm) => (
              <th key={pm.id} className={`border ${C.borderGray300} px-1 py-0.5 text-right`}>
                {pm.name}
              </th>
            ))}
            <th className={`border ${C.borderGray300} px-1 py-0.5 text-right`}>合計</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr key={row.label}>
              <td className={`border ${C.borderGray300} px-1 py-0.5`}>{row.label}</td>
              <td className={`border ${C.borderGray300} px-1 py-0.5 text-right`}>
                {formatClosingCount(row.count)}
              </td>
              {paymentMethods.map((pm) => (
                <td key={pm.id} className={`border ${C.borderGray300} px-1 py-0.5 text-right`}>
                  {row.byMethod[pm.name] != null ? formatCurrency(row.byMethod[pm.name]) : "—"}
                </td>
              ))}
              <td className={`border ${C.borderGray300} px-1 py-0.5 text-right font-medium`}>
                {formatCurrency(row.rowTotal)}
              </td>
            </tr>
          ))}
        </tbody>
        <tfoot>
          <tr className={`${C.bgMuted} font-semibold`}>
            <td className={`border ${C.borderGray300} px-1 py-0.5`}>合計</td>
            <td className={`border ${C.borderGray300} px-1 py-0.5 text-right`}>
              {formatClosingCount(totals.count)}
            </td>
            {paymentMethods.map((pm) => (
              <td key={pm.id} className={`border ${C.borderGray300} px-1 py-0.5 text-right`}>
                {totals.byMethod[pm.name] != null ? formatCurrency(totals.byMethod[pm.name]) : "—"}
              </td>
            ))}
            <td className={`border ${C.borderGray300} px-1 py-0.5 text-right font-bold`}>
              {formatCurrency(totals.grandTotal)}
            </td>
          </tr>
        </tfoot>
      </table>

      {/* 消費税内訳 + 現金照合 */}
      <div className="flex gap-4 mb-3 text-[9pt]">
        <div className="flex-1">
          <p className="font-semibold mb-1">消費税内訳</p>
          <table className="w-full border-collapse">
            <thead>
              <tr className={`${C.bgGray100}`}>
                <th className={`border ${C.borderGray300} px-1 py-0.5 text-left`}>区分</th>
                <th className={`border ${C.borderGray300} px-1 py-0.5 text-right`}>課税対象額</th>
                <th className={`border ${C.borderGray300} px-1 py-0.5 text-right`}>税額</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td className={`border ${C.borderGray300} px-1 py-0.5`}>標準税率（10%）</td>
                <td className={`border ${C.borderGray300} px-1 py-0.5 text-right`}>{formatCurrency(standard.taxableAmount)}</td>
                <td className={`border ${C.borderGray300} px-1 py-0.5 text-right`}>{formatCurrency(standard.taxAmount)}</td>
              </tr>
              <tr>
                <td className={`border ${C.borderGray300} px-1 py-0.5`}>軽減税率（8%）</td>
                <td className={`border ${C.borderGray300} px-1 py-0.5 text-right`}>{formatCurrency(reduced.taxableAmount)}</td>
                <td className={`border ${C.borderGray300} px-1 py-0.5 text-right`}>{formatCurrency(reduced.taxAmount)}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <div className="flex-1">
          <p className="font-semibold mb-1">現金照合</p>
          <table className="w-full border-collapse">
            <tbody>
              <tr>
                <td className={`border ${C.borderGray300} px-1 py-0.5`}>理論現金</td>
                <td className={`border ${C.borderGray300} px-1 py-0.5 text-right`}>{formatCurrency(theoreticalCash)}</td>
              </tr>
              {actualCash !== null && difference !== null ? (
                <>
                  <tr>
                    <td className={`border ${C.borderGray300} px-1 py-0.5`}>実際の現金</td>
                    <td className={`border ${C.borderGray300} px-1 py-0.5 text-right`}>{formatCurrency(actualCash)}</td>
                  </tr>
                  <tr className="font-semibold">
                    <td className={`border ${C.borderGray300} px-1 py-0.5`}>差額</td>
                    <td className={`border ${C.borderGray300} px-1 py-0.5 text-right`}>
                      {`${difference >= 0 ? "+" : ""}${difference.toLocaleString()}`}
                    </td>
                  </tr>
                </>
              ) : null}
            </tbody>
          </table>
        </div>
      </div>

      {/* 個別会計明細 */}
      <p className="font-semibold text-[9pt] mb-1">個別会計明細（{billingDetails.length}件）</p>
      <table className="w-full text-[8pt] border-collapse">
        <thead>
          <tr className={`${C.bgGray100}`}>
            <th className={`border ${C.borderGray300} px-1 py-0.5 text-left`}>時刻</th>
            <th className={`border ${C.borderGray300} px-1 py-0.5 text-left`}>飼主 / ペット</th>
            <th className={`border ${C.borderGray300} px-1 py-0.5 text-left`}>部門</th>
            <th className={`border ${C.borderGray300} px-1 py-0.5 text-left`}>支払方法</th>
            <th className={`border ${C.borderGray300} px-1 py-0.5 text-right`}>請求額</th>
            <th className={`border ${C.borderGray300} px-1 py-0.5 text-right`}>返金額</th>
            <th className={`border ${C.borderGray300} px-1 py-0.5 text-right`}>純額</th>
          </tr>
        </thead>
        <tbody>
          {billingDetails.map((d) => (
            <tr key={d.billingId}>
              <td className={`border ${C.borderGray300} px-1 py-0.5`}>
                {d.paidAt ? formatJSTTime(d.paidAt) : "—"}
              </td>
              <td className={`border ${C.borderGray300} px-1 py-0.5`}>
                {d.ownerName} / {d.petName}
                {d.isHospitalization ? "（入院）" : ""}
              </td>
              <td className={`border ${C.borderGray300} px-1 py-0.5`}>
                {CATEGORY_LABELS[d.category] ?? d.category}
              </td>
              <td className={`border ${C.borderGray300} px-1 py-0.5`}>{d.paymentMethodName}</td>
              <td className={`border ${C.borderGray300} px-1 py-0.5 text-right`}>{formatCurrency(d.billingAmount)}</td>
              <td className={`border ${C.borderGray300} px-1 py-0.5 text-right`}>
                {d.refundAmount > 0 ? `-${formatCurrency(d.refundAmount)}` : "—"}
              </td>
              <td className={`border ${C.borderGray300} px-1 py-0.5 text-right font-medium`}>{formatCurrency(d.netAmount)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </PrintPortal>
  );
}
