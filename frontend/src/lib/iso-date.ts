import { todayJSTISO } from "@/lib/jst-date";

/** JST の今日を yyyy-mm-dd 形式で返す */
export function todayISODate(): string {
  return todayJSTISO();
}

/** baseISO (yyyy-mm-dd) に days を加算した日付を yyyy-mm-dd で返す */
export function addDaysISO(baseISO: string, days: number): string {
  const [year, month, dayOfMonth] = baseISO.split("-").map(Number);
  const d = new Date(Date.UTC(year, month - 1, dayOfMonth));
  d.setUTCDate(d.getUTCDate() + days);
  const y = d.getUTCFullYear();
  const m = String(d.getUTCMonth() + 1).padStart(2, "0");
  const day = String(d.getUTCDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}
