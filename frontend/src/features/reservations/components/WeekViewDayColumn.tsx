import { memo, useCallback, useMemo, useRef } from "react";
import type { MouseEvent as ReactMouseEvent, TouchEvent as ReactTouchEvent } from "react";
import { isSameDay } from "date-fns";

import { C, PALETTE } from "@/lib/design-tokens";
import { toJSTWallDate } from "@/lib/jst-date";
import type { Reservation } from "@/types";

import { AppointmentCard } from "./WeekViewAppointmentCard";
import {
  calculateEventLayout,
  HOUR_HEIGHT,
  HOURS,
  type ReservationTypeColor,
} from "../lib/week-view-grid-constants";

/** クリック位置から時刻を計算する際のスナップ単位（分） */
const TIME_SLOT_SNAP_MINUTES = 15;
/** タップ判定: この移動量(px)未満ならスクロールではなくタップとみなす */
const TAP_MOVEMENT_THRESHOLD_PX = 10;
/** タップ判定: この時間(ms)未満ならタップとみなす */
const TAP_DURATION_THRESHOLD_MS = 300;

interface DayColumnProps {
  date: Date;
  appointments: Reservation[];
  onAppointmentClick: (reservation: Reservation) => void;
  onTimeSlotClick?: (date: Date) => void;
  onAppointmentUpdate?: (reservation: Reservation, newStart: Date, newEnd: Date) => void;
  dynamicColorMap?: Map<string, ReservationTypeColor>;
  minWidth?: string;
}

export const DayColumn = memo(function DayColumn({
  date,
  appointments,
  onAppointmentClick,
  onTimeSlotClick,
  onAppointmentUpdate,
  dynamicColorMap,
  minWidth,
}: DayColumnProps) {
  const layoutStyles = useMemo(() => calculateEventLayout(appointments), [appointments]);
  const now = toJSTWallDate(new Date());
  const isToday = isSameDay(date, now);
  const currentTimeTop = ((now.getHours() * 60 + now.getMinutes()) / 60) * HOUR_HEIGHT;

  const computeTimeFromY = useCallback(
    (y: number): Date => {
      const hoursFromStart = y / HOUR_HEIGHT;
      const totalMinutes =
        Math.round(Math.floor(hoursFromStart * 60) / TIME_SLOT_SNAP_MINUTES) *
        TIME_SLOT_SNAP_MINUTES;
      const clickedDate = new Date(date);
      clickedDate.setHours(0, 0, 0, 0);
      clickedDate.setMinutes(totalMinutes);
      return clickedDate;
    },
    [date],
  );

  const handleColumnClick = (e: ReactMouseEvent<HTMLDivElement>) => {
    if (!onTimeSlotClick) return;
    const rect = e.currentTarget.getBoundingClientRect();
    const y = e.clientY - rect.top;
    onTimeSlotClick(computeTimeFromY(y));
  };

  const touchStartY = useRef<number | null>(null);
  const touchStartTime = useRef<number>(0);

  const handleTouchStart = useCallback((e: ReactTouchEvent<HTMLDivElement>) => {
    touchStartY.current = e.touches[0].clientY;
    touchStartTime.current = Date.now();
  }, []);

  const handleTouchEnd = useCallback(
    (e: ReactTouchEvent<HTMLDivElement>) => {
      if (!onTimeSlotClick || touchStartY.current === null) return;

      const touch = e.changedTouches[0];
      const dy = Math.abs(touch.clientY - touchStartY.current);
      const dt = Date.now() - touchStartTime.current;

      if (dy < TAP_MOVEMENT_THRESHOLD_PX && dt < TAP_DURATION_THRESHOLD_MS) {
        const rect = e.currentTarget.getBoundingClientRect();
        const y = touch.clientY - rect.top;
        onTimeSlotClick(computeTimeFromY(y));
      }

      touchStartY.current = null;
    },
    [computeTimeFromY, onTimeSlotClick],
  );

  return (
    <div
      className={`flex-1 flex-shrink-0 border-r ${C.borderMedium} relative ${
        isToday ? C.bgBrandLight8 : ""
      }`}
      style={{ minWidth: minWidth ?? "20%" }}
      onClick={handleColumnClick}
      onTouchStart={handleTouchStart}
      onTouchEnd={handleTouchEnd}
    >
      <div className="absolute inset-0 flex flex-col pointer-events-none z-0">
        {HOURS.map((_, h) => (
          <div
            key={h}
            className="relative w-full flex-shrink-0"
            style={{ height: `${HOUR_HEIGHT}px` }}
          >
            <div
              className={`absolute left-0 right-0 border-t border-dotted ${C.borderDivider}`}
              style={{ top: `${HOUR_HEIGHT / 4}px`, opacity: 0.3 }}
            />
            <div
              className={`absolute left-0 right-0 border-t border-dashed ${C.borderDivider}`}
              style={{ top: `${HOUR_HEIGHT / 2}px`, opacity: 0.55 }}
            />
            <div
              className={`absolute left-0 right-0 border-t border-dotted ${C.borderDivider}`}
              style={{ top: `${(HOUR_HEIGHT * 3) / 4}px`, opacity: 0.3 }}
            />
            <div className={`absolute bottom-0 left-0 right-0 border-b ${C.borderDivider}`} />
          </div>
        ))}
      </div>

      {isToday ? (
        <div
          className="absolute w-full border-t-2 z-20 pointer-events-none"
          style={{ borderColor: PALETTE.danger, top: `${currentTimeTop}px` }}
        >
          <div className={`absolute -left-1 -top-1.5 w-3 h-3 rounded-full ${C.bgNotionRed}`} />
        </div>
      ) : null}

      {appointments.map((appointment) => (
        <AppointmentCard
          key={appointment.id}
          appointment={appointment}
          layoutStyle={layoutStyles[appointment.id] || { left: "0%", width: "100%" }}
          onClick={onAppointmentClick}
          onUpdate={onAppointmentUpdate}
          dynamicColorMap={dynamicColorMap}
        />
      ))}
    </div>
  );
});
