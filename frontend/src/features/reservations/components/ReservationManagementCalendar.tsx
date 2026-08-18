import { lazy, Suspense, useMemo } from "react";
import type React from "react";
import { format } from "date-fns";
import { ja } from "date-fns/locale";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { SearchableSelect, type SearchableSelectOption } from "@/components/ui/searchable-select";
import { CalendarNavToolbar } from "@/components/shared/CalendarNavToolbar";
import type { LegendEntry, ReservationTypeColor } from "@/hooks/use-reservation-type-color-map";
import { C, ICON } from "@/lib/design-tokens";
import { typedSetter } from "@/lib/type-utils";
import { getCalendarViewLabel } from "@/lib/status-helpers";
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
  holidayDates?: ReadonlySet<string>;
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
  holidayDates,
}: ReservationManagementCalendarProps) {
  const doctorFilterOptions = useMemo<SearchableSelectOption[]>(
    () => [
      { value: "all", label: "すべての医師" },
      ...doctorNames.map((name) => ({ value: name, label: name })),
    ],
    [doctorNames],
  );

  return (
    <div className="flex-1 flex flex-col p-4 overflow-hidden w-full min-w-0">
      <div
        className="flex flex-col items-stretch gap-2 mb-4 sm:flex-row sm:flex-wrap sm:items-center sm:justify-between"
        data-testid="reservation-toolbar"
      >
        <CalendarNavToolbar
          onPrev={onNavigatePrevious}
          onToday={onNavigateToday}
          onNext={onNavigateNext}
          prevAriaLabel="前の月"
          nextAriaLabel="次の月"
          label={
            <h2 className={`text-xl font-bold ${C.text} flex items-center gap-2`}>
              {format(currentDate, "yyyy年 M月", { locale: ja })}
            </h2>
          }
        />

        <div
          className="flex w-full flex-wrap items-center gap-2 sm:w-auto"
          data-testid="reservation-toolbar-filters"
        >
          <Select value={sourceFilter} onValueChange={onSourceFilterChange}>
            <SelectTrigger aria-label="予約ソース" className={`w-[140px] ${C.bgWhite} ${C.borderMedium} h-11 text-base`}>
              <SelectValue placeholder="予約ソース" />
            </SelectTrigger>
            <SelectContent>{SOURCE_FILTER_SELECT_ITEMS}</SelectContent>
          </Select>

          <SearchableSelect
            value={doctorFilter}
            onValueChange={onDoctorFilterChange}
            options={doctorFilterOptions}
            ariaLabel="担当医絞り込み"
            placeholder="担当医で絞込"
            searchPlaceholder="担当医を検索..."
            className="w-[160px] h-11"
          />

          <Select value={view} onValueChange={typedSetter(onViewChange, CALENDAR_VIEW_VALUES)}>
            <SelectTrigger aria-label="カレンダー表示切替" className={`w-[140px] ${C.bgWhite} ${C.borderMedium} h-11 text-base`}>
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
              holidayDates={holidayDates}
            />
          ) : (
            <WeekView
              currentDate={currentDate}
              appointments={appointments}
              onAppointmentClick={onAppointmentClick}
              onTimeSlotClick={
                canCreate
                  ? (date) => {
                      if (holidayDates?.has(format(date, "yyyy-MM-dd"))) return;
                      onTimeSlotClick(date);
                    }
                  : undefined
              }
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
