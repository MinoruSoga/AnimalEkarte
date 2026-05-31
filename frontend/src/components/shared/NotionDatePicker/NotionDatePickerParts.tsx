import { ChevronLeft, ChevronRight, X } from "lucide-react";
import type { Calendar } from "@/components/ui/calendar";
import { C, ICON } from "@/lib/design-tokens";
import { cn } from "@/lib/utils";

export interface SingleDatePickerProps {
  mode?: "single";
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  className?: string;
  id?: string;
  disabledDays?: React.ComponentProps<typeof Calendar>["disabled"];
}

export interface RangeDatePickerProps {
  mode: "range";
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  className?: string;
  id?: string;
}

export type NotionDatePickerProps = SingleDatePickerProps | RangeDatePickerProps;

export function parseLocalDate(iso: string): Date | undefined {
  if (!iso) return undefined;
  const [year, month, day] = iso.split("-").map(Number);
  if (!year || !month || !day) return undefined;
  return new Date(year, month - 1, day, 12, 0, 0);
}

export function formatIso(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

export function formatDisplay(date: Date): string {
  const year = date.getFullYear();
  const month = date.getMonth() + 1;
  const day = date.getDate();
  const weekdays = ["日", "月", "火", "水", "木", "金", "土"];
  const weekday = weekdays[date.getDay()];
  return `${year}年${month}月${day}日（${weekday}）`;
}

export function formatShort(date: Date): string {
  const year = date.getFullYear();
  const month = date.getMonth() + 1;
  const day = date.getDate();
  return `${year}/${month}/${day}`;
}

export function formatSlash(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}/${month}/${day}`;
}

export function parseRangeValue(value: string): { from?: Date; to?: Date } {
  if (!value) return {};
  const [fromStr, toStr] = value.split("~");
  return {
    from: parseLocalDate(fromStr),
    to: toStr ? parseLocalDate(toStr) : undefined,
  };
}

export function parseDateInput(input: string): Date | null {
  const trimmed = input.trim();
  if (!trimmed) return null;

  let year: number;
  let month: number;
  let day: number;

  const compact = trimmed.match(/^(\d{4})(\d{2})(\d{2})$/);
  if (compact) {
    year = Number(compact[1]);
    month = Number(compact[2]);
    day = Number(compact[3]);
  } else {
    const separated = trimmed.match(/^(\d{4})[/-](\d{1,2})[/-](\d{1,2})$/);
    if (!separated) return null;
    year = Number(separated[1]);
    month = Number(separated[2]);
    day = Number(separated[3]);
  }

  if (month < 1 || month > 12 || day < 1 || day > 31) return null;
  const date = new Date(year, month - 1, day, 12, 0, 0);
  if (date.getMonth() !== month - 1 || date.getDate() !== day) return null;
  return date;
}

const MONTH_LABELS = [
  "1月", "2月", "3月", "4月", "5月", "6月",
  "7月", "8月", "9月", "10月", "11月", "12月",
];

export const SINGLE_CALENDAR_CLASSES = {
  selected:
    `${C.bgPrimary} ${C.textWhite} hover:${C.bgPrimary} hover:text-white focus:${C.bgPrimary} focus:text-white`,
  today: `${C.bgPage} ${C.text}`,
  nav: "hidden",
  month_caption: "hidden",
};

export const RANGE_CALENDAR_CLASSES = {
  selected:
    `${C.bgPrimary} ${C.textWhite} hover:${C.bgPrimary} hover:text-white focus:${C.bgPrimary} focus:text-white`,
  today: `${C.bgPage} ${C.text}`,
  nav: "hidden",
};

export const TRIGGER_BASE =
  `flex h-11 w-full items-center justify-between rounded-md border ${C.borderMedium} bg-white px-3 text-sm ${C.text} transition-colors ${C.hoverBgPage} focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring`;

const NAV_BTN =
  `inline-flex items-center justify-center rounded p-1 ${C.text50} ${C.hoverBgPrimary10} ${C.hoverText} transition-colors`;

export function CalendarNav({
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
        <ChevronLeft className={ICON.action} />
      </button>
      <button
        type="button"
        onClick={onTitleClick}
        className={`rounded px-2 py-1 text-sm font-medium ${C.text} ${C.hoverBgPage} transition-colors`}
      >
        {year}年 {month}月
      </button>
      <button type="button" onClick={onNext} className={NAV_BTN} aria-label="次の月">
        <ChevronRight className={ICON.action} />
      </button>
    </div>
  );
}

export function YearNav({
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
        <ChevronLeft className={ICON.action} />
      </button>
      <span className={`text-sm font-medium ${C.text}`}>{year}年</span>
      <button type="button" onClick={onNextYear} className={NAV_BTN} aria-label="次の年">
        <ChevronRight className={ICON.action} />
      </button>
    </div>
  );
}

export function MonthGrid({
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
              ? `${C.bgPrimary} ${C.textWhite} font-medium`
              : `${C.text} ${C.hoverBgPage}`,
          )}
        >
          {label}
        </button>
      ))}
    </div>
  );
}

export function ClearButton({ onClick }: { onClick: (e: React.MouseEvent<HTMLButtonElement>) => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`ml-1 shrink-0 rounded p-0.5 ${C.text40} ${C.hoverBgPrimary10} ${C.hoverText}/70 cursor-pointer`}
      aria-label="日付をクリア"
    >
      <X className={ICON.action} />
    </button>
  );
}
