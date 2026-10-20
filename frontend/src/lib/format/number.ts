/**
 * 数値フォーマットユーティリティ
 */

/**
 * 空値の表記（design-patterns.md §10・2026-10-20 決裁）。U+2014 EM DASH 1 文字。
 * 負の金額の `-¥1,234` と視覚的に区別するため半角ハイフンは使わない。画面側で `"-"` を直書きせずこれを使う。
 */
export const EMPTY_VALUE = "—";

/**
 * 体重をフォーマット（単位付き）
 * @param weight 体重の数値または文字列
 * @returns フォーマットされた体重 (e.g., "26.5kg") または未設定の場合は EMPTY_VALUE（"—"）
 */
export function formatWeight(weight: number | string | undefined | null): string {
  if (weight === undefined || weight === null || weight === "") return EMPTY_VALUE;

  const numWeight = typeof weight === "string" ? parseFloat(weight) : weight;

  if (isNaN(numWeight)) return EMPTY_VALUE;

  // 小数点以下1桁まで表示（整数の場合は.0を省略）
  const formatted = numWeight % 1 === 0 ? numWeight.toString() : numWeight.toFixed(1);

  return `${formatted}kg`;
}

/**
 * 通貨をフォーマット（円記号付き）
 * @param amount 金額
 * @returns フォーマットされた金額 (e.g., "¥1,234")。null/undefined は EMPTY_VALUE（"—"）
 */
export function formatCurrency(amount: number | undefined | null): string {
  if (amount === undefined || amount === null) return EMPTY_VALUE;

  return `${amount < 0 ? "-" : ""}¥${Math.abs(amount).toLocaleString("ja-JP")}`;
}

/**
 * 通貨フォーマット（正の金額のみ表示、0・負値・null/undefined は EMPTY_VALUE "—"）。
 * 画面の `x > 0 ? ¥... : EMPTY_VALUE` インライン分岐の正本（FE5-12）
 */
export function formatCurrencyOrDash(amount: number | undefined | null): string {
  if (amount === undefined || amount === null || amount <= 0) return EMPTY_VALUE;
  return formatCurrency(amount);
}

/**
 * 記録済み金額の空欄表示。0 / 未設定は EMPTY_VALUE "—"、負額は符号のまま出す。
 * 正の金額だけ見せる `formatCurrencyOrDash` とは別（移行の赤伝・負額用）。
 */
export function formatCurrencyIfNonzero(amount: number | undefined | null): string {
  if (amount === undefined || amount === null || amount === 0) return EMPTY_VALUE;
  return formatCurrency(amount);
}
