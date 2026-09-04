import { useCallback, useMemo, useState } from "react";
import { CalendarIcon } from "lucide-react";
import { ja } from "date-fns/locale";
import type { DateRange } from "react-day-picker";

import { Calendar } from "@/components/ui/calendar";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { C, ICON } from "@/lib/design-tokens";
import { cn } from "@/lib/utils";
import {
  CalendarNav,
  ClearButton,
  MonthGrid,
  YearNav,
  type RangeDatePickerProps,
} from "./DatePickerParts";
import {
  formatIso,
  formatShort,
  parseRangeValue,
  RANGE_CALENDAR_CLASSES,
  TRIGGER_BASE,
} from "./DatePickerModel";

export function RangePicker({
  value,
  onChange,
  placeholder = "期間を選択…",
  className,
  id,
}: RangeDatePickerProps) {
  const [open, setOpen] = useState(false);
  const [view, setView] = useState<"calendar" | "monthGrid">("calendar");
  const [displayMonth, setDisplayMonth] = useState<Date>(() => new Date());

  const range = useMemo(() => parseRangeValue(value), [value]);
  const dateRange: DateRange | undefined = range.from
    ? { from: range.from, to: range.to }
    : undefined;

  const handleOpenChange = useCallback(
    (nextOpen: boolean) => {
      setOpen(nextOpen);
      if (nextOpen) {
        setView("calendar");
        setDisplayMonth(range.from ?? new Date());
      }
    },
    [range.from],
  );

  const handleSelect = useCallback(
    (selected: DateRange | undefined) => {
      if (!selected?.from) {
        onChange("");
        return;
      }
      const fromIso = formatIso(selected.from);
      const toIso = selected.to ? formatIso(selected.to) : "";
      onChange(`${fromIso}~${toIso}`);

      if (selected.from && selected.to) {
        setOpen(false);
      }
    },
    [onChange],
  );

  const handleClear = useCallback(
    (e: React.MouseEvent<HTMLButtonElement>) => {
      e.stopPropagation();
      onChange("");
    },
    [onChange],
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

  const displayLabel = range.from
    ? range.to
      ? `${formatShort(range.from)} 〜 ${formatShort(range.to)}`
      : `${formatShort(range.from)} 〜`
    : placeholder;

  return (
    <Popover open={open} onOpenChange={handleOpenChange}>
      <div className="relative w-full">
        <PopoverTrigger asChild>
          <button
            id={id}
            type="button"
            className={cn(TRIGGER_BASE, value && "pr-12", !value && C.text40, className)}
          >
            <span className="flex items-center gap-2 truncate">
              <CalendarIcon className={`${ICON.action} shrink-0 ${C.text40}`} />
              {displayLabel}
            </span>
          </button>
        </PopoverTrigger>
        {value ? (
          <span className="absolute inset-y-0 right-0 flex items-center">
            <ClearButton onClick={handleClear} />
          </span>
        ) : null}
      </div>

      <PopoverContent className="w-auto p-0" align="start">
        {view === "calendar" ? (
          <div className="pt-2">
            <CalendarNav
              displayMonth={displayMonth}
              onPrev={handlePrevMonth}
              onNext={handleNextMonth}
              onTitleClick={() => setView("monthGrid")}
            />
          </div>
        ) : (
          <div className="pt-2">
            <YearNav
              year={displayMonth.getFullYear()}
              onPrevYear={() => handleYearDelta(-1)}
              onNextYear={() => handleYearDelta(1)}
            />
          </div>
        )}

        {view === "calendar" ? (
          <Calendar
            mode="range"
            month={displayMonth}
            onMonthChange={setDisplayMonth}
            selected={dateRange}
            onSelect={handleSelect}
            numberOfMonths={2}
            locale={ja}
            fixedWeeks
            className="rounded-md pt-0"
            classNames={RANGE_CALENDAR_CLASSES}
          />
        ) : (
          <MonthGrid currentMonth={displayMonth.getMonth()} onSelect={handleMonthSelect} />
        )}
      </PopoverContent>
    </Popover>
  );
}
