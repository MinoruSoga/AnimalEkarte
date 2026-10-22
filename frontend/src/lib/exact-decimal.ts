/**
 * EMR-447: exact decimal 文字列 wire（`^(0|[1-9][0-9]*)(\.[0-9]+)?$`、api.yaml Medicine）と
 * 画面内 number の明示変換。wire 外の値は黙って NaN/0 へ落とさず例外にする。
 */
const EXACT_DECIMAL_RE = /^(0|[1-9][0-9]*)(\.[0-9]+)?$/;

/** 応答の decimal 文字列を number へ変換する（表示・計算用）。形式違反は例外。 */
export function exactDecimalToNumber(value: string, field: string): number {
  if (!EXACT_DECIMAL_RE.test(value)) {
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
