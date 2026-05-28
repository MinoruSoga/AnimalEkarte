import { lazy, Suspense, useMemo } from "react";
import type React from "react";
import { format } from "date-fns";
import { ja } from "date-fns/locale";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import type { LegendEntry, ReservationTypeColor } from "@/hooks/use-reservation-type-color-map";
import { C, ICON } from "@/lib/design-tokens";
import { typedSetter } from "@/lib/type-utils";
import { getCalendarViewLabel } from "@/utils/status-helpers";
import type { CalendarView, Reservation } from "../types";
import { CALENDAR_VIEW_VALUES } from "../types";
import { DaysRangeToggle } from "./DaysRangeToggle";

const MonthView = lazy(() =>
  import("./MonthView").then((module) => ({ default: module.MonthView })),
);
const WeekView = lazy(() =>
  import("./WeekView").then((module) => ({ default: module.WeekView })),
);

const SOURCE_FILTER_SELECT_ITEMS = (
  <>
    <SelectItem value="all">すべて</SelectItem>
    <SelectItem value="manual">手動予約</SelectItem>
    <SelectItem value="line">LINE予約</SelectItem>
  </>
);

const CALENDAR_VIEW_SELECT_ITEMS = CALENDAR_VIEW_VALUES.map((view) => (
  <SelectItem key={view} value={view}>
    {getCalendarViewLabel(view)}
  </SelectItem>
));

interface ReservationManagementCalendarProps {
  currentDate: Date;
  view: CalendarView;
  days: 5 | 7;
  doctorFilter: string;
  sourceFilter: string;
  doctorNames: string[];
  appointments: Reservation[];
  activeEntries: LegendEntry[];
  dynamicColorMap: Map<string, ReservationTypeColor>;
  canCreate: boolean;
  onViewChange: React.Dispatch<React.SetStateAction<CalendarView>>;
  onDoctorFilterChange: (value: string) => void;
  onSourceFilterChange: (value: string) => void;
  onDaysChange: (days: 5 | 7) => void;
  onNavigatePrevious: () => void;
  onNavigateToday: () => void;
  onNavigateNext: () => void;
  onAppointmentClick: (reservation: Reservation) => void;
  onMonthDateClick: (date: Date) => void;
  onTimeSlotClick: (date: Date) => void;
  onAppointmentUpdate: (reservation: Reservation, newStart: Date, newEnd: Date) => void;
}

export function ReservationManagementCalendar({
  currentDate,
  view,
  days,
  doctorFilter,
  sourceFilter,
  doctorNames,
  appointments,
  activeEntries,
  dynamicColorMap,
  canCreate,
  onViewChange,
  onDoctorFilterChange,
  onSourceFilterChange,
  onDaysChange,
  onNavigatePrevious,
  onNavigateToday,
  onNavigateNext,
  onAppointmentClick,
  onMonthDateClick,
  onTimeSlotClick,
  onAppointmentUpdate,
}: ReservationManagementCalendarProps) {
  const doctorNameSelectItems = useMemo(
    () => doctorNames.map((name) => (
      <SelectItem key={name} value={name}>{name}</SelectItem>
    )),
    [doctorNames],
  );

  return (
    <div className="flex-1 flex flex-col p-4 overflow-hidden w-full min-w-0">
      <div className="flex flex-wrap items-center justify-between gap-2 mb-4">
        <div className="flex items-center gap-4">
          <div className={`flex items-center ${C.bgWhite} rounded-md border ${C.borderMedium} p-1 shadow-sm`}>
            <Button variant="ghost" size="icon" className="h-10 w-10" onClick={onNavigatePrevious}>
              <ChevronLeft className={ICON.page} />
            </Button>
            <Button variant="ghost" size="sm" className="h-10 px-4 text-base font-medium" onClick={onNavigateToday}>
              今日
            </Button>
            <Button variant="ghost" size="icon" className="h-10 w-10" onClick={onNavigateNext}>
              <ChevronRight className={ICON.page} />
            </Button>
          </div>
          <h2 className={`text-xl font-bold ${C.text} flex items-center gap-2`}>
            {format(currentDate, "yyyy年 M月", { locale: ja })}
          </h2>
        </div>

        <div className="flex items-center gap-2">
          <Select value={sourceFilter} onValueChange={onSourceFilterChange}>
            <SelectTrigger className={`w-[140px] ${C.bgWhite} ${C.borderMedium} h-10 text-base`}>
              <SelectValue placeholder="予約ソース" />
            </SelectTrigger>
            <SelectContent>{SOURCE_FILTER_SELECT_ITEMS}</SelectContent>
          </Select>

          <Select value={doctorFilter} onValueChange={onDoctorFilterChange}>
            <SelectTrigger className={`w-[160px] ${C.bgWhite} ${C.borderMedium} h-10 text-base`}>
              <SelectValue placeholder="担当医で絞込" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">すべての医師</SelectItem>
              {doctorNameSelectItems}
            </SelectContent>
          </Select>

          <Select value={view} onValueChange={typedSetter(onViewChange, CALENDAR_VIEW_VALUES)}>
            <SelectTrigger className={`w-[140px] ${C.bgWhite} ${C.borderMedium} h-10 text-base`}>
              <SelectValue placeholder="表示切替" />
            </SelectTrigger>
            <SelectContent>
              {CALENDAR_VIEW_SELECT_ITEMS}
            </SelectContent>
          </Select>

          {view === "week" ? (
            <DaysRangeToggle days={days} onChange={onDaysChange} />
          ) : null}
        </div>
      </div>

      <div className="flex items-center gap-3 mb-3 flex-wrap">
        {activeEntries.map((entry) => (
          <div key={entry.name} className="flex items-center gap-1.5">
            <span className={`${ICON.dotMd} rounded-full`} style={entry.color.dotStyle} />
            <span className={`text-base ${C.text60}`}>{entry.name}</span>
          </div>
        ))}
      </div>

      <div className="flex-1 overflow-hidden flex flex-col">
        <Suspense
          fallback={
            <div className="flex-1 flex items-center justify-center">
              <div className="text-center">
                <div className={`inline-block animate-spin rounded-full h-8 w-8 border-b-2 ${C.borderPrimary}`} />
                <p className={`mt-2 ${C.text60} text-base`}>読み込み中...</p>
              </div>
            </div>
          }
        >
          {view === "month" ? (
            <MonthView
              currentDate={currentDate}
              appointments={appointments}
              onAppointmentClick={onAppointmentClick}
              onDateClick={onMonthDateClick}
              dynamicColorMap={dynamicColorMap}
            />
          ) : (
            <WeekView
              currentDate={currentDate}
              appointments={appointments}
              onAppointmentClick={onAppointmentClick}
              onTimeSlotClick={canCreate ? onTimeSlotClick : undefined}
              onAppointmentUpdate={onAppointmentUpdate}
              dynamicColorMap={dynamicColorMap}
              days={days}
            />
          )}
        </Suspense>
      </div>
    </div>
  );
}
