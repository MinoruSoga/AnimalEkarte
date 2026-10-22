/**
 * 数値フォーマットユーティリティ
 */

import { EMPTY_VALUE } from "@/constants/empty-value";

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
 * 整数・件数などの数値を 3 桁区切りで整形する（"1,234"）。
 * 画面側で `toLocaleString()` を直書きしない（design-patterns §10）。単位は付けない。
 */
export function formatNumber(value: number): string {
  return value.toLocaleString("ja-JP");
}

/**
 * 件数表記（"1,234件"）。数字と単位の間に空白を入れない（design-patterns §10）。
 * @param unit 単位。既定は「件」。名・頭・日などは呼び出し側で指定する
 */
export function formatCount(count: number, unit = "件"): string {
  return `${formatNumber(count)}${unit}`;
}

/**
 * 小数桁を固定して整形する（"38.5"）。画面側で `toFixed` を直書きしない。
 * 単位（℃ / mg / %）は呼び出し側で付ける。
 */
export function formatDecimal(value: number, fractionDigits: number): string {
  return value.toFixed(fractionDigits);
}

/**
 * 比率を百分率で整形する（0.123 → "12.3%"）。
 */
export function formatPercent(rate: number, fractionDigits = 1): string {
  return `${formatDecimal(rate * 100, fractionDigits)}%`;
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
