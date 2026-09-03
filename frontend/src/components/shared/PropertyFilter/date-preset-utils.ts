import {
  addDays,
  addMonths,
  addWeeks,
  addYears,
  endOfDay,
  endOfMonth,
  endOfWeek,
  endOfYear,
  startOfDay,
  startOfMonth,
  startOfWeek,
  startOfYear,
  subDays,
  subMonths,
  subWeeks,
  subYears,
} from "date-fns";
import { ja } from "date-fns/locale";

import { toJSTWallDate } from "@/lib/jst-date";
import type { RelativePoint, RelativeUnit } from "./types";

export type DatePreset =
  | { label: string; type: "relative"; point: RelativePoint; unit: RelativeUnit }
  | { label: string; type: "last_n_days"; n: number };

export const DATE_PRESETS: DatePreset[] = [
  { label: "今日", type: "relative", point: "this", unit: "day" },
  { label: "昨日", type: "relative", point: "last", unit: "day" },
  { label: "今週", type: "relative", point: "this", unit: "week" },
  { label: "先週", type: "relative", point: "last", unit: "week" },
  { label: "今月", type: "relative", point: "this", unit: "month" },
  { label: "先月", type: "relative", point: "last", unit: "month" },
  { label: "直近7日", type: "last_n_days", n: 7 },
  { label: "直近30日", type: "last_n_days", n: 30 },
];

function resolveRelativeDate(point: RelativePoint, unit: RelativeUnit): { from: Date; to: Date } {
  const now = toJSTWallDate(new Date());
  const today = startOfDay(now);

  switch (unit) {
    case "day": {
      const target =
        point === "this" ? today : point === "last" ? subDays(today, 1) : addDays(today, 1);
      return { from: target, to: endOfDay(target) };
    }
    case "week": {
      const base =
        point === "this" ? today : point === "last" ? subWeeks(today, 1) : addWeeks(today, 1);
      return {
        from: startOfWeek(base, { locale: ja }),
        to: endOfWeek(base, { locale: ja }),
      };
    }
    case "month": {
      const base =
        point === "this" ? today : point === "last" ? subMonths(today, 1) : addMonths(today, 1);
      return { from: startOfMonth(base), to: endOfMonth(base) };
    }
    case "year": {
      const base =
        point === "this" ? today : point === "last" ? subYears(today, 1) : addYears(today, 1);
      return { from: startOfYear(base), to: endOfYear(base) };
    }
  }
}

export function resolvePreset(preset: DatePreset): { from: Date; to: Date } {
  if (preset.type === "last_n_days") {
    const today = startOfDay(toJSTWallDate(new Date()));
    return { from: subDays(today, preset.n - 1), to: endOfDay(today) };
  }
  return resolveRelativeDate(preset.point, preset.unit);
}
