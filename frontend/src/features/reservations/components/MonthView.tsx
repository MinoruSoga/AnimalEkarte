// React/Framework
import { memo, useMemo } from "react";
import type React from "react";

// External
import {
  startOfMonth,
  endOfMonth,
  startOfWeek,
  endOfWeek,
  format,
  isSameMonth,
  isSameDay,
  addDays,
} from "date-fns";
import { ja } from "date-fns/locale";

// Internal
import { BirthdayBadge } from "@/components/shared/BirthdayBadge";
import { C } from "@/lib/design-tokens";
import { isBirthMonthJST } from "@/lib/jst-date";
import { toJSTWallDate } from "@/lib/jst-date";
import { getReservationTypeColor } from "@/lib/status-helpers";
import { DAY_OF_WEEK_LABEL_LIST } from "@/constants/day-of-week";
import { isTentativeReservation, TENTATIVE_APPOINTMENT_LABEL } from "../lib/tentative";

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
  /** BUG-343: 休診日 (YYYY-MM-DD) */
  holidayDates?: ReadonlySet<string>;
}

const HEADER_ROW = (
  <div className={`grid grid-cols-7 border-b ${C.borderMedium} ${C.bgPage}`}>
    {DAY_OF_WEEK_LABEL_LIST.map((d, i) => (
      <div
        key={d}
        className={`py-3 text-sm font-semibold text-center ${
          i === 0 ? C.textNotionRed : i === 6 ? C.textBrand : C.text60
        }`}
      >
        {d}
      </div>
    ))}
  </div>
);

export const MonthView = memo(function MonthView({
  currentDate,
  appointments,
  onAppointmentClick,
  onDateClick,
  dynamicColorMap,
  holidayDates,
}: MonthViewProps) {
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

        const dayAppointments = appointments.filter((app) => isSameDay(app.start, cloneDay));
        const dateKey = format(cloneDay, "yyyy-MM-dd");
        const isHoliday = holidayDates?.has(dateKey) ?? false;

        days.push(
          <div
            key={day.toString()}
            className={`h-full min-h-[140px] ${C.bgWhite} border-b border-r ${C.borderLight} p-2 transition-colors ${C.hoverBgPage} flex flex-col
              ${!isSameMonth(day, monthStart) ? `${C.bgPage30} ${C.text30}` : C.text}
              ${isSameDay(day, today) ? C.bgBrand8 : ""}
              ${isHoliday ? `${C.bgPage} opacity-70` : ""}
            `}
            title={isHoliday ? "休診日" : undefined}
          >
            <div className="flex justify-between items-start mb-2">
              <button
                type="button"
                className={`text-base font-semibold min-h-11 min-w-11 flex items-center justify-center rounded-full transition-colors ${isSameDay(day, today) ? `${C.bgActionPrimarySolid} ${C.textOnActionPrimary} ${C.hoverBgActionPrimarySolid} ${C.hoverTextOnActionPrimary} ${C.activeBgActionPrimarySolid} ${C.activeTextOnActionPrimary}` : `${C.hoverBgBrandLight} ${C.hoverTextBrand}`}`}
                onClick={(e) => {
                  e.stopPropagation();
                  onDateClick?.(cloneDay);
                }}
                aria-label={`${format(cloneDay, "M月d日")}の週表示へ${isHoliday ? "（休診日）" : ""}`}
              >
                {formattedDate}
              </button>
              {isHoliday ? (
                <span className={`text-2xs font-medium ${C.textNotionRed}`}>休診</span>
              ) : null}
            </div>
            <div className="space-y-1.5 flex-1 overflow-hidden">
              {dayAppointments.slice(0, 4).map((app) => {
                const colorStyle = getReservationTypeColor(app.type, dynamicColorMap);
                const isClassNameColor = typeof colorStyle === "string";
                // EMR-308: tentative（次回予定・未予約）は色だけでなく明示テキストで識別する
                const isTentative = isTentativeReservation(app);
                // EMR-301: static 変種バッジの title はチップのボタン title/aria-label に
                // 上書きされて到達不能。生年月日は親の tooltip/accessible name に含める。
                const isDeceasedPet = app.petStatus === "deceased" || app.petDeceasedAt != null;
                const birthdayLabel =
                  !isDeceasedPet && isBirthMonthJST(app.petBirthDate)
                    ? `今月誕生日（生年月日: ${app.petBirthDate}）`
                    : "";
                const chipTitle = [isTentative ? TENTATIVE_APPOINTMENT_LABEL : "", birthdayLabel]
                  .filter(Boolean)
                  .join(" / ");
                // PR#614 再レビュー: 子テキストで見える意味情報（tentative・飼主名・
                // 非空の担当医）を accessible name から隠さない。WeekViewAppointmentCard
                // と同方針で全情報を合成する（MonthView では担当医も含める）。
                const accessibleLabel = [
                  isTentative ? TENTATIVE_APPOINTMENT_LABEL : "",
                  app.visitType === "first" ? "初診" : "再診",
                  app.petName,
                  app.ownerName,
                  app.doctor,
                  birthdayLabel,
                ]
                  .filter(Boolean)
                  .join(" ");
                return (
                  <button
                    key={app.id}
                    type="button"
                    className={`block w-full min-h-11 text-left text-sm px-2 py-1.5 rounded border cursor-pointer hover:opacity-80 leading-tight ${isClassNameColor ? colorStyle : ""} ${isTentative ? "border-dashed" : ""}`}
                    style={isClassNameColor ? undefined : (colorStyle as React.CSSProperties)}
                    title={chipTitle || undefined}
                    aria-label={accessibleLabel}
                    onClick={(e) => {
                      e.stopPropagation();
                      onAppointmentClick(app);
                    }}
                  >
                    <div className="flex items-center gap-1 min-w-0">
                      {app.visitType === "first" ? (
                        <span
                          className={`${C.bgRedLight} ${C.textNotionRed} text-2xs px-1 rounded flex-shrink-0`}
                        >
                          初
                        </span>
                      ) : (
                        <span
                          className={`${C.bgBrandLight} ${C.textBrandDark} text-2xs px-1 rounded flex-shrink-0`}
                        >
                          再
                        </span>
                      )}
                      <span className="truncate text-xs font-medium">{app.petName}</span>
                      {/* EMR-301: 誕生月 Cake 小アイコン（チップ自体が button のため static 変種） */}
                      {!isDeceasedPet ? (
                        <BirthdayBadge
                          birthDate={app.petBirthDate}
                          subjectName={app.petName}
                          variant="static"
                        />
                      ) : null}
                    </div>
                    {isTentative ? (
                      <div className={`text-2xs font-semibold ${C.textWarning} truncate mt-0.5`}>
                        {TENTATIVE_APPOINTMENT_LABEL}
                      </div>
                    ) : null}
                    <div className="text-2xs opacity-70 truncate mt-0.5">
                      {app.ownerName}
                      {app.doctor ? ` / ${app.doctor}` : ""}
                    </div>
                  </button>
                );
              })}
              {dayAppointments.length > 4 ? (
                <div className={`text-sm ${C.text60} pl-1`}>
                  他 {(dayAppointments.length - 4).toLocaleString()} 件
                </div>
              ) : null}
            </div>
          </div>,
        );
        day = addDays(day, 1);
      }
      result.push(
        <div className="grid grid-cols-7 flex-1" key={day.toString()}>
          {days}
        </div>,
      );
      days = [];
    }

    return result;
  }, [currentDate, appointments, dynamicColorMap, onAppointmentClick, onDateClick, holidayDates]);

  return (
    <div
      className={`flex flex-col h-full border-l border-t ${C.borderMedium} rounded-lg overflow-hidden ${C.bgWhite}`}
    >
      {HEADER_ROW}
      <div className="flex-1 flex flex-col">{rows}</div>
    </div>
  );
});
