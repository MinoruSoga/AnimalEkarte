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
  // PR #186 review (P2-1): FormData は Object.entries で空の plain object に変換されて
  // しまい、multipart アップロード（CSV 等）のファイルパートが失われる。File/Blob エントリは
  // そのまま通し、同居するテキストフィールド（例: purpose, owner_id 等）は再帰的にサニタイズする
  // （セキュリティレビュー指摘: FormData 全体をパススルーすると同居テキストフィールドの
  // NULL バイト除去が漏れ、BUG-067 の再発経路になる）。
  if (typeof FormData !== "undefined" && value instanceof FormData) {
    const sanitized = new FormData();
    for (const [key, entryValue] of value.entries()) {
      sanitized.append(
        key,
        typeof entryValue === "string" ? (sanitizeNullBytes(entryValue) as string) : entryValue,
      );
    }
    return sanitized;
  }
  // Blob/File 単体（FormData に包まれていない生の値）はバイナリデータなのでそのまま返す。
  if (
    (typeof Blob !== "undefined" && value instanceof Blob) ||
    (typeof File !== "undefined" && value instanceof File)
  ) {
    return value;
  }
  if (value !== null && typeof value === "object") {
    return Object.fromEntries(
      Object.entries(value as Record<string, unknown>).map(([k, v]) => [k, sanitizeNullBytes(v)]),
    );
  }
  return value;
}
