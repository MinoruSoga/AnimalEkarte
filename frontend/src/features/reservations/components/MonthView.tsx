// React/Framework
import { memo, useMemo } from "react";
import type React from "react";

// External
import { startOfMonth, endOfMonth, startOfWeek, endOfWeek, format, isSameMonth, isSameDay, addDays } from "date-fns";
import { ja } from "date-fns/locale";

// Internal
import { C } from "@/lib/design-tokens";
import { toJSTWallDate } from "@/lib/jst-date";
import { getReservationTypeColor } from "@/lib/status-helpers";
import { DAY_OF_WEEK_LABEL_LIST } from "@/constants/day-of-week";

// Types
import type { Reservation } from "@/types";

interface ReservationTypeColor {
  style: React.CSSProperties;
  dotStyle: React.CSSProperties;
  hex: string;
}

interface MonthViewProps {
  currentDate: Date;
  appointments: Reservation[];
  onAppointmentClick: (reservation: Reservation) => void;
  /** BUG-076: 日付セルクリックで週表示に遷移するコールバック */
  onDateClick?: (date: Date) => void;
  dynamicColorMap?: Map<string, ReservationTypeColor>;
}

const HEADER_ROW = (
  <div className={`grid grid-cols-7 border-b ${C.borderMedium} ${C.bgPage}`}>
    {DAY_OF_WEEK_LABEL_LIST.map((d, i) => (
      <div
        key={d}
        className={`py-3 text-sm font-bold text-center ${
          i === 0 ? C.danger : i === 6 ? C.textBrand : C.text60
        }`}
      >
        {d}
      </div>
    ))}
  </div>
);

export const MonthView = memo(function MonthView({ currentDate, appointments, onAppointmentClick, onDateClick, dynamicColorMap }: MonthViewProps) {

  const rows = useMemo(() => {
    const today = toJSTWallDate(new Date());
    const monthStart = startOfMonth(currentDate);
    const monthEnd = endOfMonth(monthStart);
    const startDate = startOfWeek(monthStart, { locale: ja });
    const endDate = endOfWeek(monthEnd, { locale: ja });

    const dateFormat = "d";
    const result = [];
    let days = [];
    let day = startDate;

    while (day <= endDate) {
      for (let i = 0; i < 7; i++) {
        const formattedDate = format(day, dateFormat);
        const cloneDay = day;

        const dayAppointments = appointments.filter(app => isSameDay(app.start, cloneDay));

        days.push(
          <div
            key={day.toString()}
            className={`h-full min-h-[140px] ${C.bgWhite} border-b border-r ${C.borderLight} p-2 transition-colors ${C.hoverBgPage} cursor-pointer flex flex-col
              ${!isSameMonth(day, monthStart) ? `${C.bgPage30} ${C.text30}` : C.text}
              ${isSameDay(day, today) ? C.bgBrand8 : ""}
            `}
          >
            <div className="flex justify-between items-start mb-2">
                <button
                  type="button"
                  className={`text-base font-bold size-7 flex items-center justify-center rounded-full transition-colors ${isSameDay(day, today) ? `${C.bgBrand} ${C.textOnBrand}` : `${C.hoverBgBrandLight} ${C.hoverTextBrand}`}`}
                  onClick={(e) => {
                    e.stopPropagation();
                    onDateClick?.(cloneDay);
                  }}
                  aria-label={`${format(cloneDay, "M月d日")}の週表示へ`}
                >
                  {formattedDate}
                </button>
            </div>
            <div className="space-y-1.5 flex-1 overflow-hidden">
                {dayAppointments.slice(0, 4).map(app => {
                    const colorStyle = getReservationTypeColor(app.type, dynamicColorMap);
                    const isClassNameColor = typeof colorStyle === "string";
                    return (
                    <button
                        key={app.id}
                        type="button"
                        className={`block w-full text-left text-sm px-2 py-1.5 rounded border cursor-pointer hover:opacity-80 leading-tight ${isClassNameColor ? colorStyle : ""}`}
                        style={isClassNameColor ? undefined : (colorStyle as React.CSSProperties)}
                        onClick={(e) => {
                            e.stopPropagation();
                            onAppointmentClick(app);
                        }}
                    >
                        <div className="flex items-center gap-1 min-w-0">
                            {app.visitType === "first"
                                ? <span className={`${C.bgRedLight} ${C.danger} text-2xs px-1 rounded flex-shrink-0`}>初</span>
                                : <span className={`${C.bgBrandLight} ${C.textBrandDark} text-2xs px-1 rounded flex-shrink-0`}>再</span>
                            }
                            <span className="truncate text-xs font-medium">{app.petName}</span>
                        </div>
                        <div className="text-2xs opacity-70 truncate mt-0.5">
                            {app.ownerName}{app.doctor ? ` / ${app.doctor}` : ""}
                        </div>
                    </button>
                    );
                })}
                {dayAppointments.length > 4 ? (
                    <div className={`text-sm ${C.text60} pl-1`}>
                        他 {dayAppointments.length - 4} 件
                    </div>
                ) : null}
            </div>
          </div>
        );
        day = addDays(day, 1);
      }
      result.push(
        <div className="grid grid-cols-7 flex-1" key={day.toString()}>
          {days}
        </div>
      );
      days = [];
    }

    return result;
  }, [currentDate, appointments, dynamicColorMap, onAppointmentClick, onDateClick]);

  return (
    <div className={`flex flex-col h-full border-l border-t ${C.borderMedium} rounded-lg overflow-hidden ${C.bgWhite}`}>
      {HEADER_ROW}
      <div className="flex-1 flex flex-col">
        {rows}
      </div>
    </div>
  );
});
