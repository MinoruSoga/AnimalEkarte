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
  const latest = [...examinations]
    .filter((examination) => isValidISODate(examination.date))
    .sort((left, right) => right.date.localeCompare(left.date))[0];

  return latest?.items?.filter((item) => item.isAbnormal).length ?? 0;
}

/** 今日以降の接種予定から最も近い1件を選ぶ。 */
export function selectUpcomingVaccination<T extends VaccinationForSummary>(
  vaccinations: ReadonlyArray<T>,
  today: string = todayJSTISO(),
): T | undefined {
  return [...vaccinations]
    .filter((vaccination) => {
      return isValidISODate(vaccination.nextDate) && vaccination.nextDate >= today;
    })
    .sort((left, right) => left.nextDate.localeCompare(right.nextDate))[0];
}

/** 未来予定がない場合に、最も直近で期限を過ぎた接種予定を選ぶ。 */
export function selectLatestOverdueVaccination<T extends VaccinationForSummary>(
  vaccinations: ReadonlyArray<T>,
  today: string = todayJSTISO(),
): T | undefined {
  return [...vaccinations]
    .filter((vaccination) => {
      return isValidISODate(vaccination.nextDate) && vaccination.nextDate < today;
    })
    .sort((left, right) => right.nextDate.localeCompare(left.nextDate))[0];
}
