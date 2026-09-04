import { formatJSTDate } from "@/lib/jst-date";
import { DAY_OF_WEEK_LABEL_LIST } from "@/constants/day-of-week";

/** 表示用の標準日付・時刻フォーマット(PO決定2026-07-12。wire用 "yyyy-MM-dd" とは別物) */
export const DISPLAY_DATE_FORMAT = "yyyy/MM/dd";
export const DISPLAY_TIME_FORMAT = "HH:mm";

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

/**
 * Date を「Y年M月D日（曜）」形式にフォーマット（月・日は 0 パディングしない）。
 * FE4-8: DatePickerModel.formatDisplay の逐語移設（ローカル getter 実装）。
 * "YYYY-MM-DD"→Date の parse 契約は 2 系統併存する（DatePicker=ローカル正午 / line-reserve=UTC 深夜）。
 * 本関数はローカル getter で整形するため、渡す Date は DatePickerModel.parseLocalDate 等の
 * ローカル正午 parse から得たものを想定する（line-reserve 側の UTC 深夜 Date を渡すと TZ 依存でずれる）。
 */
/** 表示用のスラッシュ日付 + 曜日（予約モーダル等）。正本は yyyy/MM/dd。 */
export function formatDateWithWeekday(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  const weekday = DAY_OF_WEEK_LABEL_LIST[date.getDay()];
  return `${year}/${month}/${day} (${weekday})`;
}

export function formatJapaneseDate(date: Date): string {
  const year = date.getFullYear();
  const month = date.getMonth() + 1;
  const day = date.getDate();
  const weekday = DAY_OF_WEEK_LABEL_LIST[date.getDay()];
  return `${year}年${month}月${day}日（${weekday}）`;
}
