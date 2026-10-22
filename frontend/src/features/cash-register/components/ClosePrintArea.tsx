import { useEffect } from "react";
import { PrintPortal } from "@/components/shared/PrintPortal";
import { C } from "@/lib/design-tokens";
import { formatCurrency, formatNumber } from "@/lib/format/number";
import type { PaymentMethodMaster } from "@/types/generated/models";
import type { CloseBillingDetail } from "../api/get-cash-register-preview";
import {
  buildUnifiedClosingRows,
  buildUnifiedClosingTotals,
  formatClosingCount,
  type UnclassifiedOtherCountInput,
} from "../lib/closing-summary";
import { projectBillingDetailRows } from "../lib/billing-detail-rows";
import { PERIOD_LABELS, type CashRegisterPeriod } from "../lib/constants";
import { BillingDetailTable } from "./BillingDetailTable";

// EMR-354: close 印刷の mount 中に html 要素へ付与する専用クラス。
// globals.css の html/body { overflow:hidden; height:100% } を @media print で
// このクラスに限り解除するためのフック（レガシーブラウザターゲットを考慮し
// :has() へ依存しない class 管理とした）。
const CLOSE_PRINT_ROOT_CLASS = "close-print-active";

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
 * #184 と同じ PrintPortal（共通印刷基盤）を再利用し、A4 横に
 * 統合テーブル（部門別集計）＋消費税内訳＋個別会計明細＋現金照合を縦に並べる。
 * 個別会計明細は行数に応じて複数ページへ改ページされる（EMR-354）。
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

  // EMR-354: globals.css の html/body { height:100%; overflow:hidden } が
  // 印刷時の複数ページ描画を 1 面に切り詰めるため、close 印刷の mount 中だけ
  // html に専用クラスを付け、下の scoped @media print で解除する。
  // アンマウント時に必ず除去し、画面表示・他印刷面への影響はない。
  useEffect(() => {
    const root = document.documentElement;
    root.classList.add(CLOSE_PRINT_ROOT_CLASS);
    return () => {
      root.classList.remove(CLOSE_PRINT_ROOT_CLASS);
    };
  }, []);

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
                <td className={`border ${C.borderGray300} px-1 py-0.5 text-right`}>
                  {formatCurrency(standard.taxableAmount)}
                </td>
                <td className={`border ${C.borderGray300} px-1 py-0.5 text-right`}>
                  {formatCurrency(standard.taxAmount)}
                </td>
              </tr>
              <tr>
                <td className={`border ${C.borderGray300} px-1 py-0.5`}>軽減税率（8%）</td>
                <td className={`border ${C.borderGray300} px-1 py-0.5 text-right`}>
                  {formatCurrency(reduced.taxableAmount)}
                </td>
                <td className={`border ${C.borderGray300} px-1 py-0.5 text-right`}>
                  {formatCurrency(reduced.taxAmount)}
                </td>
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
                <td className={`border ${C.borderGray300} px-1 py-0.5 text-right`}>
                  {formatCurrency(theoreticalCash)}
                </td>
              </tr>
              {actualCash !== null && difference !== null ? (
                <>
                  <tr>
                    <td className={`border ${C.borderGray300} px-1 py-0.5`}>実際の現金</td>
                    <td className={`border ${C.borderGray300} px-1 py-0.5 text-right`}>
                      {formatCurrency(actualCash)}
                    </td>
                  </tr>
                  <tr className="font-semibold">
                    <td className={`border ${C.borderGray300} px-1 py-0.5`}>差額</td>
                    <td className={`border ${C.borderGray300} px-1 py-0.5 text-right`}>
                      {`${difference >= 0 ? "+" : ""}${formatNumber(difference)}`}
                    </td>
                  </tr>
                </>
              ) : null}
            </tbody>
          </table>
        </div>
      </div>

      {/* 個別会計明細（EMR-354: 画面と同一の1会計=1行投影 + フッタ再集計） */}
      <p className="font-semibold text-[9pt] mb-1">
        個別会計明細（{formatNumber(projectBillingDetailRows(billingDetails).length)}件）
      </p>
      <BillingDetailTable details={billingDetails} showFooter footerLabel="個別会計明細 合計" />

      {/* EMR-354: 印刷時の1ページ切り詰め修正。PrintPortal ルートの
          position:fixed/overflow:auto と globals.css の html/body
          { height:100%; overflow:hidden } が印刷時も先頭画面分しか描画しないため、
          close ポータル mount 中に付与する html クラス + この scoped CSS で
          静的配置・可視オーバーフローへ戻す（共有 PrintPortal / globals.css 自体は
          変更しない）。明細表は 7.5pt・固定レイアウト・通常改行で A4 横に収め、
          thead は各ページで繰り返し、tfoot 合計行は最終ページに1回だけ、
          行途中の分割は避ける。 */}
      <style type="text/css">
        {`
          @media print {
            html.${CLOSE_PRINT_ROOT_CLASS},
            html.${CLOSE_PRINT_ROOT_CLASS} body {
              height: auto !important;
              overflow: visible !important;
            }
            [data-testid="close-print-area"] {
              position: static !important;
              inset: auto !important;
              overflow: visible !important;
              height: auto !important;
              padding: 0 !important;
            }
            [data-testid="close-print-area"] .overflow-x-auto {
              overflow: visible !important;
            }
            [data-testid="close-print-area"] .billing-detail-table {
              table-layout: fixed;
              width: 100%;
              font-size: 7.5pt;
            }
            [data-testid="close-print-area"] .billing-detail-table th,
            [data-testid="close-print-area"] .billing-detail-table td {
              padding: 1px 3px;
              /* 共有 TableCell/TableHead の画面用 text-* ユーティリティを
                 印刷では 7.5pt へ上書き（R2-1: セルが画面サイズのまま
                 8桁領収Noを分割していた）。!important は utilities 上書き用。 */
              font-size: 7.5pt !important;
              line-height: 1.3;
              /* break-word は「1行に収まらない単語のみ」途中改行するため、
                 通常の領収No・¥金額はセル内に収まり、超長い氏名等だけ折り返す */
              overflow-wrap: break-word;
              white-space: normal !important;
            }
            [data-testid="close-print-area"] thead {
              display: table-header-group;
            }
            [data-testid="close-print-area"] tfoot {
              display: table-row-group;
            }
            [data-testid="close-print-area"] tr {
              break-inside: avoid;
              page-break-inside: avoid;
            }
          }
        `}
      </style>
    </PrintPortal>
  );
}
