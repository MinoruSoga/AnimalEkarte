import { useState, useCallback } from "react";
import { CalendarIcon, X } from "lucide-react";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Calendar } from "@/components/ui/calendar";
import { cn } from "@/lib/utils";

interface NotionDatePickerProps {
  /** ISO date string (YYYY-MM-DD) or empty string */
  value: string;
  /** Called with ISO date string or empty string when cleared */
  onChange: (value: string) => void;
  /** Placeholder shown when no date is selected */
  placeholder?: string;
  /** Additional className for the trigger button */
  className?: string;
  /** HTML id for the trigger button */
  id?: string;
}

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

/**
 * Notion-style date picker.
 *
 * Renders a trigger button that opens a popover calendar.
 * The selected date is displayed in Japanese format.
 */
export function NotionDatePicker({
  value,
  onChange,
  placeholder = "日付を選択…",
  className,
  id,
}: NotionDatePickerProps) {
  const [open, setOpen] = useState(false);

  const selected = parseLocalDate(value);

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
    (e: React.MouseEvent) => {
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
          className={cn(
            "flex h-10 w-full items-center justify-between rounded-md border border-[rgba(55,53,47,0.16)] bg-white px-3 text-sm text-[#37352F] transition-colors",
            "hover:bg-[#F7F6F3] focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring",
            !value && "text-[#37352F]/40",
            className,
          )}
        >
          <span className="flex items-center gap-2 truncate">
            <CalendarIcon className="h-4 w-4 shrink-0 text-[#37352F]/40" />
            {selected ? formatDisplay(selected) : placeholder}
          </span>
          {value && (
            <span
              role="button"
              tabIndex={-1}
              onClick={handleClear}
              className="ml-1 shrink-0 rounded p-0.5 text-[#37352F]/40 hover:bg-[#37352F]/10 hover:text-[#37352F]/70"
            >
              <X className="h-3.5 w-3.5" />
            </span>
          )}
        </button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0" align="start">
        <Calendar
          mode="single"
          selected={selected}
          onSelect={handleSelect}
          defaultMonth={selected}
          className="rounded-md"
          classNames={{
            day_selected:
              "bg-[#37352F] text-white hover:bg-[#37352F] hover:text-white focus:bg-[#37352F] focus:text-white",
            day_today: "bg-[#F7F6F3] text-[#37352F]",
          }}
        />
      </PopoverContent>
    </Popover>
  );
}
