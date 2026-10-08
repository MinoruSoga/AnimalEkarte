import { isPastJSTDate, todayJSTISO } from "@/lib/jst-date";

/**
 * EMR-308: 次回予定（tentative appointment）の意図日ルール。
 * medical-records（ヘッダー Popover）と共有 NextVisitResetDialog（受付/予約管理の
 * キャンセル後再設定・期限超過パネル）の両方が使うため feature に依存しない
 * shared 層に置く。
 * BE 側も reservation domain で「過去日 NG・上限2年」を検証する（reservation_intent_service.go）。
 */

export function addYearsIso(iso: string, years: number): string {
  const [yearPart, monthPart, dayPart] = iso.split("-");
  const year = Number(yearPart) + years;
  const month = Number(monthPart);
  const day = Number(dayPart);
  const isLeap = year % 4 === 0 && (year % 100 !== 0 || year % 400 === 0);
  const clampedDay = month === 2 && day === 29 && !isLeap ? 28 : day;
  return `${String(year).padStart(4, "0")}-${monthPart}-${String(clampedDay).padStart(2, "0")}`;
}

/** 意図日の上限（今日から2年、YYYY-MM-DD） */
export function nextVisitIntentMaxDate(): string {
  return addYearsIso(todayJSTISO(), 2);
}

/** 次回予定の意図日を検証する。問題なければ null、NG ならエラーメッセージを返す。 */
export function validateNextVisitIntentDate(value: string): string | null {
  if (value === "") return null;
  if (isPastJSTDate(value)) return "今日より前の日付は設定できません";
  if (value > nextVisitIntentMaxDate()) return "今日から2年以内の日付を設定してください";
  return null;
}
