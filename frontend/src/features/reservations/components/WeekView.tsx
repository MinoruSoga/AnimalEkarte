import { memo, useMemo } from "react";
import { format, startOfWeek } from "date-fns";
import { ja } from "date-fns/locale";

import type { Reservation } from "@/types";
import { WeekViewGrid, type ReservationTypeColor } from "./WeekViewParts";
import { isTentativeReservation } from "../lib/tentative";

interface WeekViewProps {
  currentDate: Date;
  appointments: Reservation[];
  onAppointmentClick: (reservation: Reservation) => void;
  onTimeSlotClick?: (date: Date) => void;
  onAppointmentUpdate?: (reservation: Reservation, newStart: Date, newEnd: Date) => void;
  dynamicColorMap?: Map<string, ReservationTypeColor>;
  days?: 5 | 7;
}

export const WeekView = memo(function WeekView({
  currentDate,
  appointments,
  onAppointmentClick,
  onTimeSlotClick,
  onAppointmentUpdate,
  dynamicColorMap,
  days = 5,
}: WeekViewProps) {
  const columnWidth = days === 7 ? `${(100 / 7).toFixed(4)}%` : "20%";

  const startDate = useMemo(() => startOfWeek(currentDate, { locale: ja }), [currentDate]);

  // Codex PR606 #9: tentative は実時間を持たない sentinel（意図日 00:00 JST）で、
  // タイムグリッドでは 0:00 位置の小カードになり viewport 外に押し出され得る。
  // タイムグリッド用（appointmentsByDay）と sticky 見出し下の常時表示帯用
  // （tentativeByDay）に振り分ける。ヘッダー件数は合算で表示する。
  const { appointmentsByDay, tentativeByDay } = useMemo(() => {
    const timed = new Map<string, Reservation[]>();
    const tentative = new Map<string, Reservation[]>();
    for (const appointment of appointments) {
      const key = format(appointment.start, "yyyy-MM-dd");
      const map = isTentativeReservation(appointment) ? tentative : timed;
      const existing = map.get(key);
      if (existing) {
        existing.push(appointment);
      } else {
        map.set(key, [appointment]);
      }
    }
    return { appointmentsByDay: timed, tentativeByDay: tentative };
  }, [appointments]);

  return (
    <WeekViewGrid
      startDate={startDate}
      appointmentsByDay={appointmentsByDay}
      tentativeByDay={tentativeByDay}
      onAppointmentClick={onAppointmentClick}
      onTimeSlotClick={onTimeSlotClick}
      onAppointmentUpdate={onAppointmentUpdate}
      dynamicColorMap={dynamicColorMap}
      columnWidth={columnWidth}
    />
  );
});
