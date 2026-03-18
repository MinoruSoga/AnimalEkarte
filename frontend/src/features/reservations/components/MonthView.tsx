// React/Framework
import { memo, useMemo } from "react";
import type React from "react";

// External
import { startOfMonth, endOfMonth, startOfWeek, endOfWeek, format, isSameMonth, isSameDay, addDays } from "date-fns";
import { ja } from "date-fns/locale";

// Internal
import { getReservationTypeColor } from "@/utils/status-helpers";

// Types
import type { ReservationAppointment } from "@/types";

interface ServiceTypeColor {
  style: React.CSSProperties;
  dotStyle: React.CSSProperties;
  hex: string;
}

interface MonthViewProps {
  currentDate: Date;
  appointments: ReservationAppointment[];
  onAppointmentClick: (appointment: ReservationAppointment) => void;
  dynamicColorMap?: Map<string, ServiceTypeColor>;
}

const DAYS_OF_WEEK = ["日", "月", "火", "水", "木", "金", "土"] as const;

const HEADER_ROW = (
  <div className="grid grid-cols-7 border-b border-[rgba(55,53,47,0.16)] bg-[#F7F6F3]">
    {DAYS_OF_WEEK.map((d, i) => (
      <div
        key={d}
        className={`py-3 text-sm font-bold text-center ${
          i === 0 ? "text-red-500" : i === 6 ? "text-blue-500" : "text-[#37352F]/60"
        }`}
      >
        {d}
      </div>
    ))}
  </div>
);

export const MonthView = memo(function MonthView({ currentDate, appointments, onAppointmentClick, dynamicColorMap }: MonthViewProps) {

  const rows = useMemo(() => {
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
            className={`h-full min-h-[140px] bg-white border-b border-r border-[rgba(55,53,47,0.09)] p-2 transition-colors hover:bg-[#F7F6F3] cursor-pointer flex flex-col
              ${!isSameMonth(day, monthStart) ? "bg-[#F7F6F3]/30 text-[#37352F]/30" : "text-[#37352F]"}
              ${isSameDay(day, new Date()) ? "bg-blue-50/30" : ""}
            `}
          >
            <div className="flex justify-between items-start mb-2">
                <span className={`text-base font-bold size-7 flex items-center justify-center rounded-full ${isSameDay(day, new Date()) ? "bg-blue-600 text-white shadow-sm" : ""}`}>
                    {formattedDate}
                </span>
            </div>
            <div className="space-y-1.5 flex-1 overflow-hidden">
                {dayAppointments.slice(0, 4).map(app => {
                    const colorStyle = getReservationTypeColor(app.type, dynamicColorMap);
                    const isClassNameColor = typeof colorStyle === "string";
                    return (
                    <div
                        key={app.id}
                        className={`text-sm px-2 py-1.5 rounded border cursor-pointer hover:opacity-80 leading-tight ${isClassNameColor ? colorStyle : ""}`}
                        style={isClassNameColor ? undefined : (colorStyle as React.CSSProperties)}
                        onClick={(e) => {
                            e.stopPropagation();
                            onAppointmentClick(app);
                        }}
                    >
                        <div className="flex items-center gap-1 min-w-0">
                            {app.visitType === "first"
                                ? <span className="bg-red-100 text-red-600 text-[10px] px-1 rounded flex-shrink-0">初</span>
                                : <span className="bg-blue-100 text-blue-600 text-[10px] px-1 rounded flex-shrink-0">再</span>
                            }
                            <span className="truncate text-xs font-medium">{app.petName}</span>
                        </div>
                        <div className="text-[11px] opacity-70 truncate mt-0.5">
                            {app.ownerName}{app.doctor ? ` / ${app.doctor}` : ""}
                        </div>
                    </div>
                    );
                })}
                {dayAppointments.length > 4 ? (
                    <div className="text-sm text-[#37352F]/60 pl-1">
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
  }, [currentDate, appointments, dynamicColorMap, onAppointmentClick]);

  return (
    <div className="flex flex-col h-full border-l border-t border-[rgba(55,53,47,0.16)] rounded-lg overflow-hidden bg-white shadow-sm">
      {HEADER_ROW}
      <div className="flex-1 flex flex-col">
        {rows}
      </div>
    </div>
  );
});
