import { formatJSTDate, todayJSTISO } from "@/lib/jst-date";

interface ExaminationForSummary {
  date: string;
  items?: ReadonlyArray<{ isAbnormal: boolean }>;
}

interface VaccinationForSummary {
  nextDate: string;
}

function isValidISODate(value: string): boolean {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) return false;
  return formatJSTDate(`${value}T00:00:00+09:00`) === value;
}

/** 最新の検査だけを「現在確認すべき検査」として扱い、過去の異常を累積しない。 */
export function countLatestExaminationAbnormalResults(
  examinations: ReadonlyArray<ExaminationForSummary>,
): number {
  let latest: ExaminationForSummary | undefined;
  for (const examination of examinations) {
    if (!isValidISODate(examination.date)) continue;
    if (!latest || examination.date > latest.date) latest = examination;
  }

  return latest?.items?.filter((item) => item.isAbnormal).length ?? 0;
}

/** 今日以降の接種予定から最も近い1件を選ぶ。 */
export function selectUpcomingVaccination<T extends VaccinationForSummary>(
  vaccinations: ReadonlyArray<T>,
  today: string = todayJSTISO(),
): T | undefined {
  let upcoming: T | undefined;
  for (const vaccination of vaccinations) {
    if (!isValidISODate(vaccination.nextDate) || vaccination.nextDate < today) continue;
    if (!upcoming || vaccination.nextDate < upcoming.nextDate) upcoming = vaccination;
  }
  return upcoming;
}

/** 未来予定がない場合に、最も直近で期限を過ぎた接種予定を選ぶ。 */
export function selectLatestOverdueVaccination<T extends VaccinationForSummary>(
  vaccinations: ReadonlyArray<T>,
  today: string = todayJSTISO(),
): T | undefined {
  let overdue: T | undefined;
  for (const vaccination of vaccinations) {
    if (!isValidISODate(vaccination.nextDate) || vaccination.nextDate >= today) continue;
    if (!overdue || vaccination.nextDate > overdue.nextDate) overdue = vaccination;
  }
  return overdue;
}
