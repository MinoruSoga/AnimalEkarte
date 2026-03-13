// React/Framework
import { useState, useCallback } from "react";

// External
import { CalendarIcon, X, ArrowRight } from "lucide-react";

// Internal
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Calendar } from "@/components/ui/calendar";
import { cn } from "@/lib/utils";

// ─────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────

/** Formats a Date as "YYYY/M/D" for display. */
function formatDisplay(date: Date): string {
  const y = date.getFullYear();
  const m = date.getMonth() + 1;
  const d = date.getDate();
  return `${y}/${m}/${d}`;
}

// ─────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────

export interface DateRangePickerProps {
  startDate: Date | null;
  endDate: Date | null;
  onStartDateChange: (date: Date | null) => void;
  onEndDateChange: (date: Date | null) => void;
  placeholder?: string;
  className?: string;
}

// ─────────────────────────────────────────────────
// Component
// ─────────────────────────────────────────────────

/**
 * DateRangePicker
 *
 * Notion-style date range selector using shadcn/ui Popover + Calendar.
 * Clicking opens a popover with two months side by side (start / end).
 * The selection flow is: first click sets startDate, second click sets endDate.
 * If the user selects an end date before the start date, the two values swap.
 */
export function DateRangePicker({
  startDate,
  endDate,
  onStartDateChange,
  onEndDateChange,
  placeholder = "期間を選択...",
  className,
}: DateRangePickerProps) {
  const [open, setOpen] = useState(false);
  /** Which leg of the range is being selected next. */
  const [selecting, setSelecting] = useState<"start" | "end">("start");

  // Default the calendar month to startDate, or today.
  const defaultMonth = startDate ?? new Date();

  const handleDayClick = useCallback(
    (day: Date | undefined) => {
      if (!day) return;

      if (selecting === "start") {
        onStartDateChange(day);
        // If end is before new start, clear it.
        if (endDate && day > endDate) {
          onEndDateChange(null);
        }
        setSelecting("end");
      } else {
        // Swap if end < start
        if (startDate && day < startDate) {
          onEndDateChange(startDate);
          onStartDateChange(day);
        } else {
          onEndDateChange(day);
        }
        setOpen(false);
        setSelecting("start");
      }
    },
    [selecting, startDate, endDate, onStartDateChange, onEndDateChange],
  );

  const handleClearStart = useCallback(
    (e: React.MouseEvent) => {
      e.stopPropagation();
      onStartDateChange(null);
      setSelecting("start");
    },
    [onStartDateChange],
  );

  const handleClearEnd = useCallback(
    (e: React.MouseEvent) => {
      e.stopPropagation();
      onEndDateChange(null);
      setSelecting("end");
    },
    [onEndDateChange],
  );

  const hasRange = startDate !== null || endDate !== null;

  // Build the date range for Calendar's `selected` (for range highlighting).
  const calendarSelected =
    startDate || endDate
      ? {
          from: startDate ?? undefined,
          to: endDate ?? undefined,
        }
      : undefined;

  return (
    <Popover
      open={open}
      onOpenChange={(v) => {
        setOpen(v);
        if (!v) setSelecting("start");
      }}
    >
      <PopoverTrigger asChild>
        <button
          type="button"
          className={cn(
            "flex h-11 items-center gap-1.5 rounded-md border border-[rgba(55,53,47,0.16)] bg-white px-3 text-sm text-[#37352F] transition-colors",
            "hover:bg-[#F7F6F3] focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring",
            !hasRange && "text-[#37352F]/40",
            className,
          )}
        >
          <CalendarIcon className="size-4 shrink-0 text-[#37352F]/40" />

          {/* Start date */}
          <span className="flex items-center gap-1">
            {startDate ? (
              <>
                <span>{formatDisplay(startDate)}</span>
                <span
                  role="button"
                  tabIndex={-1}
                  onClick={handleClearStart}
                  className="ml-0.5 shrink-0 rounded p-0.5 text-[#37352F]/40 hover:bg-[#37352F]/10 hover:text-[#37352F]/70"
                >
                  <X className="size-3" />
                </span>
              </>
            ) : (
              <span className="text-[#37352F]/40">開始日</span>
            )}
          </span>

          <ArrowRight className="size-3.5 shrink-0 text-[#37352F]/30" />

          {/* End date */}
          <span className="flex items-center gap-1">
            {endDate ? (
              <>
                <span>{formatDisplay(endDate)}</span>
                <span
                  role="button"
                  tabIndex={-1}
                  onClick={handleClearEnd}
                  className="ml-0.5 shrink-0 rounded p-0.5 text-[#37352F]/40 hover:bg-[#37352F]/10 hover:text-[#37352F]/70"
                >
                  <X className="size-3" />
                </span>
              </>
            ) : (
              <span className="text-[#37352F]/40">
                {hasRange ? "終了日" : placeholder}
              </span>
            )}
          </span>
        </button>
      </PopoverTrigger>

      <PopoverContent className="w-auto p-3" align="start">
        {/* Instruction label */}
        <p className="mb-2 text-xs text-[#37352F]/50 select-none">
          {selecting === "start" ? "開始日を選択" : "終了日を選択"}
        </p>

        <Calendar
          mode="range"
          selected={calendarSelected}
          onSelect={(range) => {
            // Calendar's range mode fires onSelect with { from, to }.
            // We intercept individual clicks via onDayClick for precise control.
            if (!range) return;
          }}
          onDayClick={handleDayClick}
          defaultMonth={defaultMonth}
          numberOfMonths={2}
          className="rounded-md"
          classNames={{
            day_selected:
              "bg-[#37352F] text-white hover:bg-[#37352F] hover:text-white focus:bg-[#37352F] focus:text-white",
            day_today: "bg-[#F7F6F3] text-[#37352F]",
            day_range_middle:
              "bg-[#37352F]/10 text-[#37352F] rounded-none",
            day_range_start:
              "bg-[#37352F] text-white rounded-l-md",
            day_range_end:
              "bg-[#37352F] text-white rounded-r-md",
          }}
        />
      </PopoverContent>
    </Popover>
  );
}
