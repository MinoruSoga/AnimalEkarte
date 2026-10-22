/**
 * EMR-447: exact decimal 文字列 wire（`^(0|[1-9][0-9]*)(\.[0-9]+)?$`、api.yaml Medicine）と
 * 画面内 number の明示変換。wire 外の値は黙って NaN/0 へ落とさず例外にする。
 */
const EXACT_DECIMAL_RE = /^(0|[1-9][0-9]*)(\.[0-9]+)?$/;

/** 符号・指数のない非負 decimal 文字列か（wire の字句検査）。 */
export function isExactDecimalString(value: string): boolean {
  return EXACT_DECIMAL_RE.test(value);
}

/** 応答の decimal 文字列を number へ変換する（表示・計算用）。形式違反は例外。 */
export function exactDecimalToNumber(value: string, field: string): number {
  // 型宣言は string だが、実行時の JSON number（旧 wire）は RegExp の文字列化で素通りするため明示的に弾く。
  if (typeof value !== "string" || !EXACT_DECIMAL_RE.test(value)) {
    throw new Error(`${field} is not an exact decimal string`);
  }
  return Number(value);
}

/** 非負の有限 number を送信用の decimal 文字列へ変換する。指数表記になる値・負値・非有限は例外。 */
export function numberToExactDecimal(value: number, field: string): string {
  const text = Number.isFinite(value) ? String(value) : "";
  if (!EXACT_DECIMAL_RE.test(text)) {
    throw new Error(`${field} cannot be sent as an exact decimal string`);
  }
  return text;
}
