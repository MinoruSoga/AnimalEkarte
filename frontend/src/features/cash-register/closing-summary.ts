import type { CloseBillingDetail } from "./api/get-cash-register-preview";
import { DISPLAY_CATEGORIES } from "./constants";

/**
 * #153: レジ締めサマリーの「統合テーブル」1行分。
 * 旧 CategoryPaymentMatrix（部門×支払方法の金額）に件数（会計件数）を統合した。
 */
export interface UnifiedClosingRow {
  /** 部門の表示ラベル（例: 診療 / 外科 / トリミング） */
  label: string;
  /** その部門に属する会計件数 */
  count: number;
  /** 支払方法名 → 金額 */
  byMethod: Record<string, number>;
  /** 部門合計金額 */
  rowTotal: number;
}

export interface UnifiedClosingTotals {
  count: number;
  byMethod: Record<string, number>;
  grandTotal: number;
}

/**
 * 部門別集計（categories: 部門→支払方法→金額）と個別会計明細（billingDetails）を
 * 統合し、部門ごとに「件数 + 支払方法別金額 + 合計」を持つ統合行を表示順で返す。
 *
 * 画面表示（UnifiedClosingSummaryTable）と印刷／PDF（ClosePrintArea）の双方が
 * この単一関数を描画源とすることで、表示値とPDF出力値の一致を構造的に保証する。
 *
 * 注: 件数は billingDetails、金額は categories マトリクスという別系統のサーバ値から
 * 導出する（クライアントでは突合しない）。両者の整合はサーバ側の集計責務とする。
 */
export function buildUnifiedClosingRows(
  categories: Record<string, Record<string, number>>,
  billingDetails: readonly CloseBillingDetail[],
): UnifiedClosingRow[] {
  return DISPLAY_CATEGORIES.map((group) => {
    const byMethod: Record<string, number> = {};
    let rowTotal = 0;
    for (const key of group.keys) {
      const methodMap = categories[key] ?? {};
      for (const [methodName, amount] of Object.entries(methodMap)) {
        byMethod[methodName] = (byMethod[methodName] ?? 0) + amount;
        rowTotal += amount;
      }
    }
    const count = billingDetails.filter((d) => group.keys.includes(d.category)).length;
    return { label: group.label, count, byMethod, rowTotal };
  }).filter((row) => row.rowTotal > 0 || row.count > 0);
}

/**
 * 統合行から列合計（件数合計・支払方法別合計・総合計）を集計する。
 */
export function buildUnifiedClosingTotals(
  rows: readonly UnifiedClosingRow[],
): UnifiedClosingTotals {
  const byMethod: Record<string, number> = {};
  let count = 0;
  let grandTotal = 0;
  for (const row of rows) {
    count += row.count;
    grandTotal += row.rowTotal;
    for (const [methodName, amount] of Object.entries(row.byMethod)) {
      byMethod[methodName] = (byMethod[methodName] ?? 0) + amount;
    }
  }
  return { count, byMethod, grandTotal };
}
