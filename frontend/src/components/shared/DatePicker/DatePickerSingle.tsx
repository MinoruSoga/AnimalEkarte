import { useCallback, useMemo, useState } from "react";
import { CalendarIcon } from "lucide-react";
import { ja } from "date-fns/locale";

import { Calendar } from "@/components/ui/calendar";
import { Popover, PopoverAnchor, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { C, ICON } from "@/lib/design-tokens";
import { cn } from "@/lib/utils";
import {
  CalendarNav,
  ClearButton,
  MonthGrid,
  YearNav,
  type SingleDatePickerProps,
} from "./DatePickerParts";
import {
  formatDisplay,
  formatIso,
  formatSlash,
  parseDateInput,
  parseLocalDate,
  SINGLE_CALENDAR_CLASSES,
  TRIGGER_BASE,
} from "./DatePickerModel";

export function SinglePicker({
  value,
  onChange,
  placeholder = "日付を選択…",
  className,
  id,
  name,
  disabledDays,
}: SingleDatePickerProps) {
  const [open, setOpen] = useState(false);
  const [view, setView] = useState<"calendar" | "monthGrid">("calendar");
  const [displayMonth, setDisplayMonth] = useState<Date>(() => new Date());
  const [focused, setFocused] = useState(false);
  const [inputText, setInputText] = useState("");

  const selected = useMemo(() => parseLocalDate(value), [value]);
  const displayText = focused ? inputText : selected ? formatDisplay(selected) : "";

  const handleOpenChange = useCallback(
    (nextOpen: boolean) => {
      setOpen(nextOpen);
      if (nextOpen) {
        setView("calendar");
        setDisplayMonth(selected ?? new Date());
      }
    },
    [selected],
  );

  const handleSelect = useCallback(
    (day: Date | undefined) => {
      if (day) {
        onChange(formatIso(day));
        setInputText(formatSlash(day));
      }
      setOpen(false);
    },
    [onChange],
  );

  const handleClear = useCallback(
    (e: React.MouseEvent<HTMLButtonElement>) => {
      e.stopPropagation();
      onChange("");
      setInputText("");
    },
    [onChange],
  );

  const handleToday = useCallback(() => {
    const today = new Date();
    onChange(formatIso(today));
    setInputText(formatSlash(today));
    setOpen(false);
  }, [onChange]);

  const commitInput = useCallback(() => {
    const parsed = parseDateInput(inputText);
    if (parsed) {
      onChange(formatIso(parsed));
      setInputText(formatSlash(parsed));
    } else if (inputText.trim() === "") {
      onChange("");
      setInputText("");
    } else {
      setInputText(selected ? formatSlash(selected) : "");
    }
  }, [inputText, onChange, selected]);

  const handleInputKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLInputElement>) => {
      if (e.key === "Enter") {
        e.preventDefault();
        commitInput();
        setOpen(false);
      }
    },
    [commitInput],
  );

  const handleFocus = useCallback(() => {
    setFocused(true);
    setInputText(selected ? formatSlash(selected) : "");
  }, [selected]);

  const handleBlur = useCallback(() => {
    setFocused(false);
    commitInput();
  }, [commitInput]);

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
    <Popover open={open} onOpenChange={handleOpenChange}>
      <PopoverAnchor asChild>
        <div className={cn(TRIGGER_BASE, !value && !focused && C.text40, className)}>
          <PopoverTrigger asChild>
            <button
              type="button"
              className="-my-px min-h-11 min-w-11 shrink-0 cursor-pointer bg-transparent border-none p-0"
              aria-label="カレンダーを開く"
            >
              <CalendarIcon className={`${ICON.action} ${C.text40}`} />
            </button>
          </PopoverTrigger>
          <input
            id={id}
            name={name}
            type="text"
            value={displayText}
            onChange={(e) => setInputText(e.target.value)}
            onFocus={handleFocus}
            onBlur={handleBlur}
            onKeyDown={handleInputKeyDown}
            placeholder={placeholder}
            className={`-my-px min-h-11 flex-1 min-w-0 bg-transparent outline-none text-sm ${C.text} ${C.textPlaceholder} focus-visible:ring-2 ${C.focusRingAccent40}`}
          />
          {value ? <ClearButton onClick={handleClear} /> : null}
        </div>
      </PopoverAnchor>

      <PopoverContent className="w-auto p-0" align="start">
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
            mode="single"
            month={displayMonth}
            onMonthChange={setDisplayMonth}
            selected={selected}
            onSelect={handleSelect}
            disabled={disabledDays}
            locale={ja}
            fixedWeeks
            className="rounded-md pt-0"
            classNames={SINGLE_CALENDAR_CLASSES}
          />
        ) : (
          <MonthGrid
            currentMonth={displayMonth.getMonth()}
            onSelect={handleMonthSelect}
          />
        )}

        <div className={`border-t ${C.borderLight} px-3 py-1.5`}>
          <button
            type="button"
            onClick={handleToday}
            className={`inline-flex min-h-11 min-w-11 items-center justify-center rounded px-2 py-1 text-sm ${C.text60} ${C.hoverBgPage} ${C.hoverText} transition-colors`}
          >
            Today
          </button>
        </div>
      </PopoverContent>
    </Popover>
  );
}
