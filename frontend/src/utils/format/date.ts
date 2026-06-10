import { formatJSTDate } from "@/lib/jst-date";

/**
 * 日付をYYYY/MM/DD形式にフォーマット
 * @param dateString ISO形式の日付文字列 (e.g., "2024-01-15" or "2024-01-15T00:00:00+09:00")
 * @returns フォーマットされた日付文字列 (e.g., "2024/01/15") または未設定の場合は "-"
 */
export function formatDate(dateString: string | undefined | null): string {
  if (!dateString) return "-";

  try {
    const formatted = formatJSTDate(dateString);
    if (formatted === "NaN-NaN-NaN") return "-";

    const [year, month, day] = formatted.split("-");

    return `${year}/${month}/${day}`;
  } catch {
    return "-";
  }
}
