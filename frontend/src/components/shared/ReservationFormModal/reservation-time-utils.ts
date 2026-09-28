import { format } from "date-fns";
import type { ReservationTypeUnavailableTime } from "@/hooks/use-reservation-type-unavailable-times";
import type { ReservationSlotVacancyStatus } from "@/hooks/use-reservation-types";

// EMR-170: 1枠の空き状況（〇△✕）
export interface SlotVacancy {
  status: ReservationSlotVacancyStatus;
  /** 受け入れ可能な残り枠数（上限なしは null） */
  remaining: number | null;
}

const SLOT_VACANCY_LABEL: Record<ReservationSlotVacancyStatus, string> = {
  available: "〇 空きあり",
  low: "△ 残り1枠",
  full: "✕ 満員",
};

/** 空き状況の表示ラベル。記号＋テキストで色だけに依存しない（WCAG） */
export function slotVacancyLabel(status: ReservationSlotVacancyStatus): string {
  return SLOT_VACANCY_LABEL[status];
}

function generateTimeOptions(): string[] {
  const times: string[] = [];
  for (let h = 0; h < 24; h++) {
    const hh = String(h).padStart(2, "0");
    times.push(`${hh}:00`);
    times.push(`${hh}:15`);
    times.push(`${hh}:30`);
    times.push(`${hh}:45`);
  }
  return times;
}

export const TIME_OPTIONS = generateTimeOptions();

/**
 * BUG-015/EMR-191: Select の options に現在値が無ければ末尾へ追加する。
 * Radix Select は options に存在しない value を trigger に表示できないため、
 * スロット外・非15分刻みの既存値でも空白表示しないようにする。
 */
export function withCurrentTimeOption(options: string[], current: string | undefined): string[] {
  if (current === undefined || options.includes(current)) return options;
  return [...options, current];
}

/** EMR-191: 予約区分の duration_minutes が未設定/0 のときの既定所要時間 */
export const DEFAULT_RESERVATION_DURATION_MINUTES = 15;

/**
 * EMR-191: 開始時刻の変更に連動する終了時刻。
 * LINE 空き枠スロットの終了時刻があればそれを優先し、無い場合は
 * 予約区分の durationMinutes（既定 15 分）を開始時刻へ加算する。
 */
export function resolveEndTimeOnStartChange(
  start: Date,
  durationMinutes: number | undefined,
  slotEnd: string | undefined,
): Date {
  const end = new Date(start);
  if (slotEnd) {
    const [hours, minutes] = slotEnd.split(":").map(Number);
    end.setHours(hours, minutes, 0, 0);
    return end;
  }
  const duration =
    durationMinutes !== undefined && durationMinutes > 0
      ? durationMinutes
      : DEFAULT_RESERVATION_DURATION_MINUTES;
  end.setMinutes(end.getMinutes() + duration);
  return end;
}

function timeToMinutes(time: string): number {
  const [hours, minutes] = time.split(":").map(Number);
  return hours * 60 + minutes;
}

export function slotTimeToSelectValue(time: string): string {
  if (time.includes(":")) {
    const [hour, minute] = time.split(":");
    const normalizedMinute = minute ? minute.padStart(2, "0").slice(0, 2) : "00";
    return `${hour.padStart(2, "0")}:${normalizedMinute}`;
  }
  return `${time.slice(0, 2)}:${time.slice(2, 4)}`;
}

export function getApplicableUnavailableTimes(
  times: ReservationTypeUnavailableTime[],
  date: Date | undefined,
): ReservationTypeUnavailableTime[] {
  if (!date) return [];
  const dateStr = format(date, "yyyy-MM-dd");
  const specific = times.filter(
    (time) => time.unavailableType === "specific" && time.specificDate === dateStr,
  );
  if (specific.length > 0) return specific;
  const dayOfWeek = date.getDay();
  return times.filter((time) => time.unavailableType === "weekly" && time.dayOfWeek === dayOfWeek);
}

export function isStartTimeUnavailable(
  time: string,
  unavailableTimes: ReservationTypeUnavailableTime[],
): boolean {
  const minutes = timeToMinutes(time);
  return unavailableTimes.some(
    (unavailableTime) =>
      minutes >= timeToMinutes(unavailableTime.startTime) &&
      minutes < timeToMinutes(unavailableTime.endTime),
  );
}
