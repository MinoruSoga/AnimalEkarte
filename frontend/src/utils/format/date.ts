/**
 * 日付フォーマットユーティリティ
 * ISO形式の日付文字列を日本語形式に変換
 */

/**
 * 日付をYYYY/MM/DD形式にフォーマット
 * @param dateString ISO形式の日付文字列 (e.g., "2024-01-15" or "2024-01-15T00:00:00Z")
 * @returns フォーマットされた日付文字列 (e.g., "2024/01/15") または未設定の場合は "-"
 */
export function formatDate(dateString: string | undefined | null): string {
  if (!dateString) return "-";

  try {
    const date = new Date(dateString);
    if (isNaN(date.getTime())) return "-";

    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, "0");
    const day = String(date.getDate()).padStart(2, "0");

    return `${year}/${month}/${day}`;
  } catch {
    return "-";
  }
}

/**
 * 日付をYYYY年MM月DD日形式にフォーマット
 */
export function formatDateJapanese(
  dateString: string | undefined | null
): string {
  if (!dateString) return "-";

  try {
    const date = new Date(dateString);
    if (isNaN(date.getTime())) return "-";

    const year = date.getFullYear();
    const month = date.getMonth() + 1;
    const day = date.getDate();

    return `${year}年${month}月${day}日`;
  } catch {
    return "-";
  }
}
