import { format } from "date-fns";
import type { ReservationTypeUnavailableTime } from "@/hooks/use-reservation-type-unavailable-times";

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
