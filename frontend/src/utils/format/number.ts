/**
 * 数値フォーマットユーティリティ
 */

/**
 * 体重をフォーマット（単位付き）
 * @param weight 体重の数値または文字列
 * @returns フォーマットされた体重 (e.g., "26.5kg") または未設定の場合は "-"
 */
export function formatWeight(weight: number | string | undefined | null): string {
  if (weight === undefined || weight === null || weight === "") return "-";

  const numWeight = typeof weight === "string" ? parseFloat(weight) : weight;

  if (isNaN(numWeight)) return "-";

  // 小数点以下1桁まで表示（整数の場合は.0を省略）
  const formatted = numWeight % 1 === 0 ? numWeight.toString() : numWeight.toFixed(1);

  return `${formatted}kg`;
}

/**
 * 通貨をフォーマット（円記号付き）
 * @param amount 金額
 * @returns フォーマットされた金額 (e.g., "¥1,234")
 */
export function formatCurrency(amount: number | undefined | null): string {
  if (amount === undefined || amount === null) return "-";

  return `¥${amount.toLocaleString("ja-JP")}`;
}
