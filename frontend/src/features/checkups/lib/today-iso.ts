/** ローカルタイムゾーンの今日を yyyy-mm-dd 形式で返す */
export function todayISODate(): string {
  return formatLocalISO(new Date());
}

/** baseISO (yyyy-mm-dd) に days を加算した日付を yyyy-mm-dd で返す */
export function addDaysISO(baseISO: string, days: number): string {
  const d = new Date(`${baseISO}T00:00:00`);
  d.setDate(d.getDate() + days);
  return formatLocalISO(d);
}

function formatLocalISO(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}
