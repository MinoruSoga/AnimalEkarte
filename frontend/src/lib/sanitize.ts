/**
 * BUG-067: NULL バイト (U+0000) を再帰的に除去する。
 * PostgreSQL は NULL バイトを含む文字列を受け付けないため 500 エラーになる。
 * React に依存しないため main アプリ (axios.ts) と line-reserve/liff 等の
 * 別 axios インスタンスから共有利用できる。
 */
const NULL_BYTE_PATTERN = new RegExp(String.fromCharCode(0), "g");

export function sanitizeNullBytes(value: unknown): unknown {
  if (typeof value === "string") {
    return value.replace(NULL_BYTE_PATTERN, "");
  }
  if (Array.isArray(value)) {
    return value.map(sanitizeNullBytes);
  }
  if (value !== null && typeof value === "object") {
    return Object.fromEntries(
      Object.entries(value as Record<string, unknown>).map(([k, v]) => [k, sanitizeNullBytes(v)])
    );
  }
  return value;
}
