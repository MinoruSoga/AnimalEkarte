import { useState, useCallback, useMemo } from "react";
import { CalendarIcon, ChevronLeft, ChevronRight, X } from "lucide-react";
import { ja } from "date-fns/locale";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Calendar } from "@/components/ui/calendar";
import { cn } from "@/lib/utils";
import type { DateRange } from "react-day-picker";

// ─── Props ────────────────────────────────────────────────────

/** Single date picker props */
interface SingleDatePickerProps {
  mode?: "single";
  /** ISO date string (YYYY-MM-DD) or empty string */
  value: string;
  /** Called with ISO date string or empty string when cleared */
  onChange: (value: string) => void;
  placeholder?: string;
  className?: string;
  id?: string;
}

/** Range date picker props */
interface RangeDatePickerProps {
  mode: "range";
  /** ISO date range string "YYYY-MM-DD~YYYY-MM-DD" or partial "YYYY-MM-DD~" or empty string */
  value: string;
  /** Called with ISO date range string or empty string when cleared */
  onChange: (value: string) => void;
  placeholder?: string;
  className?: string;
  id?: string;
}

type NotionDatePickerProps = SingleDatePickerProps | RangeDatePickerProps;

// ─── Utilities ────────────────────────────────────────────────

/** Parses "YYYY-MM-DD" into a local Date (noon to avoid timezone shifts). */
function parseLocalDate(iso: string): Date | undefined {
  if (!iso) return undefined;
  const [y, m, d] = iso.split("-").map(Number);
  if (!y || !m || !d) return undefined;
  return new Date(y, m - 1, d, 12, 0, 0);
}

/** Formats a Date as "YYYY-MM-DD". */
function formatIso(date: Date): string {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, "0");
  const d = String(date.getDate()).padStart(2, "0");
  return `${y}-${m}-${d}`;
}

/** Formats a Date as a human-readable Japanese date string. */
function formatDisplay(date: Date): string {
  const y = date.getFullYear();
  const m = date.getMonth() + 1;
  const d = date.getDate();
  const weekdays = ["日", "月", "火", "水", "木", "金", "土"];
  const w = weekdays[date.getDay()];
  return `${y}年${m}月${d}日（${w}）`;
}

/** Formats a Date as a short Japanese date string for range display. */
function formatShort(date: Date): string {
  const y = date.getFullYear();
  const m = date.getMonth() + 1;
  const d = date.getDate();
  return `${y}/${m}/${d}`;
}

/** Parses "YYYY-MM-DD~YYYY-MM-DD" into { from, to }. */
function parseRangeValue(value: string): { from?: Date; to?: Date } {
  if (!value) return {};
  const [fromStr, toStr] = value.split("~");
  return {
    from: parseLocalDate(fromStr),
    to: toStr ? parseLocalDate(toStr) : undefined,
  };
}

/** Parses user text input into a Date. Supports YYYY/MM/DD, YYYY-MM-DD, YYYYMMDD. */
function parseDateInput(input: string): Date | null {
  const trimmed = input.trim();
  if (!trimmed) return null;

  let y: number;
  let m: number;
  let d: number;

  // YYYYMMDD
  const compact = trimmed.match(/^(\d{4})(\d{2})(\d{2})$/);
  if (compact) {
    y = Number(compact[1]);
    m = Number(compact[2]);
    d = Number(compact[3]);
  } else {
    // YYYY/MM/DD or YYYY-MM-DD
    const separated = trimmed.match(/^(\d{4})[/-](\d{1,2})[/-](\d{1,2})$/);
    if (!separated) return null;
    y = Number(separated[1]);
    m = Number(separated[2]);
    d = Number(separated[3]);
  }

  if (m < 1 || m > 12 || d < 1 || d > 31) return null;
  const date = new Date(y, m - 1, d, 12, 0, 0);
  // Validate the date is real (e.g., Feb 30 → invalid)
  if (date.getMonth() !== m - 1 || date.getDate() !== d) return null;
  return date;
}

// ─── Constants ────────────────────────────────────────────────

const MONTH_LABELS = [
  "1月", "2月", "3月", "4月", "5月", "6月",
  "7月", "8月", "9月", "10月", "11月", "12月",
];

/** Hide built-in nav + caption; our custom header replaces them */
const SINGLE_CALENDAR_CLASSES = {
  selected:
    "bg-[#37352F] text-white hover:bg-[#37352F] hover:text-white focus:bg-[#37352F] focus:text-white",
  today: "bg-[#F7F6F3] text-[#37352F]",
  nav: "hidden",
  month_caption: "hidden",
};

/** Hide built-in nav but keep per-month caption for 2-month range view */
const RANGE_CALENDAR_CLASSES = {
  selected:
    "bg-[#37352F] text-white hover:bg-[#37352F] hover:text-white focus:bg-[#37352F] focus:text-white",
  today: "bg-[#F7F6F3] text-[#37352F]",
  nav: "hidden",
};

const TRIGGER_BASE =
  "flex h-11 w-full items-center justify-between rounded-md border border-[rgba(55,53,47,0.16)] bg-white px-3 text-sm text-[#37352F] transition-colors hover:bg-[#F7F6F3] focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring";

const NAV_BTN =
  "inline-flex items-center justify-center rounded p-1 text-[#37352F]/50 hover:bg-[#37352F]/8 hover:text-[#37352F] transition-colors";

// ─── NotionDatePicker (entry) ─────────────────────────────────

/**
 * Notion-style date picker with single and range modes.
 *
 * - `mode="single"` (default): single date selection with text input + Today button
 * - `mode="range"`: date range selection (from ~ to)
 */
export function NotionDatePicker(props: NotionDatePickerProps) {
  const isRange = props.mode === "range";

  if (isRange) {
    return <RangePicker {...props} />;
  }
  return <SinglePicker {...props} />;
}

// ─── Single Mode ──────────────────────────────────────────────

function SinglePicker({
  value,
  onChange,
  placeholder = "日付を選択…",
  className,
  id,
}: SingleDatePickerProps) {
  const [open, setOpen] = useState(false);
  const [view, setView] = useState<"calendar" | "monthGrid">("calendar");
  const [displayMonth, setDisplayMonth] = useState<Date>(() => new Date());
  const [textInput, setTextInput] = useState("");

  const selected = useMemo(() => parseLocalDate(value), [value]);

  const handleOpenChange = useCallback(
    (nextOpen: boolean) => {
      setOpen(nextOpen);
      if (nextOpen) {
        setView("calendar");
        setDisplayMonth(selected ?? new Date());
        setTextInput(value ? value.replace(/-/g, "/") : "");
      }
    },
    [selected, value],
  );

  const handleSelect = useCallback(
    (day: Date | undefined) => {
      if (day) {
        onChange(formatIso(day));
      }
      setOpen(false);
    },
    [onChange],
  );

  const handleClear = useCallback(
    (e: React.MouseEvent<HTMLSpanElement>) => {
      e.stopPropagation();
      onChange("");
    },
    [onChange],
  );

  const handleToday = useCallback(() => {
    onChange(formatIso(new Date()));
    setOpen(false);
  }, [onChange]);

  const handleTextKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLInputElement>) => {
      if (e.key === "Enter") {
        const parsed = parseDateInput(textInput);
        if (parsed) {
          onChange(formatIso(parsed));
          setOpen(false);
        }
      }
    },
    [textInput, onChange],
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
    <Popover open={open} onOpenChange={handleOpenChange}>
      <PopoverTrigger asChild>
        <button
          id={id}
          type="button"
          className={cn(TRIGGER_BASE, !value && "text-[#37352F]/40", className)}
        >
          <span className="flex items-center gap-2 truncate">
            <CalendarIcon className="h-4 w-4 shrink-0 text-[#37352F]/40" />
            {selected ? formatDisplay(selected) : placeholder}
          </span>
          {value ? <ClearButton onClick={handleClear} /> : null}
        </button>
      </PopoverTrigger>

      <PopoverContent className="w-auto p-0" align="start">
        {/* ── Text input ── */}
        <div className="px-3 pt-3 pb-1">
          <input
            type="text"
            value={textInput}
            onChange={(e) => setTextInput(e.target.value)}
            onKeyDown={handleTextKeyDown}
            placeholder="YYYY/MM/DD"
            className="w-full rounded border border-[rgba(55,53,47,0.16)] bg-white px-2.5 py-1.5 text-sm text-[#37352F] outline-none placeholder:text-[#37352F]/30 focus:border-[#37352F]/40 focus:ring-1 focus:ring-[#37352F]/20"
            autoFocus
          />
        </div>

        {/* ── Navigation header ── */}
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

        {/* ── Calendar / Month grid ── */}
        {view === "calendar" ? (
          <Calendar
            mode="single"
            month={displayMonth}
            onMonthChange={setDisplayMonth}
            selected={selected}
            onSelect={handleSelect}
            locale={ja}
            className="rounded-md pt-0"
            classNames={SINGLE_CALENDAR_CLASSES}
          />
        ) : (
          <MonthGrid
            currentMonth={displayMonth.getMonth()}
            onSelect={handleMonthSelect}
          />
        )}

        {/* ── Today button ── */}
        <div className="border-t border-[rgba(55,53,47,0.09)] px-3 py-1.5">
          <button
            type="button"
            onClick={handleToday}
            className="rounded px-2 py-1 text-sm text-[#37352F]/60 hover:bg-[#F7F6F3] hover:text-[#37352F] transition-colors"
          >
            Today
          </button>
        </div>
      </PopoverContent>
    </Popover>
  );
}

// ─── Range Mode ───────────────────────────────────────────────

function RangePicker({
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
    (e: React.MouseEvent<HTMLSpanElement>) => {
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
      <PopoverTrigger asChild>
        <button
          id={id}
          type="button"
          className={cn(TRIGGER_BASE, !value && "text-[#37352F]/40", className)}
        >
          <span className="flex items-center gap-2 truncate">
            <CalendarIcon className="h-4 w-4 shrink-0 text-[#37352F]/40" />
            {displayLabel}
          </span>
          {value ? <ClearButton onClick={handleClear} /> : null}
        </button>
      </PopoverTrigger>

      <PopoverContent className="w-auto p-0" align="start">
        {/* ── Navigation header ── */}
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

        {/* ── Calendar / Month grid ── */}
        {view === "calendar" ? (
          <Calendar
            mode="range"
            month={displayMonth}
            onMonthChange={setDisplayMonth}
            selected={dateRange}
            onSelect={handleSelect}
            numberOfMonths={2}
            locale={ja}
            className="rounded-md pt-0"
            classNames={RANGE_CALENDAR_CLASSES}
          />
        ) : (
          <MonthGrid
            currentMonth={displayMonth.getMonth()}
            onSelect={handleMonthSelect}
          />
        )}
      </PopoverContent>
    </Popover>
  );
}

// ─── Shared Components ────────────────────────────────────────

/** Calendar view header: ◀ 2026年 3月 ▶ — clickable title opens month grid */
function CalendarNav({
  displayMonth,
  onPrev,
  onNext,
  onTitleClick,
}: {
  displayMonth: Date;
  onPrev: () => void;
  onNext: () => void;
  onTitleClick: () => void;
}) {
  const year = displayMonth.getFullYear();
  const month = displayMonth.getMonth() + 1;

  return (
    <div className="flex items-center justify-between px-3 pb-1">
      <button type="button" onClick={onPrev} className={NAV_BTN} aria-label="前の月">
        <ChevronLeft className="h-4 w-4" />
      </button>
      <button
        type="button"
        onClick={onTitleClick}
        className="rounded px-2 py-1 text-sm font-medium text-[#37352F] hover:bg-[#F7F6F3] transition-colors"
      >
        {year}年 {month}月
      </button>
      <button type="button" onClick={onNext} className={NAV_BTN} aria-label="次の月">
        <ChevronRight className="h-4 w-4" />
      </button>
    </div>
  );
}

/** Month grid view header: ◀ 2026年 ▶ */
function YearNav({
  year,
  onPrevYear,
  onNextYear,
}: {
  year: number;
  onPrevYear: () => void;
  onNextYear: () => void;
}) {
  return (
    <div className="flex items-center justify-between px-3 pb-2 pt-1">
      <button type="button" onClick={onPrevYear} className={NAV_BTN} aria-label="前の年">
        <ChevronLeft className="h-4 w-4" />
      </button>
      <span className="text-sm font-medium text-[#37352F]">{year}年</span>
      <button type="button" onClick={onNextYear} className={NAV_BTN} aria-label="次の年">
        <ChevronRight className="h-4 w-4" />
      </button>
    </div>
  );
}

/** 4×3 month selection grid */
function MonthGrid({
  currentMonth,
  onSelect,
}: {
  currentMonth: number;
  onSelect: (month: number) => void;
}) {
  return (
    <div className="grid grid-cols-4 gap-1 px-3 pb-3">
      {MONTH_LABELS.map((label, i) => (
        <button
          key={label}
          type="button"
          onClick={() => onSelect(i)}
          className={cn(
            "rounded px-2 py-2 text-sm transition-colors",
            i === currentMonth
              ? "bg-[#37352F] text-white font-medium"
              : "text-[#37352F] hover:bg-[#F7F6F3]",
          )}
        >
          {label}
        </button>
      ))}
    </div>
  );
}

/** Clear (×) button for the trigger */
function ClearButton({ onClick }: { onClick: (e: React.MouseEvent<HTMLSpanElement>) => void }) {
  return (
    <span
      role="button"
      tabIndex={0}
      onClick={onClick}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ")
          onClick(e as unknown as React.MouseEvent<HTMLSpanElement>);
      }}
      className="ml-1 shrink-0 rounded p-0.5 text-[#37352F]/40 hover:bg-[#37352F]/10 hover:text-[#37352F]/70 cursor-pointer"
      aria-label="日付をクリア"
    >
      <X className="h-3.5 w-3.5" />
    </span>
  );
}
