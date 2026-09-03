import { memo, useEffect, useMemo, useRef } from "react";
import { addDays, format, isSameDay } from "date-fns";
import { ja } from "date-fns/locale";

import { C } from "@/lib/design-tokens";
import type { Reservation } from "@/types";

import { DayColumn } from "./WeekViewDayColumn";
import {
  HOUR_HEIGHT,
  HOURS,
  WEEK_DAYS,
  type ReservationTypeColor,
} from "../lib/week-view-grid-constants";

const INITIAL_SCROLL_HOUR = 9;

export type { ReservationTypeColor } from "../lib/week-view-grid-constants";

interface WeekViewGridProps {
  startDate: Date;
  appointmentsByDay: Map<string, Reservation[]>;
  onAppointmentClick: (reservation: Reservation) => void;
  onTimeSlotClick?: (date: Date) => void;
  onAppointmentUpdate?: (reservation: Reservation, newStart: Date, newEnd: Date) => void;
  dynamicColorMap?: Map<string, ReservationTypeColor>;
  columnWidth: string;
}

export function WeekViewGrid({
  startDate,
  appointmentsByDay,
  onAppointmentClick,
  onTimeSlotClick,
  onAppointmentUpdate,
  dynamicColorMap,
  columnWidth,
}: WeekViewGridProps) {
  const scrollRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = INITIAL_SCROLL_HOUR * HOUR_HEIGHT;
    }
  }, []);

  return (
    <div
      ref={scrollRef}
      className={`flex-1 border ${C.borderMedium} rounded-lg ${C.bgWhite} overflow-auto relative`}
    >
      <div style={{ minWidth: "100%" }}>
        <div className={`flex border-b ${C.borderMedium} sticky top-0 z-30 ${C.bgWhite}`}>
          <div
            className={`w-12 flex-shrink-0 ${C.bgPage} border-r border-b ${C.borderMedium} sticky left-0 z-40`}
          />
          <WeekHeader
            startDate={startDate}
            appointmentsByDay={appointmentsByDay}
            columnWidth={columnWidth}
          />
        </div>

        <div className="flex relative">
          <TimeSidebar />
          <div className="flex relative flex-1">
            {WEEK_DAYS.map((offset) => {
              const day = addDays(startDate, offset);
              const dayKey = format(day, "yyyy-MM-dd");
              const dayAppointments = appointmentsByDay.get(dayKey) ?? [];

              return (
                <DayColumn
                  key={offset}
                  date={day}
                  appointments={dayAppointments}
                  onAppointmentClick={onAppointmentClick}
                  onTimeSlotClick={onTimeSlotClick}
                  onAppointmentUpdate={onAppointmentUpdate}
                  dynamicColorMap={dynamicColorMap}
                  minWidth={columnWidth}
                />
              );
            })}
          </div>
        </div>
      </div>
    </div>
  );
}

interface WeekHeaderProps {
  startDate: Date;
  appointmentsByDay: Map<string, Reservation[]>;
  columnWidth: string;
}

function WeekHeader({ startDate, appointmentsByDay, columnWidth }: WeekHeaderProps) {
  const headerDays = useMemo(
    () =>
      WEEK_DAYS.map((offset) => {
        const day = addDays(startDate, offset);
        const isToday = isSameDay(day, new Date());
        const count = appointmentsByDay.get(format(day, "yyyy-MM-dd"))?.length ?? 0;
        const dow = day.getDay();
        return (
          <div
            key={offset}
            data-testid="week-header-day"
            className={`flex-1 flex-shrink-0 text-center py-2 border-r border-b ${C.borderMedium} ${C.bgWhite} ${
              isToday ? C.bgBrand8 : ""
            }`}
            style={{ minWidth: columnWidth }}
          >
            <div
              className={`text-sm mb-1 ${
                dow === 0 ? C.textNotionRed : dow === 6 ? C.textBrand : C.text60
              }`}
            >
              {format(day, "E", { locale: ja })}
            </div>
            <div className="flex items-center justify-center">
              <div className={`relative text-xl font-bold ${isToday ? C.textBrand : C.text}`}>
                {format(day, "d")}
                {count > 0 ? (
                  <span
                    className={`absolute -right-7 bottom-0 text-xs whitespace-nowrap ${C.textBrand}`}
                  >
                    {count}件
                  </span>
                ) : null}
              </div>
            </div>
          </div>
        );
      }),
    [appointmentsByDay, columnWidth, startDate],
  );

  return <div className="flex flex-1">{headerDays}</div>;
}

const TimeSidebar = memo(function TimeSidebar() {
  return (
    <div
      className={`w-12 flex-shrink-0 flex flex-col ${C.bgPage} border-r ${C.borderMedium} z-30 sticky left-0`}
    >
      {HOURS.map((hour) => (
        <div
          key={hour}
          className={`relative flex-shrink-0 text-xs ${C.text40} text-right pr-2 pt-0.5 leading-none`}
          style={{ height: `${HOUR_HEIGHT}px` }}
        >
          {hour}:00
          <div className={`absolute top-0 right-0 w-1.5 h-[1px] ${C.bgLight}`} />
          <div
            className="absolute right-0 pr-2 text-xs leading-none"
            style={{ top: `${HOUR_HEIGHT / 2}px`, transform: "translateY(-50%)", opacity: 0.45 }}
          >
            :30
          </div>
          <div
            className={`absolute right-0 w-[5px] h-[1px] ${C.bgLight}`}
            style={{ top: `${HOUR_HEIGHT / 2}px` }}
          />
        </div>
      ))}
    </div>
  );
});
