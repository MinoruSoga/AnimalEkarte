import { useState, useCallback, useMemo } from "react";
import { CalendarIcon, X } from "lucide-react";
import { ja } from "date-fns/locale";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Calendar } from "@/components/ui/calendar";
import { cn } from "@/lib/utils";
import type { DateRange } from "react-day-picker";

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

// Calendar date range for dropdown navigation (100 years back → current year)
const CALENDAR_START_MONTH = new Date(new Date().getFullYear() - 100, 0);
const CALENDAR_END_MONTH = new Date(new Date().getFullYear(), 11);

// Calendar classNames shared between modes
const CALENDAR_CLASS_NAMES = {
  selected:
    "bg-[#37352F] text-white hover:bg-[#37352F] hover:text-white focus:bg-[#37352F] focus:text-white",
  today: "bg-[#F7F6F3] text-[#37352F]",
};

const TRIGGER_BASE =
  "flex h-11 w-full items-center justify-between rounded-md border border-[rgba(55,53,47,0.16)] bg-white px-3 text-sm text-[#37352F] transition-colors hover:bg-[#F7F6F3] focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring";

/**
 * Notion-style date picker with single and range modes.
 *
 * - `mode="single"` (default): single date selection
 * - `mode="range"`: date range selection (from ~ to)
 */
export function NotionDatePicker(props: NotionDatePickerProps) {
  const isRange = props.mode === "range";

  if (isRange) {
    return <RangePicker {...props} />;
  }
  return <SinglePicker {...props} />;
}

// ─── Single Mode ───────────────────────────────────────────────

function SinglePicker({
  value,
  onChange,
  placeholder = "日付を選択…",
  className,
  id,
}: SingleDatePickerProps) {
  const [open, setOpen] = useState(false);

  const selected = useMemo(() => parseLocalDate(value), [value]);

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

  return (
    <Popover open={open} onOpenChange={setOpen}>
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
          {value ? (
            <ClearButton onClick={handleClear} />
          ) : null}
        </button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0" align="start">
        <Calendar
          mode="single"
          captionLayout="dropdown"
          startMonth={CALENDAR_START_MONTH}
          endMonth={CALENDAR_END_MONTH}
          reverseYears
          selected={selected}
          onSelect={handleSelect}
          defaultMonth={selected ?? new Date()}
          locale={ja}
          className="rounded-md"
          classNames={CALENDAR_CLASS_NAMES}
        />
      </PopoverContent>
    </Popover>
  );
}

// ─── Range Mode ────────────────────────────────────────────────

function RangePicker({
  value,
  onChange,
  placeholder = "期間を選択…",
  className,
  id,
}: RangeDatePickerProps) {
  const [open, setOpen] = useState(false);

  const range = useMemo(() => parseRangeValue(value), [value]);
  const dateRange: DateRange | undefined = range.from
    ? { from: range.from, to: range.to }
    : undefined;

  const handleSelect = useCallback(
    (selected: DateRange | undefined) => {
      if (!selected?.from) {
        onChange("");
        return;
      }
      const fromIso = formatIso(selected.from);
      const toIso = selected.to ? formatIso(selected.to) : "";
      onChange(`${fromIso}~${toIso}`);

      // 両方選択されたら閉じる
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

  const displayLabel = range.from
    ? range.to
      ? `${formatShort(range.from)} 〜 ${formatShort(range.to)}`
      : `${formatShort(range.from)} 〜`
    : placeholder;

  return (
    <Popover open={open} onOpenChange={setOpen}>
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
          {value ? (
            <ClearButton onClick={handleClear} />
          ) : null}
        </button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0" align="start">
        <Calendar
          mode="range"
          captionLayout="dropdown"
          startMonth={CALENDAR_START_MONTH}
          endMonth={CALENDAR_END_MONTH}
          reverseYears
          selected={dateRange}
          onSelect={handleSelect}
          numberOfMonths={2}
          defaultMonth={range.from ?? new Date()}
          locale={ja}
          className="rounded-md"
          classNames={CALENDAR_CLASS_NAMES}
        />
      </PopoverContent>
    </Popover>
  );
}

// ─── Shared ────────────────────────────────────────────────────

function ClearButton({ onClick }: { onClick: (e: React.MouseEvent<HTMLSpanElement>) => void }) {
  return (
    <span
      role="button"
      tabIndex={0}
      onClick={onClick}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") onClick(e as unknown as React.MouseEvent<HTMLSpanElement>);
      }}
      className="ml-1 shrink-0 rounded p-0.5 text-[#37352F]/40 hover:bg-[#37352F]/10 hover:text-[#37352F]/70 cursor-pointer"
      aria-label="日付をクリア"
    >
      <X className="h-3.5 w-3.5" />
    </span>
  );
}
