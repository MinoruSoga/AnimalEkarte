import type { CloseBillingDetail } from "./api/get-cash-register-preview";
import { DISPLAY_CATEGORIES } from "./constants";

/**
 * #153: レジ締めサマリーの「統合テーブル」1行分。
 * 旧 CategoryPaymentMatrix（部門×支払方法の金額）に件数（会計件数）を統合した。
 */
export interface UnifiedClosingRow {
  /** 部門の表示ラベル（例: 診療 / 外科 / トリミング） */
  label: string;
  /**
   * その部門に属する会計件数。
   * DEC-40: null は旧 snapshot で件数未記録（表示は「記録なし」）。0 扱いや live 再計算をしない。
   */
  count: number | null;
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
 * DEC-40: 未分類・要確認（other）の会計 distinct 件数。
 * - number: サーバ集計値（preview / 新 snapshot）
 * - null: 旧 snapshot でフィールド欠落 → 「記録なし」
 * - undefined: 呼び出し側が件数を持たない（other 行は billingDetails 由来にフォールバック）
 */
export type UnclassifiedOtherCountInput = number | null | undefined;

/**
 * 部門別集計（categories: 部門→支払方法→金額）と個別会計明細（billingDetails）を
 * 統合し、部門ごとに「件数 + 支払方法別金額 + 合計」を持つ統合行を表示順で返す。
 *
 * 画面表示（UnifiedClosingSummaryTable）と印刷／PDF（ClosePrintArea）の双方が
 * この単一関数を描画源とすることで、表示値とPDF出力値の一致を構造的に保証する。
 *
 * 件数優先順位 (#247 DEC-16⑥ / DEC-40):
 * 1. other 行: unclassifiedOtherCount（独立 distinct）
 * 2. 一般行: categoryCounts（サーバ会計 distinct、split 二重計上なし）
 * 3. フォールバック: billingDetails 行数（旧経路・MIN(category)/split 歪みあり）
 */
export function buildUnifiedClosingRows(
  categories: Record<string, Record<string, number>>,
  billingDetails: readonly CloseBillingDetail[],
  unclassifiedOtherCount?: UnclassifiedOtherCountInput,
  categoryCounts?: Record<string, number>,
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

    const isOther = group.keys.includes("other");
    let count: number | null;
    if (isOther && unclassifiedOtherCount !== undefined) {
      // null → 記録なし / number → 独立集計値
      count = unclassifiedOtherCount;
      // サーバ件数が 0 でも金額がある場合は明細から件数を復元する（BUG-006）
      if (count === 0 && rowTotal > 0) {
        count = billingDetails.filter((d) => group.keys.includes(d.category)).length;
      }
    } else if (categoryCounts) {
      // #247: 会計 distinct をカテゴリキー合算（同一会計が複数 key を持つ場合は OR 近似として合算。
      // DISPLAY_CATEGORIES の key は排他グループなので二重計上しない）
      count = group.keys.reduce((sum, key) => sum + (categoryCounts[key] ?? 0), 0);
    } else {
      count = billingDetails.filter((d) => group.keys.includes(d.category)).length;
    }

    return { label: group.label, count, byMethod, rowTotal };
  }).filter((row) => {
    if (row.rowTotal > 0) return true;
    // 件数未記録のみ・金額 0 の行は出さない（旧 snapshot で other 金額も無い場合）
    if (row.count === null) return false;
    return row.count > 0;
  });
}

/**
 * 統合行から列合計（件数合計・支払方法別合計・総合計）を集計する。
 * 件数合計は「記録なし」(null) 行を加算しない。
 */
export function buildUnifiedClosingTotals(
  rows: readonly UnifiedClosingRow[],
): UnifiedClosingTotals {
  const byMethod: Record<string, number> = {};
  let count = 0;
  let grandTotal = 0;
  for (const row of rows) {
    if (row.count != null) count += row.count;
    grandTotal += row.rowTotal;
    for (const [methodName, amount] of Object.entries(row.byMethod)) {
      byMethod[methodName] = (byMethod[methodName] ?? 0) + amount;
    }
  }
  return { count, byMethod, grandTotal };
}

/** 件数セル表示文言（DEC-40: null → 記録なし） */
export function formatClosingCount(count: number | null): string {
  if (count === null) return "記録なし";
  return `${count}件`;
}
