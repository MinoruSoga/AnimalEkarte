import { memo, useEffect, useMemo, useRef } from "react";
import { addDays, format, isSameDay } from "date-fns";
import { ja } from "date-fns/locale";

import { C } from "@/lib/design-tokens";
import type { Reservation } from "@/types";

import { DayColumn } from "./WeekViewDayColumn";
import { TENTATIVE_APPOINTMENT_LABEL } from "../lib/tentative";
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
  /** Codex PR606 #9: 日付別の tentative（次回予定・未予約）。sticky 帯に常時表示する。 */
  tentativeByDay: Map<string, Reservation[]>;
  onAppointmentClick: (reservation: Reservation) => void;
  onTimeSlotClick?: (date: Date) => void;
  onAppointmentUpdate?: (reservation: Reservation, newStart: Date, newEnd: Date) => void;
  dynamicColorMap?: Map<string, ReservationTypeColor>;
  columnWidth: string;
}

export function WeekViewGrid({
  startDate,
  appointmentsByDay,
  tentativeByDay,
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
        {/* Codex PR606 #9: ヘッダー行と tentative 帯を同一 sticky ブロックに置き、
            縦スクロールしても tentative が常に表示されるようにする */}
        <div className={`sticky top-0 z-30 ${C.bgWhite}`}>
          <div className={`flex border-b ${C.borderMedium}`}>
            <div
              className={`w-12 flex-shrink-0 ${C.bgPage} border-r border-b ${C.borderMedium} sticky left-0 z-40`}
            />
            <WeekHeader
              startDate={startDate}
              appointmentsByDay={appointmentsByDay}
              tentativeByDay={tentativeByDay}
              columnWidth={columnWidth}
            />
          </div>
          <WeekTentativeRow
            startDate={startDate}
            tentativeByDay={tentativeByDay}
            columnWidth={columnWidth}
            onAppointmentClick={onAppointmentClick}
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
  tentativeByDay: Map<string, Reservation[]>;
  columnWidth: string;
}

function WeekHeader({
  startDate,
  appointmentsByDay,
  tentativeByDay,
  columnWidth,
}: WeekHeaderProps) {
  const headerDays = useMemo(
    () =>
      WEEK_DAYS.map((offset) => {
        const day = addDays(startDate, offset);
        const isToday = isSameDay(day, new Date());
        const dayKey = format(day, "yyyy-MM-dd");
        // #9: tentative は帯側に出すが件数には含める（その日の予定として数える）
        const count =
          (appointmentsByDay.get(dayKey)?.length ?? 0) + (tentativeByDay.get(dayKey)?.length ?? 0);
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
              <div className={`relative text-xl font-semibold ${isToday ? C.textBrand : C.text}`}>
                {format(day, "d")}
                {count > 0 ? (
                  <span
                    className={`absolute -right-7 bottom-0 text-xs whitespace-nowrap ${C.textBrand}`}
                  >
                    {count.toLocaleString()}件
                  </span>
                ) : null}
              </div>
            </div>
          </div>
        );
      }),
    [appointmentsByDay, tentativeByDay, columnWidth, startDate],
  );

  return <div className="flex flex-1">{headerDays}</div>;
}

/**
 * Codex PR606 #9: tentative（次回予定・未予約）の常時表示帯。
 * sentinel 00:00 のタイムグリッド位置に依存せず、sticky ヘッダー直下の
 * 日付セルに chip で出す。ドラッグ不可・クリックで詳細を開く。
 * 1日あたりの表示は最大2行分で内部スクロール（帯が viewport を占有しないよう制限）。
 */
interface WeekTentativeRowProps {
  startDate: Date;
  tentativeByDay: Map<string, Reservation[]>;
  columnWidth: string;
  onAppointmentClick: (reservation: Reservation) => void;
}

function WeekTentativeRow({
  startDate,
  tentativeByDay,
  columnWidth,
  onAppointmentClick,
}: WeekTentativeRowProps) {
  const hasAny = useMemo(
    () => [...tentativeByDay.values()].some((list) => list.length > 0),
    [tentativeByDay],
  );
  if (!hasAny) return null;
  return (
    <div className={`flex border-b ${C.borderMedium}`} data-testid="week-tentative-row">
      <div
        className={`w-12 flex-shrink-0 ${C.bgPage} border-r ${C.borderMedium} sticky left-0 z-40 flex items-center justify-center`}
      >
        <span className={`text-2xs ${C.text60} leading-tight`}>次回予定</span>
      </div>
      <div className="flex flex-1">
        {WEEK_DAYS.map((offset) => {
          const day = addDays(startDate, offset);
          const dayTentatives = tentativeByDay.get(format(day, "yyyy-MM-dd")) ?? [];
          return (
            <div
              key={offset}
              className={`flex-1 flex-shrink-0 border-r ${C.borderMedium} p-1 flex flex-col gap-1 max-h-18 overflow-y-auto`}
              style={{ minWidth: columnWidth }}
            >
              {dayTentatives.map((reservation) => (
                <TentativeChip
                  key={reservation.id}
                  reservation={reservation}
                  onClick={onAppointmentClick}
                />
              ))}
            </div>
          );
        })}
      </div>
    </div>
  );
}

interface TentativeChipProps {
  reservation: Reservation;
  onClick: (reservation: Reservation) => void;
}

function TentativeChip({ reservation, onClick }: TentativeChipProps) {
  const label = `${TENTATIVE_APPOINTMENT_LABEL} ${reservation.petName} ${reservation.ownerName}`;
  return (
    <button
      type="button"
      className={`flex items-center gap-1 min-h-7 w-full rounded border border-dashed px-1 py-0.5 text-left cursor-pointer ${C.bgWarning50} ${C.borderWarning20} ${C.textWarning}`}
      aria-label={label}
      title={label}
      onClick={(e) => {
        e.stopPropagation();
        onClick(reservation);
      }}
    >
      <span className="font-semibold whitespace-nowrap text-2xs flex-shrink-0">
        {TENTATIVE_APPOINTMENT_LABEL}
      </span>
      <span className="truncate text-xs font-medium">{reservation.petName}</span>
    </button>
  );
}

const TimeSidebar = memo(function TimeSidebar() {
  return (
    <div
      className={`w-12 flex-shrink-0 flex flex-col ${C.bgPage} border-r ${C.borderMedium} z-30 sticky left-0`}
    >
      {HOURS.map((hour) => (
        <div
          key={hour}
          className={`relative flex-shrink-0 text-xs ${C.text60} text-right pr-2 pt-0.5 leading-none`}
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
