import { memo, useCallback, useState } from "react";
import { ja } from "date-fns/locale";
import type { DateRange } from "react-day-picker";

import {
  formatIso,
  formatShort,
  parseLocalDate,
  RANGE_CALENDAR_CLASSES,
} from "@/components/shared/DatePicker/DatePickerModel";
import {
  CalendarNav,
  MonthGrid,
  RangeEndpointNav,
  YearNav,
  type RangeEditTarget,
} from "@/components/shared/DatePicker/DatePickerParts";
import { Calendar } from "@/components/ui/calendar";
import { cn } from "@/components/ui/utils";
import { C } from "@/lib/design-tokens";
import { toJSTWallDate } from "@/lib/jst-date";

import { DATE_PRESETS, resolvePreset } from "./date-preset-utils";

interface DateValueEditorProps {
  currentValue?: { from?: string; to?: string };
  onApply: (value: { from?: string; to?: string }, displayValue: string) => void;
}

type EditTarget = RangeEditTarget;
type EditorView = "calendar" | "monthGrid";

const EDITOR_CALENDAR_CLASSES = {
  ...RANGE_CALENDAR_CLASSES,
  month_caption: "hidden",
};

export const DateValueEditor = memo(function DateValueEditor({
  currentValue,
  onApply,
}: DateValueEditorProps) {
  const [dateRange, setDateRange] = useState<DateRange | undefined>(() => {
    const from = currentValue?.from ? parseLocalDate(currentValue.from) : undefined;
    if (!from) return undefined;
    return {
      from,
      to: currentValue?.to ? parseLocalDate(currentValue.to) : undefined,
    };
  });
  const [editTarget, setEditTarget] = useState<EditTarget>(() =>
    currentValue?.from ? "to" : "from",
  );
  const [view, setView] = useState<EditorView>("calendar");
  const [displayMonth, setDisplayMonth] = useState<Date>(
    () =>
      (currentValue?.from ? parseLocalDate(currentValue.from) : undefined) ??
      toJSTWallDate(new Date()),
  );

  const applyRange = useCallback(
    (from: Date, to: Date) => {
      const fromIso = formatIso(from);
      const toIso = formatIso(to);
      onApply(
        { from: fromIso, to: toIso },
        fromIso === toIso ? formatShort(from) : `${formatShort(from)}〜${formatShort(to)}`,
      );
    },
    [onApply],
  );

  const handlePresetClick = useCallback(
    (from: Date, to: Date, label: string) => {
      setDateRange({ from, to });
      setEditTarget("to");
      setDisplayMonth(from);
      setView("calendar");
      onApply({ from: formatIso(from), to: formatIso(to) }, label);
    },
    [onApply],
  );

  const handleDayClick = useCallback(
    (day: Date) => {
      const from = dateRange?.from;
      const to = dateRange?.to;

      if (editTarget === "from" || !from) {
        const nextTo = to && day <= to ? to : undefined;
        setDateRange({ from: day, to: nextTo });
        setEditTarget("to");
        if (nextTo) applyRange(day, nextTo);
        return;
      }

      const next = day < from ? { from: day, to: from } : { from, to: day };
      setDateRange(next);
      applyRange(next.from, next.to);
    },
    [applyRange, dateRange, editTarget],
  );

  const handlePrevMonth = useCallback(() => {
    setDisplayMonth((prev) => new Date(prev.getFullYear(), prev.getMonth() - 1, 1));
  }, []);
  const handleNextMonth = useCallback(() => {
    setDisplayMonth((prev) => new Date(prev.getFullYear(), prev.getMonth() + 1, 1));
  }, []);
  const handleMonthSelect = useCallback((month: number) => {
    setDisplayMonth((prev) => new Date(prev.getFullYear(), month, 1));
    setView("calendar");
  }, []);
  const handleYearDelta = useCallback((delta: number) => {
    setDisplayMonth((prev) => new Date(prev.getFullYear() + delta, prev.getMonth(), 1));
  }, []);

  return (
    <div className={`flex divide-x ${C.divideDivider}`}>
      <div className="w-[108px] py-1 shrink-0">
        {DATE_PRESETS.map((preset) => (
          <button
            key={preset.label}
            type="button"
            onClick={() => {
              const { from, to } = resolvePreset(preset);
              handlePresetClick(from, to, preset.label);
            }}
            className={cn(
              `w-full text-left px-3 min-h-11 text-sm ${C.bgMutedBadge} ${C.hoverBgMutedBadge} transition-colors`,
              C.text,
            )}
          >
            {preset.label}
          </button>
        ))}
      </div>

      <div className="p-3">
        <div className="mb-2">
          <RangeEndpointNav
            from={dateRange?.from}
            to={dateRange?.to}
            editTarget={editTarget}
            onSelectTarget={setEditTarget}
          />
        </div>

        {view === "calendar" ? (
          <CalendarNav
            displayMonth={displayMonth}
            onPrev={handlePrevMonth}
            onNext={handleNextMonth}
            onTitleClick={() => setView("monthGrid")}
          />
        ) : (
          <YearNav
            year={displayMonth.getFullYear()}
            onPrevYear={() => handleYearDelta(-1)}
            onNextYear={() => handleYearDelta(1)}
          />
        )}

        {view === "calendar" ? (
          <Calendar
            mode="range"
            month={displayMonth}
            onMonthChange={setDisplayMonth}
            selected={dateRange}
            onDayClick={handleDayClick}
            numberOfMonths={1}
            locale={ja}
            fixedWeeks
            className="rounded-md pt-0"
            classNames={EDITOR_CALENDAR_CLASSES}
          />
        ) : (
          <MonthGrid currentMonth={displayMonth.getMonth()} onSelect={handleMonthSelect} />
        )}
      </div>
    </div>
  );
});
