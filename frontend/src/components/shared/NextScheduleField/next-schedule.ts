import { addMonths, addWeeks, addYears, format } from "date-fns";

export const NEXT_SCHEDULE_TYPES = [
  "3weeks",
  "4weeks",
  "6weeks",
  "6months",
  "1year",
  "other",
] as const;

export type NextScheduleType = (typeof NEXT_SCHEDULE_TYPES)[number];

export interface NextScheduleOption {
  value: NextScheduleType;
  label: string;
}

/** 予防接種・次回来院の共通プリセット（BUG-032） */
export const NEXT_SCHEDULE_OPTIONS: readonly NextScheduleOption[] = [
  { value: "3weeks", label: "3週後" },
  { value: "4weeks", label: "4週後" },
  { value: "1year", label: "1年後" },
  { value: "other", label: "以外（手動）" },
];

export function isNextScheduleType(value: string): value is NextScheduleType {
  return (NEXT_SCHEDULE_TYPES as readonly string[]).includes(value);
}

export function calculateNextDate(baseDate: string, scheduleType: string): string {
  if (!baseDate || scheduleType === "other") return "";
  const date = new Date(`${baseDate}T00:00:00`);
  if (Number.isNaN(date.getTime())) return "";
  switch (scheduleType) {
    case "3weeks":
      return format(addWeeks(date, 3), "yyyy-MM-dd");
    case "4weeks":
      return format(addWeeks(date, 4), "yyyy-MM-dd");
    case "6weeks":
      return format(addWeeks(date, 6), "yyyy-MM-dd");
    case "6months":
      return format(addMonths(date, 6), "yyyy-MM-dd");
    case "1year":
      return format(addYears(date, 1), "yyyy-MM-dd");
    default:
      return "";
  }
}

export function resolveScheduleTypeAfterManualDate(
  baseDate: string,
  currentType: string,
  nextDate: string,
): string {
  if (currentType !== "other" && baseDate && nextDate) {
    const calculated = calculateNextDate(baseDate, currentType);
    if (calculated && calculated === nextDate) return currentType;
  }
  return "other";
}
