import { memo, useMemo } from "react";
import { format, startOfWeek } from "date-fns";
import { ja } from "date-fns/locale";

import type { Reservation } from "@/types";
import { WeekViewGrid, type ReservationTypeColor } from "./WeekViewParts";

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

  const appointmentsByDay = useMemo(() => {
    const map = new Map<string, Reservation[]>();
    for (const appointment of appointments) {
      const key = format(appointment.start, "yyyy-MM-dd");
      const existing = map.get(key);
      if (existing) {
        existing.push(appointment);
      } else {
        map.set(key, [appointment]);
      }
    }
    return map;
  }, [appointments]);

  return (
    <WeekViewGrid
      startDate={startDate}
      appointmentsByDay={appointmentsByDay}
      onAppointmentClick={onAppointmentClick}
      onTimeSlotClick={onTimeSlotClick}
      onAppointmentUpdate={onAppointmentUpdate}
      dynamicColorMap={dynamicColorMap}
      columnWidth={columnWidth}
    />
  );
});
