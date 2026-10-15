import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { CalendarIcon } from "lucide-react";
import { dateMatchModifiers } from "react-day-picker";
import { ja } from "date-fns/locale";

import { Calendar } from "@/components/ui/calendar";
import { Popover, PopoverAnchor, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { C, ICON, STYLE } from "@/lib/design-tokens";
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

/**
 * Controlled single-date picker using an ISO `YYYY-MM-DD` value.
 * Calendar selections and typed commits respect `min`, `max`, and `disabledDays`; invalid values
 * are reported through the input's native validity state.
 */
export function SinglePicker({
  value,
  onChange,
  placeholder = "日付を選択…",
  className,
  id,
  name,
  ariaLabel,
  required,
  disabled,
  disabledDays,
  min,
  max,
  onInvalid,
  validationMessage,
  ariaInvalid = false,
  ariaDescribedBy,
  ariaRequired,
  showClearButton = true,
}: SingleDatePickerProps) {
  const [popoverState, setPopoverState] = useState({ disabled, open: false });
  if (popoverState.disabled !== disabled) {
    setPopoverState({ disabled, open: false });
  }
  const open = popoverState.open;
  const setOpen = useCallback(
    (nextOpen: boolean) => setPopoverState({ disabled, open: disabled ? false : nextOpen }),
    [disabled],
  );
  const [view, setView] = useState<"calendar" | "monthGrid">("calendar");
  const [displayMonth, setDisplayMonth] = useState<Date>(() => new Date());
  const [focused, setFocused] = useState(false);
  const [inputText, setInputText] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);
  const lastInputCommitRef = useRef<string | null>(null);

  const selected = useMemo(() => parseLocalDate(value), [value]);
  const displayText = focused ? inputText : selected ? formatDisplay(selected) : "";
  const isDateOutOfRange = useCallback(
    (date: Date) => {
      const isoDate = formatIso(date);
      return (min !== undefined && isoDate < min) || (max !== undefined && isoDate > max);
    },
    [max, min],
  );
  const isDateDisabled = useCallback(
    (date: Date) =>
      (disabledDays ? dateMatchModifiers(date, disabledDays) : false) || isDateOutOfRange(date),
    [disabledDays, isDateOutOfRange],
  );
  const validityDate = value ? selected : (parseDateInput(inputText) ?? undefined);
  const nativeValidityMessage = disabled
    ? ""
    : validationMessage ||
      (validityDate && isDateOutOfRange(validityDate)
        ? min !== undefined && formatIso(validityDate) < min
          ? `日付は ${min} 以降を指定してください`
          : `日付は ${max} 以前を指定してください`
        : validityDate && disabledDays && dateMatchModifiers(validityDate, disabledDays)
          ? "入力した日付は選択できません"
          : "");

  useEffect(() => {
    inputRef.current?.setCustomValidity(nativeValidityMessage);
  }, [nativeValidityMessage]);

  const handleOpenChange = useCallback(
    (nextOpen: boolean) => {
      setOpen(disabled ? false : nextOpen);
      if (nextOpen && !disabled) {
        setView("calendar");
        setDisplayMonth(selected ?? new Date());
      }
    },
    [disabled, selected, setOpen],
  );

  const handleSelect = useCallback(
    (day: Date | undefined) => {
      if (!disabled && day && !isDateDisabled(day)) {
        onChange(formatIso(day));
        setInputText(formatSlash(day));
        setOpen(false);
      }
    },
    [disabled, isDateDisabled, onChange, setOpen],
  );

  const handleClear = useCallback(
    (e: React.MouseEvent<HTMLButtonElement>) => {
      e.stopPropagation();
      if (disabled) return;
      onChange("");
      setInputText("");
    },
    [disabled, onChange],
  );

  const handleToday = useCallback(() => {
    const today = new Date();
    if (disabled || isDateDisabled(today)) return;
    onChange(formatIso(today));
    setInputText(formatSlash(today));
    setOpen(false);
  }, [disabled, isDateDisabled, onChange, setOpen]);

  const commitInput = useCallback(() => {
    if (disabled) return;
    const parsed = parseDateInput(inputText);
    if (parsed || inputText.trim() === "") {
      const candidate = parsed ? formatIso(parsed) : "";
      if (candidate !== lastInputCommitRef.current) {
        lastInputCommitRef.current = candidate;
        onChange(candidate);
      }
      setInputText(parsed ? formatSlash(parsed) : "");
    } else {
      setInputText(selected ? formatSlash(selected) : "");
    }
  }, [disabled, inputText, onChange, selected]);

  const handleInputKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLInputElement>) => {
      if (e.key === "Enter") {
        e.preventDefault();
        commitInput();
        setOpen(false);
      }
    },
    [commitInput, setOpen],
  );

  const handleFocus = useCallback(() => {
    lastInputCommitRef.current = null;
    setFocused(true);
    setInputText(selected ? formatSlash(selected) : "");
  }, [selected]);

  const handleBlur = useCallback(() => {
    setFocused(false);
    commitInput();
    lastInputCommitRef.current = null;
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
        <div
          className={cn(
            TRIGGER_BASE,
            !value && !focused && C.text40,
            ariaInvalid && STYLE.formInputError,
            className,
          )}
        >
          <PopoverTrigger asChild>
            <button
              type="button"
              disabled={disabled}
              className="-my-px min-h-11 min-w-11 shrink-0 cursor-pointer bg-transparent border-none p-0"
              aria-label={ariaLabel ? `${ariaLabel}を選択` : "カレンダーを開く"}
            >
              <CalendarIcon className={`${ICON.action} ${C.text60}`} />
            </button>
          </PopoverTrigger>
          <input
            ref={inputRef}
            id={id}
            type="text"
            required={required}
            disabled={disabled}
            aria-label={ariaLabel}
            aria-invalid={ariaInvalid || undefined}
            aria-required={ariaRequired || undefined}
            aria-describedby={ariaDescribedBy}
            onInvalid={onInvalid}
            value={displayText}
            onChange={(e) => {
              lastInputCommitRef.current = null;
              setInputText(e.target.value);
            }}
            onFocus={handleFocus}
            onBlur={handleBlur}
            onKeyDown={handleInputKeyDown}
            placeholder={placeholder}
            className={`-my-px min-h-11 flex-1 min-w-0 bg-transparent outline-none text-sm ${C.text} ${C.textPlaceholder} focus-visible:ring-2 ${C.focusRingAccent40}`}
          />
          {value && showClearButton ? (
            <ClearButton onClick={handleClear} disabled={disabled} />
          ) : null}
        </div>
      </PopoverAnchor>

      {name !== undefined ? (
        <input type="hidden" name={name} value={value} disabled={disabled} />
      ) : null}

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
            disabled={isDateDisabled}
            locale={ja}
            fixedWeeks
            className="rounded-md pt-0"
            classNames={SINGLE_CALENDAR_CLASSES}
          />
        ) : (
          <MonthGrid currentMonth={displayMonth.getMonth()} onSelect={handleMonthSelect} />
        )}

        <div className={`border-t ${C.borderLight} px-3 py-1.5`}>
          <button
            type="button"
            onClick={handleToday}
            disabled={disabled || isDateDisabled(new Date())}
            className={`inline-flex min-h-11 min-w-11 items-center justify-center rounded px-2 py-1 text-sm ${C.text60} ${C.hoverBgPage} ${C.hoverText} focus-visible:ring-2 focus-visible:ring-offset-1 ${C.focusRingAccent40} transition-colors`}
          >
            Today
          </button>
        </div>
      </PopoverContent>
    </Popover>
  );
}
