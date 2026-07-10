import { format } from "date-fns";
import type { ReservationTypeUnavailableTime } from "@/features/master";

export function generateTimeOptions(): string[] {
  const times: string[] = [];
  for (let h = 0; h < 24; h++) {
    times.push(`${h}:00`);
    times.push(`${h}:15`);
    times.push(`${h}:30`);
    times.push(`${h}:45`);
  }
  return times;
}

export const TIME_OPTIONS = generateTimeOptions();

export function timeToMinutes(time: string): number {
  const [hours, minutes] = time.split(":").map(Number);
  return hours * 60 + minutes;
}

export function slotTimeToSelectValue(time: string): string {
  if (time.includes(":")) return time.replace(/^0/, "");
  const hour = Number(time.slice(0, 2));
  const minute = time.slice(2, 4);
  return `${hour}:${minute}`;
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
  return times.filter(
    (time) => time.unavailableType === "weekly" && time.dayOfWeek === dayOfWeek,
  );
}

export function isStartTimeUnavailable(time: string, unavailableTimes: ReservationTypeUnavailableTime[]): boolean {
  const minutes = timeToMinutes(time);
  return unavailableTimes.some(
    (unavailableTime) =>
      minutes >= timeToMinutes(unavailableTime.startTime) &&
      minutes < timeToMinutes(unavailableTime.endTime),
  );
}
